package charts

// range_eval_test.go - Unified range evaluation matrix for diverse, valid datasets.
//
// Run with: go test -v -run "TestRangeEvalMatrix"
//
// This file defines one golden dataset catalog and evaluates it in five modes:
//   - single_false     (PreferNiceIntervals=false)
//   - single_true      (PreferNiceIntervals=true)
//   - dual_false_false (left=false, right=false)
//   - dual_true_false  (left=true, right=false)
//   - dual_true_true   (left=true, right=true)
//
// Dual-axis evaluation uses ordered pairs (A,B) for all A != B to provide full
// symmetry coverage across axis-role assignment.

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const (
	evalLabelCountLowTarget  = 5
	evalLabelCountHighTarget = 10
	evalPadPctWarn           = 20.0
	evalPadPctBad            = 50.0

	// Dual-axis integration (defaultRender) is significantly more expensive than
	// exercising range coordination directly. Keep only this path bounded to ensure
	// the commit-matrix script remains usable.
	evalDualIntegrationMaxPairs = 250
)

// ---------------------------------------------------------------------------
// Quality helpers
// ---------------------------------------------------------------------------

// evalNiceScore classifies how "nice" an interval is.
//
//	T0 = standard nice set  {1, 2, 2.5, 5} × 10^n
//	T1 = extended nice set  {1, 2, 2.5, 3, 4, 5, 6, 8} × 10^n
//	T2 = not nice
func evalNiceScore(interval float64) string {
	if interval <= 0 {
		return "-"
	}
	const eps = 1e-10
	exp := math.Floor(math.Log10(interval))
	pow := math.Pow(10, exp)
	frac := interval / pow
	for _, n := range []float64{1, 2, 2.5, 5} {
		if math.Abs(frac-n) < eps {
			return "T0"
		}
	}
	for _, n := range []float64{1, 2, 2.5, 3, 4, 5, 6, 8} {
		if math.Abs(frac-n) < eps {
			return "T1"
		}
	}
	return "T2"
}

// evalAxisQuality computes easy-to-compare quality metrics for one resolved axisRange.
type evalAxisQuality struct {
	dataMin, dataMax  float64
	dataSpan          float64
	axisMin, axisMax  float64
	axisSpan          float64
	labelCount        int
	labelBucket       string
	interval          float64
	niceScore         string
	coverageMiss      bool
	missLeft          float64 // >0 means axisMin > dataMin
	missRight         float64 // >0 means axisMax < dataMax
	padExcessPct      float64 // NaN when dataSpan<=0 or axis misses data
	tightness         float64 // dataSpan/axisSpan for non-zero spans (higher is better)
	zeroSpanExpansion float64 // axis span when dataSpan==0
	padWarn           bool    // padExcessPct > evalPadPctWarn
	padBad            bool    // padExcessPct > evalPadPctBad
}

func computeEvalQuality(ar axisRange, dataMin, dataMax float64) evalAxisQuality {
	dataSpan := dataMax - dataMin
	axisSpan := ar.max - ar.min

	interval := math.NaN()
	if ar.labelCount > 1 {
		interval = axisSpan / float64(ar.labelCount-1)
	}
	niceScore := "-"
	if !math.IsNaN(interval) && interval > 0 {
		niceScore = evalNiceScore(interval)
	}

	leftPad := dataMin - ar.min
	rightPad := ar.max - dataMax
	coverageMiss := leftPad < -1e-9 || rightPad < -1e-9
	missLeft := math.Max(ar.min-dataMin, 0)
	missRight := math.Max(dataMax-ar.max, 0)

	padExcessPct := math.NaN()
	tightness := math.NaN()
	zeroSpanExpansion := math.NaN()
	if dataSpan > 0 {
		if !coverageMiss && axisSpan > 0 {
			padExcess := math.Max(leftPad, 0) + math.Max(rightPad, 0)
			padExcessPct = padExcess / dataSpan * 100
			tightness = dataSpan / axisSpan
		}
	} else {
		zeroSpanExpansion = axisSpan
	}

	labelBucket := "LT5"
	if ar.labelCount > evalLabelCountHighTarget {
		labelBucket = "GT10"
	} else if ar.labelCount >= evalLabelCountLowTarget {
		labelBucket = "5to10"
	}

	padWarn := !math.IsNaN(padExcessPct) && padExcessPct > evalPadPctWarn
	padBad := !math.IsNaN(padExcessPct) && padExcessPct > evalPadPctBad

	return evalAxisQuality{
		dataMin:           dataMin,
		dataMax:           dataMax,
		dataSpan:          dataSpan,
		axisMin:           ar.min,
		axisMax:           ar.max,
		axisSpan:          axisSpan,
		labelCount:        ar.labelCount,
		labelBucket:       labelBucket,
		interval:          interval,
		niceScore:         niceScore,
		coverageMiss:      coverageMiss,
		missLeft:          missLeft,
		missRight:         missRight,
		padExcessPct:      padExcessPct,
		tightness:         tightness,
		zeroSpanExpansion: zeroSpanExpansion,
		padWarn:           padWarn,
		padBad:            padBad,
	}
}

func evalMinMaxIgnoreNull(values []float64) (mn, mx float64, ok bool) {
	mn = math.Inf(1)
	mx = math.Inf(-1)
	nv := GetNullValue()
	for _, v := range values {
		if v == nv || math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		if v < mn {
			mn = v
		}
		if v > mx {
			mx = v
		}
		ok = true
	}
	if !ok {
		return 0, 0, false
	}
	return mn, mx, true
}

// ---------------------------------------------------------------------------
// Golden dataset catalog
// ---------------------------------------------------------------------------

type evalValueScenario struct {
	name   string
	values []float64
}

type dualScenario struct {
	leftName    string
	rightName   string
	leftValues  []float64
	rightValues []float64
}

type evalSeriesScenario struct {
	name      string
	series    seriesList
	yAxisIdx  int
	stack     bool
	baseValue string // optional: underlying value scenario name
}

func evalHasValidData(sl seriesList, yAxisIndex int) bool {
	nv := GetNullValue()
	for i := 0; i < sl.len(); i++ {
		s := sl.getSeries(i)
		if s.getYAxisIndex() != yAxisIndex {
			continue
		}
		for _, v := range s.getValues() {
			if v == nv || math.IsNaN(v) || math.IsInf(v, 0) {
				continue
			}
			return true
		}
	}
	return false
}

func evalDataRangeForSeries(sl seriesList, yAxisIndex int, stackSeries bool) (dataMin, dataMax float64, ok bool) {
	ok = evalHasValidData(sl, yAxisIndex)
	minVal, maxVal, sumMax := getSeriesMinMaxSumMax(sl, yAxisIndex, stackSeries)
	if stackSeries {
		return minVal, sumMax, ok
	}
	return minVal, maxVal, ok
}

