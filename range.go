package charts

import (
	"math"
	"strconv"

	"github.com/go-analyze/charts/chartdraw"
	"github.com/go-analyze/charts/chartdraw/matrix"
)

const rangeMinPaddingPercentMin = 0.0 // increasing could result in forced negative y-axis minimum
const rangeMinPaddingPercentMax = 20.0
const rangeMaxPaddingPercentMin = 5.0 // set minimum spacing at the top of the graph
const rangeMaxPaddingPercentMax = 20.0
const zeroSpanAdjustment = 1

// axisRange represents the calculated range for the axis, as well as values for fitting labels on the range.
type axisRange struct {
	isCategory bool
	// labels are the rendered labels: 1:1 for categories or range value labels to render.
	labels []string
	// dataStartIndex specifies the starting index for label values.
	dataStartIndex int
	tickCount      int
	divideCount    int
	labelCount     int
	min, max       float64 // only valid if !isCategory
	size           int
	textMaxWidth   int
	textMaxHeight  int
	labelRotation  float64
	labelFontStyle FontStyle
}

// valueAxisPrep captures intermediate state between preparation and resolution of a value axis range.
type valueAxisPrep struct {
	// data range from series
	minVal, maxVal           float64
	minPadScale, maxPadScale float64
	padLabelCount            int // estimated label count after collision check
	maxLabelCount            int // max labels that fit the axis pixel size
	// carry-through for resolution and finalization
	labelsCfg      []string
	valueFormatter ValueFormatter
	labelCountCfg  int // user's explicit count (0 = auto)
	labelUnit      float64
	minCfg, maxCfg *float64
	dataStartIndex int
	labelRotation  float64
	fontStyle      FontStyle
	axisSize       int
	// measured labels from preparation
	labels         []string
	labelW, labelH int
}

// prepareValueAxisRange gathers data range and estimates label count, returning intermediate state.
func prepareValueAxisRange(p *Painter, isVertical bool, axisSize int,
	minCfg, maxCfg, rangeValuePaddingScale *float64,
	labelsCfg []string, dataStartIndex int,
	labelCountCfg int, labelUnit float64, labelCountAdjustment int,
	seriesList seriesList, yAxisIndex int, stackSeries bool,
	valueFormatter ValueFormatter,
	labelRotation float64, fontStyle FontStyle) valueAxisPrep {
	minVal, maxVal, sumMax := getSeriesMinMaxSumMax(seriesList, yAxisIndex, stackSeries)
	if stackSeries {
		if minVal > 0 {
			minVal--
		}
		maxVal = sumMax
	}
	minPadScale, maxPadScale := 1.0, 1.0
	if rangeValuePaddingScale != nil {
		minPadScale = *rangeValuePaddingScale
		maxPadScale = minPadScale
	}
	if minCfg != nil && *minCfg < minVal {
		minVal = *minCfg
		minPadScale = 0.0
	}
	if maxCfg != nil && *maxCfg > maxVal {
		maxVal = *maxCfg
		maxPadScale = 0.0
	}
	decimalData := minVal != math.Floor(minVal) || (maxVal-minVal) != math.Floor(maxVal-minVal)

	// Label counts and range padding are linked together to produce a user-friendly graph.
	// First when considering padding we want to prefer a zero axis start if reasonable, and add a slight
	// padding to the maxVal so there is a little space at the top of the graph. We also want to pick
	// a maxVal value and label count that will result in round intervals on the axis, or match the user
	// supplied unit (if provided).
	//
	// In order to accomplish this, we estimate the label count (if necessary), produce some labels to measure,
	// calculate our label limit, pad the range, then once the label count is finalized produce the final labels.
	initialLabelCount := labelCountCfg
	if initialLabelCount < 1 {
		if labelUnit > 0 {
			initialLabelCount = int((maxVal-minVal)/labelUnit) + 1
		} else {
			initialLabelCount =
				chartdraw.MinInt(chartdraw.MaxInt(int(maxVal-minVal)+1, defaultYAxisLabelCountLow),
					defaultYAxisLabelCountHigh)
			if decimalData {
				initialLabelCount = chartdraw.MinInt(initialLabelCount*2, defaultYAxisLabelCountHigh)
			}
		}
	}
	initialLabelCount = chartdraw.MaxInt(initialLabelCount+labelCountAdjustment, minimumAxisLabels)
	labels := valueLabels(labelsCfg, valueFormatter, minVal, maxVal, initialLabelCount)
	labelW, labelH := p.measureTextMaxWidthHeight(labels, labelRotation, fontStyle)

	// If user gave an explicit LabelCount, then we do NOT do a collision check.
	// For default logic we want to make sure we choose a label count that is visually appealing.
	padLabelCount := initialLabelCount
	maxLabelCount := padLabelCount
	if labelCountCfg == 0 {
		if isVertical {
			if labelH > 0 {
				maxLabelCount = axisSize / labelH
			}
		} else {
			if labelW > 0 {
				maxLabelCount = axisSize / (labelW + chartdraw.MinInt(20, labelW))
			}
		}
		if maxLabelCount < padLabelCount {
			padLabelCount = chartdraw.MaxInt(maxLabelCount, minimumAxisLabels)
		}
		if labelUnit > 0 && padLabelCount > minimumAxisLabels {
			padLabelCount--
		}
	}
	return valueAxisPrep{
		minVal:         minVal,
		maxVal:         maxVal,
		minPadScale:    minPadScale,
		maxPadScale:    maxPadScale,
		padLabelCount:  padLabelCount,
		maxLabelCount:  maxLabelCount,
		labelsCfg:      labelsCfg,
		valueFormatter: valueFormatter,
		labelCountCfg:  labelCountCfg,
		labelUnit:      labelUnit,
		minCfg:         minCfg,
		maxCfg:         maxCfg,
		dataStartIndex: dataStartIndex,
		labelRotation:  labelRotation,
		fontStyle:      fontStyle,
		axisSize:       axisSize,
		labels:         labels,
		labelW:         labelW,
		labelH:         labelH,
	}
}

