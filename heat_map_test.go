package charts

import (
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHeatMapOptionWithData(t *testing.T) {
	data := [][]float64{
		{10, 20},
		{30, 40},
	}
	opt := NewHeatMapOptionWithData(data)
	require.Equal(t, data, opt.Values)

	p := NewPainter(PainterOptions{
		OutputFormat: ChartOutputSVG,
		Width:        600,
		Height:       400,
	})
	err := p.HeatMapChart(opt)
	require.NoError(t, err)
}

func makeBasicHeatMapOption() HeatMapOption {
	return HeatMapOption{
		Title: TitleOption{Text: "Heat Map"},
		Values: [][]float64{
			{1, 2, 3},
			{4, 5, 6},
			{7, 8, 9},
		},
		XAxis: HeatMapAxis{
			Title:  "X-Axis",
			Labels: []string{"A", "B", "C"},
		},
		YAxis: HeatMapAxis{
			Title:  "Y-Axis",
			Labels: []string{"Row1", "Row2", "Row3"},
		},
	}
}

func makeMinimalHeatMapOption() HeatMapOption {
	return NewHeatMapOptionWithData([][]float64{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	})
}

func makeDenseHeatMapOption() HeatMapOption {
	const size = 24
	values := make([][]float64, size)
	for i := range values {
		values[i] = make([]float64, size)
	}
	// create a grid of varying intensity
	for _, index := range []int{6, 12, 18} {
		for i := 0; i < size; i++ {
			inc := 1.0
			if i%2 == 0 {
				inc *= 2
			}
			if index == 12 { // make center line more intense
				inc *= 2
			}
			values[index][i] += inc
			values[i][index] += inc
		}
	}

	opt := NewHeatMapOptionWithData(values)
	opt.BaseColorIndex = 1
	opt.XAxis.LabelCount = size
	opt.YAxis.LabelCount = size
	return opt
}

func TestHeatMapChart(t *testing.T) {
	tests := []struct {
		name        string
		themed      bool
		makeOptions func() HeatMapOption
		pngCRC      uint32
	}{
		{
			name:        "basic_themed",
			themed:      true,
			makeOptions: makeBasicHeatMapOption,
			pngCRC:      0x658d3e50,
		},
		{
			name: "scale_override",
			makeOptions: func() HeatMapOption {
				opt := makeMinimalHeatMapOption()
				minVal, maxVal := 0.0, 20.0
				opt.ScaleMinValue = &minVal
				opt.ScaleMaxValue = &maxVal
				return opt
			},
			pngCRC: 0x2170ba57,
		},
		{
			name: "values_label",
			makeOptions: func() HeatMapOption {
				opt := makeMinimalHeatMapOption()
				opt.ValuesLabel = SeriesLabel{
					Show: Ptr(true),
					FontStyle: FontStyle{
						FontSize:  14,
						FontColor: ColorBlue,
					},
					ValueFormatter: func(f float64) string {
						return strconv.FormatFloat(f, 'f', 0, 64)
					},
				}
				return opt
			},
			pngCRC: 0x311a12e,
		},
		{
			name: "varying_row_lengths",
			makeOptions: func() HeatMapOption {
				return NewHeatMapOptionWithData([][]float64{
					{1, 2, 3, 4},
					{5, 6},
					{7, 8, 9},
					nil,
				})
			},
			pngCRC: 0xc5996908,
		},
		{
			name:        "dense_data",
			makeOptions: makeDenseHeatMapOption,
			pngCRC:      0xecc2a3e9,
		},
		{
			name: "empty_values",
			makeOptions: func() HeatMapOption {
				return HeatMapOption{
					Padding: NewBoxEqual(10),
					Values:  [][]float64{},
					XAxis:   HeatMapAxis{Title: "X-Axis", Labels: []string{"A", "B", "C"}},
					YAxis:   HeatMapAxis{Title: "Y-Axis", Labels: []string{"Row1", "Row2"}},
				}
			},
			pngCRC: 0x8548d55a,
		},
		{
			name: "no_columns",
			makeOptions: func() HeatMapOption {
				return HeatMapOption{
					Padding: NewBoxEqual(10),
					Values:  [][]float64{{}, {}},
					XAxis:   HeatMapAxis{Title: "X-Axis", Labels: []string{"A", "B", "C"}},
					YAxis:   HeatMapAxis{Title: "Y-Axis", Labels: []string{"Row1", "Row2"}},
				}
			},
			pngCRC: 0x8548d55a,
		},
		{
			name: "non_square",
			makeOptions: func() HeatMapOption {
				opt := NewHeatMapOptionWithData([][]float64{
					{1, 2, 3, 4, 5},
					{6, 7, 8, 9, 10},
					{11, 12, 13, 14, 15},
				})
				opt.XAxis.Labels = []string{"C1", "C2", "C3", "C4", "C5"}
				opt.YAxis.Labels = []string{"R1", "R2", "R3"}
				return opt
			},
			pngCRC: 0x2c8b2bab,
		},
		{
			name: "null_values",
			makeOptions: func() HeatMapOption {
				opt := NewHeatMapOptionWithData([][]float64{
					{1, 2, GetNullValue(), 4},
					{5, math.NaN(), 7, 8},
					{9, 10, GetNullValue(), 12},
				})
				opt.XAxis.Labels = []string{"A", "B", "C", "D"}
				opt.YAxis.Labels = []string{"Row1", "Row2", "Row3"}
				return opt
			},
			pngCRC: 0xb8664598,
		},
	}

	for i, tt := range tests {
		painterOptions := PainterOptions{
			OutputFormat: ChartOutputSVG,
			Width:        600,
			Height:       400,
		}
		rasterOptions := PainterOptions{
			OutputFormat: ChartOutputPNG,
			Width:        600,
			Height:       400,
		}
		if !tt.themed {
			t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
				p := NewPainter(painterOptions)
				rp := NewPainter(rasterOptions)
				validateHeatMapChartRender(t, p, rp, tt.makeOptions(), tt.pngCRC)
			})
		} else {
			theme := GetTheme(ThemeVividDark)
			t.Run(strconv.Itoa(i)+"-"+tt.name+"-theme_painter", func(t *testing.T) {
				p := NewPainter(painterOptions, PainterThemeOption(theme))
				rp := NewPainter(rasterOptions, PainterThemeOption(theme))
				validateHeatMapChartRender(t, p, rp, tt.makeOptions(), tt.pngCRC)
			})
			t.Run(strconv.Itoa(i)+"-"+tt.name+"-theme_opt", func(t *testing.T) {
				p := NewPainter(painterOptions)
				rp := NewPainter(rasterOptions)
				opt := tt.makeOptions()
				opt.Theme = theme
				validateHeatMapChartRender(t, p, rp, opt, tt.pngCRC)
			})
		}
	}
}

