package charts

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeBasicRadarChartOption() RadarChartOption {
	values := [][]float64{
		{4200, 3000, 20000, 35000, 50000, 18000},
		{5000, 14000, 28000, 26000, 42000, 21000},
	}
	return RadarChartOption{
		SeriesList: NewSeriesListRadar(values),
		Title: TitleOption{
			Text: "Basic Radar Chart",
		},
		Legend: LegendOption{
			SeriesNames: []string{"Allocated Budget", "Actual Spending"},
		},
		RadarIndicators: NewRadarIndicators([]string{
			"Sales",
			"Administration",
			"Information Technology",
			"Customer Support",
			"Development",
			"Marketing",
		}, []float64{
			6500, 16000, 30000, 38000, 52000, 25000,
		}),
	}
}

func TestNewRadarChartOptionWithData(t *testing.T) {
	t.Parallel()

	opt := NewRadarChartOptionWithData([][]float64{
		{4200, 3000, 20000, 35000, 50000, 18000},
		{5000, 14000, 28000, 26000, 42000, 21000},
	}, []string{
		"Sales",
		"Administration",
		"Information Technology",
		"Customer Support",
		"Development",
		"Marketing",
	}, []float64{
		6500, 16000, 30000, 38000, 52000, 25000,
	})

	assert.Len(t, opt.SeriesList, 2)
	assert.Equal(t, ChartTypeRadar, opt.SeriesList[0].getType())
	assert.Equal(t, defaultPadding, opt.Padding)

	p := NewPainter(PainterOptions{})
	assert.NoError(t, p.RadarChart(opt))
}

func TestRadarChart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		themed      bool
		makeOptions func() RadarChartOption
		pngCRC      uint32
	}{
		{
			name:        "basic_themed",
			themed:      true,
			makeOptions: makeBasicRadarChartOption,
			pngCRC:      0x2aaa6088,
		},
		{
			name: "empty_series",
			makeOptions: func() RadarChartOption {
				opt := NewRadarChartOptionWithData([][]float64{}, []string{"Sales", "Admin", "IT"}, []float64{100, 200, 300})
				opt.Padding = NewBoxEqual(10)
				opt.Legend = LegendOption{
					Show:        Ptr(true),
					SeriesNames: []string{"Budget", "Spending"},
				}
				return opt
			},
			pngCRC: 0xdbb70938,
		},
		{
			name: "empty_values_series_skipped",
			makeOptions: func() RadarChartOption {
				opt := makeBasicRadarChartOption()
				opt.SeriesList = append(opt.SeriesList, RadarSeries{Name: "Empty"})
				return opt
			},
			pngCRC: 0xb163320f,
		},
		{
			name: "all_series_empty_values",
			makeOptions: func() RadarChartOption {
				return NewRadarChartOptionWithData([][]float64{{}},
					[]string{"Sales", "Admin", "IT"}, []float64{100, 200, 300})
			},
			pngCRC: 0x1138d17a,
		},
		{
			name: "values_longer_than_indicators",
			makeOptions: func() RadarChartOption {
				opt := NewRadarChartOptionWithData([][]float64{{100, 200, 300, 400}},
					[]string{"Sales", "Admin", "IT"}, []float64{100, 200, 300})
				opt.SeriesList.SetSeriesLabels(SeriesLabel{Show: Ptr(true)})
				return opt
			},
			pngCRC: 0x6e141545,
		},
		{
			name: "null_values",
			makeOptions: func() RadarChartOption {
				opt := makeBasicRadarChartOption()
				// one null spoke in each series to demo the polygon break and skipped label
				values := [][]float64{
					{4200, GetNullValue(), 20000, 35000, GetNullValue(), 18000},
					{5000, 14000, GetNullValue(), 26000, 22000, 21000},
				}
				opt.SeriesList = NewSeriesListRadar(values)
				opt.SeriesList.SetSeriesLabels(SeriesLabel{Show: Ptr(true)})
				return opt
			},
			pngCRC: 0xcd355d85,
		},
		{
			name: "null_values_auto_max",
			makeOptions: func() RadarChartOption {
				opt := makeBasicRadarChartOption()
				// a null spoke must not poison the auto indicator max, which would otherwise
				// collapse every real series toward the center on that spoke
				values := [][]float64{
					{GetNullValue(), 5, 7, 9, 11, 13},
					{0, 9, 11, 13, 15, 17},
				}
				opt.SeriesList = NewSeriesListRadar(values)
				for i := range opt.RadarIndicators {
					opt.RadarIndicators[i].Max = 0 // auto-compute from the series values
				}
				return opt
			},
			pngCRC: 0xceebd205,
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
		if tt.themed {
			t.Run(strconv.Itoa(i)+"-"+tt.name+"-painter", func(t *testing.T) {
				p := NewPainter(painterOptions, PainterThemeOption(GetTheme(ThemeVividDark)))
				rp := NewPainter(rasterOptions, PainterThemeOption(GetTheme(ThemeVividDark)))

				validateRadarChartRender(t, p, rp, tt.makeOptions(), tt.pngCRC)
			})
			t.Run(strconv.Itoa(i)+"-"+tt.name+"-options", func(t *testing.T) {
				p := NewPainter(painterOptions)
				rp := NewPainter(rasterOptions)
				opt := tt.makeOptions()
				opt.Theme = GetTheme(ThemeVividDark)

				validateRadarChartRender(t, p, rp, opt, tt.pngCRC)
			})
		} else {
			t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
				p := NewPainter(painterOptions)
				rp := NewPainter(rasterOptions)

				validateRadarChartRender(t, p.Child(PainterPaddingOption(NewBoxEqual(20))),
					rp.Child(PainterPaddingOption(NewBoxEqual(20))), tt.makeOptions(), tt.pngCRC)
			})
		}
	}
}