// resolveValueAxisRange computes the padded range and label count from a prepared axis.
// When targetLabelCount > 0, it overrides padLabelCount and disables flex.
func resolveValueAxisRange(prep *valueAxisPrep, flexCount bool, targetLabelCount int) (float64, float64, int) {
	padLabelCount := prep.padLabelCount
	maxLabelCount := prep.maxLabelCount
	if targetLabelCount > 0 {
		padLabelCount = targetLabelCount
		maxLabelCount = targetLabelCount
		flexCount = false
	}

	minPadded, maxPadded, adjustedCount := padRange(padLabelCount, prep.minVal, prep.maxVal,
		prep.minPadScale, prep.maxPadScale, flexCount)
	labelCount := adjustedCount

	// if the user set only a unit, we may need to refine again after padding to meet the unit
	if prep.labelCountCfg == 0 && prep.labelUnit > 0 {
		labelUnit := prep.labelUnit
		if dataSpan := maxPadded - minPadded; labelUnit >= dataSpan {
			labelCount = minimumAxisLabels
		} else {
			down := func(v float64) float64 { return math.Floor(v/labelUnit) * labelUnit }
			up := func(v float64) float64 { return math.Ceil(v/labelUnit) * labelUnit }

			var bestCount int
			bestMin, bestMax := minPadded, maxPadded
			bestPad := math.Inf(1)
			bestDeltaC := math.MaxInt

			accept := func(c int, mn, mx float64) {
				deltaAbs := int(math.Abs(float64(padLabelCount - c)))
				pad := (minPadded - mn) + (mx - maxPadded)
				if pad < bestPad-matrix.DefaultEpsilon ||
					(math.Abs(pad-bestPad) < matrix.DefaultEpsilon &&
						(deltaAbs < bestDeltaC || (deltaAbs == bestDeltaC && c > bestCount))) {
					if bestCount == 0 || down(mn)-mn < matrix.DefaultEpsilon && mx-up(mx) < matrix.DefaultEpsilon {
						bestPad = pad
						bestMin = mn
						bestMax = mx
						bestCount = c
						bestDeltaC = deltaAbs
					}
				}
			}

			maxDelta := chartdraw.MaxInt(padLabelCount-minimumAxisLabels, maxLabelCount-padLabelCount)
			if targetLabelCount > 0 {
				maxDelta = 0 // only try the target count
			}
			for delta := 0; delta <= maxDelta; delta++ {
				try := func(c int) {
					if c < minimumAxisLabels || c > maxLabelCount {
						return
					}
					spanCount := float64(c - 1)
					span := spanCount * labelUnit

					// snapped min and max to the current label count
					snappedMin := down(minPadded)
					snappedMax := up(maxPadded)
					snappedInterval := up((snappedMax - snappedMin) / spanCount)
					flip := true
					for snappedMin+(snappedInterval*spanCount)-snappedMax > matrix.DefaultEpsilon {
						if snappedMin-snappedInterval >= 0 && flip {
							flip = false
							snappedMin -= snappedInterval
						} else {
							flip = true
							snappedMax += snappedInterval
						}
					}
					snappedMax = math.Ceil(snappedMax/snappedInterval) * snappedInterval
					accept(c, snappedMin, snappedMax)

					// shift MIN downward
					if prep.minCfg == nil {
						candMax := up(maxPadded)
						candMin := candMax - span
						if (minPadded < 0 || candMin >= 0.0-matrix.DefaultEpsilon) &&
							candMin <= minPadded+matrix.DefaultEpsilon {
							accept(c, candMin, candMax)
						}
					}

					// split padding (both free)
					if prep.minCfg == nil && prep.maxCfg == nil {
						candMin := down(minPadded - (span-dataSpan)/2)
						candMax := candMin + span
						if (minPadded < 0 || candMin >= 0.0-matrix.DefaultEpsilon) &&
							candMin <= minPadded+matrix.DefaultEpsilon &&
							candMax >= maxPadded-matrix.DefaultEpsilon {
							accept(c, candMin, candMax)
						}
					}

					// grow the MAX upward
					if prep.maxCfg == nil {
						candMin := down(minPadded)
						candMax := candMin + span
						if candMax >= maxPadded-matrix.DefaultEpsilon {
							accept(c, candMin, candMax)
						}
					}
				}

				if cand := padLabelCount - delta; cand >= minimumAxisLabels {
					try(cand)
				}
				if delta != 0 {
					if cand := padLabelCount + delta; cand <= maxLabelCount {
						try(cand)
					}
				}

				if bestPad < matrix.DefaultEpsilon {
					break
				}
			}

			if bestCount > 0 {
				labelCount = bestCount
				minPadded = bestMin
				maxPadded = bestMax
			}
		}
	}

	return minPadded, maxPadded, labelCount
}