func evalTransformValues(values []float64, f func(float64) float64) []float64 {
	nv := GetNullValue()
	out := make([]float64, len(values))
	for i, v := range values {
		if v == nv || math.IsNaN(v) || math.IsInf(v, 0) {
			out[i] = v
			continue
		}
		out[i] = f(v)
	}
	return out
}

func buildSingleAxisSeriesScenarios(values []evalValueScenario) []evalSeriesScenario {
	out := make([]evalSeriesScenario, 0, len(values)*2)

	// Base: every value scenario as a single series (axis 0).
	for _, sc := range values {
		out = append(out, evalSeriesScenario{
			name:      sc.name,
			series:    GenericSeriesList{{Type: ChartTypeLine, Name: sc.name, Values: sc.values, YAxisIndex: 0}},
			yAxisIdx:  0,
			stack:     false,
			baseValue: sc.name,
		})
	}

	// Add a limited set of multi-series and stacked variants to exercise:
	// - min/max over multiple series
	// - sumMax when stacking
	// - uneven series lengths
	scCopy := append([]evalValueScenario(nil), values...)
	sort.Slice(scCopy, func(i, j int) bool { return scCopy[i].name < scCopy[j].name })
	limit := 25
	if len(scCopy) < limit {
		limit = len(scCopy)
	}

	added := 0
	for _, sc := range scCopy {
		if added >= limit {
			break
		}
		st := evalStatsForScenario(sc)
		// Prefer positive-ish scenarios for stacked tests (sumMax is meaningful).
		if st.signBucket != "pos" || st.span <= 0 {
			continue
		}
		added++

		// Second series: scale + offset (keeps it in same magnitude but shifts min/max).
		s2 := evalTransformValues(sc.values, func(v float64) float64 { return v*0.6 + st.span*0.05 })

		multi := GenericSeriesList{
			{Type: ChartTypeLine, Name: sc.name + "_a", Values: sc.values, YAxisIndex: 0},
			{Type: ChartTypeLine, Name: sc.name + "_b", Values: s2, YAxisIndex: 0},
		}
		out = append(out, evalSeriesScenario{
			name:      "multi2_" + sc.name,
			series:    multi,
			yAxisIdx:  0,
			stack:     false,
			baseValue: sc.name,
		})
		out = append(out, evalSeriesScenario{
			name:      "stack2_" + sc.name,
			series:    multi,
			yAxisIdx:  0,
			stack:     true,
			baseValue: sc.name,
		})

		// Uneven lengths: drop the last element of series B.
		if len(s2) > 2 {
			uneven := GenericSeriesList{
				{Type: ChartTypeLine, Name: sc.name + "_a", Values: sc.values, YAxisIndex: 0},
				{Type: ChartTypeLine, Name: sc.name + "_b_short", Values: s2[:len(s2)-1], YAxisIndex: 0},
			}
			out = append(out, evalSeriesScenario{
				name:      "stack2_uneven_" + sc.name,
				series:    uneven,
				yAxisIdx:  0,
				stack:     true,
				baseValue: sc.name,
			})
		}
	}

	return out
}