func validateHeatMapChartRender(t *testing.T, svgP, pngP *Painter, opt HeatMapOption, expectedCRC uint32) {
	t.Helper()

	err := svgP.HeatMapChart(opt)
	require.NoError(t, err)
	data, err := svgP.Bytes()
	require.NoError(t, err)
	assertTestdataSVG(t, data)

	err = pngP.HeatMapChart(opt)
	require.NoError(t, err)
	rasterData, err := pngP.Bytes()
	require.NoError(t, err)
	assertEqualPNGCRC(t, expectedCRC, rasterData)
}

func TestHeatMapChartError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		makeOptions      func() HeatMapOption
		errorMsgContains string
	}{
		{
			name: "insufficient_space",
			makeOptions: func() HeatMapOption {
				return HeatMapOption{
					Padding: NewBoxEqual(2),
					Values: [][]float64{
						{1, 2, 3},
						{4, 5, 6},
						{7, 8, 9},
					},
					XAxis: HeatMapAxis{
						Title:          "X-Axis",
						Labels:         []string{"A", "B", "C"},
						LabelFontStyle: FontStyle{FontSize: 10, Font: GetDefaultFont(), FontColor: ColorBlack},
					},
					YAxis: HeatMapAxis{
						Title:          "Y-Axis",
						Labels:         []string{"Row1", "Row2", "Row3"},
						LabelFontStyle: FontStyle{FontSize: 10, Font: GetDefaultFont(), FontColor: ColorBlack},
					},
				}
			},
			errorMsgContains: "insufficient space for heat map cells",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			var p *Painter
			if tt.name == "insufficient_space" {
				p = NewPainter(PainterOptions{
					Width:  40,
					Height: 40,
				})
			} else {
				p = NewPainter(PainterOptions{
					Width:  600,
					Height: 400,
				})
			}
			err := p.HeatMapChart(tt.makeOptions())
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errorMsgContains)
		})
	}
}

// cellFillColors returns the distinct rgb fill colors used by heat map cells, excluding axis/text fills.
func cellFillColors(svgData string) []string {
	fillRe := regexp.MustCompile(`fill:(rgb\([^)]*\)|#[0-9a-fA-F]+)`)
	var seen []string
	for _, m := range fillRe.FindAllStringSubmatch(svgData, -1) {
		if strings.HasPrefix(m[1], "rgb(70,") { // axis and text use the gray label color
			continue
		}
		if !slices.Contains(seen, m[1]) {
			seen = append(seen, m[1])
		}
	}
	return seen
}

