package charts

import (
	"strconv"
	"testing"

	"github.com/dustin/go-humanize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeBasicHorizontalBarChartOption() HorizontalBarChartOption {
	return HorizontalBarChartOption{
		Padding: NewBoxEqual(10),
		SeriesList: NewSeriesListHorizontalBar([][]float64{
			{18203, 23489, 29034, 104970, 131744, 630230},
			{19325, 23438, 31000, 121594, 134141, 681807},
		}),
		Title: TitleOption{
			Text: "World Population",
		},
		Legend: LegendOption{
			SeriesNames: []string{"2011", "2012"},
			Symbol:      SymbolDot,
		},
		YAxis: YAxisOption{
			Labels: []string{"Brazil", "Indonesia", "USA", "India", "China", "World"},
		},
	}
}

func makeMinimalHorizontalBarChartOption() HorizontalBarChartOption {
	opt := NewHorizontalBarChartOptionWithData([][]float64{
		{12, 24},
		{24, 48},
	})
	opt.YAxis = YAxisOption{
		Show:   Ptr(false),
		Labels: []string{"A", "B"},
	}
	opt.XAxis.Show = Ptr(false)
	return opt
}

func makeFullHorizontalBarChartStackedOption() HorizontalBarChartOption {
	seriesList := NewSeriesListHorizontalBar([][]float64{
		{4.9, 23.2, 25.6, 102.6, 142.2, 32.6, 20.0, 3.3},
		{19.0, 26.4, 28.7, 144.6, 122.2, 48.7, 28.8, 22.3},
		{80.0, 40.4, 28.4, 28.8, 24.4, 24.2, 40.8, 80.8},
	}, BarSeriesOption{
		Label: SeriesLabel{
			Show: Ptr(true),
			ValueFormatter: func(f float64) string {
				return strconv.Itoa(int(f))
			},
		},
	})
	return HorizontalBarChartOption{
		Padding:     NewBoxEqual(20),
		SeriesList:  seriesList,
		StackSeries: Ptr(true),
		Legend: LegendOption{
			Symbol: SymbolDot,
		},
		YAxis: YAxisOption{
			Labels: []string{"1", "2", "3", "4", "5", "6", "7", "8"},
		},
	}
}

func TestNewHorizontalBarChartOptionWithData(t *testing.T) {
	t.Parallel()

	opt := NewHorizontalBarChartOptionWithData([][]float64{
		{12, 24},
		{24, 48},
	})

	assert.Len(t, opt.SeriesList, 2)
	assert.Equal(t, ChartTypeHorizontalBar, opt.SeriesList[0].getType())
	assert.Equal(t, defaultPadding, opt.Padding)

	p := NewPainter(PainterOptions{})
	assert.NoError(t, p.HorizontalBarChart(opt))
}

