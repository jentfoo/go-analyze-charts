package charts

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertToArray(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []byte(`[1]`), convertToArray([]byte("1")))
	assert.Equal(t, []byte(`[1]`), convertToArray([]byte("[1]")))
}

func TestEChartsPosition(t *testing.T) {
	t.Parallel()

	var p EChartsPosition
	require.NoError(t, p.UnmarshalJSON([]byte("1")))
	assert.Equal(t, EChartsPosition("1"), p)
	require.NoError(t, p.UnmarshalJSON([]byte(`"left"`)))
	assert.Equal(t, EChartsPosition("left"), p)
}

func TestEChartsSeriesDataValue(t *testing.T) {
	t.Parallel()

	es := EChartsSeriesDataValue{}
	require.NoError(t, es.UnmarshalJSON([]byte(`[1, 2]`)))
	assert.Equal(t, EChartsSeriesDataValue{
		values: []float64{1, 2},
	}, es)
	assert.Equal(t, EChartsSeriesDataValue{values: []float64{1, 2}}, es)
	assert.InDelta(t, 1.0, es.First(), 0)
}

func TestEChartsSeriesData(t *testing.T) {
	t.Parallel()

	es := EChartsSeriesData{}
	require.NoError(t, es.UnmarshalJSON([]byte("1.1")))
	assert.Equal(t, EChartsSeriesDataValue{
		values: []float64{1.1},
	}, es.Value)

	require.NoError(t, es.UnmarshalJSON([]byte(`{"value":200,"itemStyle":{"color":"#a90000"}}`)))
	assert.Equal(t, EChartsSeriesData{
		Value: EChartsSeriesDataValue{
			values: []float64{200.0},
		},
		ItemStyle: EChartStyle{
			Color: "#a90000",
		},
	}, es)
}

func TestEChartsXAxis(t *testing.T) {
	t.Parallel()

	ex := EChartsXAxis{}
	require.NoError(t, ex.UnmarshalJSON([]byte(`{"boundaryGap": true, "splitNumber": 5, "data": ["a", "b"], "type": "value"}`)))

	assert.Equal(t, EChartsXAxis{
		Data: []EChartsXAxisData{
			{
				BoundaryGap: Ptr(true),
				SplitNumber: 5,
				Data:        []string{"a", "b"},
				Type:        "value",
			},
		},
	}, ex)
}

func TestEChartsPadding(t *testing.T) {
	t.Parallel()

	eb := EChartsPadding{}

	require.NoError(t, eb.UnmarshalJSON([]byte(`1`)))
	assert.Equal(t, NewBoxEqual(1), eb.Box)

	require.NoError(t, eb.UnmarshalJSON([]byte(`[2, 3]`)))
	assert.Equal(t, Box{
		Left:   3,
		Top:    2,
		Right:  3,
		Bottom: 2,
		IsSet:  true,
	}, eb.Box)

	require.NoError(t, eb.UnmarshalJSON([]byte(`[4, 5, 6]`)))
	assert.Equal(t, Box{
		Left:   5,
		Top:    4,
		Right:  5,
		Bottom: 6,
		IsSet:  true,
	}, eb.Box)

	require.NoError(t, eb.UnmarshalJSON([]byte(`[4, 5, 6, 7]`)))
	assert.Equal(t, Box{
		Left:   7,
		Top:    4,
		Right:  5,
		Bottom: 6,
		IsSet:  true,
	}, eb.Box)
}

func TestEChartsMarkPoint(t *testing.T) {
	t.Parallel()

	emp := EChartsMarkPoint{
		SymbolSize: 30,
		Data: []EChartsMarkData{
			{
				Type: "test",
			},
		},
	}
	assert.Equal(t, SeriesMarkPoint{
		SymbolSize: 30,
		Points: []SeriesMark{
			{
				Type: "test",
			},
		},
	}, emp.ToSeriesMarkPoint())
}

func TestEChartsMarkLine(t *testing.T) {
	t.Parallel()

	eml := EChartsMarkLine{
		Data: []EChartsMarkData{
			{
				Type: "min",
			},
			{
				Type: "max",
			},
		},
	}
	assert.Equal(t, SeriesMarkLine{
		Lines: []SeriesMark{
			{
				Type: "min",
			},
			{
				Type: "max",
			},
		},
	}, eml.ToSeriesMarkLine())
}