// finalizeValueAxisRange produces the final axisRange, regenerating labels if the range changed.
func finalizeValueAxisRange(p *Painter, prep *valueAxisPrep, minPadded, maxPadded float64, labelCount int) axisRange {
	labels := prep.labels
	labelW, labelH := prep.labelW, prep.labelH

	if len(labels) != labelCount || prep.minVal-minPadded > matrix.DefaultEpsilon || maxPadded-prep.maxVal > matrix.DefaultEpsilon {
		labels = valueLabels(prep.labelsCfg, prep.valueFormatter, minPadded, maxPadded, labelCount)
		labelW, labelH = p.measureTextMaxWidthHeight(labels, prep.labelRotation, prep.fontStyle)
	}

	return axisRange{
		isCategory:     false,
		labels:         labels,
		dataStartIndex: prep.dataStartIndex,
		divideCount:    len(labels),
		tickCount:      labelCount,
		labelCount:     labelCount,
		min:            minPadded,
		max:            maxPadded,
		size:           prep.axisSize,
		textMaxWidth:   labelW,
		textMaxHeight:  labelH,
		labelRotation:  prep.labelRotation,
		labelFontStyle: prep.fontStyle,
	}
}

// coordinateValueAxisRanges finds a shared label count for multiple value axes so that grid lines
// align. When at least one secondary axis has PreferNiceIntervals, a search finds the best shared
// count. Otherwise, secondary axes adopt the primary's resolved count directly.
func coordinateValueAxisRanges(p *Painter, preps []*valueAxisPrep, preferNice []*bool) []axisRange {
	n := len(preps)
	if n == 0 {
		return nil
	}
	if n == 1 {
		flexCount := flagIs(true, preferNice[0])
		mn, mx, count := resolveValueAxisRange(preps[0], flexCount, 0)
		return []axisRange{finalizeValueAxisRange(p, preps[0], mn, mx, count)}
	}

	// if any axis has an explicit labelCountCfg, that count wins; others adapt
	var forcedCount int
	var hasConflict bool
	for _, prep := range preps {
		if prep.labelCountCfg > 0 {
			if forcedCount > 0 && forcedCount != prep.labelCountCfg {
				hasConflict = true
				break
			}
			forcedCount = prep.labelCountCfg
		}
	}
	if hasConflict {
		// conflicting explicit counts - resolve independently
		result := make([]axisRange, n)
		for i, prep := range preps {
			flexCount := flagIs(true, preferNice[i])
			mn, mx, count := resolveValueAxisRange(prep, flexCount, 0)
			result[i] = finalizeValueAxisRange(p, prep, mn, mx, count)
		}
		return result
	}
	if forcedCount > 0 {
		result := make([]axisRange, n)
		for i, prep := range preps {
			if prep.labelCountCfg == forcedCount {
				flexCount := flagIs(true, preferNice[i])
				mn, mx, count := resolveValueAxisRange(prep, flexCount, 0)
				result[i] = finalizeValueAxisRange(p, prep, mn, mx, count)
			} else {
				mn, mx, count := resolveValueAxisRange(prep, false, forcedCount)
				result[i] = finalizeValueAxisRange(p, prep, mn, mx, count)
			}
		}
		return result
	}

	// resolve primary axis independently
	primaryFlex := flagIs(true, preferNice[0])
	primaryMin, primaryMax, primaryCount := resolveValueAxisRange(preps[0], primaryFlex, 0)

	// check if any secondary axis opts into coordination via PreferNiceIntervals
	anySecondaryNice := false
	for i := 1; i < n; i++ {
		if flagIs(true, preferNice[i]) {
			anySecondaryNice = true
			break
		}
	}

	if !anySecondaryNice {
		// secondary axes adopt the primary's label count directly
		result := make([]axisRange, n)
		result[0] = finalizeValueAxisRange(p, preps[0], primaryMin, primaryMax, primaryCount)
		for i := 1; i < n; i++ {
			mn, mx, count := resolveValueAxisRange(preps[i], false, primaryCount)
			result[i] = finalizeValueAxisRange(p, preps[i], mn, mx, count)
		}
		return result
	}

	// at least one secondary has PreferNiceIntervals - resolve all independently then search
	type naturalResult struct {
		min, max   float64
		labelCount int
	}
	naturals := make([]naturalResult, n)
	naturals[0] = naturalResult{primaryMin, primaryMax, primaryCount}
	for i := 1; i < n; i++ {
		flexCount := flagIs(true, preferNice[i])
		mn, mx, count := resolveValueAxisRange(preps[i], flexCount, 0)
		naturals[i] = naturalResult{mn, mx, count}
	}

	// if counts already match, finalize with natural resolutions
	allMatch := true
	for i := 1; i < n; i++ {
		if naturals[i].labelCount != naturals[0].labelCount {
			allMatch = false
			break
		}
	}
	if allMatch {
		result := make([]axisRange, n)
		for i, prep := range preps {
			result[i] = finalizeValueAxisRange(p, prep, naturals[i].min, naturals[i].max, naturals[i].labelCount)
		}
		return result
	}

	// search for the best shared count
	minNatural, maxNatural := naturals[0].labelCount, naturals[0].labelCount
	minMaxLabel := preps[0].maxLabelCount
	for i := 1; i < n; i++ {
		if naturals[i].labelCount < minNatural {
			minNatural = naturals[i].labelCount
		}
		if naturals[i].labelCount > maxNatural {
			maxNatural = naturals[i].labelCount
		}
		if preps[i].maxLabelCount < minMaxLabel {
			minMaxLabel = preps[i].maxLabelCount
		}
	}

	searchMin := chartdraw.MaxInt(minNatural-3, minimumAxisLabels)
	searchMax := chartdraw.MinInt(maxNatural+3, minMaxLabel)

	bestCount := naturals[0].labelCount
	bestScore := math.Inf(1)

	for c := searchMin; c <= searchMax; c++ {
		var score float64
		for i, prep := range preps {
			mn, mx, count := resolveValueAxisRange(prep, false, c)
			if count > 1 {
				interval := (mx - mn) / float64(count-1)
				if interval > 0 {
					ni := niceNum(interval)
					if math.Abs(ni-interval) > 1e-10 {
						score += 100
					}
				}
			}
			if excess := (mx - prep.maxVal) + (prep.minVal - mn); excess > 0 {
				score += excess
			}
			score += math.Abs(float64(c-naturals[i].labelCount)) * 10
		}

		if score < bestScore-1e-10 {
			bestScore = score
			bestCount = c
		}
	}

	result := make([]axisRange, n)
	for i, prep := range preps {
		mn, mx, count := resolveValueAxisRange(prep, false, bestCount)
		result[i] = finalizeValueAxisRange(p, prep, mn, mx, count)
	}
	return result
}

