package charts

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeFullScatterChartOption() ScatterChartOption {
	values := [][]float64{
		{120, 132, 101, 134, 90, 230, 210},
		{220, 182, 191, 234, 290, 330, 310},
		{150, 232, 201, 154, 190, 330, 410},
		{320, 332, 301, 334, 390, 330, 320},
		{820, 932, 901, 934, 1290, 1330, 1320},
	}
	return ScatterChartOption{
		Title: TitleOption{
			Text: "Scatter",
		},
		Padding: NewBoxEqual(10),
		XAxis: XAxisOption{
			Labels: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
		},
		YAxis: make([]YAxisOption, 1),
		Legend: LegendOption{
			SeriesNames: []string{"Email", "Union Ads", "Video Ads", "Direct", "Search Engine"},
		},
		SeriesList: NewSeriesListScatter(values),
	}
}

func makeBasicScatterChartOption() ScatterChartOption {
	values := [][]float64{
		{120, 132, 101, 134, 90, 230, 210},
		{820, 932, 901, 934, 1290, 1330, 1320},
	}
	return ScatterChartOption{
		Padding: NewBoxEqual(10),
		XAxis: XAxisOption{
			Labels: []string{"A", "B", "C", "D", "E", "F", "G"},
		},
		YAxis: make([]YAxisOption, 1),
		Legend: LegendOption{
			SeriesNames: []string{"1", "2"},
		},
		SeriesList: NewSeriesListScatter(values),
	}
}

func makeMinimalScatterChartOption() ScatterChartOption {
	values := [][]float64{
		{120, 132, 101, 134, 90, 230, 210},
		{820, 932, 901, 934, 1290, 1330, 1320},
	}
	return ScatterChartOption{
		Padding: NewBoxEqual(10),
		XAxis: XAxisOption{
			Labels: []string{"1", "2", "3", "4", "5", "6", "7"},
			Show:   Ptr(false),
		},
		YAxis: []YAxisOption{
			{
				Show: Ptr(false),
			},
		},
		SeriesList: NewSeriesListScatter(values),
	}
}

func makeMinimalMultiValueScatterChartOption() ScatterChartOption {
	values := [][][]float64{
		{{120, GetNullValue()}, {132}, {101, 20}, {134}, {90, 28}, {230}, {210}},
		{{820, GetNullValue()}, {932}, {901, 600}, {934}, {1290}, {1330}, {1320}},
	}
	return ScatterChartOption{
		Padding: NewBoxEqual(10),
		XAxis: XAxisOption{
			Labels: []string{"1", "2", "3", "4", "5", "6", "7"},
			Show:   Ptr(false),
		},
		YAxis: []YAxisOption{
			{
				Show: Ptr(false),
			},
		},
		SeriesList: NewSeriesListScatterMultiValue(values),
	}
}

func generateRandomScatterData(seriesCount int, dataPointCount int, maxVariationPercentage float64) [][][]float64 {
	data := make([][][]float64, seriesCount)
	for i := 0; i < seriesCount; i++ {
		data[i] = make([][]float64, dataPointCount)
	}
	r := rand.New(rand.NewSource(1))

	for i := 0; i < seriesCount; i++ {
		for j := 0; j < dataPointCount; j++ {
			if j == 0 {
				// Set the initial value for the line
				data[i][j] = []float64{r.Float64() * 100}
			} else {
				// Calculate the allowed variation range
				variationRange := data[i][j-1][0] * maxVariationPercentage / 100
				min := data[i][j-1][0] - variationRange
				max := data[i][j-1][0] + variationRange

				// Generate a random value within the allowed range
				values := []float64{min + r.Float64()*(max-min)}
				if j%2 == 0 {
					values = append(values, min+r.Float64()*(max-min))
				}
				if j%10 == 0 {
					values = append(values, min+r.Float64()*(max-min))
				}
				data[i][j] = values
			}
		}
	}

	return data
}

func makeDenseScatterChartOption() ScatterChartOption {
	const dataPointCount = 100
	values := generateRandomScatterData(3, dataPointCount, 10)

	xAxisLabels := make([]string, dataPointCount)
	for i := 0; i < dataPointCount; i++ {
		xAxisLabels[i] = strconv.Itoa(i)
	}

	return ScatterChartOption{
		SeriesList: NewSeriesListScatterMultiValue(values, ScatterSeriesOption{
			TrendLine: NewTrendLine(SeriesMarkTypeAverage),
			Label: SeriesLabel{
				ValueFormatter: func(f float64) string {
					return FormatValueHumanizeShort(f, 0, false)
				},
			},
		}),
		Padding: NewBoxEqual(20),
		Theme:   GetTheme(ThemeLight),
		YAxis: []YAxisOption{
			{
				Min:            Ptr(0.0), // force min to be zero
				Max:            Ptr(200.0),
				Unit:           10,
				LabelSkipCount: 1,
			},
		},
		XAxis: XAxisOption{
			Labels:        xAxisLabels,
			BoundaryGap:   Ptr(false),
			LabelCount:    10,
			LabelRotation: DegreesToRadians(45),
		},
	}
}

func TestNewScatterChartOptionWithData(t *testing.T) {
	t.Parallel()

	opt := NewScatterChartOptionWithData([][]float64{
		{12, 24},
		{24, 48},
	})

	assert.Len(t, opt.SeriesList, 2)
	assert.Equal(t, ChartTypeScatter, opt.SeriesList[0].getType())
	assert.Len(t, opt.YAxis, 1)
	assert.Equal(t, defaultPadding, opt.Padding)

	p := NewPainter(PainterOptions{})
	assert.NoError(t, p.ScatterChart(opt))
}

