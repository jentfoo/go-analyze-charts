package charts

import (
	"math"
	"strconv"
	"testing"

	"github.com/dustin/go-humanize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeBasicFunnelChartOption() FunnelChartOption {
	return FunnelChartOption{
		SeriesList: NewSeriesListFunnel([]float64{
			100, 80, 60, 40, 20,
		}),
		Legend: LegendOption{
			SeriesNames: []string{"Show", "Click", "Visit", "Inquiry", "Order"},
		},
		Title: TitleOption{
			Text: "Funnel",
		},
	}
}

func TestNewFunnelChartOptionWithData(t *testing.T) {
	t.Parallel()

	opt := NewFunnelChartOptionWithData([]float64{12, 24, 48})

	assert.Len(t, opt.SeriesList, 3)
	assert.Equal(t, ChartTypeFunnel, opt.SeriesList[0].getType())
	assert.Equal(t, defaultPadding, opt.Padding)

	p := NewPainter(PainterOptions{})
	assert.NoError(t, p.FunnelChart(opt))
}

func TestFunnelChart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		themed      bool
		makeOptions func() FunnelChartOption
		pngCRC      uint32
	}{
		{
			name:        "basic_themed",
			themed:      true,
			makeOptions: makeBasicFunnelChartOption,
			pngCRC:      0x304e4d0d,
		},
		{
			name: "custom_legend",
			makeOptions: func() FunnelChartOption {
				opt := makeBasicFunnelChartOption()
				opt.Legend.Symbol = SymbolDot
				opt.Legend.FontStyle = NewFontStyleWithSize(4.0)
				opt.Legend.Vertical = Ptr(true)
				opt.Legend.Offset = OffsetStr{
					Left: PositionRight,
					Top:  PositionBottom,
				}
				opt.Title.Show = Ptr(false)
				return opt
			},
			pngCRC: 0x18b3583c,
		},
		{
			name: "with_label_formatter",
			makeOptions: func() FunnelChartOption {
				return FunnelChartOption{
					SeriesList: NewSeriesListFunnel([]float64{
						100, 80, 60, 40, 20,
					}, FunnelSeriesOption{
						Names: []string{"Show", "Click", "Visit", "Inquiry", "Order"},
						Label: SeriesLabel{
							Show: Ptr(true),
							LabelFormatter: func(index int, name string, val float64) (string, *LabelStyle) {
								if index == 1 || index == 3 { // highlight 2nd and 4th items
									return "⭐ " + name + ": " + strconv.FormatFloat(val, 'f', 0, 64), nil
								}
								return "", nil // hide other labels
							},
						},
					}),
					Legend: LegendOption{
						Show: Ptr(false),
					},
				}
			},
			pngCRC: 0x1b70bafa,
		},
		{
			name: "with_styled_labels",
			makeOptions: func() FunnelChartOption {
				return FunnelChartOption{
					SeriesList: NewSeriesListFunnel([]float64{
						100, 80, 60, 40, 20,
					}, FunnelSeriesOption{
						Names: []string{"Show", "Click", "Visit", "Inquiry", "Order"},
						Label: SeriesLabel{
							Show: Ptr(true),
							LabelFormatter: func(index int, name string, val float64) (string, *LabelStyle) {
								switch index {
								case 0: // first item - red background with rounded corners
									return name, &LabelStyle{
										FontStyle:       FontStyle{FontColor: ColorWhite, FontSize: 14},
										BackgroundColor: ColorRed,
										CornerRadius:    5,
									}
								case 1: // second item - blue background, larger font
									return name, &LabelStyle{
										FontStyle:       FontStyle{FontColor: ColorWhite, FontSize: 16},
										BackgroundColor: ColorBlue,
										CornerRadius:    3,
										BorderColor:     ColorPurple,
										BorderWidth:     2,
									}
								case 2: // third item - green background, square corners
									return "🟢 " + name, &LabelStyle{
										FontStyle:       FontStyle{FontColor: ColorBlack, FontSize: 12},
										BackgroundColor: ColorLime,
										BorderColor:     ColorRed,
										BorderWidth:     2,
									}
								case 3: // fourth item - no background, custom color
									return "⭐ " + name, &LabelStyle{
										FontStyle: FontStyle{FontColor: ColorOrange, FontSize: 15},
									}
								default: // last item - no label
									return "", nil
								}
							},
						},
					}),
					Legend: LegendOption{
						Show: Ptr(false),
					},
				}
			},
			pngCRC: 0x58d2b5a6,
		},
		{
			name: "border_without_background",
			makeOptions: func() FunnelChartOption {
				values := []float64{100, 50}
				opt := NewFunnelChartOptionWithData(values)
				opt.SeriesList[0].Label = SeriesLabel{
					Show: Ptr(true),
					LabelFormatter: func(index int, name string, val float64) (string, *LabelStyle) {
						return "test label", &LabelStyle{
							BorderColor: ColorRed,
							BorderWidth: 2.5,
						}
					},
				}
				return opt
			},
			pngCRC: 0xfaef52c9,
		},
		{
			name: "unsorted",
			makeOptions: func() FunnelChartOption {
				return NewFunnelChartOptionWithData([]float64{50, 100, 25})
			},
			pngCRC: 0x7a147f14,
		},
		{
			name: "empty_series",
			makeOptions: func() FunnelChartOption {
				opt := NewFunnelChartOptionWithData([]float64{})
				opt.Padding = NewBoxEqual(10)
				opt.Legend = LegendOption{
					Show:        Ptr(true),
					SeriesNames: []string{"Series A"},
				}
				return opt
			},
			pngCRC: 0x83721fd3,
		},
		{
			name: "null_values",
			makeOptions: func() FunnelChartOption {
				opt := makeBasicFunnelChartOption()
				// a middle and a trailing null: neither band is drawn, and the band above each
				// keeps a flat bottom rather than tapering into the gap
				opt.SeriesList = NewSeriesListFunnel([]float64{
					100, GetNullValue(), 60, 40, GetNullValue(),
				})
				return opt
			},
			pngCRC: 0x33680261,
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

				validateFunnelChartRender(t, p, rp, tt.makeOptions(), tt.pngCRC)
			})
			t.Run(strconv.Itoa(i)+"-"+tt.name+"-options", func(t *testing.T) {
				p := NewPainter(painterOptions)
				rp := NewPainter(rasterOptions)
				opt := tt.makeOptions()
				opt.Theme = GetTheme(ThemeVividDark)

				validateFunnelChartRender(t, p, rp, opt, tt.pngCRC)
			})
		} else {
			t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
				p := NewPainter(painterOptions)
				rp := NewPainter(rasterOptions)

				validateFunnelChartRender(t, p, rp, tt.makeOptions(), tt.pngCRC)
			})
		}
	}
}