// calculateValueAxisRange centralizes numeric axis logic, selecting human-friendly scale and label count.
func calculateValueAxisRange(p *Painter, isVertical bool, axisSize int,
	minCfg, maxCfg, rangeValuePaddingScale *float64,
	labelsCfg []string, dataStartIndex int,
	labelCountCfg int, labelUnit float64, labelCountAdjustment int,
	seriesList seriesList, yAxisIndex int, stackSeries bool,
	valueFormatter ValueFormatter,
	labelRotation float64, fontStyle FontStyle,
	preferNiceIntervals *bool) axisRange {
	prep := prepareValueAxisRange(p, isVertical, axisSize,
		minCfg, maxCfg, rangeValuePaddingScale,
		labelsCfg, dataStartIndex,
		labelCountCfg, labelUnit, labelCountAdjustment,
		seriesList, yAxisIndex, stackSeries,
		valueFormatter, labelRotation, fontStyle)
	// TODO - in v0.6.0 default flexCount if labelCountCfg == 0 && !flagIs(false, preferNiceIntervals)
	flexCount := flagIs(true, preferNiceIntervals)
	minPadded, maxPadded, labelCount := resolveValueAxisRange(&prep, flexCount, 0)
	return finalizeValueAxisRange(p, &prep, minPadded, maxPadded, labelCount)
}