func validateRadarChartRender(t *testing.T, svgP, pngP *Painter, opt RadarChartOption, expectedCRC uint32) {
	t.Helper()

	err := svgP.RadarChart(opt)
	require.NoError(t, err)
	data, err := svgP.Bytes()
	require.NoError(t, err)
	assertTestdataSVG(t, data)

	err = pngP.RadarChart(opt)
	require.NoError(t, err)
	rasterData, err := pngP.Bytes()
	require.NoError(t, err)
	assertEqualPNGCRC(t, expectedCRC, rasterData)
}

func TestRadarChartNullValue(t *testing.T) {
	t.Parallel()

	render := func(values [][]float64) string {
		opt := makeBasicRadarChartOption()
		opt.SeriesList = NewSeriesListRadar(values)
		opt.SeriesList.SetSeriesLabels(SeriesLabel{Show: Ptr(true)})
		opt.RadarIndicators = opt.RadarIndicators[:3]
		p := NewPainter(PainterOptions{OutputFormat: ChartOutputSVG, Width: 600, Height: 400})
		req := require.New(t)
		req.NoError(p.RadarChart(opt))
		data, err := p.Bytes()
		req.NoError(err)
		return string(data)
	}

	t.Run("partial_null", func(t *testing.T) {
		s := render([][]float64{{4200, GetNullValue(), 5000}})

		// the null spoke gets no label, so no sentinel text reaches the output
		assert.NotContains(t, s, "e+308")
		// real spokes still render their value labels
		assert.Contains(t, s, ">4200<")
		assert.Contains(t, s, ">5000<")
	})
	t.Run("all_null", func(t *testing.T) {
		s := render([][]float64{{GetNullValue(), GetNullValue(), GetNullValue()}})

		// only the legend swatch carries the series color; no polygon vertices or dots are drawn
		assert.Equal(t, 1, strings.Count(s, "rgb(84,112,198"))
		assert.NotContains(t, s, `r="2" style="stroke-width:2;stroke`)
	})
}

func TestRadarChartError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		makeOptions      func() RadarChartOption
		errorMsgContains string
	}{
		{
			name: "too_few_indicators",
			makeOptions: func() RadarChartOption {
				return NewRadarChartOptionWithData([][]float64{{0.0}}, []string{"foo", "bar"}, []float64{1, 2})
			},
			errorMsgContains: "indicator count",
		},
		{
			name: "indicator_name_value_mismatch",
			makeOptions: func() RadarChartOption {
				return NewRadarChartOptionWithData([][]float64{{1, 2, 3}}, []string{"foo", "bar"}, []float64{1, 2, 3})
			},
			errorMsgContains: "indicator count",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			p := NewPainter(PainterOptions{
				OutputFormat: ChartOutputSVG,
				Width:        600,
				Height:       400,
			})

			err := p.RadarChart(tt.makeOptions())
			require.Error(t, err)
			require.ErrorContains(t, err, tt.errorMsgContains)
		})
	}
}