// buildGoldenEvalValueScenarios returns one deduplicated, diverse scenario catalog
// for all range evaluation modes.
func buildGoldenEvalValueScenarios() []evalValueScenario {
	candidates := []evalValueScenario{
		// Powers / friendly anchors
		{"pow10_0_1", []float64{0, 1}},
		{"pow10_0_10", []float64{0, 10}},
		{"pow10_0_100", []float64{0, 100}},
		{"pow10_0_1k", []float64{0, 1000}},
		{"pow10_0_1m", []float64{0, 1e6}},
		{"pow10_0_1b", []float64{0, 1e9}},
		{"pow10_0_1t", []float64{0, 1e12}},

		// Wide ranges / cross zero
		{"wide_1_to_1e9", []float64{1, 1e9}},
		{"wide_1e3_to_1e12", []float64{1e3, 1e12}},
		{"wide_neg1e6_to_1e6", []float64{-1e6, 1e6}},
		{"wide_neg1e12_to_1e12", []float64{-1e12, 1e12}},
		{"cross_minus10_90", []float64{-10, 90}},
		{"cross_minus0p001_0p001", []float64{-0.001, 0.001}},
		{"cross_minus1e_9_1e_9", []float64{-1e-9, 1e-9}},
		{"neg_only_small", []float64{-0.9, -0.1}},
		{"pos_only_small", []float64{0.1, 0.9}},

		// Offset + tiny span
		{"offset_1m_span10", []float64{1e6, 1e6 + 10}},
		{"offset_1b_span1k", []float64{1e9, 1e9 + 1000}},
		{"offset_1t_span100", []float64{1e12, 1e12 + 100}},
		{"offset_1e15_span1", []float64{1e15, 1e15 + 1}},
		{"precision_offset_9e8", []float64{999999999.123456, 999999999.123457}},
		{"precision_neg_offset", []float64{-999999999.123457, -999999999.123456}},

		// Awkward / fractional
		{"prime_13_97", []float64{13, 97}},
		{"awkward_37_113", []float64{37, 113}},
		{"awkward_frac_0p13_0p89", []float64{0.13, 0.89}},
		{"irrational_pi", []float64{0, 3.14159}},
		{"irrational_e", []float64{0, 2.7182818}},
		{"irrational_sqrt2", []float64{0, 1.41421356}},
		{"frac_0_0p073", []float64{0, 0.073}},
		{"frac_1p37_4p83", []float64{1.37, 4.83}},
		{"frac_0p125_0p875", []float64{0.125, 0.875}},
		{"frac_thirds", []float64{0, 0.333, 0.667, 1.0}},
		{"around10_pos", []float64{9.99, 10.01}},
		{"around10_neg", []float64{-10.01, -9.99}},
		{"around10_cross", []float64{-9.99, 10.01}},

		// Tiny precision spans
		{"tinyspan_1e12_eps", []float64{1e12, 1e12 + 1e-6}},
		{"tinyspan_1e9_eps", []float64{1e9, 1e9 + 1e-9}},
		{"tinyspan_micro", []float64{1.23e-6, 1.29e-6}},
		{"precision_0p1_0p1000001", []float64{0.1, 0.1000001}},

		// Outlier / skew
		{"outlier_high", []float64{0.9, 1.0, 1.05, 1.1, 1000}},
		{"outlier_low", []float64{-1000, -1.2, -1.0, -0.9, -0.8}},
		{"skew_bi_tail", []float64{-5000, -5, -4.5, -4.2, 9000}},

		// Real-world style
		{"rw_latency_ms", []float64{12, 15, 19, 35, 44}},
		{"rw_cpu_percent", []float64{22.4, 31.8, 45.6, 87.3}},
		{"rw_revenue_usd", []float64{1.2e6, 2.1e6, 4.9e6, 8.2e6}},
		{"rw_fx_rate", []float64{1.0423, 1.0551, 1.0689, 1.0337}},
		{"rw_temperature_c", []float64{-12.5, -3.2, 4.8, 17.1, 28.6}},

		// Boundary values: just below and above power-of-10 transitions.
		// These stress the interval selection logic at scale boundaries.
		{"bound_below10", []float64{9.7, 9.9}},
		{"bound_above10", []float64{10.1, 10.3}},
		{"bound_cross10", []float64{9.9, 10.1}},
		{"bound_below100", []float64{98, 99.9}},
		{"bound_above100", []float64{100.1, 102}},
		{"bound_cross100", []float64{99.5, 100.5}},
		{"bound_below1k", []float64{998, 999.9}},
		{"bound_above1k", []float64{1000.1, 1002}},
		{"bound_cross1k", []float64{999, 1001}},
		{"bound_below1m", []float64{999990, 999999}},
		{"bound_above1m", []float64{1000001, 1000010}},
		{"bound_neg_below10", []float64{-10.3, -10.1}},
		{"bound_neg_above10", []float64{-9.9, -9.7}},
		{"bound_neg_cross10", []float64{-10.1, -9.9}},
		{"bound_neg_cross100", []float64{-100.5, -99.5}},
		{"bound_neg_cross1k", []float64{-1001, -999}},

		// Percent-style: common 0-100% and sub-percent ranges.
		{"pct_full", []float64{0, 100}},
		{"pct_high", []float64{85, 100}},
		{"pct_low", []float64{0, 15}},
		{"pct_mid", []float64{40, 60}},
		{"pct_sub", []float64{0.0, 1.0}},
		{"pct_sub_tight", []float64{0.15, 0.85}},
		{"pct_over100", []float64{0, 150}},

		// Finance / stock-like: close prices with small relative swings.
		{"finance_stock_lo", []float64{142.30, 143.10, 141.80, 144.60, 142.95}},
		{"finance_stock_hi", []float64{1824.50, 1831.20, 1818.75, 1837.00}},
		{"finance_bps", []float64{-0.0025, 0.0010, -0.0015, 0.0030}},

		// Monotonic: steadily increasing / decreasing at various rates.
		{"mono_inc_small", []float64{1, 2, 3, 4, 5}},
		{"mono_inc_large", []float64{100, 200, 300, 400, 500}},
		{"mono_dec_small", []float64{5, 4, 3, 2, 1}},
		{"mono_dec_neg", []float64{-1, -2, -3, -4, -5}},
		{"mono_inc_frac", []float64{0.1, 0.2, 0.3, 0.4, 0.5}},
		{"mono_exp", []float64{1, 10, 100, 1000, 10000}},

		// Label-count stress: ranges designed to produce extreme label counts
		// when using a default axis size, to stress the clamping logic.
		{"label_stress_very_tight", []float64{0, 0.0001}},
		{"label_stress_very_wide", []float64{0, 1e15}},

		// Scientific measurements: typical instrument output ranges.
		{"sci_voltage_mv", []float64{-3.3, -2.1, 1.5, 3.3}},
		{"sci_ph", []float64{6.8, 7.0, 7.2, 7.4}},
		{"sci_nano_meter", []float64{380e-9, 440e-9, 550e-9, 700e-9}},
		{"sci_gigahertz", []float64{2.4e9, 5.0e9}},
	}

	// Generated permutations for broader and more systematic coverage.
	scales := []float64{1e-12, 1e-9, 1e-6, 1e-3, 1, 1e3, 1e6, 1e9, 1e12}
	for i, scale := range scales {
		candidates = append(candidates,
			evalValueScenario{name: fmt.Sprintf("gen_pos_%02d", i), values: []float64{0, 1.3 * scale}},
			evalValueScenario{name: fmt.Sprintf("gen_cross_%02d", i), values: []float64{-0.8 * scale, 1.2 * scale}},
			evalValueScenario{name: fmt.Sprintf("gen_frac_%02d", i), values: []float64{0.11 * scale, 0.89 * scale}},
		)
	}

	offsets := []float64{1e3, 1e6, 1e9, 1e12, 1e15}
	spans := []float64{1e-3, 1, 10, 1e3, 1e6}
	for i := range offsets {
		offset := offsets[i]
		span := spans[i]
		candidates = append(candidates,
			evalValueScenario{name: fmt.Sprintf("gen_offset_pos_%02d", i), values: []float64{offset, offset + span}},
			evalValueScenario{name: fmt.Sprintf("gen_offset_neg_%02d", i), values: []float64{-offset - span, -offset}},
		)
	}

	// Boundary-crossing spans: values that straddle a power-of-10 by a small margin.
	// These are generated at each decade to ensure the scale-transition logic is robust.
	for i, scale := range scales {
		frac := 0.05 * scale
		candidates = append(candidates,
			evalValueScenario{name: fmt.Sprintf("gen_bound_below_%02d", i), values: []float64{scale - frac, scale - frac*0.1}},
			evalValueScenario{name: fmt.Sprintf("gen_bound_above_%02d", i), values: []float64{scale + frac*0.1, scale + frac}},
			evalValueScenario{name: fmt.Sprintf("gen_bound_cross_%02d", i), values: []float64{scale - frac*0.5, scale + frac*0.5}},
		)
	}

	progressionBases := []float64{1e-6, 1e-3, 1, 1e3}
	for i, base := range progressionBases {
		candidates = append(candidates,
			evalValueScenario{name: fmt.Sprintf("gen_geo_%02d", i), values: []float64{base, base * 3, base * 9, base * 27, base * 81}},
			evalValueScenario{name: fmt.Sprintf("gen_outlier_%02d", i), values: []float64{base, base * 1.02, base * 1.05, base * 1.08, base * 300}},
		)
	}

	// Sequence permutations across different lengths and scales.
	lengths := []int{2, 3, 5, 8}
	seqScales := []float64{1e-9, 1e-6, 1e-3, 1, 1e3, 1e6, 1e9}

	// Monotonic series: n equally spaced points at each scale.
	// These ensure that strictly increasing and decreasing data is handled well.
	monoLens := []int{3, 6, 10}
	for li, ln := range monoLens {
		for si, scale := range seqScales {
			inc := make([]float64, ln)
			dec := make([]float64, ln)
			for i := 0; i < ln; i++ {
				inc[i] = float64(i+1) * scale
				dec[i] = float64(ln-i) * scale
			}
			candidates = append(candidates,
				evalValueScenario{name: fmt.Sprintf("gen_mono_inc_l%02d_s%02d", li, si), values: inc},
				evalValueScenario{name: fmt.Sprintf("gen_mono_dec_l%02d_s%02d", li, si), values: dec},
			)
		}
	}

	for li, ln := range lengths {
		for si, scale := range seqScales {
			pos := make([]float64, ln)
			neg := make([]float64, ln)
			cross := make([]float64, ln)
			center := float64(ln-1) / 2
			for i := 0; i < ln; i++ {
				x := float64(i)
				pos[i] = (1.0 + x*0.37) * scale
				neg[i] = -(1.0 + x*0.37) * scale
				cross[i] = (x-center)*0.91*scale + 0.07*scale
			}
			candidates = append(candidates,
				evalValueScenario{name: fmt.Sprintf("gen_seq_pos_l%02d_s%02d", li, si), values: pos},
				evalValueScenario{name: fmt.Sprintf("gen_seq_neg_l%02d_s%02d", li, si), values: neg},
				evalValueScenario{name: fmt.Sprintf("gen_seq_cross_l%02d_s%02d", li, si), values: cross},
			)
		}
	}

	degenerateVals := []float64{-1e12, -1e6, -1, 0, 1, 1e6, 1e12}
	for i, v := range degenerateVals {
		candidates = append(candidates, evalValueScenario{
			name:   fmt.Sprintf("gen_degenerate_%02d", i),
			values: []float64{v, v, v},
		})
	}

	// Single-value series are common (e.g. a metric with 1 point).
	candidates = append(candidates,
		evalValueScenario{name: "single_point_zero", values: []float64{0}},
		evalValueScenario{name: "single_point_pos", values: []float64{42}},
		evalValueScenario{name: "single_point_neg", values: []float64{-42}},
	)

	// Null-only series should not propagate sentinel values.
	nv := GetNullValue()
	candidates = append(candidates,
		evalValueScenario{name: "null_only_2", values: []float64{nv, nv}},
		evalValueScenario{name: "null_only_5", values: []float64{nv, nv, nv, nv, nv}},
	)

	// Order/sparsity permutations for a deterministic subset to exercise
	// order invariance and sparse holes.
	{
		base := append([]evalValueScenario(nil), candidates...)
		sort.Slice(base, func(i, j int) bool { return base[i].name < base[j].name })
		limit := 24
		if len(base) < limit {
			limit = len(base)
		}
		for i := 0; i < limit; i++ {
			v := append([]float64(nil), base[i].values...)
			if len(v) < 3 {
				continue
			}
			hasInvalid := false
			for _, x := range v {
				if x == nv || math.IsNaN(x) || math.IsInf(x, 0) {
					hasInvalid = true
					break
				}
			}
			if hasInvalid {
				continue
			}

			rev := append([]float64(nil), v...)
			for l, r := 0, len(rev)-1; l < r; l, r = l+1, r-1 {
				rev[l], rev[r] = rev[r], rev[l]
			}
			candidates = append(candidates, evalValueScenario{name: "rev_" + base[i].name, values: rev})

			sparse := append([]float64(nil), v...)
			for k := 1; k < len(sparse); k += 2 {
				sparse[k] = nv
			}
			candidates = append(candidates, evalValueScenario{name: "sparse_alt_" + base[i].name, values: sparse})

			nanMid := append([]float64(nil), v...)
			nanMid[len(nanMid)/2] = math.NaN()
			candidates = append(candidates, evalValueScenario{name: "nan_mid_" + base[i].name, values: nanMid})

			infLast := append([]float64(nil), v...)
			infLast[len(infLast)-1] = math.Inf(1)
			candidates = append(candidates, evalValueScenario{name: "inf_last_" + base[i].name, values: infLast})
		}
	}

	// Null-sprinkled variants: exercise min/max skipping and "missing endpoints" patterns
	// without exploding the catalog size.
	{
		base := append([]evalValueScenario(nil), candidates...)
		sort.Slice(base, func(i, j int) bool { return base[i].name < base[j].name })
		limit := 10
		if len(base) < limit {
			limit = len(base)
		}
		for i := 0; i < limit; i++ {
			v := append([]float64(nil), base[i].values...)
			if len(v) < 3 {
				continue
			}
			hasNull := false
			for _, x := range v {
				if x == nv {
					hasNull = true
					break
				}
			}
			if hasNull {
				continue
			}
			mid := len(v) / 2
			v0 := append([]float64(nil), v...)
			v0[0] = nv
			candidates = append(candidates, evalValueScenario{name: "null_first_" + base[i].name, values: v0})

			v1 := append([]float64(nil), v...)
			v1[len(v1)-1] = nv
			candidates = append(candidates, evalValueScenario{name: "null_last_" + base[i].name, values: v1})

			v2 := append([]float64(nil), v...)
			v2[0] = nv
			v2[len(v2)-1] = nv
			candidates = append(candidates, evalValueScenario{name: "null_ends_" + base[i].name, values: v2})

			v3 := append([]float64(nil), v...)
			v3[mid] = nv
			candidates = append(candidates, evalValueScenario{name: "null_mid_" + base[i].name, values: v3})
		}
	}

	return dedupeEvalValueScenarios(candidates)
}