func validateFunnelChartRender(t *testing.T, svgP, pngP *Painter, opt FunnelChartOption, expectedCRC uint32) {
	t.Helper()

	err := svgP.FunnelChart(opt)
	require.NoError(t, err)
	data, err := svgP.Bytes()
	require.NoError(t, err)
	assertTestdataSVG(t, data)

	err = pngP.FunnelChart(opt)
	require.NoError(t, err)
	rasterData, err := pngP.Bytes()
	require.NoError(t, err)
	assertEqualPNGCRC(t, expectedCRC, rasterData)
}

func TestFunnelChartError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		makeOptions      func() FunnelChartOption
		errorMsgContains string
	}{
		{
			name: "negative_value",
			makeOptions: func() FunnelChartOption {
				return NewFunnelChartOptionWithData([]float64{10.0, -1.0})
			},
			errorMsgContains: "unsupported negative value",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			p := NewPainter(PainterOptions{
				OutputFormat: ChartOutputSVG,
				Width:        600,
				Height:       400,
			})

			err := p.FunnelChart(tt.makeOptions())
			require.Error(t, err)
			require.ErrorContains(t, err, tt.errorMsgContains)
		})
	}
}

func TestFunnelChartNullValue(t *testing.T) {
	t.Parallel()

	t.Run("all_null", func(t *testing.T) {
		p := NewPainter(PainterOptions{OutputFormat: ChartOutputSVG, Width: 600, Height: 400})
		require.NoError(t, p.FunnelChart(NewFunnelChartOptionWithData([]float64{GetNullValue(), GetNullValue()})))
		data, err := p.Bytes()
		require.NoError(t, err)

		assert.NotContains(t, string(data), "(0%)") // no series rendered
	})
	t.Run("non_finite_skipped", func(t *testing.T) {
		for _, v := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
			p := NewPainter(PainterOptions{OutputFormat: ChartOutputSVG, Width: 600, Height: 400})
			req := require.New(t)
			req.NoError(p.FunnelChart(NewFunnelChartOptionWithData([]float64{v, 5})))
			data, err := p.Bytes()
			req.NoError(err)

			// the non-finite band gets no label, like a null; only the real band's label renders
			assert.NotContains(t, string(data), "(0%)")
		}
	})
	t.Run("null_skips_formatters", func(t *testing.T) {
		var received []float64
		opt := NewFunnelChartOptionWithData([]float64{GetNullValue(), 5})
		for i := range opt.SeriesList {
			opt.SeriesList[i].Label.ValueFormatter = func(v float64) string {
				received = append(received, v)
				return humanize.FtoaWithDigits(v, 2)
			}
			opt.SeriesList[i].Label.LabelFormatter = func(_ int, name string, v float64) (string, *LabelStyle) {
				received = append(received, v)
				return name + humanize.FtoaWithDigits(v, 2), nil
			}
		}
		p := NewPainter(PainterOptions{OutputFormat: ChartOutputSVG, Width: 600, Height: 400})
		req := require.New(t)
		req.NoError(p.FunnelChart(opt))

		// the null series is never passed to a user formatter; only the real value arrives
		assert.NotContains(t, received, GetNullValue())
		assert.Contains(t, received, 5.0)
	})
}