// calculateCategoryAxisRange does the same for category axes (common for x-axis in line/bar charts).
func calculateCategoryAxisRange(p *Painter, axisSize int, isVertical bool, extraSpace bool,
	labels []string, dataStartIndex int,
	labelCountCfg int, labelCountAdjustment int, labelUnit float64,
	seriesList seriesList, labelRotation float64, fontStyle FontStyle) axisRange {
	// If user provided no labels, use series names.
	// If provided only partially, fill in the remaining labels.
	for i := len(labels); i < getSeriesMaxDataCount(seriesList); i++ {
		labels = append(labels, strconv.Itoa(i+1))
	}
	dataCount := len(labels)

	textW, textH := p.measureTextMaxWidthHeight(labels, labelRotation, fontStyle)

	labelCount := labelCountCfg
	if labelCount <= 0 {
		labelCount = dataCount
	} else if labelCount > dataCount {
		labelCount = dataCount
	}
	labelCount = chartdraw.MaxInt(labelCount+labelCountAdjustment, minimumAxisLabels)
	// validate the labels fit, otherwise reduce the count
	if labelCountCfg == 0 {
		maxLabelCount := labelCount
		if isVertical {
			if textH > 0 {
				var extra int
				if extraSpace {
					extra = 10
				}
				maxLabelCount = chartdraw.MaxInt(axisSize/(textH+extra), minimumAxisLabels)
			}
		} else {
			if textW > 0 {
				// add a little extra padding for horizontal layouts
				extra := textW
				if !extraSpace {
					extra /= 2
				}
				maxLabelCount = chartdraw.MaxInt(axisSize/(textW+extra), minimumAxisLabels)
			}
		}
		if labelUnit > 0 {
			// If the user gave a 'unit', figure out how many 'units' fit
			multiplier := 1.0
			for {
				count := ceilFloatToInt(float64(dataCount) / (labelUnit * multiplier))
				if count > maxLabelCount {
					multiplier++
				} else {
					labelCount = chartdraw.MaxInt(count, minimumAxisLabels)
					break
				}
			}
		} else if maxLabelCount < labelCount {
			// Instead of a slight reduction, we choose a skip factor (step) so that we skip every other label until
			// we are within our limit.
			step := 1
			candidateCount := 2 + (dataCount-2)/step
			for candidateCount > maxLabelCount {
				step++
				candidateCount = 2 + (dataCount-2)/step
			}
			labelCount = chartdraw.MaxInt(candidateCount, minimumAxisLabels)
		}
	}
	// ensure there are not too many ticks, we want them relative and related to the label positions
	tickCount := dataCount
	if tickCount > labelCount*2 {
		// it's difficult to choose a tick count that allows multiple ticks while staying lined up with the labels
		// TODO - I would like to improve this, but for simplicity we will match the label count if ticks are too dense
		tickCount = labelCount
	}

	return axisRange{
		isCategory:     true,
		labels:         labels,
		dataStartIndex: dataStartIndex,
		divideCount:    dataCount,
		tickCount:      tickCount,
		labelCount:     labelCount,
		size:           axisSize,
		textMaxWidth:   textW,
		textMaxHeight:  textH,
		labelRotation:  labelRotation,
		labelFontStyle: fontStyle,
	}
}