func TestHorizontalBarChart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		themed      bool
		makeOptions func() HorizontalBarChartOption
		svg         string
		pngCRC      uint32
	}{
		{
			name:        "basic_themed",
			themed:      true,
			makeOptions: makeBasicHorizontalBarChartOption,
			svg:         "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:rgb(40,40,40)\"/><text x=\"10\" y=\"26\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">World Population</text><path d=\"M 224 19\nL 254 19\" style=\"stroke-width:3;stroke:rgb(255,100,100);fill:none\"/><circle cx=\"239\" cy=\"19\" r=\"5\" style=\"stroke-width:3;stroke:rgb(255,100,100);fill:rgb(255,100,100)\"/><text x=\"256\" y=\"25\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2011</text><path d=\"M 311 19\nL 341 19\" style=\"stroke-width:3;stroke:rgb(255,210,100);fill:none\"/><circle cx=\"326\" cy=\"19\" r=\"5\" style=\"stroke-width:3;stroke:rgb(255,210,100);fill:rgb(255,210,100)\"/><text x=\"343\" y=\"25\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2012</text><path d=\"M 87 46\nL 87 366\" style=\"stroke-width:1;stroke:rgb(185,184,206);fill:none\"/><path d=\"M 82 46\nL 87 46\" style=\"stroke-width:1;stroke:rgb(185,184,206);fill:none\"/><path d=\"M 82 99\nL 87 99\" style=\"stroke-width:1;stroke:rgb(185,184,206);fill:none\"/><path d=\"M 82 152\nL 87 152\" style=\"stroke-width:1;stroke:rgb(185,184,206);fill:none\"/><path d=\"M 82 206\nL 87 206\" style=\"stroke-width:1;stroke:rgb(185,184,206);fill:none\"/><path d=\"M 82 259\nL 87 259\" style=\"stroke-width:1;stroke:rgb(185,184,206);fill:none\"/><path d=\"M 82 312\nL 87 312\" style=\"stroke-width:1;stroke:rgb(185,184,206);fill:none\"/><path d=\"M 82 366\nL 87 366\" style=\"stroke-width:1;stroke:rgb(185,184,206);fill:none\"/><text x=\"36\" y=\"78\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">World</text><text x=\"37\" y=\"131\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">China</text><text x=\"43\" y=\"184\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">India</text><text x=\"47\" y=\"237\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">USA</text><text x=\"9\" y=\"290\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Indonesia</text><text x=\"38\" y=\"343\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Brazil</text><text x=\"87\" y=\"385\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">0</text><text x=\"212\" y=\"385\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">200k</text><text x=\"338\" y=\"385\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">400k</text><text x=\"463\" y=\"385\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">600k</text><text x=\"555\" y=\"385\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">800k</text><path d=\"M 213 46\nL 213 362\" style=\"stroke-width:1;stroke:rgb(72,71,83);fill:none\"/><path d=\"M 339 46\nL 339 362\" style=\"stroke-width:1;stroke:rgb(72,71,83);fill:none\"/><path d=\"M 464 46\nL 464 362\" style=\"stroke-width:1;stroke:rgb(72,71,83);fill:none\"/><path d=\"M 590 46\nL 590 362\" style=\"stroke-width:1;stroke:rgb(72,71,83);fill:none\"/><path d=\"M 88 322\nL 99 322\nL 99 336\nL 88 336\nL 88 322\" style=\"stroke:none;fill:rgb(255,100,100)\"/><path d=\"M 88 269\nL 102 269\nL 102 283\nL 88 283\nL 88 269\" style=\"stroke:none;fill:rgb(255,100,100)\"/><path d=\"M 88 216\nL 106 216\nL 106 230\nL 88 230\nL 88 216\" style=\"stroke:none;fill:rgb(255,100,100)\"/><path d=\"M 88 162\nL 153 162\nL 153 176\nL 88 176\nL 88 162\" style=\"stroke:none;fill:rgb(255,100,100)\"/><path d=\"M 88 109\nL 170 109\nL 170 123\nL 88 123\nL 88 109\" style=\"stroke:none;fill:rgb(255,100,100)\"/><path d=\"M 88 56\nL 483 56\nL 483 70\nL 88 70\nL 88 56\" style=\"stroke:none;fill:rgb(255,100,100)\"/><path d=\"M 88 341\nL 100 341\nL 100 355\nL 88 355\nL 88 341\" style=\"stroke:none;fill:rgb(255,210,100)\"/><path d=\"M 88 288\nL 102 288\nL 102 302\nL 88 302\nL 88 288\" style=\"stroke:none;fill:rgb(255,210,100)\"/><path d=\"M 88 235\nL 107 235\nL 107 249\nL 88 249\nL 88 235\" style=\"stroke:none;fill:rgb(255,210,100)\"/><path d=\"M 88 181\nL 164 181\nL 164 195\nL 88 195\nL 88 181\" style=\"stroke:none;fill:rgb(255,210,100)\"/><path d=\"M 88 128\nL 172 128\nL 172 142\nL 88 142\nL 88 128\" style=\"stroke:none;fill:rgb(255,210,100)\"/><path d=\"M 88 75\nL 515 75\nL 515 89\nL 88 89\nL 88 75\" style=\"stroke:none;fill:rgb(255,210,100)\"/></svg>",
			pngCRC:      0xf27d8d42,
		},
		{
			name: "custom_fonts",
			makeOptions: func() HorizontalBarChartOption {
				opt := makeBasicHorizontalBarChartOption()
				customFont := NewFontStyleWithSize(4.0).WithColor(ColorBlue)
				opt.Legend.FontStyle = customFont
				opt.XAxis.FontStyle = customFont
				opt.YAxis.FontStyle = customFont
				opt.Title.FontStyle = customFont
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"10\" y=\"16\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">World Population</text><path d=\"M 247 19\nL 277 19\" style=\"stroke-width:3;stroke:rgb(84,112,198);fill:none\"/><circle cx=\"262\" cy=\"19\" r=\"5\" style=\"stroke-width:3;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><text x=\"279\" y=\"25\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">2011</text><path d=\"M 311 19\nL 341 19\" style=\"stroke-width:3;stroke:rgb(145,204,117);fill:none\"/><circle cx=\"326\" cy=\"19\" r=\"5\" style=\"stroke-width:3;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><text x=\"343\" y=\"25\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">2012</text><path d=\"M 42 36\nL 42 366\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 37 36\nL 42 36\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 37 91\nL 42 91\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 37 146\nL 42 146\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 37 201\nL 42 201\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 37 256\nL 42 256\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 37 311\nL 42 311\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 37 366\nL 42 366\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"18\" y=\"64\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">World</text><text x=\"18\" y=\"118\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">China</text><text x=\"20\" y=\"173\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">India</text><text x=\"22\" y=\"228\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">USA</text><text x=\"9\" y=\"282\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">Indonesia</text><text x=\"19\" y=\"337\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">Brazil</text><text x=\"42\" y=\"375\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">0</text><text x=\"110\" y=\"375\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">100k</text><text x=\"178\" y=\"375\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">200k</text><text x=\"247\" y=\"375\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">300k</text><text x=\"315\" y=\"375\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">400k</text><text x=\"383\" y=\"375\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">500k</text><text x=\"452\" y=\"375\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">600k</text><text x=\"520\" y=\"375\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">700k</text><text x=\"578\" y=\"375\" style=\"stroke:none;fill:blue;font-size:5.1px;font-family:'Roboto Medium',sans-serif\">800k</text><path d=\"M 111 36\nL 111 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 179 36\nL 179 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 248 36\nL 248 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 316 36\nL 316 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 384 36\nL 384 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 453 36\nL 453 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 521 36\nL 521 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 590 36\nL 590 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 321\nL 55 321\nL 55 336\nL 43 336\nL 43 321\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 43 266\nL 59 266\nL 59 281\nL 43 281\nL 43 266\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 43 211\nL 62 211\nL 62 226\nL 43 226\nL 43 211\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 43 156\nL 114 156\nL 114 171\nL 43 171\nL 43 156\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 43 101\nL 133 101\nL 133 116\nL 43 116\nL 43 101\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 43 46\nL 473 46\nL 473 61\nL 43 61\nL 43 46\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 43 341\nL 56 341\nL 56 356\nL 43 356\nL 43 341\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 43 286\nL 59 286\nL 59 301\nL 43 301\nL 43 286\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 43 231\nL 64 231\nL 64 246\nL 43 246\nL 43 231\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 43 176\nL 126 176\nL 126 191\nL 43 191\nL 43 176\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 43 121\nL 134 121\nL 134 136\nL 43 136\nL 43 121\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 43 66\nL 509 66\nL 509 81\nL 43 81\nL 43 66\" style=\"stroke:none;fill:rgb(145,204,117)\"/></svg>",
			pngCRC: 0x216be35d,
		},
		{
			name: "value_labels",
			makeOptions: func() HorizontalBarChartOption {
				opt := makeBasicHorizontalBarChartOption()
				series := opt.SeriesList
				for i := range series {
					series[i].Label.Show = Ptr(true)
					series[i].Label.ValueFormatter = func(f float64) string {
						return humanize.FtoaWithDigits(f, 2)
					}
				}
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"10\" y=\"26\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">World Population</text><path d=\"M 224 19\nL 254 19\" style=\"stroke-width:3;stroke:rgb(84,112,198);fill:none\"/><circle cx=\"239\" cy=\"19\" r=\"5\" style=\"stroke-width:3;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><text x=\"256\" y=\"25\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2011</text><path d=\"M 311 19\nL 341 19\" style=\"stroke-width:3;stroke:rgb(145,204,117);fill:none\"/><circle cx=\"326\" cy=\"19\" r=\"5\" style=\"stroke-width:3;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><text x=\"343\" y=\"25\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2012</text><path d=\"M 87 46\nL 87 366\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 82 46\nL 87 46\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 82 99\nL 87 99\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 82 152\nL 87 152\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 82 206\nL 87 206\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 82 259\nL 87 259\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 82 312\nL 87 312\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 82 366\nL 87 366\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"36\" y=\"78\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">World</text><text x=\"37\" y=\"131\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">China</text><text x=\"43\" y=\"184\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">India</text><text x=\"47\" y=\"237\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">USA</text><text x=\"9\" y=\"290\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Indonesia</text><text x=\"38\" y=\"343\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Brazil</text><text x=\"87\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">0</text><text x=\"212\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">200k</text><text x=\"338\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">400k</text><text x=\"463\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">600k</text><text x=\"555\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">800k</text><path d=\"M 213 46\nL 213 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 339 46\nL 339 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 464 46\nL 464 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 590 46\nL 590 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 88 322\nL 99 322\nL 99 336\nL 88 336\nL 88 322\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 88 269\nL 102 269\nL 102 283\nL 88 283\nL 88 269\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 88 216\nL 106 216\nL 106 230\nL 88 230\nL 88 216\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 88 162\nL 153 162\nL 153 176\nL 88 176\nL 88 162\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 88 109\nL 170 109\nL 170 123\nL 88 123\nL 88 109\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 88 56\nL 483 56\nL 483 70\nL 88 70\nL 88 56\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 88 341\nL 100 341\nL 100 355\nL 88 355\nL 88 341\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 88 288\nL 102 288\nL 102 302\nL 88 302\nL 88 288\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 88 235\nL 107 235\nL 107 249\nL 88 249\nL 88 235\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 88 181\nL 164 181\nL 164 195\nL 88 195\nL 88 181\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 88 128\nL 172 128\nL 172 142\nL 88 142\nL 88 128\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 88 75\nL 515 75\nL 515 89\nL 88 89\nL 88 75\" style=\"stroke:none;fill:rgb(145,204,117)\"/><text x=\"104\" y=\"333\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">18203</text><text x=\"107\" y=\"280\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">23489</text><text x=\"111\" y=\"227\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">29034</text><text x=\"158\" y=\"173\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">104970</text><text x=\"175\" y=\"120\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">131744</text><text x=\"488\" y=\"67\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">630230</text><text x=\"105\" y=\"352\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">19325</text><text x=\"107\" y=\"299\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">23438</text><text x=\"112\" y=\"246\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">31000</text><text x=\"169\" y=\"192\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">121594</text><text x=\"177\" y=\"139\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">134141</text><text x=\"520\" y=\"86\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">681807</text></svg>",
			pngCRC: 0xaaa09e64,
		},
		{
			name: "value_formatter",
			makeOptions: func() HorizontalBarChartOption {
				opt := makeBasicHorizontalBarChartOption()
				opt.ValueFormatter = func(f float64) string {
					return "f"
				}
				series := opt.SeriesList
				for i := range series {
					series[i].Label.Show = Ptr(true)
					series[i].Label.ValueFormatter = opt.ValueFormatter
				}
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"10\" y=\"26\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">World Population</text><path d=\"M 224 19\nL 254 19\" style=\"stroke-width:3;stroke:rgb(84,112,198);fill:none\"/><circle cx=\"239\" cy=\"19\" r=\"5\" style=\"stroke-width:3;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><text x=\"256\" y=\"25\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2011</text><path d=\"M 311 19\nL 341 19\" style=\"stroke-width:3;stroke:rgb(145,204,117);fill:none\"/><circle cx=\"326\" cy=\"19\" r=\"5\" style=\"stroke-width:3;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><text x=\"343\" y=\"25\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2012</text><path d=\"M 87 46\nL 87 366\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 82 46\nL 87 46\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 82 99\nL 87 99\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 82 152\nL 87 152\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 82 206\nL 87 206\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 82 259\nL 87 259\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 82 312\nL 87 312\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 82 366\nL 87 366\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"36\" y=\"78\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">World</text><text x=\"37\" y=\"131\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">China</text><text x=\"43\" y=\"184\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">India</text><text x=\"47\" y=\"237\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">USA</text><text x=\"9\" y=\"290\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Indonesia</text><text x=\"38\" y=\"343\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Brazil</text><text x=\"87\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"149\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"212\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"275\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"338\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"400\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"463\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"526\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"584\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">f</text><path d=\"M 150 46\nL 150 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 213 46\nL 213 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 276 46\nL 276 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 339 46\nL 339 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 401 46\nL 401 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 464 46\nL 464 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 527 46\nL 527 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 590 46\nL 590 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 88 322\nL 99 322\nL 99 336\nL 88 336\nL 88 322\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 88 269\nL 102 269\nL 102 283\nL 88 283\nL 88 269\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 88 216\nL 106 216\nL 106 230\nL 88 230\nL 88 216\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 88 162\nL 153 162\nL 153 176\nL 88 176\nL 88 162\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 88 109\nL 170 109\nL 170 123\nL 88 123\nL 88 109\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 88 56\nL 483 56\nL 483 70\nL 88 70\nL 88 56\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 88 341\nL 100 341\nL 100 355\nL 88 355\nL 88 341\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 88 288\nL 102 288\nL 102 302\nL 88 302\nL 88 288\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 88 235\nL 107 235\nL 107 249\nL 88 249\nL 88 235\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 88 181\nL 164 181\nL 164 195\nL 88 195\nL 88 181\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 88 128\nL 172 128\nL 172 142\nL 88 142\nL 88 128\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 88 75\nL 515 75\nL 515 89\nL 88 89\nL 88 75\" style=\"stroke:none;fill:rgb(145,204,117)\"/><text x=\"104\" y=\"333\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"107\" y=\"280\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"111\" y=\"227\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"158\" y=\"173\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"175\" y=\"120\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"488\" y=\"67\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"105\" y=\"352\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"107\" y=\"299\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"112\" y=\"246\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"169\" y=\"192\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"177\" y=\"139\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">f</text><text x=\"520\" y=\"86\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">f</text></svg>",
			pngCRC: 0xd2977ce7,
		},
		{
			name: "bar_height_truncate",
			makeOptions: func() HorizontalBarChartOption {
				opt := makeBasicHorizontalBarChartOption()
				opt.Title.Show = Ptr(false)
				opt.XAxis.Show = Ptr(false)
				opt.YAxis.Show = Ptr(false)
				opt.Legend.Show = Ptr(false)
				opt.BarHeight = 1000
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><path d=\"M 10 336\nL 23 336\nL 23 355\nL 10 355\nL 10 336\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 10 273\nL 27 273\nL 27 292\nL 10 292\nL 10 273\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 10 210\nL 31 210\nL 31 229\nL 10 229\nL 10 210\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 10 146\nL 86 146\nL 86 165\nL 10 165\nL 10 146\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 10 83\nL 105 83\nL 105 102\nL 10 102\nL 10 83\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 10 20\nL 466 20\nL 466 39\nL 10 39\nL 10 20\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 10 360\nL 24 360\nL 24 379\nL 10 379\nL 10 360\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 10 297\nL 26 297\nL 26 316\nL 10 316\nL 10 297\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 10 234\nL 32 234\nL 32 253\nL 10 253\nL 10 234\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 10 170\nL 98 170\nL 98 189\nL 10 189\nL 10 170\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 10 107\nL 107 107\nL 107 126\nL 10 126\nL 10 107\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 10 44\nL 504 44\nL 504 63\nL 10 63\nL 10 44\" style=\"stroke:none;fill:rgb(145,204,117)\"/></svg>",
			pngCRC: 0x2aa51c50,
		},
		{
			name: "mark_line",
			makeOptions: func() HorizontalBarChartOption {
				opt := makeBasicHorizontalBarChartOption()
				opt.SeriesList[0].MarkLine = NewMarkLine(SeriesMarkTypeMax, SeriesMarkTypeAverage)
				opt.YAxis.Show = Ptr(false)
				opt.Legend.Show = Ptr(false)
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"10\" y=\"26\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">World Population</text><text x=\"9\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">0</text><text x=\"154\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">200k</text><text x=\"299\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">400k</text><text x=\"444\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">600k</text><text x=\"555\" y=\"385\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">800k</text><path d=\"M 155 41\nL 155 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 300 41\nL 300 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 445 41\nL 445 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 590 41\nL 590 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 10 321\nL 23 321\nL 23 335\nL 10 335\nL 10 321\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 10 267\nL 27 267\nL 27 281\nL 10 281\nL 10 267\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 10 213\nL 31 213\nL 31 227\nL 10 227\nL 10 213\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 10 159\nL 86 159\nL 86 173\nL 10 173\nL 10 159\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 10 105\nL 105 105\nL 105 119\nL 10 119\nL 10 105\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 10 51\nL 466 51\nL 466 65\nL 10 65\nL 10 51\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 10 340\nL 24 340\nL 24 354\nL 10 354\nL 10 340\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 10 286\nL 26 286\nL 26 300\nL 10 300\nL 10 286\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 10 232\nL 32 232\nL 32 246\nL 10 246\nL 10 232\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 10 178\nL 98 178\nL 98 192\nL 10 192\nL 10 178\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 10 124\nL 107 124\nL 107 138\nL 10 138\nL 10 124\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 10 70\nL 504 70\nL 504 84\nL 10 84\nL 10 70\" style=\"stroke:none;fill:rgb(145,204,117)\"/><circle cx=\"466\" cy=\"363\" r=\"3\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 466 43\nL 466 366\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 461 59\nL 466 43\nL 471 59\nL 466 54\nL 461 59\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><text x=\"442\" y=\"41\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">630.23k</text><circle cx=\"123\" cy=\"363\" r=\"3\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 123 43\nL 123 366\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 118 59\nL 123 43\nL 128 59\nL 123 54\nL 118 59\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><text x=\"99\" y=\"41\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">156.28k</text></svg>",
			pngCRC: 0xb04ecd33,
		},
		{
			name: "bar_height_thin",
			makeOptions: func() HorizontalBarChartOption {
				opt := makeMinimalHorizontalBarChartOption()
				opt.BarHeight = 2
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><path d=\"M 20 285\nL 48 285\nL 48 287\nL 20 287\nL 20 285\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 20 105\nL 216 105\nL 216 107\nL 20 107\nL 20 105\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 20 292\nL 216 292\nL 216 294\nL 20 294\nL 20 292\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 20 112\nL 552 112\nL 552 114\nL 20 114\nL 20 112\" style=\"stroke:none;fill:rgb(145,204,117)\"/></svg>",
			pngCRC: 0x9cef6723,
		},
		{
			name: "bar_margin_narrow",
			makeOptions: func() HorizontalBarChartOption {
				opt := makeMinimalHorizontalBarChartOption()
				opt.BarMargin = Ptr(0.0)
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><path d=\"M 20 210\nL 48 210\nL 48 290\nL 20 290\nL 20 210\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 20 30\nL 216 30\nL 216 110\nL 20 110\nL 20 30\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 20 290\nL 216 290\nL 216 370\nL 20 370\nL 20 290\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 20 110\nL 552 110\nL 552 190\nL 20 190\nL 20 110\" style=\"stroke:none;fill:rgb(145,204,117)\"/></svg>",
			pngCRC: 0x5ff4a8b8,
		},
		{
			name: "bar_margin_wide",
			makeOptions: func() HorizontalBarChartOption {
				opt := makeMinimalHorizontalBarChartOption()
				opt.BarMargin = Ptr(1000.0) // will be limited to fit graph
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><path d=\"M 20 210\nL 48 210\nL 48 245\nL 20 245\nL 20 210\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 20 30\nL 216 30\nL 216 65\nL 20 65\nL 20 30\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 20 335\nL 216 335\nL 216 370\nL 20 370\nL 20 335\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 20 155\nL 552 155\nL 552 190\nL 20 190\nL 20 155\" style=\"stroke:none;fill:rgb(145,204,117)\"/></svg>",
			pngCRC: 0x92e8a09c,
		},
		{
			name: "bar_height_and_narrow_margin",
			makeOptions: func() HorizontalBarChartOption {
				opt := makeMinimalHorizontalBarChartOption()
				opt.BarHeight = 10
				opt.BarMargin = Ptr(0.0)
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><path d=\"M 20 280\nL 48 280\nL 48 290\nL 20 290\nL 20 280\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 20 100\nL 216 100\nL 216 110\nL 20 110\nL 20 100\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 20 290\nL 216 290\nL 216 300\nL 20 300\nL 20 290\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 20 110\nL 552 110\nL 552 120\nL 20 120\nL 20 110\" style=\"stroke:none;fill:rgb(145,204,117)\"/></svg>",
			pngCRC: 0xc73ff61b,
		},
		{
			name: "bar_height_and_wide_margin",
			makeOptions: func() HorizontalBarChartOption {
				opt := makeMinimalHorizontalBarChartOption()
				opt.BarHeight = 10
				opt.BarMargin = Ptr(1000.0) // will be limited for readability
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><path d=\"M 20 240\nL 48 240\nL 48 250\nL 20 250\nL 20 240\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 20 60\nL 216 60\nL 216 70\nL 20 70\nL 20 60\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 20 330\nL 216 330\nL 216 340\nL 20 340\nL 20 330\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 20 150\nL 552 150\nL 552 160\nL 20 160\nL 20 150\" style=\"stroke:none;fill:rgb(145,204,117)\"/></svg>",
			pngCRC: 0x6e639c12,
		},
		{
			name:        "stack_series",
			makeOptions: makeFullHorizontalBarChartStackedOption,
			svg:         "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><path d=\"M 38 20\nL 38 356\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 33 20\nL 38 20\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 33 62\nL 38 62\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 33 104\nL 38 104\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 33 146\nL 38 146\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 33 188\nL 38 188\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 33 230\nL 38 230\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 33 272\nL 38 272\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 33 314\nL 38 314\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 33 356\nL 38 356\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"19\" y=\"46\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">8</text><text x=\"19\" y=\"88\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">7</text><text x=\"19\" y=\"130\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">6</text><text x=\"19\" y=\"172\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">5</text><text x=\"19\" y=\"213\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><text x=\"19\" y=\"255\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"19\" y=\"297\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"19\" y=\"339\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"38\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">0</text><text x=\"115\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">50</text><text x=\"192\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"269\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">150</text><text x=\"347\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">200</text><text x=\"424\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">250</text><text x=\"501\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">300</text><text x=\"553\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">350</text><path d=\"M 116 20\nL 116 352\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 193 20\nL 193 352\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 270 20\nL 270 352\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 348 20\nL 348 352\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 425 20\nL 425 352\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 502 20\nL 502 352\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 580 20\nL 580 352\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 39 319\nL 46 319\nL 46 351\nL 39 351\nL 39 319\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 39 277\nL 74 277\nL 74 309\nL 39 309\nL 39 277\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 39 235\nL 78 235\nL 78 267\nL 39 267\nL 39 235\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 39 193\nL 197 193\nL 197 225\nL 39 225\nL 39 193\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 39 151\nL 258 151\nL 258 183\nL 39 183\nL 39 151\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 39 109\nL 89 109\nL 89 141\nL 39 141\nL 39 109\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 39 67\nL 69 67\nL 69 99\nL 39 99\nL 39 67\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 39 25\nL 44 25\nL 44 57\nL 39 57\nL 39 25\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 46 319\nL 75 319\nL 75 351\nL 46 351\nL 46 319\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 74 277\nL 114 277\nL 114 309\nL 74 309\nL 74 277\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 78 235\nL 122 235\nL 122 267\nL 78 267\nL 78 235\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 197 193\nL 420 193\nL 420 225\nL 197 225\nL 197 193\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 258 151\nL 446 151\nL 446 183\nL 258 183\nL 258 151\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 89 109\nL 164 109\nL 164 141\nL 89 141\nL 89 109\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 69 67\nL 113 67\nL 113 99\nL 69 99\nL 69 67\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 44 25\nL 78 25\nL 78 57\nL 44 57\nL 44 25\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 75 319\nL 198 319\nL 198 351\nL 75 351\nL 75 319\" style=\"stroke:none;fill:rgb(250,200,88)\"/><path d=\"M 114 277\nL 176 277\nL 176 309\nL 114 309\nL 114 277\" style=\"stroke:none;fill:rgb(250,200,88)\"/><path d=\"M 122 235\nL 165 235\nL 165 267\nL 122 267\nL 122 235\" style=\"stroke:none;fill:rgb(250,200,88)\"/><path d=\"M 420 193\nL 464 193\nL 464 225\nL 420 225\nL 420 193\" style=\"stroke:none;fill:rgb(250,200,88)\"/><path d=\"M 446 151\nL 483 151\nL 483 183\nL 446 183\nL 446 151\" style=\"stroke:none;fill:rgb(250,200,88)\"/><path d=\"M 164 109\nL 201 109\nL 201 141\nL 164 141\nL 164 109\" style=\"stroke:none;fill:rgb(250,200,88)\"/><path d=\"M 113 67\nL 176 67\nL 176 99\nL 113 99\nL 113 67\" style=\"stroke:none;fill:rgb(250,200,88)\"/><path d=\"M 78 25\nL 202 25\nL 202 57\nL 78 57\nL 78 25\" style=\"stroke:none;fill:rgb(250,200,88)\"/><text x=\"51\" y=\"339\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">4</text><text x=\"79\" y=\"297\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">23</text><text x=\"83\" y=\"255\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">25</text><text x=\"202\" y=\"213\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">102</text><text x=\"263\" y=\"171\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">142</text><text x=\"94\" y=\"129\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">32</text><text x=\"74\" y=\"87\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">20</text><text x=\"49\" y=\"45\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"80\" y=\"339\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">19</text><text x=\"119\" y=\"297\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">26</text><text x=\"127\" y=\"255\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">28</text><text x=\"425\" y=\"213\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">144</text><text x=\"451\" y=\"171\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">122</text><text x=\"169\" y=\"129\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">48</text><text x=\"118\" y=\"87\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">28</text><text x=\"83\" y=\"45\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">22</text><text x=\"203\" y=\"339\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">80</text><text x=\"181\" y=\"297\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">40</text><text x=\"170\" y=\"255\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">28</text><text x=\"469\" y=\"213\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">28</text><text x=\"488\" y=\"171\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">24</text><text x=\"206\" y=\"129\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">24</text><text x=\"181\" y=\"87\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">40</text><text x=\"207\" y=\"45\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">80</text></svg>",
			pngCRC:      0x540d27f,
		},
		{
			name: "stack_series_simple",
			makeOptions: func() HorizontalBarChartOption {
				opt := NewHorizontalBarChartOptionWithData([][]float64{{4.0}, {1.0}})
				opt.StackSeries = Ptr(true)
				opt.XAxis.Unit = 1
				// disable extra
				opt.YAxis.Show = Ptr(false)
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"19\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">0</text><text x=\"205\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"392\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><text x=\"571\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">6</text><path d=\"M 206 20\nL 206 352\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 393 20\nL 393 352\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 580 20\nL 580 352\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 20 30\nL 393 30\nL 393 346\nL 20 346\nL 20 30\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 393 30\nL 486 30\nL 486 346\nL 393 346\nL 393 30\" style=\"stroke:none;fill:rgb(145,204,117)\"/></svg>",
			pngCRC: 0x2f4f3f65,
		},
		{
			name: "stack_series_with_mark",
			makeOptions: func() HorizontalBarChartOption {
				opt := makeFullHorizontalBarChartStackedOption()
				opt.SeriesList[0].MarkLine = NewMarkLine(SeriesMarkTypeMax, SeriesMarkTypeAverage)
				opt.SeriesList[len(opt.SeriesList)-1].MarkLine = NewMarkLine(SeriesMarkTypeMax)
				opt.SeriesList[len(opt.SeriesList)-1].MarkLine.Lines[0].Global = true
				opt.YAxis.Show = Ptr(false)
				opt.Legend.Show = Ptr(false)
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"19\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">0</text><text x=\"99\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">50</text><text x=\"179\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"259\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">150</text><text x=\"339\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">200</text><text x=\"419\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">250</text><text x=\"499\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">300</text><text x=\"553\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">350</text><path d=\"M 100 20\nL 100 352\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 180 20\nL 180 352\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 260 20\nL 260 352\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 340 20\nL 340 352\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 420 20\nL 420 352\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 500 20\nL 500 352\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 580 20\nL 580 352\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 20 319\nL 27 319\nL 27 351\nL 20 351\nL 20 319\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 20 277\nL 57 277\nL 57 309\nL 20 309\nL 20 277\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 20 235\nL 60 235\nL 60 267\nL 20 267\nL 20 235\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 20 193\nL 184 193\nL 184 225\nL 20 225\nL 20 193\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 20 151\nL 247 151\nL 247 183\nL 20 183\nL 20 151\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 20 109\nL 72 109\nL 72 141\nL 20 141\nL 20 109\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 20 67\nL 52 67\nL 52 99\nL 20 99\nL 20 67\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 20 25\nL 25 25\nL 25 57\nL 20 57\nL 20 25\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 27 319\nL 57 319\nL 57 351\nL 27 351\nL 27 319\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 57 277\nL 99 277\nL 99 309\nL 57 309\nL 57 277\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 60 235\nL 105 235\nL 105 267\nL 60 267\nL 60 235\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 184 193\nL 415 193\nL 415 225\nL 184 225\nL 184 193\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 247 151\nL 442 151\nL 442 183\nL 247 183\nL 247 151\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 72 109\nL 149 109\nL 149 141\nL 72 141\nL 72 109\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 52 67\nL 98 67\nL 98 99\nL 52 99\nL 52 67\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 25 25\nL 60 25\nL 60 57\nL 25 57\nL 25 25\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 57 319\nL 185 319\nL 185 351\nL 57 351\nL 57 319\" style=\"stroke:none;fill:rgb(250,200,88)\"/><path d=\"M 99 277\nL 163 277\nL 163 309\nL 99 309\nL 99 277\" style=\"stroke:none;fill:rgb(250,200,88)\"/><path d=\"M 105 235\nL 150 235\nL 150 267\nL 105 267\nL 105 235\" style=\"stroke:none;fill:rgb(250,200,88)\"/><path d=\"M 415 193\nL 461 193\nL 461 225\nL 415 225\nL 415 193\" style=\"stroke:none;fill:rgb(250,200,88)\"/><path d=\"M 442 151\nL 481 151\nL 481 183\nL 442 183\nL 442 151\" style=\"stroke:none;fill:rgb(250,200,88)\"/><path d=\"M 149 109\nL 187 109\nL 187 141\nL 149 141\nL 149 109\" style=\"stroke:none;fill:rgb(250,200,88)\"/><path d=\"M 98 67\nL 163 67\nL 163 99\nL 98 99\nL 98 67\" style=\"stroke:none;fill:rgb(250,200,88)\"/><path d=\"M 60 25\nL 189 25\nL 189 57\nL 60 57\nL 60 25\" style=\"stroke:none;fill:rgb(250,200,88)\"/><circle cx=\"247\" cy=\"353\" r=\"3\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 247 22\nL 247 356\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 242 38\nL 247 22\nL 252 38\nL 247 33\nL 242 38\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><text x=\"235\" y=\"20\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">142</text><circle cx=\"90\" cy=\"353\" r=\"3\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 90 22\nL 90 356\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 85 38\nL 90 22\nL 95 38\nL 90 33\nL 85 38\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><text x=\"82\" y=\"20\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">44</text><circle cx=\"482\" cy=\"353\" r=\"3\" style=\"stroke-width:1;stroke:rgb(211,211,211);fill:rgb(211,211,211)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 482 22\nL 482 356\" style=\"stroke-width:1;stroke:rgb(211,211,211);fill:rgb(211,211,211)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 477 38\nL 482 22\nL 487 38\nL 482 33\nL 477 38\" style=\"stroke-width:1;stroke:rgb(211,211,211);fill:rgb(211,211,211)\"/><text x=\"470\" y=\"20\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">288</text><text x=\"32\" y=\"339\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">4</text><text x=\"62\" y=\"297\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">23</text><text x=\"65\" y=\"255\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">25</text><text x=\"189\" y=\"213\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">102</text><text x=\"252\" y=\"171\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">142</text><text x=\"77\" y=\"129\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">32</text><text x=\"57\" y=\"87\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">20</text><text x=\"30\" y=\"45\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"62\" y=\"339\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">19</text><text x=\"104\" y=\"297\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">26</text><text x=\"110\" y=\"255\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">28</text><text x=\"420\" y=\"213\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">144</text><text x=\"447\" y=\"171\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">122</text><text x=\"154\" y=\"129\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">48</text><text x=\"103\" y=\"87\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">28</text><text x=\"65\" y=\"45\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">22</text><text x=\"190\" y=\"339\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">80</text><text x=\"168\" y=\"297\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">40</text><text x=\"155\" y=\"255\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">28</text><text x=\"466\" y=\"213\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">28</text><text x=\"486\" y=\"171\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">24</text><text x=\"192\" y=\"129\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">24</text><text x=\"168\" y=\"87\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">40</text><text x=\"194\" y=\"45\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">80</text></svg>",
			pngCRC: 0xb65b3da0,
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

				validateHorizontalBarChartRender(t, p, rp, tt.makeOptions(), tt.svg, tt.pngCRC)
			})
			t.Run(strconv.Itoa(i)+"-"+tt.name+"-options", func(t *testing.T) {
				p := NewPainter(painterOptions)
				rp := NewPainter(rasterOptions)
				opt := tt.makeOptions()
				opt.Theme = GetTheme(ThemeVividDark)

				validateHorizontalBarChartRender(t, p, rp, opt, tt.svg, tt.pngCRC)
			})
		} else {
			t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
				p := NewPainter(painterOptions)
				rp := NewPainter(rasterOptions)

				validateHorizontalBarChartRender(t, p, rp, tt.makeOptions(), tt.svg, tt.pngCRC)
			})
		}
	}
}

func validateHorizontalBarChartRender(t *testing.T, svgP, pngP *Painter, opt HorizontalBarChartOption, expectedSVG string, expectedCRC uint32) {
	t.Helper()

	err := svgP.HorizontalBarChart(opt)
	require.NoError(t, err)
	data, err := svgP.Bytes()
	require.NoError(t, err)
	assertEqualSVG(t, expectedSVG, data)

	err = pngP.HorizontalBarChart(opt)
	require.NoError(t, err)
	rasterData, err := pngP.Bytes()
	require.NoError(t, err)
	assertEqualPNGCRC(t, expectedCRC, rasterData)
}

func TestHorizontalBarChartError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		makeOptions      func() HorizontalBarChartOption
		errorMsgContains string
	}{
		{
			name: "empty_series",
			makeOptions: func() HorizontalBarChartOption {
				return NewHorizontalBarChartOptionWithData([][]float64{})
			},
			errorMsgContains: "empty series list",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			p := NewPainter(PainterOptions{
				OutputFormat: ChartOutputSVG,
				Width:        600,
				Height:       400,
			})

			err := p.HorizontalBarChart(tt.makeOptions())
			require.Error(t, err)
			require.ErrorContains(t, err, tt.errorMsgContains)
		})
	}
}