func dedupeEvalValueScenarios(in []evalValueScenario) []evalValueScenario {
	out := make([]evalValueScenario, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, sc := range in {
		if len(sc.values) == 0 {
			continue
		}
		key := evalValuesKey(sc.values)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, sc)
	}
	return out
}

func evalValuesKey(values []float64) string {
	var sb strings.Builder
	for i, v := range values {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.FormatFloat(v, 'g', -1, 64))
	}
	return sb.String()
}

type evalScenarioStats struct {
	min, max, span float64
	signBucket     string // pos, neg, cross, zero
	spanOOM        int    // floor(log10(span)) bucketed, clamped for bucketing
}

func evalOOMBucket(oom int) int {
	// Bucket OOMs in steps of 3 to reduce dual-axis representative explosion while
	// still spanning many orders of magnitude. For negatives we want floor-style
	// bucketing (Go division truncates toward 0).
	if oom >= 0 {
		return (oom / 3) * 3
	}
	return -(((-oom + 2) / 3) * 3)
}

func evalStatsForScenario(sc evalValueScenario) evalScenarioStats {
	mn, mx, ok := evalMinMaxIgnoreNull(sc.values)
	if !ok {
		return evalScenarioStats{min: 0, max: 0, span: 0, signBucket: "zero", spanOOM: -999}
	}
	span := mx - mn
	signBucket := "cross"
	switch {
	case span == 0:
		signBucket = "zero"
	case mn >= 0:
		signBucket = "pos"
	case mx <= 0:
		signBucket = "neg"
	}
	oom := -999
	if span > 0 {
		oom = evalOOMBucket(int(math.Floor(math.Log10(span))))
		if oom < -18 {
			oom = -18
		}
		if oom > 18 {
			oom = 18
		}
	}
	return evalScenarioStats{min: mn, max: mx, span: span, signBucket: signBucket, spanOOM: oom}
}