func valueLabels(labelsCfg []string, valueFormatter ValueFormatter, min, max float64, labelCount int) []string {
	labels := make([]string, labelCount)
	offset := (max - min) / float64(labelCount-1)
	for i := range labels {
		if i < len(labelsCfg) {
			labels[i] = labelsCfg[i]
		} else {
			labels[i] = valueFormatter(min + float64(i)*offset)
		}
	}
	return labels
}

var niceNums = [...]float64{1, 2, 2.5, 5}

// niceNum returns the smallest "nice" number >= val from {1, 2, 2.5, 5} × 10^n.
func niceNum(val float64) float64 {
	if val <= 0 {
		return 0
	}
	exp := math.Floor(math.Log10(val))
	frac := val / math.Pow(10, exp)
	for _, n := range niceNums {
		if n >= frac-1e-10 {
			return n * math.Pow(10, exp)
		}
	}
	return math.Pow(10, exp+1)
}

func padRange(divideCount int, min, max, minPaddingScale, maxPaddingScale float64, flexCount bool) (float64, float64, int) {
	if minPaddingScale <= 0.0 && maxPaddingScale <= 0.0 {
		return min, max, divideCount
	}
	// scale percents for min value
	scaledMinPadPercentMin := rangeMinPaddingPercentMin * minPaddingScale
	scaledMinPadPercentMax := rangeMinPaddingPercentMax * minPaddingScale
	// scale percents for max value
	scaledMaxPadPercentMin := rangeMaxPaddingPercentMin * maxPaddingScale
	scaledMaxPadPercentMax := rangeMaxPaddingPercentMax * maxPaddingScale
	minResult := min
	spanIncrement := (max - min) * 0.01 // must be 1% of the span
	var spanIncrementMultiplier float64
	// find a min value to start our range from
	// we prefer (in order, negative if necessary), 0, 1, 10, 100, ..., 2, 20, ..., 5, 50, ...
	var updatedMin bool
	var passedLowTarget bool // tracks if we continued past a too-low target for positive min
rootLoop:
	for _, multiple := range []float64{1.0, 2.0, 5.0} {
		if min < 0 {
			multiple *= -1 // convert multiple sign to adjust targetVal correctly
		}
	expoLoop:
		for expo := -1.0; expo < 6; expo++ {
			if expo == -1.0 && multiple != 1.0 {
				continue expoLoop // we only want to test targetVal 0 once
			}
			// use 10^expo so that we prefer 0, 10, 100, etc numbers
			targetVal := math.Floor(math.Pow(10, expo)) * multiple // Math.Floor to convert 0.1 from -1 expo into 0
			if targetVal < min-(spanIncrement*scaledMinPadPercentMax) {
				if min < 0 {
					break expoLoop // negative min means targets decrease, no match possible
				}
				passedLowTarget = true
			} else if targetVal <= min-(spanIncrement*scaledMinPadPercentMin) {
				// targetVal can be between our span increment increases, calculate and set result
				updatedMin = true
				spanIncrementMultiplier = (min - targetVal) / spanIncrement
				minResult = targetVal
				break rootLoop
			} else if min >= 0 {
				break expoLoop // past acceptable range for positive data
			}
		}
	}
	if !updatedMin {
		minResult, spanIncrementMultiplier =
			friendlyRound(min, spanIncrement, scaledMinPadPercentMin,
				scaledMinPadPercentMin, scaledMinPadPercentMax, false)
	} else if passedLowTarget && math.Abs(max) >= 10 {
		// the target was only reachable because we continued past too-low targets for positive min,
		// verify this produces a clean interval — a target far from the data can inflate the max
		// rounding and produce ugly label intervals
		frMin, frMult := friendlyRound(min, spanIncrement, scaledMinPadPercentMin,
			scaledMinPadPercentMin, scaledMinPadPercentMax, false)
		if frTrunk := math.Trunc(frMin); frTrunk <= min-(spanIncrement*scaledMinPadPercentMin) {
			frMin = frTrunk
		}
		targetMin := minResult
		if trunk := math.Trunc(targetMin); trunk <= min-(spanIncrement*scaledMinPadPercentMin) {
			targetMin = trunk
		}
		spanCount := float64(divideCount - 1)
		incrementPerLabel := spanIncrement / spanCount
		intervalQuality := func(candidateMin, mult float64) float64 {
			interval := (max - candidateMin) / spanCount
			ri, _ := friendlyRound(interval, incrementPerLabel,
				math.Max(mult, scaledMaxPadPercentMin),
				scaledMaxPadPercentMin, scaledMaxPadPercentMax, true)
			finalMax := candidateMin + ri*spanCount
			if trunk := math.Trunc(finalMax); trunk >= max+(spanIncrement*scaledMaxPadPercentMin) {
				finalMax = trunk
			}
			fi := (finalMax - candidateMin) / spanCount
			// score: fractional part at 1 decimal place — 0 is a clean integer interval
			scaled := fi * 10
			return math.Abs(scaled - math.Round(scaled))
		}
		targetQ := intervalQuality(targetMin, spanIncrementMultiplier)
		frQ := intervalQuality(frMin, frMult)
		if frQ < targetQ-1e-6 {
			minResult = frMin
			spanIncrementMultiplier = frMult
		}
	}
	if minTrunk := math.Trunc(minResult); minTrunk <= min-(spanIncrement*scaledMinPadPercentMin) {
		minResult = minTrunk // remove possible float multiplication inaccuracies
	}

	if max == minResult {
		// no adjustment was made and there is no span, because of that the max calculation below can't function
		// for that reason we apply a default constant span, still wanting to prefer a zero start if possible
		if minResult == 0 {
			return minResult, minResult + (2 * zeroSpanAdjustment), divideCount
		}
		return minResult - zeroSpanAdjustment, minResult + zeroSpanAdjustment, divideCount
	} else if maxPaddingScale <= 0.0 {
		return minResult, max, divideCount
	}

	adjustedCount := divideCount

	// always compute the friendlyRound baseline as the default and cap for the flex path
	var baselineMax float64
	if math.Abs(max) < 10 {
		baselineMax = math.Ceil(max) + 1
	} else {
		interval := (max - minResult) / float64(divideCount-1)
		roundedInterval, _ := friendlyRound(interval, spanIncrement/float64(divideCount-1),
			math.Max(spanIncrementMultiplier, scaledMaxPadPercentMin),
			scaledMaxPadPercentMin, scaledMaxPadPercentMax, true)
		baselineMax = minResult + (roundedInterval * float64(divideCount-1))
		if maxTrunk := math.Trunc(baselineMax); maxTrunk >= max+(spanIncrement*scaledMaxPadPercentMin) {
			baselineMax = maxTrunk // remove possible float multiplication inaccuracies
		}
	}
	maxResult := baselineMax

	if flexCount {
		// attempt to find a nice-number interval by flexing divideCount ±3, capped by baseline
		minPadRequired := max + spanIncrement*scaledMaxPadPercentMin
		baselineExcess := baselineMax - max
		// cap is the tighter of a percentage-based limit and the baseline-derived limit
		maxPadLimit := math.Min(
			max+spanIncrement*scaledMaxPadPercentMax*1.4,
			baselineMax+baselineExcess*8,
		)
		bestScore := math.Inf(1)
		for delta := -3; delta <= 3; delta++ {
			dc := divideCount + delta
			if dc < minimumAxisLabels {
				continue
			}
			ni := niceNum((max - minResult) / float64(dc-1))
			if ni <= 0 {
				continue
			}
			candidateMax := minResult + ni*float64(dc-1)
			if candidateMax < minPadRequired-1e-10 || candidateMax > maxPadLimit+1e-10 {
				continue
			}
			absDelta := int(math.Abs(float64(delta)))
			excess := candidateMax - max
			// prefer ±1 label count change over larger changes, then minimize excess
			score := excess
			if absDelta > 1 {
				score = 1e15 + float64(absDelta)*1e10 + excess
			}
			if score < bestScore-1e-10 {
				bestScore = score
				maxResult = candidateMax
				adjustedCount = dc
			}
		}
	}

	return minResult, maxResult, adjustedCount
}