func TestHeatMapChartNullValue(t *testing.T) {
	t.Parallel()

	render := func(values [][]float64, label bool) string {
		opt := NewHeatMapOptionWithData(values)
		if label {
			opt.ValuesLabel = SeriesLabel{Show: Ptr(true)}
		}
		p := NewPainter(PainterOptions{OutputFormat: ChartOutputSVG, Width: 600, Height: 400})
		req := require.New(t)
		req.NoError(p.HeatMapChart(opt))
		data, err := p.Bytes()
		req.NoError(err)
		return string(data)
	}

	t.Run("partial_null", func(t *testing.T) {
		full := render([][]float64{{1, 4}, {2, 3}}, false)
		part := render([][]float64{{1, GetNullValue()}, {2, 3}}, false)

		// the null cell draws no rect, so exactly one fewer filled path than the all-valid grid
		assert.Equal(t, strings.Count(full, "<path")-1, strings.Count(part, "<path"))
		// real cells still resolve to distinct shades, proving the scale did not flatten
		assert.GreaterOrEqual(t, len(cellFillColors(part)), 3)
	})
	t.Run("all_null", func(t *testing.T) {
		full := render([][]float64{{1, 4}, {2, 3}}, false)
		nulls := render([][]float64{{GetNullValue(), GetNullValue()}, {GetNullValue(), GetNullValue()}}, false)

		// no cells are drawn at all; only the fixed axis/background paths remain
		assert.Equal(t, strings.Count(full, "<path")-4, strings.Count(nulls, "<path"))
		assert.Empty(t, cellFillColors(nulls))
	})
	t.Run("nan_first_cell", func(t *testing.T) {
		v := render([][]float64{{math.NaN(), 2}, {3, 4}}, false)

		// a NaN in values[0][0] must not poison min/max seeding; remaining cells keep distinct shades
		assert.GreaterOrEqual(t, len(cellFillColors(v)), 3)
	})
	t.Run("no_sentinel_in_labels", func(t *testing.T) {
		s := render([][]float64{{1, GetNullValue()}, {2, math.NaN()}}, true)

		// null and NaN cells never reach the label path, so no sentinel or nan text appears
		assert.NotContains(t, s, "e+308")
		assert.NotContains(t, s, "NaN")
	})
	t.Run("non_finite_scale_override", func(t *testing.T) {
		values := [][]float64{{1, 4}, {2, 3}}
		renderScaled := func(min, max float64) string {
			opt := NewHeatMapOptionWithData(values)
			opt.ScaleMinValue, opt.ScaleMaxValue = &min, &max
			p := NewPainter(PainterOptions{OutputFormat: ChartOutputSVG, Width: 600, Height: 400})
			req := require.New(t)
			req.NoError(p.HeatMapChart(opt))
			data, err := p.Bytes()
			req.NoError(err)
			return string(data)
		}

		// a non-finite override is ignored, falling back to the computed extent rather than
		// producing a NaN ratio for every cell
		computed := render(values, false)
		for _, v := range []float64{math.NaN(), math.Inf(1), math.Inf(-1), GetNullValue()} {
			assert.Equal(t, computed, renderScaled(v, v))
		}
	})
}

func TestComputeMinMax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		values      [][]float64
		numCol      int
		expectedMin float64
		expectedMax float64
	}{
		{
			name:        "empty",
			values:      [][]float64{},
			numCol:      0,
			expectedMin: 0,
			expectedMax: 0,
		},
		{
			name: "uneven_rows",
			values: [][]float64{
				{1, 2, 3},
				{4},
			},
			numCol:      3,
			expectedMin: 0,
			expectedMax: 4,
		},
		{
			name: "negative_values",
			values: [][]float64{
				{-5, -2},
				{-3, -1},
			},
			numCol:      2,
			expectedMin: -5,
			expectedMax: -1,
		},
		{
			name: "default_column_padding", // real -1 stays in range alongside the default 0
			values: [][]float64{
				{-1},
				{},
			},
			numCol:      1,
			expectedMin: -1,
			expectedMax: 0,
		},
		{
			name: "ragged_negative",
			values: [][]float64{
				{-5, -3},
				{-2},
			},
			numCol:      2,
			expectedMin: -5,
			expectedMax: 0,
		},
		{
			name: "ragged_positive",
			values: [][]float64{
				{5, 8},
				{3},
			},
			numCol:      2,
			expectedMin: 0,
			expectedMax: 8,
		},
		{
			name: "ragged_mixed_sign",
			values: [][]float64{
				{-2, 4},
				{1},
			},
			numCol:      2,
			expectedMin: -2,
			expectedMax: 4,
		},
		{
			name: "null_value", // the sentinel is not a real value and must not poison the range
			values: [][]float64{
				{1, GetNullValue()},
				{2, 3},
			},
			numCol:      2,
			expectedMin: 1,
			expectedMax: 3,
		},
		{
			name: "nan_value", // NaN comparisons are always false, so it must be skipped explicitly
			values: [][]float64{
				{math.NaN(), 2},
				{1, 3},
			},
			numCol:      2,
			expectedMin: 1,
			expectedMax: 3,
		},
		{
			name: "inf_values", // ±Inf is not a real value either
			values: [][]float64{
				{math.Inf(1), 2},
				{1, math.Inf(-1)},
			},
			numCol:      2,
			expectedMin: 1,
			expectedMax: 2,
		},
		{
			name: "all_null", // no valid values, range defaults to zero
			values: [][]float64{
				{GetNullValue()},
				{math.NaN()},
			},
			numCol:      1,
			expectedMin: 0,
			expectedMax: 0,
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			min, max := computeMinMax(tt.values, tt.numCol)

			assert.InDelta(t, tt.expectedMin, min, 0.0)
			assert.InDelta(t, tt.expectedMax, max, 0.0)
		})
	}
}