// selectDualRepresentativeScenarios picks a deterministic, representative subset from the full
// catalog to keep dual-axis evaluation runtime reasonable while preserving diverse scale/sign coverage.
func selectDualRepresentativeScenarios(scenarios []evalValueScenario) []evalValueScenario {
	type bucketKey struct {
		sign string
		oom  int
	}
	seen := make(map[bucketKey]struct{})
	var reps []evalValueScenario

	// Stable ordering makes selection deterministic and reproducible across commits.
	scCopy := append([]evalValueScenario(nil), scenarios...)
	sort.Slice(scCopy, func(i, j int) bool { return scCopy[i].name < scCopy[j].name })

	for _, sc := range scCopy {
		st := evalStatsForScenario(sc)
		key := bucketKey{sign: st.signBucket, oom: st.spanOOM}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		reps = append(reps, sc)
	}

	// Ensure some "edge" representatives are always present if they exist.
	ensureByNamePrefix := func(prefix string) {
		for _, sc := range scCopy {
			if strings.HasPrefix(sc.name, prefix) {
				for _, r := range reps {
					if r.name == sc.name {
						return
					}
				}
				reps = append(reps, sc)
				return
			}
		}
	}
	ensureByNamePrefix("offset_")
	ensureByNamePrefix("precision_")
	ensureByNamePrefix("tinyspan_")
	ensureByNamePrefix("null_only_")

	sort.Slice(reps, func(i, j int) bool { return reps[i].name < reps[j].name })
	return reps
}

// buildSymmetricDualAxisPairs generates all ordered pairs (i != j) from scenarios.
func buildSymmetricDualAxisPairs(scenarios []evalValueScenario) []dualScenario {
	n := len(scenarios)
	pairs := make([]dualScenario, 0, n*(n-1))
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			pairs = append(pairs, dualScenario{
				leftName:    scenarios[i].name,
				rightName:   scenarios[j].name,
				leftValues:  scenarios[i].values,
				rightValues: scenarios[j].values,
			})
		}
	}
	return pairs
}

// ---------------------------------------------------------------------------
// CSV helpers
// ---------------------------------------------------------------------------

// evalOpenCSV creates a CSV file, writes the header row, and returns the file and writer.
// Returns nil, nil on failure (non-fatal).
func evalOpenCSV(filename string, header []string) (*os.File, *csv.Writer) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, nil
	}
	w := csv.NewWriter(f)
	_ = w.Write(header)
	return f, w
}

// evalFmtFloat formats a float for CSV output.
func evalFmtFloat(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}

// evalFmtMaybeFloat formats optional float values for CSV output (empty when NaN/Inf).
func evalFmtMaybeFloat(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return ""
	}
	return evalFmtFloat(v)
}

func evalAxisCSVFields(q evalAxisQuality) []string {
	return []string{
		evalFmtFloat(q.dataMin),
		evalFmtFloat(q.dataMax),
		evalFmtFloat(q.dataSpan),
		evalFmtFloat(q.axisMin),
		evalFmtFloat(q.axisMax),
		evalFmtFloat(q.axisSpan),
		strconv.Itoa(q.labelCount),
		q.labelBucket,
		evalFmtMaybeFloat(q.interval),
		q.niceScore,
		strconv.FormatBool(q.coverageMiss),
		evalFmtMaybeFloat(q.padExcessPct),
		evalFmtMaybeFloat(q.tightness),
		evalFmtMaybeFloat(q.zeroSpanExpansion),
	}
}

func evalEmptyAxisCSVFields() []string {
	return []string{"", "", "", "", "", "", "", "", "", "", "", "", "", ""}
}