func TestEChartsOption(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		option string
	}{
		{
			option: `{
				"xAxis": {
					"type": "category",
					"data": [
						"Mon",
						"Tue",
						"Wed",
						"Thu",
						"Fri",
						"Sat",
						"Sun"
					]
				},
				"yAxis": {
					"type": "value"
				},
				"series": [
					{
						"data": [
							120,
							{
								"value": 200,
								"itemStyle": {
									"color": "#a90000"
								}
							},
							150,
							80,
							70,
							110,
							130
						],
						"type": "bar"
					}
				]
			}`,
		},
		{
			option: `{
				"title": {
					"text": "Referer of a Website",
					"subtext": "Fake Data",
					"left": "center"
				},
				"tooltip": {
					"trigger": "item"
				},
				"legend": {
					"orient": "vertical",
					"left": "left"
				},
				"series": [
					{
						"name": "Access From",
						"type": "pie",
						"radius": "50%",
						"data": [
							{
								"value": 1048,
								"name": "Search Engine"
							},
							{
								"value": 735,
								"name": "Direct"
							},
							{
								"value": 580,
								"name": "Email"
							},
							{
								"value": 484,
								"name": "Union Ads"
							},
							{
								"value": 300,
								"name": "Video Ads"
							}
						],
						"emphasis": {
							"itemStyle": {
								"shadowBlur": 10,
								"shadowOffsetX": 0,
								"shadowColor": "rgba(0, 0, 0, 0.5)"
							}
						}
					}
				]
			}`,
		},
		{
			option: `{
				"title": {
					"text": "Rainfall vs Evaporation",
					"subtext": "Fake Data"
				},
				"tooltip": {
					"trigger": "axis"
				},
				"legend": {
					"data": [
						"Rainfall",
						"Evaporation"
					]
				},
				"toolbox": {
					"show": true,
					"feature": {
						"dataView": {
							"show": true,
							"readOnly": false
						},
						"magicType": {
							"show": true,
							"type": [
								"line",
								"bar"
							]
						},
						"restore": {
							"show": true
						},
						"saveAsImage": {
							"show": true
						}
					}
				},
				"calculable": true,
				"xAxis": [
					{
						"type": "category",
						"data": [
							"Jan",
							"Feb",
							"Mar",
							"Apr",
							"May",
							"Jun",
							"Jul",
							"Aug",
							"Sep",
							"Oct",
							"Nov",
							"Dec"
						]
					}
				],
				"yAxis": [
					{
						"type": "value"
					}
				],
				"series": [
					{
						"name": "Rainfall",
						"type": "bar",
						"data": [
							2,
							4.9,
							7,
							23.2,
							25.6,
							76.7,
							135.6,
							162.2,
							32.6,
							20,
							6.4,
							3.3
						],
						"markPoint": {
							"data": [
								{
									"type": "max",
									"name": "Max"
								},
								{
									"type": "min",
									"name": "Min"
								}
							]
						},
						"markLine": {
							"data": [
								{
									"type": "average",
									"name": "Avg"
								}
							]
						}
					},
					{
						"name": "Evaporation",
						"type": "bar",
						"data": [
							2.6,
							5.9,
							9,
							26.4,
							28.7,
							70.7,
							175.6,
							182.2,
							48.7,
							18.8,
							6,
							2.3
						],
						"markPoint": {
							"data": [
								{
									"name": "Max",
									"value": 182.2,
									"xAxis": 7,
									"yAxis": 183
								},
								{
									"name": "Min",
									"value": 2.3,
									"xAxis": 11,
									"yAxis": 3
								}
							]
						},
						"markLine": {
							"data": [
								{
									"type": "average",
									"name": "Avg"
								}
							]
						}
					}
				]
			}`,
		},
		{
			name: "basic_bar_Chart",
			option: `{
				"xAxis": { "type": "category", "data": ["Mon", "Tue", "Wed"] },
				"yAxis": { "type": "value" },
				"series": [{ "data": [120, 200, 150], "type": "bar" }]
			}`,
		},
		{
			name: "basic_pie_chart",
			option: `{
				"title": { "text": "Website Traffic", "left": "center" },
				"series": [{ "name": "Source", "type": "pie", "data": [{ "value": 100, "name": "Google" }] }]
			}`,
		},
	}

	for i, tt := range tests {
		name := strconv.Itoa(i)
		if tt.name != "" {
			name += tt.name
		}
		t.Run(name, func(t *testing.T) {
			opt := EChartsOption{}
			require.NoError(t, json.Unmarshal([]byte(tt.option), &opt))
			assert.NotEmpty(t, opt.Series)
			assert.NotEmpty(t, opt.ToOption().SeriesList)

			if len(opt.XAxis.Data) > 0 {
				assert.NotEmpty(t, opt.XAxis.Data[0].Data)
				assert.NotEmpty(t, opt.XAxis.Data[0].Type)
			}
		})
	}
}