func TestScatterChart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ignore      string // specified if the test is ignored
		themed      bool
		makeOptions func() ScatterChartOption
		svg         string
		pngCRC      uint32
	}{
		{
			name:        "basic_themed",
			themed:      true,
			makeOptions: makeFullScatterChartOption,
			svg:         "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:rgb(40,40,40)\"/><text x=\"10\" y=\"26\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Scatter</text><path d=\"M 21 35\nL 51 35\" style=\"stroke-width:3;stroke:rgb(255,100,100);fill:none\"/><circle cx=\"36\" cy=\"35\" r=\"5\" style=\"stroke-width:3;stroke:rgb(255,100,100);fill:rgb(255,100,100)\"/><text x=\"53\" y=\"41\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Email</text><path d=\"M 112 35\nL 142 35\" style=\"stroke-width:3;stroke:rgb(255,210,100);fill:none\"/><circle cx=\"127\" cy=\"35\" r=\"5\" style=\"stroke-width:3;stroke:rgb(255,210,100);fill:rgb(255,210,100)\"/><text x=\"144\" y=\"41\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Union Ads</text><path d=\"M 235 35\nL 265 35\" style=\"stroke-width:3;stroke:rgb(100,180,210);fill:none\"/><circle cx=\"250\" cy=\"35\" r=\"5\" style=\"stroke-width:3;stroke:rgb(100,180,210);fill:rgb(100,180,210)\"/><text x=\"267\" y=\"41\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Video Ads</text><path d=\"M 357 35\nL 387 35\" style=\"stroke-width:3;stroke:rgb(64,160,110);fill:none\"/><circle cx=\"372\" cy=\"35\" r=\"5\" style=\"stroke-width:3;stroke:rgb(64,160,110);fill:rgb(64,160,110)\"/><text x=\"389\" y=\"41\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Direct</text><path d=\"M 450 35\nL 480 35\" style=\"stroke-width:3;stroke:rgb(154,96,180);fill:none\"/><circle cx=\"465\" cy=\"35\" r=\"5\" style=\"stroke-width:3;stroke:rgb(154,96,180);fill:rgb(154,96,180)\"/><text x=\"482\" y=\"41\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Search Engine</text><text x=\"9\" y=\"63\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1.6k</text><text x=\"9\" y=\"101\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1.4k</text><text x=\"9\" y=\"139\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1.2k</text><text x=\"22\" y=\"177\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1k</text><text x=\"12\" y=\"215\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">800</text><text x=\"12\" y=\"253\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">600</text><text x=\"12\" y=\"291\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">400</text><text x=\"12\" y=\"329\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">200</text><text x=\"30\" y=\"368\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">0</text><path d=\"M 45 57\nL 590 57\" style=\"stroke-width:1;stroke:rgb(72,71,83);fill:none\"/><path d=\"M 45 95\nL 590 95\" style=\"stroke-width:1;stroke:rgb(72,71,83);fill:none\"/><path d=\"M 45 133\nL 590 133\" style=\"stroke-width:1;stroke:rgb(72,71,83);fill:none\"/><path d=\"M 45 172\nL 590 172\" style=\"stroke-width:1;stroke:rgb(72,71,83);fill:none\"/><path d=\"M 45 210\nL 590 210\" style=\"stroke-width:1;stroke:rgb(72,71,83);fill:none\"/><path d=\"M 45 248\nL 590 248\" style=\"stroke-width:1;stroke:rgb(72,71,83);fill:none\"/><path d=\"M 45 287\nL 590 287\" style=\"stroke-width:1;stroke:rgb(72,71,83);fill:none\"/><path d=\"M 45 325\nL 590 325\" style=\"stroke-width:1;stroke:rgb(72,71,83);fill:none\"/><path d=\"M 49 364\nL 590 364\" style=\"stroke-width:1;stroke:rgb(185,184,206);fill:none\"/><path d=\"M 49 369\nL 49 364\" style=\"stroke-width:1;stroke:rgb(185,184,206);fill:none\"/><path d=\"M 139 369\nL 139 364\" style=\"stroke-width:1;stroke:rgb(185,184,206);fill:none\"/><path d=\"M 229 369\nL 229 364\" style=\"stroke-width:1;stroke:rgb(185,184,206);fill:none\"/><path d=\"M 319 369\nL 319 364\" style=\"stroke-width:1;stroke:rgb(185,184,206);fill:none\"/><path d=\"M 409 369\nL 409 364\" style=\"stroke-width:1;stroke:rgb(185,184,206);fill:none\"/><path d=\"M 499 369\nL 499 364\" style=\"stroke-width:1;stroke:rgb(185,184,206);fill:none\"/><path d=\"M 590 369\nL 590 364\" style=\"stroke-width:1;stroke:rgb(185,184,206);fill:none\"/><text x=\"48\" y=\"390\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Mon</text><text x=\"138\" y=\"390\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Tue</text><text x=\"228\" y=\"390\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Wed</text><text x=\"318\" y=\"390\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Thu</text><text x=\"408\" y=\"390\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Fri</text><text x=\"498\" y=\"390\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Sat</text><text x=\"563\" y=\"390\" style=\"stroke:none;fill:rgb(238,238,238);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Sun</text><circle cx=\"49\" cy=\"341\" r=\"2\" style=\"stroke-width:1;stroke:rgb(255,100,100);fill:rgb(255,100,100)\"/><circle cx=\"139\" cy=\"339\" r=\"2\" style=\"stroke-width:1;stroke:rgb(255,100,100);fill:rgb(255,100,100)\"/><circle cx=\"229\" cy=\"345\" r=\"2\" style=\"stroke-width:1;stroke:rgb(255,100,100);fill:rgb(255,100,100)\"/><circle cx=\"319\" cy=\"339\" r=\"2\" style=\"stroke-width:1;stroke:rgb(255,100,100);fill:rgb(255,100,100)\"/><circle cx=\"409\" cy=\"347\" r=\"2\" style=\"stroke-width:1;stroke:rgb(255,100,100);fill:rgb(255,100,100)\"/><circle cx=\"499\" cy=\"320\" r=\"2\" style=\"stroke-width:1;stroke:rgb(255,100,100);fill:rgb(255,100,100)\"/><circle cx=\"590\" cy=\"324\" r=\"2\" style=\"stroke-width:1;stroke:rgb(255,100,100);fill:rgb(255,100,100)\"/><circle cx=\"49\" cy=\"322\" r=\"2\" style=\"stroke-width:1;stroke:rgb(255,210,100);fill:rgb(255,210,100)\"/><circle cx=\"139\" cy=\"330\" r=\"2\" style=\"stroke-width:1;stroke:rgb(255,210,100);fill:rgb(255,210,100)\"/><circle cx=\"229\" cy=\"328\" r=\"2\" style=\"stroke-width:1;stroke:rgb(255,210,100);fill:rgb(255,210,100)\"/><circle cx=\"319\" cy=\"320\" r=\"2\" style=\"stroke-width:1;stroke:rgb(255,210,100);fill:rgb(255,210,100)\"/><circle cx=\"409\" cy=\"309\" r=\"2\" style=\"stroke-width:1;stroke:rgb(255,210,100);fill:rgb(255,210,100)\"/><circle cx=\"499\" cy=\"301\" r=\"2\" style=\"stroke-width:1;stroke:rgb(255,210,100);fill:rgb(255,210,100)\"/><circle cx=\"590\" cy=\"305\" r=\"2\" style=\"stroke-width:1;stroke:rgb(255,210,100);fill:rgb(255,210,100)\"/><circle cx=\"49\" cy=\"336\" r=\"2\" style=\"stroke-width:1;stroke:rgb(100,180,210);fill:rgb(100,180,210)\"/><circle cx=\"139\" cy=\"320\" r=\"2\" style=\"stroke-width:1;stroke:rgb(100,180,210);fill:rgb(100,180,210)\"/><circle cx=\"229\" cy=\"326\" r=\"2\" style=\"stroke-width:1;stroke:rgb(100,180,210);fill:rgb(100,180,210)\"/><circle cx=\"319\" cy=\"335\" r=\"2\" style=\"stroke-width:1;stroke:rgb(100,180,210);fill:rgb(100,180,210)\"/><circle cx=\"409\" cy=\"328\" r=\"2\" style=\"stroke-width:1;stroke:rgb(100,180,210);fill:rgb(100,180,210)\"/><circle cx=\"499\" cy=\"301\" r=\"2\" style=\"stroke-width:1;stroke:rgb(100,180,210);fill:rgb(100,180,210)\"/><circle cx=\"590\" cy=\"286\" r=\"2\" style=\"stroke-width:1;stroke:rgb(100,180,210);fill:rgb(100,180,210)\"/><circle cx=\"49\" cy=\"303\" r=\"2\" style=\"stroke-width:1;stroke:rgb(64,160,110);fill:rgb(64,160,110)\"/><circle cx=\"139\" cy=\"301\" r=\"2\" style=\"stroke-width:1;stroke:rgb(64,160,110);fill:rgb(64,160,110)\"/><circle cx=\"229\" cy=\"307\" r=\"2\" style=\"stroke-width:1;stroke:rgb(64,160,110);fill:rgb(64,160,110)\"/><circle cx=\"319\" cy=\"300\" r=\"2\" style=\"stroke-width:1;stroke:rgb(64,160,110);fill:rgb(64,160,110)\"/><circle cx=\"409\" cy=\"290\" r=\"2\" style=\"stroke-width:1;stroke:rgb(64,160,110);fill:rgb(64,160,110)\"/><circle cx=\"499\" cy=\"301\" r=\"2\" style=\"stroke-width:1;stroke:rgb(64,160,110);fill:rgb(64,160,110)\"/><circle cx=\"590\" cy=\"303\" r=\"2\" style=\"stroke-width:1;stroke:rgb(64,160,110);fill:rgb(64,160,110)\"/><circle cx=\"49\" cy=\"207\" r=\"2\" style=\"stroke-width:1;stroke:rgb(154,96,180);fill:rgb(154,96,180)\"/><circle cx=\"139\" cy=\"186\" r=\"2\" style=\"stroke-width:1;stroke:rgb(154,96,180);fill:rgb(154,96,180)\"/><circle cx=\"229\" cy=\"192\" r=\"2\" style=\"stroke-width:1;stroke:rgb(154,96,180);fill:rgb(154,96,180)\"/><circle cx=\"319\" cy=\"185\" r=\"2\" style=\"stroke-width:1;stroke:rgb(154,96,180);fill:rgb(154,96,180)\"/><circle cx=\"409\" cy=\"117\" r=\"2\" style=\"stroke-width:1;stroke:rgb(154,96,180);fill:rgb(154,96,180)\"/><circle cx=\"499\" cy=\"109\" r=\"2\" style=\"stroke-width:1;stroke:rgb(154,96,180);fill:rgb(154,96,180)\"/><circle cx=\"590\" cy=\"111\" r=\"2\" style=\"stroke-width:1;stroke:rgb(154,96,180);fill:rgb(154,96,180)\"/></svg>",
			pngCRC:      0x9819ba2e,
		},
		{
			name: "boundary_gap_enable",
			makeOptions: func() ScatterChartOption {
				opt := makeMinimalScatterChartOption()
				opt.XAxis.Show = Ptr(true)
				opt.XAxis.BoundaryGap = Ptr(true)
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><path d=\"M 10 364\nL 590 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 10 369\nL 10 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 92 369\nL 92 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 175 369\nL 175 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 258 369\nL 258 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 341 369\nL 341 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 424 369\nL 424 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 507 369\nL 507 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 590 369\nL 590 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"47\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"129\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"212\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"295\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><text x=\"378\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">5</text><text x=\"461\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">6</text><text x=\"544\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">7</text><circle cx=\"51\" cy=\"338\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"133\" cy=\"335\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"216\" cy=\"342\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"299\" cy=\"335\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"382\" cy=\"345\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"465\" cy=\"314\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"548\" cy=\"318\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"51\" cy=\"183\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"133\" cy=\"158\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"216\" cy=\"165\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"299\" cy=\"158\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"382\" cy=\"79\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"465\" cy=\"70\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"548\" cy=\"72\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/></svg>",
			pngCRC: 0x7689d03a,
		},
		{
			name: "double_yaxis",
			makeOptions: func() ScatterChartOption {
				opt := makeBasicScatterChartOption()
				opt.Theme = GetTheme(ThemeLight)
				opt.SeriesList[1].YAxisIndex = 1
				opt.YAxis = append(opt.YAxis, opt.YAxis[0])
				opt.YAxis[0].Theme = opt.Theme.WithYAxisSeriesColor(0)
				opt.YAxis[1].Theme = opt.Theme.WithYAxisSeriesColor(1)
				opt.XAxis.Show = Ptr(false)
				opt.Title.Show = Ptr(false)
				opt.Legend.Show = Ptr(false)
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"561\" y=\"16\" style=\"stroke:none;fill:rgb(145,204,117);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1.4k</text><text x=\"561\" y=\"79\" style=\"stroke:none;fill:rgb(145,204,117);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1.3k</text><text x=\"561\" y=\"142\" style=\"stroke:none;fill:rgb(145,204,117);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1.2k</text><text x=\"561\" y=\"205\" style=\"stroke:none;fill:rgb(145,204,117);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1.1k</text><text x=\"561\" y=\"268\" style=\"stroke:none;fill:rgb(145,204,117);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1k</text><text x=\"561\" y=\"331\" style=\"stroke:none;fill:rgb(145,204,117);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">900</text><text x=\"561\" y=\"394\" style=\"stroke:none;fill:rgb(145,204,117);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">800</text><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(84,112,198);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">250</text><text x=\"9\" y=\"63\" style=\"stroke:none;fill:rgb(84,112,198);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">230</text><text x=\"9\" y=\"110\" style=\"stroke:none;fill:rgb(84,112,198);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">210</text><text x=\"9\" y=\"157\" style=\"stroke:none;fill:rgb(84,112,198);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">190</text><text x=\"9\" y=\"205\" style=\"stroke:none;fill:rgb(84,112,198);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">170</text><text x=\"9\" y=\"252\" style=\"stroke:none;fill:rgb(84,112,198);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">150</text><text x=\"9\" y=\"299\" style=\"stroke:none;fill:rgb(84,112,198);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"346\" style=\"stroke:none;fill:rgb(84,112,198);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"18\" y=\"394\" style=\"stroke:none;fill:rgb(84,112,198);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 42 10\nL 551 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 57\nL 551 57\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 105\nL 551 105\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 152\nL 551 152\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 200\nL 551 200\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 247\nL 551 247\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 295\nL 551 295\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 342\nL 551 342\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><circle cx=\"46\" cy=\"319\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"130\" cy=\"291\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"214\" cy=\"364\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"298\" cy=\"286\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"382\" cy=\"390\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"466\" cy=\"58\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"551\" cy=\"105\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"46\" cy=\"378\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"130\" cy=\"307\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"214\" cy=\"327\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"298\" cy=\"306\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"382\" cy=\"80\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"466\" cy=\"55\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"551\" cy=\"61\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/></svg>",
			pngCRC: 0x5f3ba140,
		},
		{
			name: "data_gap",
			makeOptions: func() ScatterChartOption {
				opt := makeMinimalScatterChartOption()
				opt.SeriesList[0].Values[4] = []float64{GetNullValue()}
				opt.SeriesList[1].Values[2] = []float64{GetNullValue()}
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><circle cx=\"10\" cy=\"362\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"106\" cy=\"359\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"203\" cy=\"367\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"300\" cy=\"359\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"396\" cy=\"2147483657\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"493\" cy=\"336\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"590\" cy=\"341\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"10\" cy=\"196\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"106\" cy=\"169\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"203\" cy=\"2147483657\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"300\" cy=\"169\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"396\" cy=\"84\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"493\" cy=\"75\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"590\" cy=\"77\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/></svg>",
			pngCRC: 0x3940204c,
		},
		{
			name: "mark_line",
			makeOptions: func() ScatterChartOption {
				opt := makeMinimalMultiValueScatterChartOption()
				opt.Padding = NewBoxEqual(40)
				opt.SymbolSize = 4.5
				for i := range opt.SeriesList {
					markLine := NewMarkLine("min", "max", "average")
					markLine.ValueFormatter = func(f float64) string {
						return FormatValueHumanizeShort(f, 0, false)
					}
					opt.SeriesList[i].MarkLine = markLine
				}
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><circle cx=\"40\" cy=\"336\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"126\" cy=\"334\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"213\" cy=\"340\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"213\" cy=\"356\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"300\" cy=\"334\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"386\" cy=\"342\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"386\" cy=\"355\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"473\" cy=\"314\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"560\" cy=\"318\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"40\" cy=\"196\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"126\" cy=\"174\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"213\" cy=\"180\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"213\" cy=\"240\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"300\" cy=\"174\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"386\" cy=\"102\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"473\" cy=\"94\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"560\" cy=\"96\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"43\" cy=\"356\" r=\"3\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 49 356\nL 542 356\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 542 351\nL 558 356\nL 542 361\nL 547 356\nL 542 351\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><text x=\"560\" y=\"360\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">20</text><circle cx=\"43\" cy=\"314\" r=\"3\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 49 314\nL 542 314\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 542 309\nL 558 314\nL 542 319\nL 547 314\nL 542 309\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><text x=\"560\" y=\"318\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">230</text><circle cx=\"43\" cy=\"337\" r=\"3\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 49 337\nL 542 337\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 542 332\nL 558 337\nL 542 342\nL 547 337\nL 542 332\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><text x=\"560\" y=\"341\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">118</text><circle cx=\"43\" cy=\"240\" r=\"3\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 49 240\nL 542 240\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 542 235\nL 558 240\nL 542 245\nL 547 240\nL 542 235\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><text x=\"560\" y=\"244\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">600</text><circle cx=\"43\" cy=\"94\" r=\"3\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 49 94\nL 542 94\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 542 89\nL 558 94\nL 542 99\nL 547 94\nL 542 89\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><text x=\"560\" y=\"98\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">1k</text><circle cx=\"43\" cy=\"157\" r=\"3\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 49 157\nL 542 157\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 542 152\nL 558 157\nL 542 162\nL 547 157\nL 542 152\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><text x=\"560\" y=\"161\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">1k</text></svg>",
			pngCRC: 0x817286fe,
		},
		{
			name: "series_label",
			makeOptions: func() ScatterChartOption {
				opt := makeMinimalMultiValueScatterChartOption()
				opt.YAxis[0].Show = Ptr(false)
				for i := range opt.SeriesList {
					opt.SeriesList[i].Label.Show = Ptr(true)
					opt.SeriesList[i].Label.FontStyle = FontStyle{
						FontSize:  12.0,
						Font:      GetDefaultFont(),
						FontColor: ColorBlue,
					}
					opt.SeriesList[i].Label.ValueFormatter = func(f float64) string {
						return FormatValueHumanizeShort(f, 2, false)
					}
				}
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><circle cx=\"10\" cy=\"362\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"106\" cy=\"359\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"203\" cy=\"367\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"203\" cy=\"386\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"300\" cy=\"359\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"396\" cy=\"369\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"396\" cy=\"384\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"493\" cy=\"336\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"590\" cy=\"341\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"10\" cy=\"196\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"106\" cy=\"169\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"203\" cy=\"177\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"203\" cy=\"248\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"300\" cy=\"169\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"396\" cy=\"84\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"493\" cy=\"75\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"590\" cy=\"77\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><text x=\"15\" y=\"368\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"111\" y=\"365\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">132</text><text x=\"208\" y=\"373\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">101</text><text x=\"208\" y=\"392\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">20</text><text x=\"305\" y=\"365\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">134</text><text x=\"401\" y=\"375\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><text x=\"401\" y=\"390\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">28</text><text x=\"498\" y=\"342\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">230</text><text x=\"573\" y=\"347\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">210</text><text x=\"15\" y=\"202\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">820</text><text x=\"111\" y=\"175\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">932</text><text x=\"208\" y=\"183\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">901</text><text x=\"208\" y=\"254\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">600</text><text x=\"305\" y=\"175\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">934</text><text x=\"401\" y=\"90\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1.29k</text><text x=\"498\" y=\"81\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1.33k</text><text x=\"561\" y=\"83\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1.32k</text></svg>",
			pngCRC: 0xa96e97c,
		},
		{
			name: "symbol_dot",
			makeOptions: func() ScatterChartOption {
				opt := makeBasicScatterChartOption()
				opt.XAxis.Labels = opt.XAxis.Labels[:5]
				for i := range opt.SeriesList {
					opt.SeriesList[i].Values = opt.SeriesList[i].Values[:5]
				}
				opt.SymbolSize = 4.5
				opt.Symbol = SymbolDot
				opt.Legend.Symbol = SymbolDot
				opt.Title.Show = Ptr(false)
				opt.XAxis.Show = Ptr(false)
				opt.YAxis[0].Show = Ptr(false)
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><path d=\"M 250 19\nL 280 19\" style=\"stroke-width:3;stroke:rgb(84,112,198);fill:none\"/><circle cx=\"265\" cy=\"19\" r=\"5\" style=\"stroke-width:3;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><text x=\"282\" y=\"25\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><path d=\"M 311 19\nL 341 19\" style=\"stroke-width:3;stroke:rgb(145,204,117);fill:none\"/><circle cx=\"326\" cy=\"19\" r=\"5\" style=\"stroke-width:3;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><text x=\"343\" y=\"25\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><circle cx=\"10\" cy=\"365\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"155\" cy=\"362\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"300\" cy=\"369\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"445\" cy=\"362\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"590\" cy=\"371\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"10\" cy=\"214\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"155\" cy=\"190\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"300\" cy=\"197\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"445\" cy=\"190\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"590\" cy=\"113\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/></svg>",
			pngCRC: 0x2c36c75c,
		},
		{
			name: "symbol_circle",
			makeOptions: func() ScatterChartOption {
				opt := makeBasicScatterChartOption()
				opt.XAxis.Labels = opt.XAxis.Labels[:5]
				for i := range opt.SeriesList {
					opt.SeriesList[i].Values = opt.SeriesList[i].Values[:5]
				}
				opt.SymbolSize = 4.5
				opt.Symbol = SymbolCircle
				opt.Legend.Symbol = SymbolCircle
				opt.Title.Show = Ptr(false)
				opt.XAxis.Show = Ptr(false)
				opt.YAxis[0].Show = Ptr(false)
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><path d=\"M 250 19\nL 280 19\" style=\"stroke-width:3;stroke:rgb(84,112,198);fill:none\"/><circle cx=\"265\" cy=\"19\" r=\"5\" style=\"stroke-width:3;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"265\" cy=\"19\" r=\"2\" style=\"stroke-width:3;stroke:white;fill:white\"/><text x=\"282\" y=\"25\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><path d=\"M 311 19\nL 341 19\" style=\"stroke-width:3;stroke:rgb(145,204,117);fill:none\"/><circle cx=\"326\" cy=\"19\" r=\"5\" style=\"stroke-width:3;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"326\" cy=\"19\" r=\"2\" style=\"stroke-width:3;stroke:white;fill:white\"/><text x=\"343\" y=\"25\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><circle cx=\"10\" cy=\"365\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:white\"/><circle cx=\"155\" cy=\"362\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:white\"/><circle cx=\"300\" cy=\"369\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:white\"/><circle cx=\"445\" cy=\"362\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:white\"/><circle cx=\"590\" cy=\"371\" r=\"5\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:white\"/><circle cx=\"10\" cy=\"214\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:white\"/><circle cx=\"155\" cy=\"190\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:white\"/><circle cx=\"300\" cy=\"197\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:white\"/><circle cx=\"445\" cy=\"190\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:white\"/><circle cx=\"590\" cy=\"113\" r=\"5\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:white\"/></svg>",
			pngCRC: 0x8d11b7f9,
		},
		{
			name: "symbol_square",
			makeOptions: func() ScatterChartOption {
				opt := makeBasicScatterChartOption()
				opt.XAxis.Labels = opt.XAxis.Labels[:5]
				for i := range opt.SeriesList {
					opt.SeriesList[i].Values = opt.SeriesList[i].Values[:5]
				}
				opt.SymbolSize = 4.5
				opt.Symbol = SymbolSquare
				opt.Legend.Symbol = SymbolSquare
				opt.Title.Show = Ptr(false)
				opt.XAxis.Show = Ptr(false)
				opt.YAxis[0].Show = Ptr(false)
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><path d=\"M 250 13\nL 280 13\nL 280 26\nL 250 26\nL 250 13\" style=\"stroke:none;fill:rgb(84,112,198)\"/><text x=\"282\" y=\"25\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><path d=\"M 311 13\nL 341 13\nL 341 26\nL 311 26\nL 311 13\" style=\"stroke:none;fill:rgb(145,204,117)\"/><text x=\"343\" y=\"25\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><path d=\"M 5 360\nL 14 360\nL 14 369\nL 5 369\nL 5 360\nM 150 357\nL 159 357\nL 159 366\nL 150 366\nL 150 357\nM 295 364\nL 304 364\nL 304 373\nL 295 373\nL 295 364\nM 440 357\nL 449 357\nL 449 366\nL 440 366\nL 440 357\nM 585 366\nL 594 366\nL 594 375\nL 585 375\nL 585 366\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path d=\"M 5 209\nL 14 209\nL 14 218\nL 5 218\nL 5 209\nM 150 185\nL 159 185\nL 159 194\nL 150 194\nL 150 185\nM 295 192\nL 304 192\nL 304 201\nL 295 201\nL 295 192\nM 440 185\nL 449 185\nL 449 194\nL 440 194\nL 440 185\nM 585 108\nL 594 108\nL 594 117\nL 585 117\nL 585 108\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/></svg>",
			pngCRC: 0xd385429a,
		},
		{
			name: "symbol_diamond",
			makeOptions: func() ScatterChartOption {
				opt := makeBasicScatterChartOption()
				opt.XAxis.Labels = opt.XAxis.Labels[:5]
				for i := range opt.SeriesList {
					opt.SeriesList[i].Values = opt.SeriesList[i].Values[:5]
				}
				opt.SymbolSize = 4.5
				opt.Symbol = SymbolDiamond
				opt.Legend.Symbol = SymbolDiamond
				opt.Title.Show = Ptr(false)
				opt.XAxis.Show = Ptr(false)
				opt.YAxis[0].Show = Ptr(false)
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><path d=\"M 265 10\nL 272 20\nL 265 30\nL 258 20\nL 265 10\" style=\"stroke:none;fill:rgb(84,112,198)\"/><text x=\"282\" y=\"25\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><path d=\"M 316 10\nL 323 20\nL 316 30\nL 309 20\nL 316 10\" style=\"stroke:none;fill:rgb(145,204,117)\"/><text x=\"333\" y=\"25\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><path d=\"M 10 359\nL 16 365\nL 10 371\nL 4 365\nL 10 359\nM 155 356\nL 161 362\nL 155 368\nL 149 362\nL 155 356\nM 300 363\nL 306 369\nL 300 375\nL 294 369\nL 300 363\nM 445 356\nL 451 362\nL 445 368\nL 439 362\nL 445 356\nM 590 365\nL 596 371\nL 590 377\nL 584 371\nL 590 365\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path d=\"M 10 208\nL 16 214\nL 10 220\nL 4 214\nL 10 208\nM 155 184\nL 161 190\nL 155 196\nL 149 190\nL 155 184\nM 300 191\nL 306 197\nL 300 203\nL 294 197\nL 300 191\nM 445 184\nL 451 190\nL 445 196\nL 439 190\nL 445 184\nM 590 107\nL 596 113\nL 590 119\nL 584 113\nL 590 107\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/></svg>",
			pngCRC: 0x8e351e73,
		},
		{
			name:   "symbol_mixed",
			ignore: "size", // svg is too big to commit
			makeOptions: func() ScatterChartOption {
				opt := makeFullScatterChartOption()
				opt.XAxis.Labels = opt.XAxis.Labels[:5]
				for i := range opt.SeriesList {
					opt.SeriesList[i].Values = opt.SeriesList[i].Values[:5]
				}
				opt.SymbolSize = 4.0
				opt.SeriesList[0].Symbol = SymbolCircle
				opt.SeriesList[1].Symbol = SymbolSquare
				opt.SeriesList[2].Symbol = SymbolDiamond
				opt.SeriesList[3].Symbol = SymbolDot
				opt.Title.Show = Ptr(false)
				opt.XAxis.Show = Ptr(false)
				opt.YAxis[0].Show = Ptr(false)
				return opt
			},
			svg:    "",
			pngCRC: 0,
		},
		{
			name:   "dense_trends",
			ignore: "size", // svg is too big to commit
			makeOptions: func() ScatterChartOption {
				opt := makeDenseScatterChartOption()
				for i := range opt.SeriesList {
					opt.SeriesList[i].TrendLine[0].StrokeSmoothingTension = 0.9 // smooth average line
					opt.SeriesList[i].TrendLine[0].Period = 5
					c1 := Color{
						R: uint8(80 + (20 * i)),
						G: uint8(80 + (20 * i)),
						B: uint8(80 + (20 * i)),
						A: 255,
					}
					c2 := c1
					if i%2 == 0 {
						c2.R = 200
					} else {
						c2.B = 200
					}
					trendLine1 := SeriesTrendLine{
						Type:      SeriesTrendTypeCubic,
						LineColor: c1,
					}
					trendLine2 := SeriesTrendLine{
						Type:      SeriesTrendTypeLinear,
						LineColor: c2,
					}
					opt.SeriesList[i].TrendLine = append(opt.SeriesList[i].TrendLine, trendLine1, trendLine2)
				}
				// disable extras
				opt.XAxis.Show = Ptr(false)
				opt.YAxis[0].Show = Ptr(false)
				return opt
			},
			svg:    "",
			pngCRC: 0,
		},
		{
			name: "trend_line_dashed",
			makeOptions: func() ScatterChartOption {
				opt := makeBasicScatterChartOption()
				opt.Theme = GetDefaultTheme()
				opt.Title.Show = Ptr(false)
				opt.XAxis.Show = Ptr(false)
				opt.YAxis[0].Show = Ptr(false)
				opt.Legend.Show = Ptr(false)
				// Add dashed trend lines to each series
				for i := range opt.SeriesList {
					opt.SeriesList[i].TrendLine = []SeriesTrendLine{
						{
							StrokeSmoothingTension: 0.8,
							Type:                   SeriesTrendTypeSMA,
							DashedLine:             Ptr(true), // Explicitly set to dashed
							LineColor:              opt.Theme.GetSeriesTrendColor(i).WithAdjustHSL(0, .2, -.2),
						},
						{
							Type:       SeriesTrendTypeCubic,
							DashedLine: Ptr(true), // Explicitly set to dashed
							LineColor:  opt.Theme.GetSeriesTrendColor(i).WithAdjustHSL(0, .4, -.4),
						},
					}
				}
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><circle cx=\"10\" cy=\"362\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"106\" cy=\"359\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"203\" cy=\"367\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"300\" cy=\"359\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"396\" cy=\"369\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"493\" cy=\"336\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"590\" cy=\"341\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"10\" cy=\"196\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"106\" cy=\"169\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"203\" cy=\"177\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"300\" cy=\"169\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"396\" cy=\"84\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"493\" cy=\"75\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"590\" cy=\"77\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><path stroke-dasharray=\"9.6, 7.7\" d=\"M 10 361\nQ106,363 144,362\nQ203,361 241,362\nQ300,365 338,361\nQ396,355 434,352\nQ493,349 531,344\nQ493,349 590,338\" style=\"stroke-width:2;stroke:rgb(12,38,115);fill:none\"/><path stroke-dasharray=\"9.6, 7.7\" d=\"M 10 360\nL 106 364\nL 203 364\nL 300 362\nL 396 357\nL 493 348\nL 590 337\" style=\"stroke-width:2;stroke:rgb(0,6,25);fill:none\"/><path stroke-dasharray=\"9.6, 7.7\" d=\"M 10 182\nQ106,180 144,176\nQ203,171 241,159\nQ300,143 338,129\nQ396,109 434,97\nQ493,79 531,77\nQ493,79 590,76\" style=\"stroke-width:2;stroke:rgb(61,146,20);fill:none\"/><path stroke-dasharray=\"9.6, 7.7\" d=\"M 10 187\nL 106 192\nL 203 173\nL 300 140\nL 396 104\nL 493 78\nL 590 73\" style=\"stroke-width:2;stroke:rgb(21,63,1);fill:none\"/></svg>",
			pngCRC: 0xb8e90568,
		},
		{
			name: "with_conditional_labels",
			makeOptions: func() ScatterChartOption {
				return ScatterChartOption{
					Padding: NewBoxEqual(10),
					XAxis: XAxisOption{
						Labels: []string{"A", "B", "C", "D", "E"},
					},
					YAxis: []YAxisOption{{}},
					SeriesList: NewSeriesListScatter([][]float64{
						{50, 150, 100, 200, 175},
						{75, 125, 90, 160, 140},
					}, ScatterSeriesOption{
						Names: []string{"Dataset1", "Dataset2"},
						Label: SeriesLabel{
							Show: Ptr(true),
							LabelFormatter: func(index int, name string, val float64) (string, *LabelStyle) {
								// Show labels only for values above 120
								if val > 120 {
									switch {
									case val >= 180: // High values - gold styling
										return "⭐ " + strconv.FormatFloat(val, 'f', 0, 64), &LabelStyle{
											FontStyle:       FontStyle{FontColor: ColorBlack, FontSize: 14},
											BackgroundColor: ColorFromHex("#FFD700"), // Gold
											CornerRadius:    6,
										}
									case val >= 150: // Medium-high values - silver styling
										return "📊 " + strconv.FormatFloat(val, 'f', 0, 64), &LabelStyle{
											FontStyle:       FontStyle{FontColor: ColorBlack, FontSize: 12},
											BackgroundColor: ColorFromHex("#C0C0C0"), // Silver
											CornerRadius:    4,
										}
									default: // Values above 120 but below 150 - simple styling
										return strconv.FormatFloat(val, 'f', 0, 64), &LabelStyle{
											FontStyle: FontStyle{FontColor: ColorBlue, FontSize: 10},
										}
									}
								}
								// Hide labels for values <= 120
								return "", nil
							},
						},
					}),
					Title: TitleOption{
						Show: Ptr(false),
					},
					Legend: LegendOption{
						Show: Ptr(false),
					},
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">220</text><text x=\"9\" y=\"60\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">195</text><text x=\"9\" y=\"104\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">170</text><text x=\"9\" y=\"148\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">145</text><text x=\"9\" y=\"192\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"18\" y=\"236\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"280\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">70</text><text x=\"18\" y=\"324\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">45</text><text x=\"18\" y=\"368\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">20</text><path d=\"M 42 10\nL 590 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 54\nL 590 54\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 98\nL 590 98\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 142\nL 590 142\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 187\nL 590 187\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 231\nL 590 231\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 275\nL 590 275\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 319\nL 590 319\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 46 364\nL 590 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 46 369\nL 46 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 182 369\nL 182 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 318 369\nL 318 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 454 369\nL 454 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 590 369\nL 590 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"45\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">A</text><text x=\"181\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">B</text><text x=\"317\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">C</text><text x=\"453\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">D</text><text x=\"581\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">E</text><circle cx=\"46\" cy=\"311\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"182\" cy=\"134\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"318\" cy=\"223\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"454\" cy=\"46\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"590\" cy=\"90\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"46\" cy=\"267\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"182\" cy=\"179\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"318\" cy=\"241\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"454\" cy=\"117\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"590\" cy=\"152\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><path d=\"M 187 120\nL 229 120\nL 229 120\nA 4 4 90.00 0 1 233 124\nL 233 140\nL 233 140\nA 4 4 90.00 0 1 229 144\nL 187 144\nL 187 144\nA 4 4 90.00 0 1 183 140\nL 183 124\nL 183 124\nA 4 4 90.00 0 1 187 120\nZ\" style=\"stroke:none;fill:silver\"/><text x=\"187\" y=\"140\" style=\"stroke:none;fill:black;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">📊 150</text><path d=\"M 461 31\nL 506 31\nL 506 31\nA 6 6 90.00 0 1 512 37\nL 512 51\nL 512 51\nA 6 6 90.00 0 1 506 57\nL 461 57\nL 461 57\nA 6 6 90.00 0 1 455 51\nL 455 37\nL 455 37\nA 6 6 90.00 0 1 461 31\nZ\" style=\"stroke:none;fill:rgb(255,215,0)\"/><text x=\"459\" y=\"53\" style=\"stroke:none;fill:black;font-size:17.9px;font-family:'Roboto Medium',sans-serif\">⭐ 200</text><path d=\"M 558 76\nL 600 76\nL 600 76\nA 4 4 90.00 0 1 604 80\nL 604 96\nL 604 96\nA 4 4 90.00 0 1 600 100\nL 558 100\nL 558 100\nA 4 4 90.00 0 1 554 96\nL 554 80\nL 554 80\nA 4 4 90.00 0 1 558 76\nZ\" style=\"stroke:none;fill:silver\"/><text x=\"558\" y=\"96\" style=\"stroke:none;fill:black;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">📊 175</text><text x=\"187\" y=\"183\" style=\"stroke:none;fill:blue;font-size:12.8px;font-family:'Roboto Medium',sans-serif\">125</text><path d=\"M 459 103\nL 501 103\nL 501 103\nA 4 4 90.00 0 1 505 107\nL 505 123\nL 505 123\nA 4 4 90.00 0 1 501 127\nL 459 127\nL 459 127\nA 4 4 90.00 0 1 455 123\nL 455 107\nL 455 107\nA 4 4 90.00 0 1 459 103\nZ\" style=\"stroke:none;fill:silver\"/><text x=\"459\" y=\"123\" style=\"stroke:none;fill:black;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">📊 160</text><text x=\"578\" y=\"156\" style=\"stroke:none;fill:blue;font-size:12.8px;font-family:'Roboto Medium',sans-serif\">140</text></svg>",
			pngCRC: 0x30d22ac6,
		},
		{
			name: "bollinger",
			makeOptions: func() ScatterChartOption {
				opt := makeBasicScatterChartOption()
				opt.SeriesList[0].TrendLine = []SeriesTrendLine{
					{Type: SeriesTrendTypeBollingerLower, Period: 3},
				}
				opt.SeriesList[1].TrendLine = []SeriesTrendLine{
					{Type: SeriesTrendTypeBollingerUpper, Period: 3},
				}
				opt.Legend.Show = Ptr(false)
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1.6k</text><text x=\"9\" y=\"60\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1.4k</text><text x=\"9\" y=\"104\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1.2k</text><text x=\"22\" y=\"148\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1k</text><text x=\"12\" y=\"192\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">800</text><text x=\"12\" y=\"236\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">600</text><text x=\"12\" y=\"280\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">400</text><text x=\"12\" y=\"324\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">200</text><text x=\"30\" y=\"368\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">0</text><path d=\"M 45 10\nL 590 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 45 54\nL 590 54\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 45 98\nL 590 98\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 45 142\nL 590 142\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 45 187\nL 590 187\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 45 231\nL 590 231\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 45 275\nL 590 275\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 45 319\nL 590 319\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 49 364\nL 590 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 49 369\nL 49 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 139 369\nL 139 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 229 369\nL 229 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 319 369\nL 319 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 409 369\nL 409 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 499 369\nL 499 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 590 369\nL 590 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"48\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">A</text><text x=\"138\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">B</text><text x=\"228\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">C</text><text x=\"318\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">D</text><text x=\"408\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">E</text><text x=\"498\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">F</text><text x=\"579\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">G</text><circle cx=\"49\" cy=\"338\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"139\" cy=\"335\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"229\" cy=\"342\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"319\" cy=\"335\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"409\" cy=\"345\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"499\" cy=\"314\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"590\" cy=\"318\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"49\" cy=\"183\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"139\" cy=\"158\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"229\" cy=\"165\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"319\" cy=\"158\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"409\" cy=\"79\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"499\" cy=\"70\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"590\" cy=\"72\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><path d=\"M 49 339\nL 139 344\nL 229 344\nL 319 349\nL 409 357\nL 499 353\nL 590 320\" style=\"stroke-width:2;stroke:rgb(46,80,184);fill:none\"/><path d=\"M 49 146\nL 139 148\nL 229 154\nL 319 56\nL 409 24\nL 499 66\nL 590 69\" style=\"stroke-width:2;stroke:rgb(111,202,67);fill:none\"/></svg>",
			pngCRC: 0x44072616,
		},
		{
			name: "rsi",
			makeOptions: func() ScatterChartOption {
				opt := makeBasicScatterChartOption()
				opt.SeriesList[0].TrendLine = []SeriesTrendLine{
					{Type: SeriesTrendTypeRSI, Period: 3},
				}
				opt.Legend.Show = Ptr(false)
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1.6k</text><text x=\"9\" y=\"60\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1.4k</text><text x=\"9\" y=\"104\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1.2k</text><text x=\"22\" y=\"148\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1k</text><text x=\"12\" y=\"192\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">800</text><text x=\"12\" y=\"236\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">600</text><text x=\"12\" y=\"280\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">400</text><text x=\"12\" y=\"324\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">200</text><text x=\"30\" y=\"368\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">0</text><path d=\"M 45 10\nL 590 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 45 54\nL 590 54\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 45 98\nL 590 98\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 45 142\nL 590 142\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 45 187\nL 590 187\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 45 231\nL 590 231\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 45 275\nL 590 275\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 45 319\nL 590 319\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 49 364\nL 590 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 49 369\nL 49 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 139 369\nL 139 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 229 369\nL 229 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 319 369\nL 319 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 409 369\nL 409 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 499 369\nL 499 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 590 369\nL 590 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"48\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">A</text><text x=\"138\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">B</text><text x=\"228\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">C</text><text x=\"318\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">D</text><text x=\"408\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">E</text><text x=\"498\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">F</text><text x=\"579\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">G</text><circle cx=\"49\" cy=\"338\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"139\" cy=\"335\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"229\" cy=\"342\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"319\" cy=\"335\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"409\" cy=\"345\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"499\" cy=\"314\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"590\" cy=\"318\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><circle cx=\"49\" cy=\"183\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"139\" cy=\"158\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"229\" cy=\"165\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"319\" cy=\"158\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"409\" cy=\"79\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"499\" cy=\"70\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><circle cx=\"590\" cy=\"72\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><path d=\"M 319 351\nL 409 357\nL 499 347\nL 590 349\" style=\"stroke-width:2;stroke:rgb(46,80,184);fill:none\"/></svg>",
			pngCRC: 0x345a2e06,
		},
	}

	for i, tt := range tests {
		if tt.ignore != "" {
			continue
		}
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

				validateScatterChartRender(t, p, rp, tt.makeOptions(), tt.svg, tt.pngCRC)
			})
		} else {
			theme := GetTheme(ThemeVividDark)
			t.Run(strconv.Itoa(i)+"-"+tt.name+"-theme_painter", func(t *testing.T) {
				p := NewPainter(painterOptions, PainterThemeOption(theme))
				rp := NewPainter(rasterOptions, PainterThemeOption(theme))

				validateScatterChartRender(t, p, rp, tt.makeOptions(), tt.svg, tt.pngCRC)
			})
			t.Run(strconv.Itoa(i)+"-"+tt.name+"-theme_opt", func(t *testing.T) {
				p := NewPainter(painterOptions)
				rp := NewPainter(rasterOptions)
				opt := tt.makeOptions()
				opt.Theme = theme

				validateScatterChartRender(t, p, rp, opt, tt.svg, tt.pngCRC)
			})
		}
	}
}

func validateScatterChartRender(t *testing.T, svgP, pngP *Painter, opt ScatterChartOption, expectedSVG string, expectedCRC uint32) {
	t.Helper()

	err := svgP.ScatterChart(opt)
	require.NoError(t, err)
	data, err := svgP.Bytes()
	require.NoError(t, err)
	assertEqualSVG(t, expectedSVG, data)

	err = pngP.ScatterChart(opt)
	require.NoError(t, err)
	rasterData, err := pngP.Bytes()
	require.NoError(t, err)
	assertEqualPNGCRC(t, expectedCRC, rasterData)
}

func TestScatterChartError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		makeOptions      func() ScatterChartOption
		errorMsgContains string
	}{
		{
			name: "empty_series",
			makeOptions: func() ScatterChartOption {
				return NewScatterChartOptionWithData([][]float64{})
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

			err := p.ScatterChart(tt.makeOptions())
			require.Error(t, err)
			require.ErrorContains(t, err, tt.errorMsgContains)
		})
	}
}