func evalPercent(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

func evalMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func evalP50(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	cp := append([]float64(nil), values...)
	sort.Float64s(cp)
	mid := len(cp) / 2
	if len(cp)%2 == 0 {
		return (cp[mid-1] + cp[mid]) / 2
	}
	return cp[mid]
}

func evalP95(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	cp := append([]float64(nil), values...)
	sort.Float64s(cp)
	idx := int(math.Ceil(0.95*float64(len(cp)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(cp) {
		idx = len(cp) - 1
	}
	return cp[idx]
}

type evalAxisMetrics struct {
	total int

	// nice interval tiers; "other" = no interval (labelCount <= 1 or constant data)
	t0, t1, t2, other int
	// singleLabel counts axes where labelCount <= 1 (subset of "other")
	singleLabel int

	labelsLT5, labels5to10, labelsGT10 int

	coverageMiss int
	padWarn      int
	padBad       int

	padExcessPcts      []float64
	tightnesses        []float64
	zeroSpanExpansions []float64

	// sign-bucketed T0 counts for diagnosing regression by data polarity
	t0Pos, totalPos   int // positive-only data (dataMin >= 0)
	t0Neg, totalNeg   int // negative-only data (dataMax <= 0)
	t0Cross, totalCross int // cross-zero data
}

func (m *evalAxisMetrics) add(q evalAxisQuality) {
	m.total++

	if q.labelCount <= 1 {
		m.singleLabel++
	}

	switch q.niceScore {
	case "T0":
		m.t0++
	case "T1":
		m.t1++
	case "T2":
		m.t2++
	default:
		m.other++
	}

	switch q.labelBucket {
	case "LT5":
		m.labelsLT5++
	case "GT10":
		m.labelsGT10++
	default:
		m.labels5to10++
	}

	if q.coverageMiss {
		m.coverageMiss++
	}
	if q.padWarn {
		m.padWarn++
	}
	if q.padBad {
		m.padBad++
	}
	if !math.IsNaN(q.padExcessPct) {
		m.padExcessPcts = append(m.padExcessPcts, q.padExcessPct)
	}
	if !math.IsNaN(q.tightness) {
		m.tightnesses = append(m.tightnesses, q.tightness)
	}
	if !math.IsNaN(q.zeroSpanExpansion) {
		m.zeroSpanExpansions = append(m.zeroSpanExpansions, q.zeroSpanExpansion)
	}

	// sign-bucketed T0 tracking (only for axes with a valid interval)
	if q.dataSpan > 0 {
		isPos := q.dataMin >= 0
		isNeg := q.dataMax <= 0
		switch {
		case isPos:
			m.totalPos++
			if q.niceScore == "T0" {
				m.t0Pos++
			}
		case isNeg:
			m.totalNeg++
			if q.niceScore == "T0" {
				m.t0Neg++
			}
		default: // cross-zero
			m.totalCross++
			if q.niceScore == "T0" {
				m.t0Cross++
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Unified matrix evaluation
// ---------------------------------------------------------------------------

type evalAxisContext struct {
	tag                    string
	isVertical             bool
	axisSize               int
	labelRotation          float64
	fontStyle              FontStyle
	rangeValuePaddingScale *float64

	// Optional, scenario-derived configuration knobs.
	labelUnitMode string // "", "span_div6", "span_div3"
	minMode       string // "", "min0_if_pos"
	maxMode       string // "", "max0_if_neg"
}

func (c evalAxisContext) deriveMinMaxUnit(dataMin, dataMax float64) (minCfg, maxCfg *float64, labelUnit float64) {
	span := dataMax - dataMin
	switch c.labelUnitMode {
	case "span_div6":
		if span > 0 {
			labelUnit = span / 6.0
		}
	case "span_div3":
		if span > 0 {
			labelUnit = span / 3.0
		}
	}

	if c.minMode == "min0_if_pos" && dataMin >= 0 {
		v := 0.0
		minCfg = &v
	}
	if c.maxMode == "max0_if_neg" && dataMax <= 0 {
		v := 0.0
		maxCfg = &v
	}
	return minCfg, maxCfg, labelUnit
}

func evalResolveDualAxisViaChartConfig(
	leftValues, rightValues []float64,
	niceL, niceR *bool,
	fs FontStyle,
) (axisRange, axisRange, error) {
	p := NewPainter(PainterOptions{Width: 800, Height: 600}, PainterThemeOption(GetDefaultTheme()))
	opt := defaultRenderOption{
		theme:   p.theme,
		padding: Box{},
		seriesList: GenericSeriesList{
			{Type: ChartTypeLine, Name: "left", Values: leftValues, YAxisIndex: 0},
			{Type: ChartTypeLine, Name: "right", Values: rightValues, YAxisIndex: 1},
		},
		xAxis: &XAxisOption{},
		yAxis: []YAxisOption{
			{LabelFontStyle: fs, PreferNiceIntervals: niceL},
			{LabelFontStyle: fs, PreferNiceIntervals: niceR},
		},
		title:  TitleOption{},
		legend: &LegendOption{Show: Ptr(false)},
	}
	res, err := defaultRender(p, opt)
	if err != nil {
		return axisRange{}, axisRange{}, err
	}
	arL, okL := res.yaxisRanges[0]
	arR, okR := res.yaxisRanges[1]
	if !okL || !okR {
		return axisRange{}, axisRange{}, errors.New("expected y-axis ranges for index 0 and 1")
	}
	return arL, arR, nil
}

func evalResolveDualAxisViaCoord(
	p *Painter,
	leftValues, rightValues []float64,
	niceL, niceR *bool,
	fs FontStyle,
) (axisRange, axisRange) {
	vf := func(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
	leftSL := GenericSeriesList{{Type: ChartTypeLine, Name: "left", Values: leftValues, YAxisIndex: 0}}
	rightSL := GenericSeriesList{{Type: ChartTypeLine, Name: "right", Values: rightValues, YAxisIndex: 1}}
	// Combine into one list as chart rendering does.
	combined := GenericSeriesList{leftSL[0], rightSL[0]}

	// Prepare preps for each axis and coordinate them directly.
	prepL := prepareValueAxisRange(p, true, 600,
		nil, nil, nil,
		nil, 0,
		0, 0, 0,
		combined, 0, false,
		vf,
		0, fs,
		niceL)
	prepR := prepareValueAxisRange(p, true, 600,
		nil, nil, nil,
		nil, 0,
		0, 0, 0,
		combined, 1, false,
		vf,
		0, fs,
		niceR)
	ars := coordinateValueAxisRanges(p, []*valueAxisPrep{&prepL, &prepR})
	return ars[0], ars[1]
}

func TestRangeEvalMatrix(t *testing.T) {
	valueScenarios := buildGoldenEvalValueScenarios()
	singleSeriesScenarios := buildSingleAxisSeriesScenarios(valueScenarios)
	dualCoordPairs := buildSymmetricDualAxisPairs(valueScenarios)
	dualIntegrationReps := selectDualRepresentativeScenarios(valueScenarios)
	dualIntegrationPairs := buildSymmetricDualAxisPairs(dualIntegrationReps)
	p := NewPainter(PainterOptions{Width: 800, Height: 600})
	fs := FontStyle{FontSize: 12}
	preferFalse := Ptr(false)
	preferTrue := Ptr(true)

	singleConfigs := []struct {
		tag  string
		nice *bool
	}{
		{"false", preferFalse},
		{"true", preferTrue},
	}

	dualConfigs := []struct {
		tag   string
		niceL *bool
		niceR *bool
	}{
		{"false_false", preferFalse, preferFalse},
		{"true_false", preferTrue, preferFalse},
		{"true_true", preferTrue, preferTrue},
	}

	// Axis size variants only: these change the label-count ceiling that the range
	// algorithm must work within, which is the key rendering-environment variable.
	axisCtxs := []evalAxisContext{
		{tag: "vert600", isVertical: true, axisSize: 600, labelRotation: 0, fontStyle: fs},
		{tag: "vert180", isVertical: true, axisSize: 180, labelRotation: 0, fontStyle: fs},
		{tag: "vert90", isVertical: true, axisSize: 90, labelRotation: 0, fontStyle: fs},
	}

	f, w := evalOpenCSV("range_eval_matrix.csv", []string{
		"mode", "ctx", "left_scenario", "right_scenario", "config", "aligned",
		"dataMin_L", "dataMax_L", "dataSpan_L", "axisMin_L", "axisMax_L", "axisSpan_L",
		"N_L", "labelBucket_L", "interval_L", "nice_L", "coverageMiss_L", "padPct_L", "tightness_L", "zeroSpanExpansion_L",
		"dataMin_R", "dataMax_R", "dataSpan_R", "axisMin_R", "axisMax_R", "axisSpan_R",
		"N_R", "labelBucket_R", "interval_R", "nice_R", "coverageMiss_R", "padPct_R", "tightness_R", "zeroSpanExpansion_R",
	})
	if f != nil {
		defer func() {
			w.Flush()
			if err := f.Close(); err != nil {
				t.Errorf("close csv file: %v", err)
			}
		}()
	}

	type singleSummary struct {
		axis evalAxisMetrics
	}
	type dualSummary struct {
		pairsTotal int
		aligned    int
		axis       evalAxisMetrics
		left       evalAxisMetrics
		right      evalAxisMetrics
	}
	singleSummaries := make([]singleSummary, len(singleConfigs))
	dualSummaries := make([]dualSummary, len(dualConfigs))
	// Optional breakdowns, useful when diagnosing regressions.
	singleCtxSummaries := make([][]evalAxisMetrics, len(singleConfigs))
	for i := range singleCtxSummaries {
		singleCtxSummaries[i] = make([]evalAxisMetrics, len(axisCtxs))
	}

	vf := func(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
	for _, sc := range singleSeriesScenarios {
		dMin, dMax, ok := evalDataRangeForSeries(sc.series, sc.yAxisIdx, sc.stack)
		if !ok {
			// Treat "no valid data" as a 0..0 dataset for evaluation purposes.
			dMin, dMax = 0, 0
		}
		for ctxi, ctx := range axisCtxs {
			minCfg, maxCfg, labelUnit := ctx.deriveMinMaxUnit(dMin, dMax)
			for ci, cfg := range singleConfigs {
				ar := calculateValueAxisRange(p, ctx.isVertical, ctx.axisSize,
					minCfg, maxCfg, ctx.rangeValuePaddingScale,
					nil, 0,
					0, labelUnit, 0,
					sc.series, sc.yAxisIdx, sc.stack,
					vf,
					ctx.labelRotation, ctx.fontStyle, cfg.nice)
				q := computeEvalQuality(ar, dMin, dMax)
				singleSummaries[ci].axis.add(q)
				singleCtxSummaries[ci][ctxi].add(q)
				if w != nil {
					row := make([]string, 0, 34)
					row = append(row, "single", ctx.tag, sc.name, "", cfg.tag, "")
					row = append(row, evalAxisCSVFields(q)...)
					row = append(row, evalEmptyAxisCSVFields()...)
					_ = w.Write(row)
				}
			}
		}
	}

	// Dual axis: evaluate coordination directly (fast path) across the full catalog.
	for _, sc := range dualCoordPairs {
		ldMin, ldMax, okL := evalMinMaxIgnoreNull(sc.leftValues)
		if !okL {
			ldMin, ldMax = 0, 0
		}
		rdMin, rdMax, okR := evalMinMaxIgnoreNull(sc.rightValues)
		if !okR {
			rdMin, rdMax = 0, 0
		}
		for ci, cfg := range dualConfigs {
			arL, arR := evalResolveDualAxisViaCoord(p, sc.leftValues, sc.rightValues, cfg.niceL, cfg.niceR, fs)
			qL := computeEvalQuality(arL, ldMin, ldMax)
			qR := computeEvalQuality(arR, rdMin, rdMax)
			aligned := arL.labelCount == arR.labelCount
			alignStr := "YES"
			if !aligned {
				alignStr = "NO"
			}

			sm := &dualSummaries[ci]
			sm.pairsTotal++
			if aligned {
				sm.aligned++
			}
			sm.axis.add(qL)
			sm.axis.add(qR)
			sm.left.add(qL)
			sm.right.add(qR)

			if w != nil {
				row := make([]string, 0, 34)
				row = append(row, "dual_coord", "vert600_fs12_r0", sc.leftName, sc.rightName, cfg.tag, alignStr)
				row = append(row, evalAxisCSVFields(qL)...)
				row = append(row, evalAxisCSVFields(qR)...)
				_ = w.Write(row)
			}
		}
	}

	// Dual axis integration: sanity-check a bounded subset via chart config (defaultRender path).
	dualIntegrationSummaries := make([]dualSummary, len(dualConfigs))
	for i := 0; i < len(dualIntegrationPairs) && i < evalDualIntegrationMaxPairs; i++ {
		sc := dualIntegrationPairs[i]
		ldMin, ldMax, okL := evalMinMaxIgnoreNull(sc.leftValues)
		if !okL {
			ldMin, ldMax = 0, 0
		}
		rdMin, rdMax, okR := evalMinMaxIgnoreNull(sc.rightValues)
		if !okR {
			rdMin, rdMax = 0, 0
		}
		for ci, cfg := range dualConfigs {
			arL, arR, err := evalResolveDualAxisViaChartConfig(sc.leftValues, sc.rightValues, cfg.niceL, cfg.niceR, fs)
			if err != nil {
				t.Fatalf("resolve dual axis via chart config failed for %s/%s (%s): %v",
					sc.leftName, sc.rightName, cfg.tag, err)
			}
			qL := computeEvalQuality(arL, ldMin, ldMax)
			qR := computeEvalQuality(arR, rdMin, rdMax)
			aligned := arL.labelCount == arR.labelCount
			alignStr := "YES"
			if !aligned {
				alignStr = "NO"
			}

			sm := &dualIntegrationSummaries[ci]
			sm.pairsTotal++
			if aligned {
				sm.aligned++
			}
			sm.axis.add(qL)
			sm.axis.add(qR)
			sm.left.add(qL)
			sm.right.add(qR)

			if w != nil {
				row := make([]string, 0, 34)
				row = append(row, "dual_chart", "800x600", sc.leftName, sc.rightName, cfg.tag, alignStr)
				row = append(row, evalAxisCSVFields(qL)...)
				row = append(row, evalAxisCSVFields(qR)...)
				_ = w.Write(row)
			}
		}
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "\n=== RANGE EVAL MATRIX ===\n")
	fmt.Fprintf(&sb, "EVAL|catalog|ValueScenarios=%d|SingleSeriesScenarios=%d|AxisContexts=%d|DualCoordPairs=%d|DualIntegrationReps=%d|DualIntegrationPairs=%d|SingleConfigs=%d|DualConfigs=%d|SingleRuns=%d|DualRuns=%d|DualIntegrationRuns=%d\n",
		len(valueScenarios), len(singleSeriesScenarios), len(axisCtxs), len(dualCoordPairs), len(dualIntegrationReps), len(dualIntegrationPairs), len(singleConfigs), len(dualConfigs),
		len(singleSeriesScenarios)*len(axisCtxs)*len(singleConfigs), len(dualCoordPairs)*len(dualConfigs), minInt(evalDualIntegrationMaxPairs, len(dualIntegrationPairs))*len(dualConfigs))

	for ci, cfg := range singleConfigs {
		sm := singleSummaries[ci].axis
		fmt.Fprintf(&sb,
			"EVAL|single_summary|Config=%s|Axes=%d|T0=%d|T1=%d|T2=%d|Other=%d|SingleLabel=%d|SingleLabelRate=%.2f|T0Rate=%.2f|T01Rate=%.2f|LblLT5=%d|Lbl5to10=%d|LblGT10=%d|LblLT5Rate=%.2f|Lbl5to10Rate=%.2f|LblGT10Rate=%.2f|CoverageMiss=%d|CoverageMissRate=%.2f|PadWarn=%d|PadWarnRate=%.2f|PadBad=%d|PadBadRate=%.2f|PadAvg=%.2f|PadP50=%.2f|PadP95=%.2f|TightAvg=%.4f|TightP50=%.4f|ZeroSpanCases=%d|ZeroSpanExpAvg=%g|T0PosRate=%.2f|T0NegRate=%.2f|T0CrossRate=%.2f\n",
			cfg.tag, sm.total, sm.t0, sm.t1, sm.t2, sm.other,
			sm.singleLabel, evalPercent(sm.singleLabel, sm.total),
			evalPercent(sm.t0, sm.total),
			evalPercent(sm.t0+sm.t1, sm.total),
			sm.labelsLT5, sm.labels5to10, sm.labelsGT10,
			evalPercent(sm.labelsLT5, sm.total),
			evalPercent(sm.labels5to10, sm.total),
			evalPercent(sm.labelsGT10, sm.total),
			sm.coverageMiss,
			evalPercent(sm.coverageMiss, sm.total),
			sm.padWarn,
			evalPercent(sm.padWarn, sm.total),
			sm.padBad,
			evalPercent(sm.padBad, sm.total),
			evalMean(sm.padExcessPcts),
			evalP50(sm.padExcessPcts),
			evalP95(sm.padExcessPcts),
			evalMean(sm.tightnesses),
			evalP50(sm.tightnesses),
			len(sm.zeroSpanExpansions),
			evalMean(sm.zeroSpanExpansions),
			evalPercent(sm.t0Pos, sm.totalPos),
			evalPercent(sm.t0Neg, sm.totalNeg),
			evalPercent(sm.t0Cross, sm.totalCross),
		)
		for ctxi, ctx := range axisCtxs {
			cm := singleCtxSummaries[ci][ctxi]
			fmt.Fprintf(&sb,
				"EVAL|single_ctx_summary|Config=%s|Ctx=%s|Axes=%d|T0Rate=%.2f|T01Rate=%.2f|LblLT5Rate=%.2f|Lbl5to10Rate=%.2f|LblGT10Rate=%.2f|CoverageMissRate=%.2f|PadWarnRate=%.2f|PadBadRate=%.2f|PadAvg=%.2f|PadP95=%.2f|TightAvg=%.4f|ZeroSpanCases=%d|ZeroSpanExpAvg=%g\n",
				cfg.tag, ctx.tag, cm.total,
				evalPercent(cm.t0, cm.total),
				evalPercent(cm.t0+cm.t1, cm.total),
				evalPercent(cm.labelsLT5, cm.total),
				evalPercent(cm.labels5to10, cm.total),
				evalPercent(cm.labelsGT10, cm.total),
				evalPercent(cm.coverageMiss, cm.total),
				evalPercent(cm.padWarn, cm.total),
				evalPercent(cm.padBad, cm.total),
				evalMean(cm.padExcessPcts),
				evalP95(cm.padExcessPcts),
				evalMean(cm.tightnesses),
				len(cm.zeroSpanExpansions),
				evalMean(cm.zeroSpanExpansions),
			)
		}
	}

	for ci, cfg := range dualConfigs {
		sm := dualSummaries[ci]
		axis := sm.axis
		left := sm.left
		right := sm.right
		fmt.Fprintf(&sb,
			"EVAL|dual_summary|Config=%s|Pairs=%d|Aligned=%d|AlignRate=%.2f|Axes=%d|T0=%d|T1=%d|T2=%d|Other=%d|SingleLabel=%d|SingleLabelRate=%.2f|T0Rate=%.2f|T01Rate=%.2f|LblLT5=%d|Lbl5to10=%d|LblGT10=%d|LblLT5Rate=%.2f|Lbl5to10Rate=%.2f|LblGT10Rate=%.2f|CoverageMiss=%d|CoverageMissRate=%.2f|PadWarn=%d|PadWarnRate=%.2f|PadBad=%d|PadBadRate=%.2f|PadAvg=%.2f|PadP50=%.2f|PadP95=%.2f|TightAvg=%.4f|TightP50=%.4f|ZeroSpanCases=%d|ZeroSpanExpAvg=%g|LeftT0Rate=%.2f|LeftLbl5to10Rate=%.2f|LeftCoverageMissRate=%.2f|RightT0Rate=%.2f|RightLbl5to10Rate=%.2f|RightCoverageMissRate=%.2f|T0PosRate=%.2f|T0NegRate=%.2f|T0CrossRate=%.2f\n",
			cfg.tag, sm.pairsTotal, sm.aligned, evalPercent(sm.aligned, sm.pairsTotal),
			axis.total, axis.t0, axis.t1, axis.t2, axis.other,
			axis.singleLabel, evalPercent(axis.singleLabel, axis.total),
			evalPercent(axis.t0, axis.total),
			evalPercent(axis.t0+axis.t1, axis.total),
			axis.labelsLT5, axis.labels5to10, axis.labelsGT10,
			evalPercent(axis.labelsLT5, axis.total),
			evalPercent(axis.labels5to10, axis.total),
			evalPercent(axis.labelsGT10, axis.total),
			axis.coverageMiss,
			evalPercent(axis.coverageMiss, axis.total),
			axis.padWarn,
			evalPercent(axis.padWarn, axis.total),
			axis.padBad,
			evalPercent(axis.padBad, axis.total),
			evalMean(axis.padExcessPcts),
			evalP50(axis.padExcessPcts),
			evalP95(axis.padExcessPcts),
			evalMean(axis.tightnesses),
			evalP50(axis.tightnesses),
			len(axis.zeroSpanExpansions),
			evalMean(axis.zeroSpanExpansions),
			evalPercent(left.t0, left.total),
			evalPercent(left.labels5to10, left.total),
			evalPercent(left.coverageMiss, left.total),
			evalPercent(right.t0, right.total),
			evalPercent(right.labels5to10, right.total),
			evalPercent(right.coverageMiss, right.total),
			evalPercent(axis.t0Pos, axis.totalPos),
			evalPercent(axis.t0Neg, axis.totalNeg),
			evalPercent(axis.t0Cross, axis.totalCross),
		)
	}

	for ci, cfg := range dualConfigs {
		sm := dualIntegrationSummaries[ci]
		axis := sm.axis
		left := sm.left
		right := sm.right
		fmt.Fprintf(&sb,
			"EVAL|dual_integration_summary|Config=%s|Pairs=%d|Aligned=%d|AlignRate=%.2f|Axes=%d|T0Rate=%.2f|T01Rate=%.2f|LblLT5Rate=%.2f|Lbl5to10Rate=%.2f|LblGT10Rate=%.2f|CoverageMissRate=%.2f|PadWarnRate=%.2f|PadBadRate=%.2f|PadAvg=%.2f|PadP50=%.2f|PadP95=%.2f|TightAvg=%.4f|TightP50=%.4f|LeftT0Rate=%.2f|LeftLbl5to10Rate=%.2f|LeftCoverageMissRate=%.2f|RightT0Rate=%.2f|RightLbl5to10Rate=%.2f|RightCoverageMissRate=%.2f\n",
			cfg.tag, sm.pairsTotal, sm.aligned, evalPercent(sm.aligned, sm.pairsTotal),
			axis.total,
			evalPercent(axis.t0, axis.total),
			evalPercent(axis.t0+axis.t1, axis.total),
			evalPercent(axis.labelsLT5, axis.total),
			evalPercent(axis.labels5to10, axis.total),
			evalPercent(axis.labelsGT10, axis.total),
			evalPercent(axis.coverageMiss, axis.total),
			evalPercent(axis.padWarn, axis.total),
			evalPercent(axis.padBad, axis.total),
			evalMean(axis.padExcessPcts),
			evalP50(axis.padExcessPcts),
			evalP95(axis.padExcessPcts),
			evalMean(axis.tightnesses),
			evalP50(axis.tightnesses),
			evalPercent(left.t0, left.total),
			evalPercent(left.labels5to10, left.total),
			evalPercent(left.coverageMiss, left.total),
			evalPercent(right.t0, right.total),
			evalPercent(right.labels5to10, right.total),
			evalPercent(right.coverageMiss, right.total),
		)
	}
	t.Log(sb.String())
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