func TestRenderEChartsToSVG(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		jsonData string
		expected string // Placeholder for expected SVG, can be empty for now
	}{
		{
			name: "detailed",
			jsonData: `{
		"title": {
			"text": "Rainfall vs Evaporation",
			"subtext": "Fake Data"
		},
		"legend": {
			"data": [
				"Rainfall",
				"Evaporation"
			]
		},
		"padding": [10, 30, 10, 10],
		"xAxis": [
			{
				"type": "category",
				"data": [
					"Jan",
					"Feb",
					"Mar",
					"Apr",
					"May",
					"Jun",
					"Jul",
					"Aug",
					"Sep",
					"Oct",
					"Nov",
					"Dec"
				]
			}
		],
		"series": [
			{
				"name": "Rainfall",
				"type": "bar",
				"data": [
					2,
					4.9,
					7,
					23.2,
					25.6,
					76.7,
					135.6,
					162.2,
					32.6,
					20,
					6.4,
					3.3
				],
				"markPoint": {
					"data": [
						{
							"type": "max"
						},
						{
							"type": "min"
						}
					]
				},
				"markLine": {
					"data": [
						{
							"type": "average"
						}
					]
				}
			},
			{
				"name": "Evaporation",
				"type": "bar",
				"data": [
					2.6,
					5.9,
					9,
					26.4,
					28.7,
					70.7,
					175.6,
					182.2,
					48.7,
					18.8,
					6,
					2.3
				],
				"markPoint": {
					"data": [
						{
							"type": "max"
						},
						{
							"type": "min"
						}
					]
				},
				"markLine": {
					"data": [
						{
							"type": "average"
						}
					]
				}
			}
		]
	}`,
			expected: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"10\" y=\"26\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Rainfall vs Evaporation</text><text x=\"54\" y=\"42\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Fake Data</text><path d=\"M 182 19\nL 212 19\" style=\"stroke-width:3;stroke:rgb(84,112,198);fill:none\"/><circle cx=\"197\" cy=\"19\" r=\"5\" style=\"stroke-width:3;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><text x=\"214\" y=\"25\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Rainfall</text><path d=\"M 286 19\nL 316 19\" style=\"stroke-width:3;stroke:rgb(145,204,117);fill:none\"/><circle cx=\"301\" cy=\"19\" r=\"5\" style=\"stroke-width:3;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><text x=\"318\" y=\"25\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Evaporation</text><text x=\"9\" y=\"63\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">225</text><text x=\"9\" y=\"96\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">200</text><text x=\"9\" y=\"130\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">175</text><text x=\"9\" y=\"164\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">150</text><text x=\"9\" y=\"198\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"232\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"266\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">75</text><text x=\"18\" y=\"300\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">50</text><text x=\"18\" y=\"334\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">25</text><text x=\"27\" y=\"368\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">0</text><path d=\"M 42 57\nL 570 57\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 91\nL 570 91\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 125\nL 570 125\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 159\nL 570 159\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 193\nL 570 193\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 227\nL 570 227\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 261\nL 570 261\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 295\nL 570 295\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 329\nL 570 329\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 46 364\nL 570 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 46 369\nL 46 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 89 369\nL 89 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 133 369\nL 133 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 177 369\nL 177 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 220 369\nL 220 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 264 369\nL 264 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 308 369\nL 308 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 351 369\nL 351 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 395 369\nL 395 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 439 369\nL 439 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 482 369\nL 482 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 526 369\nL 526 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 570 369\nL 570 364\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"54\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Jan</text><text x=\"98\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Feb</text><text x=\"141\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Mar</text><text x=\"186\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Apr</text><text x=\"227\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">May</text><text x=\"273\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Jun</text><text x=\"319\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Jul</text><text x=\"359\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Aug</text><text x=\"404\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Sep</text><text x=\"448\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Oct</text><text x=\"490\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Nov</text><text x=\"535\" y=\"390\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Dec</text><path d=\"M 51 362\nL 66 362\nL 66 363\nL 51 363\nL 51 362\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 94 358\nL 109 358\nL 109 363\nL 94 363\nL 94 358\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 138 355\nL 153 355\nL 153 363\nL 138 363\nL 138 355\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 182 333\nL 197 333\nL 197 363\nL 182 363\nL 182 333\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 225 330\nL 240 330\nL 240 363\nL 225 363\nL 225 330\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 269 260\nL 284 260\nL 284 363\nL 269 363\nL 269 260\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 313 179\nL 328 179\nL 328 363\nL 313 363\nL 313 179\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 356 143\nL 371 143\nL 371 363\nL 356 363\nL 356 143\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 400 320\nL 415 320\nL 415 363\nL 400 363\nL 400 320\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 444 337\nL 459 337\nL 459 363\nL 444 363\nL 444 337\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 487 356\nL 502 356\nL 502 363\nL 487 363\nL 487 356\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 531 360\nL 546 360\nL 546 363\nL 531 363\nL 531 360\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 69 361\nL 84 361\nL 84 363\nL 69 363\nL 69 361\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 112 356\nL 127 356\nL 127 363\nL 112 363\nL 112 356\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 156 352\nL 171 352\nL 171 363\nL 156 363\nL 156 352\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 200 328\nL 215 328\nL 215 363\nL 200 363\nL 200 328\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 243 325\nL 258 325\nL 258 363\nL 243 363\nL 243 325\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 287 268\nL 302 268\nL 302 363\nL 287 363\nL 287 268\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 331 125\nL 346 125\nL 346 363\nL 331 363\nL 331 125\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 374 116\nL 389 116\nL 389 363\nL 374 363\nL 374 116\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 418 298\nL 433 298\nL 433 363\nL 418 363\nL 418 298\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 462 339\nL 477 339\nL 477 363\nL 462 363\nL 462 339\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 505 356\nL 520 356\nL 520 363\nL 505 363\nL 505 356\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 549 361\nL 564 361\nL 564 363\nL 549 363\nL 549 361\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 359 136\nA 14 14 330.00 1 1 367 136\nL 363 122\nZ\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 349 122\nQ363,157 377,122\nZ\" style=\"stroke:none;fill:rgb(84,112,198)\"/><text x=\"350\" y=\"127\" style=\"stroke:none;fill:rgb(238,238,238);font-size:10.2px;font-family:'Roboto Medium',sans-serif\">162.2</text><path d=\"M 54 355\nA 14 14 330.00 1 1 62 355\nL 58 341\nZ\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 44 341\nQ58,376 72,341\nZ\" style=\"stroke:none;fill:rgb(84,112,198)\"/><text x=\"54\" y=\"346\" style=\"stroke:none;fill:rgb(238,238,238);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">2</text><path d=\"M 377 109\nA 14 14 330.00 1 1 385 109\nL 381 95\nZ\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 367 95\nQ381,130 395,95\nZ\" style=\"stroke:none;fill:rgb(145,204,117)\"/><text x=\"368\" y=\"100\" style=\"stroke:none;fill:rgb(70,70,70);font-size:10.2px;font-family:'Roboto Medium',sans-serif\">182.2</text><path d=\"M 552 354\nA 14 14 330.00 1 1 560 354\nL 556 340\nZ\" style=\"stroke:none;fill:rgb(145,204,117)\"/><path d=\"M 542 340\nQ556,375 570,340\nZ\" style=\"stroke:none;fill:rgb(145,204,117)\"/><text x=\"547\" y=\"345\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">2.3</text><circle cx=\"49\" cy=\"308\" r=\"3\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 55 308\nL 552 308\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 552 303\nL 568 308\nL 552 313\nL 557 308\nL 552 303\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><text x=\"570\" y=\"312\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">41.63</text><circle cx=\"49\" cy=\"299\" r=\"3\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 55 299\nL 552 299\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><path stroke-dasharray=\"4.0, 2.0\" d=\"M 552 294\nL 568 299\nL 552 304\nL 557 299\nL 552 294\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:rgb(145,204,117)\"/><text x=\"570\" y=\"303\" style=\"stroke:none;fill:rgb(70,70,70);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">48.07</text></svg>",
		},
		{
			name: "basic_bar_chart",
			jsonData: `{
				"title": { "text": "Sales" },
				"xAxis": { "type": "category", "data": ["Jan", "Feb"] },
				"yAxis": { "type": "value" },
				"series": [{ "data": [100, 200], "type": "bar" }]
			}`,
			expected: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"20\" y=\"36\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Sales</text><text x=\"19\" y=\"57\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">210</text><text x=\"19\" y=\"84\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">200</text><text x=\"19\" y=\"111\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">190</text><text x=\"19\" y=\"139\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">180</text><text x=\"19\" y=\"166\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">170</text><text x=\"19\" y=\"193\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">160</text><text x=\"19\" y=\"221\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">150</text><text x=\"19\" y=\"248\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">140</text><text x=\"19\" y=\"275\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"19\" y=\"303\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"19\" y=\"330\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"19\" y=\"358\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><path d=\"M 52 51\nL 580 51\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 52 78\nL 580 78\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 52 106\nL 580 106\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 52 133\nL 580 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 52 161\nL 580 161\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 52 188\nL 580 188\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 52 216\nL 580 216\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 52 243\nL 580 243\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 52 271\nL 580 271\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 52 298\nL 580 298\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 52 326\nL 580 326\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 56 354\nL 580 354\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 56 359\nL 56 354\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 318 359\nL 318 354\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 580 359\nL 580 354\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"174\" y=\"380\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Jan</text><text x=\"436\" y=\"380\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Feb</text><path d=\"M 66 354\nL 308 354\nL 308 353\nL 66 353\nL 66 354\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 328 79\nL 570 79\nL 570 353\nL 328 353\nL 328 79\" style=\"stroke:none;fill:rgb(84,112,198)\"/></svg>",
		},
		{
			name: "axis_styling",
			jsonData: `{
				"xAxis": { "axisLabel": { "color": "#ff0000", "fontSize": 14 } },
				"yAxis": { "axisLabel": { "color": "#00ff00", "fontSize": 12 } },
				"series": [{ "data": [10, 20], "type": "bar" }]
			}`,
			expected: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"19\" y=\"26\" style=\"stroke:none;fill:lime;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">21</text><text x=\"19\" y=\"56\" style=\"stroke:none;fill:lime;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">20</text><text x=\"19\" y=\"86\" style=\"stroke:none;fill:lime;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">19</text><text x=\"19\" y=\"116\" style=\"stroke:none;fill:lime;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">18</text><text x=\"19\" y=\"146\" style=\"stroke:none;fill:lime;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">17</text><text x=\"19\" y=\"176\" style=\"stroke:none;fill:lime;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">16</text><text x=\"19\" y=\"206\" style=\"stroke:none;fill:lime;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">15</text><text x=\"19\" y=\"236\" style=\"stroke:none;fill:lime;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">14</text><text x=\"19\" y=\"266\" style=\"stroke:none;fill:lime;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">13</text><text x=\"19\" y=\"296\" style=\"stroke:none;fill:lime;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">12</text><text x=\"19\" y=\"326\" style=\"stroke:none;fill:lime;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">11</text><text x=\"19\" y=\"356\" style=\"stroke:none;fill:lime;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">10</text><path d=\"M 43 20\nL 580 20\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 50\nL 580 50\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 80\nL 580 80\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 110\nL 580 110\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 140\nL 580 140\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 170\nL 580 170\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 201\nL 580 201\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 231\nL 580 231\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 261\nL 580 261\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 291\nL 580 291\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 321\nL 580 321\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 47 352\nL 580 352\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 47 357\nL 47 352\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 313 357\nL 313 352\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 580 357\nL 580 352\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"175\" y=\"380\" style=\"stroke:none;fill:red;font-size:17.9px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"441\" y=\"380\" style=\"stroke:none;fill:red;font-size:17.9px;font-family:'Roboto Medium',sans-serif\">2</text><path d=\"M 57 352\nL 303 352\nL 303 351\nL 57 351\nL 57 352\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 323 51\nL 569 51\nL 569 351\nL 323 351\nL 323 51\" style=\"stroke:none;fill:rgb(84,112,198)\"/></svg>",
		},
		{
			name: "title_and_axis_labels_hidden",
			jsonData: `{
				"title": {
					"show": false,
					"text": "Hidden Title"
				},
				"xAxis": { "axisLabel": { "show": false }, "type": "category", "data": ["X1", "X2"] },
				"yAxis": { "axisLabel": { "show": false }, "type": "value" },
				"series": [{ "data": [5, 15], "type": "bar" }]
			}`,
			expected: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"29\" y=\"18\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">16</text><text x=\"29\" y=\"48\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">15</text><text x=\"29\" y=\"78\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">14</text><text x=\"29\" y=\"108\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">13</text><text x=\"29\" y=\"139\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">12</text><text x=\"29\" y=\"169\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">11</text><text x=\"29\" y=\"199\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">10</text><text x=\"29\" y=\"229\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">9</text><text x=\"29\" y=\"260\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">8</text><text x=\"29\" y=\"290\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">7</text><text x=\"29\" y=\"320\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">6</text><text x=\"29\" y=\"351\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">5</text><path d=\"M 25 20\nL 580 20\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 50\nL 580 50\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 80\nL 580 80\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 111\nL 580 111\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 141\nL 580 141\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 172\nL 580 172\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 202\nL 580 202\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 233\nL 580 233\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 263\nL 580 263\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 294\nL 580 294\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 324\nL 580 324\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 29 355\nL 580 355\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 29 360\nL 29 355\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 304 360\nL 304 355\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 580 360\nL 580 355\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"166\" y=\"365\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">X1</text><text x=\"442\" y=\"365\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">X2</text><path d=\"M 39 355\nL 294 355\nL 294 354\nL 39 354\nL 39 355\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 314 51\nL 569 51\nL 569 354\nL 314 354\nL 314 51\" style=\"stroke:none;fill:rgb(84,112,198)\"/></svg>",
		},
		{
			name: "legend_border_color",
			jsonData: `{
				"legend": {
					"borderColor": "#00ff00",
					"data": ["Series1"]
				},
				"xAxis": { "axisLabel": { "show": false }, "type": "category", "data": ["A", "B"] },
				"yAxis": { "axisLabel": { "show": false }, "type": "value" },
				"series": [{ "data": [20, 30], "type": "line" }]
			}`,
			expected: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><path d=\"M 259 29\nL 289 29\" style=\"stroke-width:3;stroke:rgb(84,112,198);fill:none\"/><circle cx=\"274\" cy=\"29\" r=\"5\" style=\"stroke-width:3;stroke:rgb(84,112,198);fill:rgb(84,112,198)\"/><text x=\"291\" y=\"35\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Series1</text><path d=\"M 249 51\nL 249 10\nL 351 10\nL 351 51\nL 249 51\" style=\"stroke-width:2;stroke:lime;fill:none\"/><text x=\"29\" y=\"54\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">31</text><text x=\"29\" y=\"81\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">30</text><text x=\"29\" y=\"108\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">29</text><text x=\"29\" y=\"135\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">28</text><text x=\"29\" y=\"162\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">27</text><text x=\"29\" y=\"189\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">26</text><text x=\"29\" y=\"216\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">25</text><text x=\"29\" y=\"243\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">24</text><text x=\"29\" y=\"270\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">23</text><text x=\"29\" y=\"297\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">22</text><text x=\"29\" y=\"324\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">21</text><text x=\"29\" y=\"351\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">20</text><path d=\"M 25 56\nL 580 56\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 83\nL 580 83\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 110\nL 580 110\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 137\nL 580 137\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 164\nL 580 164\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 191\nL 580 191\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 219\nL 580 219\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 246\nL 580 246\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 273\nL 580 273\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 300\nL 580 300\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 327\nL 580 327\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 29 355\nL 580 355\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 29 360\nL 29 355\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 304 360\nL 304 355\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 580 360\nL 580 355\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"166\" y=\"365\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">A</text><text x=\"442\" y=\"365\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">B</text><path d=\"M 166 355\nL 442 84\" style=\"stroke-width:2;stroke:rgb(84,112,198);fill:none\"/><circle cx=\"166\" cy=\"355\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:white\"/><circle cx=\"442\" cy=\"84\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:white\"/></svg>",
		},
		{
			name: "yaxis_line_show",
			jsonData: `{
				"yAxis": {
					"axisLine": { "show": true, "lineStyle": { "color": "#ff0000", "opacity": 0.8 } }
				},
				"series": [{ "data": [5, 15], "type": "bar" }]
			}`,
			expected: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><path d=\"M 47 20\nL 47 354\" style=\"stroke-width:1;stroke:rgba(255,0,0,0.8);fill:none\"/><path d=\"M 42 20\nL 47 20\" style=\"stroke-width:1;stroke:rgba(255,0,0,0.8);fill:none\"/><path d=\"M 42 50\nL 47 50\" style=\"stroke-width:1;stroke:rgba(255,0,0,0.8);fill:none\"/><path d=\"M 42 80\nL 47 80\" style=\"stroke-width:1;stroke:rgba(255,0,0,0.8);fill:none\"/><path d=\"M 42 111\nL 47 111\" style=\"stroke-width:1;stroke:rgba(255,0,0,0.8);fill:none\"/><path d=\"M 42 141\nL 47 141\" style=\"stroke-width:1;stroke:rgba(255,0,0,0.8);fill:none\"/><path d=\"M 42 171\nL 47 171\" style=\"stroke-width:1;stroke:rgba(255,0,0,0.8);fill:none\"/><path d=\"M 42 202\nL 47 202\" style=\"stroke-width:1;stroke:rgba(255,0,0,0.8);fill:none\"/><path d=\"M 42 232\nL 47 232\" style=\"stroke-width:1;stroke:rgba(255,0,0,0.8);fill:none\"/><path d=\"M 42 262\nL 47 262\" style=\"stroke-width:1;stroke:rgba(255,0,0,0.8);fill:none\"/><path d=\"M 42 293\nL 47 293\" style=\"stroke-width:1;stroke:rgba(255,0,0,0.8);fill:none\"/><path d=\"M 42 323\nL 47 323\" style=\"stroke-width:1;stroke:rgba(255,0,0,0.8);fill:none\"/><path d=\"M 42 354\nL 47 354\" style=\"stroke-width:1;stroke:rgba(255,0,0,0.8);fill:none\"/><text x=\"19\" y=\"26\" style=\"stroke:none;fill:rgba(255,0,0,0.8);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">16</text><text x=\"19\" y=\"56\" style=\"stroke:none;fill:rgba(255,0,0,0.8);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">15</text><text x=\"19\" y=\"86\" style=\"stroke:none;fill:rgba(255,0,0,0.8);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">14</text><text x=\"19\" y=\"116\" style=\"stroke:none;fill:rgba(255,0,0,0.8);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">13</text><text x=\"19\" y=\"146\" style=\"stroke:none;fill:rgba(255,0,0,0.8);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">12</text><text x=\"19\" y=\"176\" style=\"stroke:none;fill:rgba(255,0,0,0.8);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">11</text><text x=\"19\" y=\"207\" style=\"stroke:none;fill:rgba(255,0,0,0.8);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">10</text><text x=\"28\" y=\"237\" style=\"stroke:none;fill:rgba(255,0,0,0.8);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">9</text><text x=\"28\" y=\"267\" style=\"stroke:none;fill:rgba(255,0,0,0.8);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">8</text><text x=\"28\" y=\"297\" style=\"stroke:none;fill:rgba(255,0,0,0.8);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">7</text><text x=\"28\" y=\"327\" style=\"stroke:none;fill:rgba(255,0,0,0.8);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">6</text><text x=\"28\" y=\"358\" style=\"stroke:none;fill:rgba(255,0,0,0.8);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">5</text><path d=\"M 43 20\nL 580 20\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 50\nL 580 50\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 80\nL 580 80\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 111\nL 580 111\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 141\nL 580 141\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 171\nL 580 171\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 202\nL 580 202\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 232\nL 580 232\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 262\nL 580 262\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 293\nL 580 293\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 323\nL 580 323\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 48 354\nL 580 354\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 48 359\nL 48 354\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 314 359\nL 314 354\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 580 359\nL 580 354\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"177\" y=\"380\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"443\" y=\"380\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><path d=\"M 58 354\nL 304 354\nL 304 353\nL 58 353\nL 58 354\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 324 51\nL 570 51\nL 570 353\nL 324 353\nL 324 51\" style=\"stroke:none;fill:rgb(84,112,198)\"/></svg>",
		},
		{
			name: "two_yaxis",
			jsonData: `{
				"xAxis": { "type": "category", "data": ["Jan", "Feb"] },
				"yAxis": [
					{ "type": "value", "axisLabel": { "color": "#ff0000" } },
					{ "type": "value", "axisLabel": { "color": "#0000ff" } }
				],
				"series": [
					{ "data": [30, 60], "type": "bar", "yAxisIndex": 0 },
					{ "data": [1.5, 3.2], "type": "line", "yAxisIndex": 1 }
				]
			}`,
			expected: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"559\" y=\"26\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3.5</text><text x=\"559\" y=\"92\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"559\" y=\"158\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2.5</text><text x=\"559\" y=\"225\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"559\" y=\"291\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1.5</text><text x=\"559\" y=\"358\" style=\"stroke:none;fill:blue;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"19\" y=\"26\" style=\"stroke:none;fill:red;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">65</text><text x=\"19\" y=\"73\" style=\"stroke:none;fill:red;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">60</text><text x=\"19\" y=\"120\" style=\"stroke:none;fill:red;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">55</text><text x=\"19\" y=\"168\" style=\"stroke:none;fill:red;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">50</text><text x=\"19\" y=\"215\" style=\"stroke:none;fill:red;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">45</text><text x=\"19\" y=\"263\" style=\"stroke:none;fill:red;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">40</text><text x=\"19\" y=\"310\" style=\"stroke:none;fill:red;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">35</text><text x=\"19\" y=\"358\" style=\"stroke:none;fill:red;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">30</text><path d=\"M 43 20\nL 549 20\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 67\nL 549 67\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 115\nL 549 115\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 163\nL 549 163\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 210\nL 549 210\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 258\nL 549 258\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 43 306\nL 549 306\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 47 354\nL 549 354\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 47 359\nL 47 354\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 298 359\nL 298 354\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 549 359\nL 549 354\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"159\" y=\"380\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Jan</text><text x=\"410\" y=\"380\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Feb</text><path d=\"M 57 354\nL 288 354\nL 288 353\nL 57 353\nL 57 354\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 308 68\nL 539 68\nL 539 353\nL 308 353\nL 308 68\" style=\"stroke:none;fill:rgb(84,112,198)\"/><path d=\"M 172 288\nL 423 61\" style=\"stroke-width:2;stroke:rgb(145,204,117);fill:none\"/><circle cx=\"172\" cy=\"288\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:white\"/><circle cx=\"423\" cy=\"61\" r=\"2\" style=\"stroke-width:1;stroke:rgb(145,204,117);fill:white\"/></svg>",
		},
		{
			name: "background_color",
			jsonData: `{
				"backgroundColor": "#e0e0e0",
				"xAxis": { "axisLabel": { "show": false }, "type": "category", "data": ["A", "B"] },
				"yAxis": { "axisLabel": { "show": false }, "type": "value" },
				"series": [{ "data": [40, 70], "type": "line" }]
			}`,
			expected: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:rgb(224,224,224)\"/><text x=\"29\" y=\"18\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">75</text><text x=\"29\" y=\"65\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">70</text><text x=\"29\" y=\"113\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">65</text><text x=\"29\" y=\"160\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">60</text><text x=\"29\" y=\"208\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">55</text><text x=\"29\" y=\"255\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">50</text><text x=\"29\" y=\"303\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">45</text><text x=\"29\" y=\"351\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">40</text><path d=\"M 25 20\nL 580 20\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 67\nL 580 67\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 115\nL 580 115\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 163\nL 580 163\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 211\nL 580 211\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 259\nL 580 259\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 307\nL 580 307\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 29 355\nL 580 355\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 29 360\nL 29 355\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 304 360\nL 304 355\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 580 360\nL 580 355\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"166\" y=\"365\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">A</text><text x=\"442\" y=\"365\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">B</text><path d=\"M 166 355\nL 442 68\" style=\"stroke-width:2;stroke:rgb(84,112,198);fill:none\"/><circle cx=\"166\" cy=\"355\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(224,224,224)\"/><circle cx=\"442\" cy=\"68\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:rgb(224,224,224)\"/></svg>",
		},
		{
			name: "title_border",
			jsonData: `{
				"title": {
					"text": "Title",
					"borderColor": "#00ff00"
				},
				"xAxis": { "axisLabel": { "show": false }, "type": "category", "data": ["A", "B"] },
				"yAxis": { "axisLabel": { "show": false }, "type": "value" },
				"series": [{ "data": [40, 70], "type": "line" }]
			}`,
			expected: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path d=\"M 0 0\nL 600 0\nL 600 400\nL 0 400\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"20\" y=\"36\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Title</text><path d=\"M 10 46\nL 10 10\nL 61 10\nL 61 46\nL 10 46\" style=\"stroke-width:2;stroke:lime;fill:none\"/><text x=\"29\" y=\"49\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">75</text><text x=\"29\" y=\"92\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">70</text><text x=\"29\" y=\"135\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">65</text><text x=\"29\" y=\"178\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">60</text><text x=\"29\" y=\"221\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">55</text><text x=\"29\" y=\"264\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">50</text><text x=\"29\" y=\"307\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">45</text><text x=\"29\" y=\"351\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">40</text><path d=\"M 25 51\nL 580 51\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 94\nL 580 94\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 137\nL 580 137\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 181\nL 580 181\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 224\nL 580 224\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 268\nL 580 268\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 25 311\nL 580 311\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 29 355\nL 580 355\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 29 360\nL 29 355\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 304 360\nL 304 355\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 580 360\nL 580 355\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"166\" y=\"365\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">A</text><text x=\"442\" y=\"365\" style=\"stroke:none;fill:none;font-size:15.3px;font-family:'Roboto Medium',sans-serif\">B</text><path d=\"M 166 355\nL 442 95\" style=\"stroke-width:2;stroke:rgb(84,112,198);fill:none\"/><circle cx=\"166\" cy=\"355\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:white\"/><circle cx=\"442\" cy=\"95\" r=\"2\" style=\"stroke-width:1;stroke:rgb(84,112,198);fill:white\"/></svg>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := RenderEChartsToSVG(tt.jsonData)
			require.NoError(t, err)
			assertEqualSVG(t, tt.expected, data)
		})
	}
}