func friendlyRound(val, increment, defaultMultiplier, minMultiplier, maxMultiplier float64, add bool) (float64, float64) {
	absVal := math.Abs(val)
	startOOM := math.Floor(math.Log10(absVal))
	// for sub-unit values extend to finer-grained rounding, but only when rounding up (add=true)
	lowerBound := 1.0
	if absVal > 0 && startOOM < 0 && add {
		lowerBound = startOOM - 1
	}
	for orderOfMagnitude := startOOM; orderOfMagnitude >= lowerBound; orderOfMagnitude-- {
		roundValue := math.Pow(10, orderOfMagnitude)
		var proposedVal float64
		var proposedMultiplier float64
		for roundAdjust := 0.0; roundAdjust < 9.0; roundAdjust++ {
			if add {
				proposedVal = (math.Ceil(absVal/roundValue) * roundValue) + (roundValue * roundAdjust)
			} else {
				proposedVal = (math.Floor(absVal/roundValue) * roundValue) + (roundValue * roundAdjust)
			}
			if val < 0 { // Apply the original sign back to proposedVal
				proposedVal *= -1
			}
			if add {
				proposedMultiplier = (proposedVal - val) / increment
			} else {
				proposedMultiplier = (val - proposedVal) / increment
			}

			if proposedMultiplier > maxMultiplier {
				break // shortcut inner loop as multiplier will only go up
			} else if proposedMultiplier > minMultiplier {
				return proposedVal, proposedMultiplier
			}
		}
		if proposedMultiplier <= minMultiplier {
			break // shortcut outer loop if multiplier is below the min after adjust check, as this will only get smaller
		}
	}
	// No match found, let's see if we can just round to the next whole number
	if (increment*maxMultiplier) >= 1.0 && val != math.Trunc(val) {
		var proposedVal float64
		var proposedMultiplier float64
		if add {
			proposedVal = math.Ceil(val)
			proposedMultiplier = (proposedVal - val) / increment
		} else {
			proposedVal = math.Floor(val)
			proposedMultiplier = (val - proposedVal) / increment
		}
		return proposedVal, proposedMultiplier
	}
	// else, just adjust based off default multiplier
	if add {
		return val + (increment * defaultMultiplier), defaultMultiplier
	} else {
		return val - (increment * defaultMultiplier), defaultMultiplier
	}
}

func (r axisRange) getHeight(value float64) int {
	if r.max <= r.min {
		return 0
	}
	v := (value - r.min) / (r.max - r.min)
	// Clamp the result to valid range to prevent infinite loops with extreme values
	result := int(v * float64(r.size))
	if result < 0 {
		return 0
	} else if result > r.size {
		return r.size
	}
	return result
}

func (r axisRange) getRestHeight(value float64) int {
	return r.size - r.getHeight(value)
}

// getRange returns a range at a given index.
func (r axisRange) getRange(index int) (float64, float64) {
	unit := float64(r.size) / float64(r.divideCount)
	return unit * float64(index), unit * float64(index+1)
}

// autoDivide divides the axis size by the configured count.
func (r axisRange) autoDivide() []int {
	return autoDivide(r.size, r.divideCount)
}
