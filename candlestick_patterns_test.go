package charts

import (
	"strconv"
	"strings"
	"testing"

	"github.com/go-analyze/bulk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDojiPattern(t *testing.T) {
	t.Parallel()

	// Valid doji: open ≈ close
	doji := OHLCData{Open: 100, High: 105, Low: 95, Close: 100.1}
	data := []OHLCData{doji}
	for _, tt := range []struct {
		name      string
		threshold float64
		expected  bool
	}{
		{"low", 0.009, false},
		{"default", 0.01, true},
		{"high", 0.011, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, detectDojiAt(data, 0, CandlestickPatternConfig{DojiThreshold: tt.threshold}))
		})
	}

	// Invalid: body too large
	notDoji := OHLCData{Open: 100, High: 105, Low: 95, Close: 103}
	data = []OHLCData{notDoji}
	assert.False(t, detectDojiAt(data, 0, CandlestickPatternConfig{DojiThreshold: 0.01}))

	// Invalid: invalid OHLC
	invalidOHLC := OHLCData{Open: 100, High: 95, Low: 105, Close: 98}
	data = []OHLCData{invalidOHLC}
	assert.False(t, detectDojiAt(data, 0, CandlestickPatternConfig{DojiThreshold: 0.01}))
}

func TestHammerPattern(t *testing.T) {
	t.Parallel()

	// Valid hammer: long lower shadow, small body at top
	hammer := OHLCData{Open: 105, High: 107, Low: 95, Close: 106}
	data := []OHLCData{hammer}
	for _, tt := range []struct {
		name     string
		ratio    float64
		expected bool
	}{
		{"low", 1.0, true},
		{"default", 2.0, true},
		{"high", 11.1, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, detectHammerAt(data, 0, CandlestickPatternConfig{ShadowRatio: tt.ratio}))
		})
	}

	// Invalid: short lower shadow
	notHammer := OHLCData{Open: 105, High: 107, Low: 104, Close: 106}
	data = []OHLCData{notHammer}
	assert.False(t, detectHammerAt(data, 0, CandlestickPatternConfig{ShadowRatio: 2.0}))

	// Invalid: long upper shadow
	notHammer2 := OHLCData{Open: 95, High: 107, Low: 94, Close: 96}
	data = []OHLCData{notHammer2}
	assert.False(t, detectHammerAt(data, 0, CandlestickPatternConfig{ShadowRatio: 2.0}))
}

func TestInvertedHammerPattern(t *testing.T) {
	t.Parallel()

	// Valid inverted hammer: long upper shadow, small body at bottom
	invertedHammer := OHLCData{Open: 95, High: 107, Low: 94, Close: 96}
	for _, tt := range []struct {
		name     string
		ratio    float64
		expected bool
	}{
		{"low", 1.0, true},
		{"default", 2.0, true},
		{"high", 11.1, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, detectInvertedHammerAt([]OHLCData{invertedHammer}, 0, CandlestickPatternConfig{ShadowRatio: tt.ratio}))
		})
	}

	// Invalid: short upper shadow
	notInvertedHammer := OHLCData{Open: 95, High: 97, Low: 94, Close: 96}
	assert.False(t, detectInvertedHammerAt([]OHLCData{notInvertedHammer}, 0, CandlestickPatternConfig{ShadowRatio: 2.0}))
}

func TestEngulfingPattern(t *testing.T) {
	t.Parallel()

	prevBearish := OHLCData{Open: 110, High: 112, Low: 105, Close: 106}
	currentBullish := OHLCData{Open: 104, High: 115, Low: 103, Close: 114}
	for _, tt := range []struct {
		name     string
		size     float64
		expected bool
	}{
		{"low", 0.5, true},
		{"default", 0.8, true},
		{"high", 2.6, false},
	} {
		t.Run("bullish_"+tt.name, func(t *testing.T) {
			detected := detectBullishEngulfingAt([]OHLCData{prevBearish, currentBullish}, 1, CandlestickPatternConfig{EngulfingMinSize: tt.size})
			assert.Equal(t, tt.expected, detected)
		})
	}
	assert.False(t, detectBearishEngulfingAt([]OHLCData{prevBearish, currentBullish}, 1, CandlestickPatternConfig{EngulfingMinSize: 0.8}))

	prevBullish := OHLCData{Open: 106, High: 112, Low: 105, Close: 110}
	currentBearish := OHLCData{Open: 114, High: 115, Low: 103, Close: 104}

	for _, tt := range []struct {
		name     string
		size     float64
		expected bool
	}{
		{"low", 0.5, true},
		{"default", 0.8, true},
		{"high", 2.6, false},
	} {
		t.Run("bearish_"+tt.name, func(t *testing.T) {
			detected := detectBearishEngulfingAt([]OHLCData{prevBullish, currentBearish}, 1, CandlestickPatternConfig{EngulfingMinSize: tt.size})
			assert.Equal(t, tt.expected, detected)
		})
	}
	assert.False(t, detectBullishEngulfingAt([]OHLCData{prevBullish, currentBearish}, 1, CandlestickPatternConfig{EngulfingMinSize: 0.8}))

	// Test non-engulfing
	nonEngulfing := OHLCData{Open: 107, High: 109, Low: 106, Close: 108}
	assert.False(t, detectBullishEngulfingAt([]OHLCData{prevBullish, nonEngulfing}, 1, CandlestickPatternConfig{EngulfingMinSize: 0.8}))
	assert.False(t, detectBearishEngulfingAt([]OHLCData{prevBullish, nonEngulfing}, 1, CandlestickPatternConfig{EngulfingMinSize: 0.8}))
}

func TestShootingStarPattern(t *testing.T) {
	t.Parallel()

	// Valid shooting star: small body at bottom, long upper shadow
	shootingStar := OHLCData{Open: 106, High: 125, Low: 105, Close: 107}
	for _, tt := range []struct {
		name     string
		ratio    float64
		expected bool
	}{
		{"low", 1.0, true},
		{"default", 2.0, true},
		{"high", 18.1, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, detectShootingStarAt([]OHLCData{shootingStar}, 0, CandlestickPatternConfig{ShadowRatio: tt.ratio}))
		})
	}

	// Invalid: body not near bottom
	notShootingStar := OHLCData{Open: 115, High: 125, Low: 105, Close: 117}
	assert.False(t, detectShootingStarAt([]OHLCData{notShootingStar}, 0, CandlestickPatternConfig{ShadowRatio: 2.0}))

	// Invalid: upper shadow too short
	shortShadow := OHLCData{Open: 106, High: 110, Low: 105, Close: 107}
	assert.False(t, detectShootingStarAt([]OHLCData{shortShadow}, 0, CandlestickPatternConfig{ShadowRatio: 2.0}))
}

func TestGravestoneDojiPattern(t *testing.T) {
	t.Parallel()

	gravestoneDoji := OHLCData{Open: 108, High: 120, Low: 107, Close: 108.1}
	for _, tt := range []struct {
		name      string
		threshold float64
		shadow    float64
		expected  bool
	}{
		{"low_threshold", 0.004, 2.0, false},
		{"default", 0.01, 2.0, true},
		{"high_shadow", 0.01, 200, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			opt := CandlestickPatternConfig{DojiThreshold: tt.threshold, ShadowRatio: tt.shadow}
			assert.Equal(t, tt.expected, detectGravestoneDojiAt([]OHLCData{gravestoneDoji}, 0, opt))
		})
	}

	// Invalid: not a doji (body too large)
	notDoji := OHLCData{Open: 108, High: 120, Low: 107, Close: 115}
	assert.False(t, detectGravestoneDojiAt([]OHLCData{notDoji}, 0, CandlestickPatternConfig{DojiThreshold: 0.01, ShadowRatio: 2.0}))

	// Invalid: doji but no long upper shadow
	dojiNoShadow := OHLCData{Open: 108, High: 109, Low: 107, Close: 108.1}
	assert.False(t, detectGravestoneDojiAt([]OHLCData{dojiNoShadow}, 0, CandlestickPatternConfig{DojiThreshold: 0.01, ShadowRatio: 2.0}))
}

func TestDragonflyDojiPattern(t *testing.T) {
	t.Parallel()

	dragonflyDoji := OHLCData{Open: 109, High: 110, Low: 90, Close: 108.9}
	for _, tt := range []struct {
		name      string
		threshold float64
		shadow    float64
		expected  bool
	}{
		{"low_threshold", 0.004, 2.0, false},
		{"default", 0.01, 2.0, true},
		{"high_shadow", 0.01, 200, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			opt := CandlestickPatternConfig{DojiThreshold: tt.threshold, ShadowRatio: tt.shadow}
			assert.Equal(t, tt.expected, detectDragonflyDojiAt([]OHLCData{dragonflyDoji}, 0, opt))
		})
	}

	// Invalid: not a doji
	notDoji := OHLCData{Open: 109, High: 110, Low: 90, Close: 102}
	assert.False(t, detectDragonflyDojiAt([]OHLCData{notDoji}, 0, CandlestickPatternConfig{DojiThreshold: 0.01, ShadowRatio: 2.0}))

	// Invalid: doji but no long lower shadow
	dojiNoShadow := OHLCData{Open: 109, High: 110, Low: 108, Close: 108.9}
	assert.False(t, detectDragonflyDojiAt([]OHLCData{dojiNoShadow}, 0, CandlestickPatternConfig{DojiThreshold: 0.01, ShadowRatio: 2.0}))
}

func TestMorningStarPattern(t *testing.T) {
	t.Parallel()

	opt := CandlestickPatternConfig{}

	// Valid morning star pattern
	first := OHLCData{Open: 120, High: 125, Low: 105, Close: 108}  // Large bearish
	second := OHLCData{Open: 102, High: 104, Low: 100, Close: 103} // Small body, gap down
	third := OHLCData{Open: 108, High: 125, Low: 106, Close: 122}  // Large bullish, gap up

	assert.True(t, detectMorningStarAt([]OHLCData{first, second, third}, 2, opt))

	// Invalid: first candle not bearish
	invalidFirst := OHLCData{Open: 108, High: 125, Low: 105, Close: 120} // Bullish
	assert.False(t, detectMorningStarAt([]OHLCData{invalidFirst, second, third}, 2, opt))

	// Invalid: no gap down between first and second
	noGapSecond := OHLCData{Open: 109, High: 111, Low: 107, Close: 110} // No gap
	assert.False(t, detectMorningStarAt([]OHLCData{first, noGapSecond, third}, 2, opt))

	// Invalid: third candle not bullish
	invalidThird := OHLCData{Open: 108, High: 110, Low: 105, Close: 107} // Bearish
	assert.False(t, detectMorningStarAt([]OHLCData{first, second, invalidThird}, 2, opt))
}

func TestEveningStarPattern(t *testing.T) {
	t.Parallel()

	opt := CandlestickPatternConfig{}

	// Valid evening star pattern
	first := OHLCData{Open: 122, High: 140, Low: 120, Close: 138}  // Large bullish
	second := OHLCData{Open: 142, High: 144, Low: 140, Close: 143} // Small body, gap up
	third := OHLCData{Open: 138, High: 140, Low: 115, Close: 118}  // Large bearish, gap down

	assert.True(t, detectEveningStarAt([]OHLCData{first, second, third}, 2, opt))

	// Invalid: first candle not bullish
	invalidFirst := OHLCData{Open: 138, High: 140, Low: 120, Close: 122} // Bearish
	assert.False(t, detectEveningStarAt([]OHLCData{invalidFirst, second, third}, 2, opt))

	// Invalid: no gap up between first and second
	noGapSecond := OHLCData{Open: 136, High: 140, Low: 134, Close: 139} // No gap
	assert.False(t, detectEveningStarAt([]OHLCData{first, noGapSecond, third}, 2, opt))

	// Invalid: third candle not bearish
	invalidThird := OHLCData{Open: 138, High: 145, Low: 135, Close: 142} // Bullish
	assert.False(t, detectEveningStarAt([]OHLCData{first, second, invalidThird}, 2, opt))
}

func newCandlestickWithPatterns(data []OHLCData, options ...CandlestickPatternConfig) CandlestickSeries {
	// Start with defaults and override with provided options
	config := &CandlestickPatternConfig{
		PreferPatternLabels: true,
		EnabledPatterns:     (&CandlestickPatternConfig{}).WithPatternsAll().EnabledPatterns,
		DojiThreshold:       0.001,
		ShadowTolerance:     0.01,
		ShadowRatio:         2.0,
		EngulfingMinSize:    0.8,
	}
	if len(options) > 0 {
		// Merge provided options with defaults
		opt := options[0]
		if opt.DojiThreshold > 0 {
			config.DojiThreshold = opt.DojiThreshold
		}
		if opt.ShadowRatio > 0 {
			config.ShadowRatio = opt.ShadowRatio
		}
		if opt.EngulfingMinSize > 0 {
			config.EngulfingMinSize = opt.EngulfingMinSize
		}
		if opt.ShadowTolerance > 0 {
			config.ShadowTolerance = opt.ShadowTolerance
		}
	}

	return CandlestickSeries{
		Data:          data,
		PatternConfig: config,
	}
}

func makePatternChartOption(data []OHLCData, config CandlestickPatternConfig) CandlestickChartOption {
	series := newCandlestickWithPatterns(data, config)
	labels := make([]string, len(data))
	for i := range labels {
		labels[i] = strconv.Itoa(i + 1)
	}
	return CandlestickChartOption{
		XAxis:      XAxisOption{Labels: labels},
		YAxis:      make([]YAxisOption, 1),
		SeriesList: CandlestickSeriesList{series},
		Padding:    NewBoxEqual(10),
	}
}

func TestMarubozuPattern(t *testing.T) {
	t.Parallel()

	// Bullish Marubozu - no shadows
	bullishMarubozu := OHLCData{Open: 100, High: 120, Low: 100, Close: 120}
	for _, tt := range []struct {
		tol      float64
		expected bool
	}{
		{0.005, true},
		{0.01, true},
		{0.02, true},
	} {
		t.Run("bullish_tol_"+strconv.FormatFloat(tt.tol, 'f', 3, 64), func(t *testing.T) {
			detected := detectBullishMarubozuAt([]OHLCData{bullishMarubozu}, 0, CandlestickPatternConfig{ShadowTolerance: tt.tol})
			assert.Equal(t, tt.expected, detected)
		})
	}
	assert.False(t, detectBearishMarubozuAt([]OHLCData{bullishMarubozu}, 0, CandlestickPatternConfig{ShadowTolerance: 0.01}))

	// Bearish Marubozu - no shadows
	bearishMarubozu := OHLCData{Open: 120, High: 120, Low: 100, Close: 100}
	for _, tt := range []struct {
		tol      float64
		expected bool
	}{
		{0.005, true},
		{0.01, true},
		{0.02, true},
	} {
		t.Run("bearish_tol_"+strconv.FormatFloat(tt.tol, 'f', 3, 64), func(t *testing.T) {
			detected := detectBearishMarubozuAt([]OHLCData{bearishMarubozu}, 0, CandlestickPatternConfig{ShadowTolerance: tt.tol})
			assert.Equal(t, tt.expected, detected)
		})
	}
	assert.False(t, detectBullishMarubozuAt([]OHLCData{bearishMarubozu}, 0, CandlestickPatternConfig{ShadowTolerance: 0.01}))

	// Not a marubozu - has significant shadows
	notMarubozu := OHLCData{Open: 105, High: 125, Low: 95, Close: 115}
	assert.False(t, detectBullishMarubozuAt([]OHLCData{notMarubozu}, 0, CandlestickPatternConfig{ShadowTolerance: 0.01}))
	assert.False(t, detectBearishMarubozuAt([]OHLCData{notMarubozu}, 0, CandlestickPatternConfig{ShadowTolerance: 0.01}))
	assert.True(t, detectBullishMarubozuAt([]OHLCData{notMarubozu}, 0, CandlestickPatternConfig{ShadowTolerance: 0.7}))
}

func TestPiercingLinePattern(t *testing.T) {
	t.Parallel()

	// Classic piercing line - bearish then bullish with gap down and close above midpoint
	prev := OHLCData{Open: 120, High: 120, Low: 110, Close: 110}    // Bearish
	current := OHLCData{Open: 108, High: 118, Low: 108, Close: 116} // Bullish, opens below prev low, closes above midpoint (115)
	detected := detectPiercingLineAt([]OHLCData{prev, current}, 1, CandlestickPatternConfig{})
	assert.True(t, detected)

	// Not piercing line - current closes below midpoint
	current = OHLCData{Open: 108, High: 114, Low: 108, Close: 112}
	detected = detectPiercingLineAt([]OHLCData{prev, current}, 1, CandlestickPatternConfig{})
	assert.False(t, detected)
}

func TestDarkCloudCoverPattern(t *testing.T) {
	t.Parallel()

	// Classic dark cloud cover - bullish then bearish with gap up and close below midpoint
	prev := OHLCData{Open: 110, High: 120, Low: 110, Close: 120}    // Bullish
	current := OHLCData{Open: 122, High: 122, Low: 112, Close: 114} // Bearish, opens above prev high, closes below midpoint (115)
	detected := detectDarkCloudCoverAt([]OHLCData{prev, current}, 1, CandlestickPatternConfig{})
	assert.True(t, detected)

	// Not dark cloud cover - current closes above midpoint
	current = OHLCData{Open: 122, High: 122, Low: 118, Close: 118}
	detected = detectDarkCloudCoverAt([]OHLCData{prev, current}, 1, CandlestickPatternConfig{})
	assert.False(t, detected)
}

func TestPatternValidation(t *testing.T) {
	t.Parallel()

	// Test with invalid OHLC data
	invalidOHLC := OHLCData{Open: 100, High: 95, Low: 105, Close: 98} // High < Low

	assert.False(t, detectDojiAt([]OHLCData{invalidOHLC}, 0, CandlestickPatternConfig{DojiThreshold: 0.01}))
	assert.False(t, detectHammerAt([]OHLCData{invalidOHLC}, 0, CandlestickPatternConfig{ShadowRatio: 2.0}))
	assert.False(t, detectShootingStarAt([]OHLCData{invalidOHLC}, 0, CandlestickPatternConfig{ShadowRatio: 2.0}))

	// Test three-candle patterns with invalid data
	validOHLC := OHLCData{Open: 100, High: 110, Low: 95, Close: 105}
	opt := CandlestickPatternConfig{}

	assert.False(t, detectMorningStarAt([]OHLCData{invalidOHLC, validOHLC, validOHLC}, 2, opt))
	assert.False(t, detectMorningStarAt([]OHLCData{validOHLC, invalidOHLC, validOHLC}, 2, opt))
	assert.False(t, detectMorningStarAt([]OHLCData{validOHLC, validOHLC, invalidOHLC}, 2, opt))

	assert.False(t, detectEveningStarAt([]OHLCData{invalidOHLC, validOHLC, validOHLC}, 2, opt))
	assert.False(t, detectEveningStarAt([]OHLCData{validOHLC, invalidOHLC, validOHLC}, 2, opt))
	assert.False(t, detectEveningStarAt([]OHLCData{validOHLC, validOHLC, invalidOHLC}, 2, opt))
}

func TestPatternScanningComprehensive(t *testing.T) {
	t.Parallel()

	data := []OHLCData{
		// Index 0: Normal candle
		{Open: 100, High: 110, Low: 95, Close: 105},
		// Index 1: Doji
		{Open: 105, High: 108, Low: 102, Close: 105.05},
		// Index 2: Hammer
		{Open: 108, High: 109, Low: 98, Close: 107},
		// Index 3: Shooting Star
		{Open: 106, High: 125, Low: 105, Close: 107},
		// Index 4: Gravestone Doji
		{Open: 108, High: 120, Low: 107, Close: 108.1},
		// Index 5: Dragonfly Doji
		{Open: 109, High: 110, Low: 90, Close: 108.9},
		// Index 6-8: Morning Star sequence
		{Open: 120, High: 125, Low: 105, Close: 108}, // 6: Large bearish
		{Open: 102, High: 104, Low: 100, Close: 103}, // 7: Small body, gap down
		{Open: 108, High: 125, Low: 106, Close: 122}, // 8: Large bullish, gap up
		// Index 9-11: Evening Star sequence
		{Open: 122, High: 140, Low: 120, Close: 138}, // 9: Large bullish
		{Open: 142, High: 144, Low: 140, Close: 143}, // 10: Small body, gap up
		{Open: 138, High: 140, Low: 115, Close: 118}, // 11: Large bearish, gap down
		// Index 12: Bullish Marubozu (no shadows)
		{Open: 120, High: 135, Low: 120, Close: 135},
		// Index 13: Bearish Marubozu (no shadows)
		{Open: 135, High: 135, Low: 115, Close: 115},
		// Index 14: Spinning Top (small body, long shadows)
		{Open: 118, High: 125, Low: 110, Close: 119},
		// Index 15: Setup for Piercing Line - bearish candle
		{Open: 120, High: 121, Low: 115, Close: 115},
		// Index 16: Piercing Line - bullish candle opening below prev low, closing above midpoint
		{Open: 112, High: 119, Low: 112, Close: 118}, // Opens below 115, closes above midpoint (117.5)
		// Index 17: Setup for Dark Cloud Cover - bullish candle
		{Open: 118, High: 125, Low: 118, Close: 125},
		// Index 18: Dark Cloud Cover - bearish candle opening above prev high, closing below midpoint
		{Open: 127, High: 127, Low: 120, Close: 121}, // Opens above 125, closes below midpoint (121.5)
		// Index 19: Setup for Tweezer Bottom - bearish with low at 100
		{Open: 125, High: 126, Low: 100, Close: 102},
		// Index 20: Tweezer Bottom - bullish with same low at 100
		{Open: 102, High: 108, Low: 100, Close: 107},
		// Index 21-23: Three White Soldiers sequence
		{Open: 110, High: 115, Low: 109, Close: 114}, // 21: First soldier
		{Open: 113, High: 118, Low: 112, Close: 117}, // 22: Second soldier
		{Open: 116, High: 121, Low: 115, Close: 120}, // 23: Third soldier
		// Index 24-26: Three Black Crows sequence
		{Open: 120, High: 121, Low: 115, Close: 116}, // 24: First crow
		{Open: 117, High: 118, Low: 112, Close: 113}, // 25: Second crow
		{Open: 114, High: 115, Low: 108, Close: 109}, // 26: Third crow
	}

	opt := (&CandlestickPatternConfig{}).WithPatternsAll()
	opt.DojiThreshold = 0.01
	opt.ShadowRatio = 2.0
	opt.EngulfingMinSize = 0.8
	indexPatterns := scanForCandlestickPatterns(data, *opt)

	// Verify specific patterns were detected
	patternsByIndex := make(map[int][]string)
	uniquePatterns := make(map[string]bool)
	for index, patterns := range indexPatterns {
		for _, pattern := range patterns {
			patternsByIndex[index] = append(patternsByIndex[index], pattern.PatternType)
			uniquePatterns[pattern.PatternType] = true
		}
	}

	// Check expected patterns
	assert.Len(t, uniquePatterns, 13)
	assert.Contains(t, patternsByIndex[1], "doji")
	assert.Contains(t, patternsByIndex[2], "hammer")
	assert.Contains(t, patternsByIndex[3], "shooting_star")
	assert.Contains(t, patternsByIndex[4], "gravestone_doji")
	assert.Contains(t, patternsByIndex[5], "dragonfly_doji")
	assert.Contains(t, patternsByIndex[8], "morning_star")
	assert.Contains(t, patternsByIndex[11], "evening_star")
	assert.Contains(t, patternsByIndex[12], "marubozu_bull")
	assert.Contains(t, patternsByIndex[13], "marubozu_bear")
	assert.Contains(t, patternsByIndex[16], "piercing_line")
	assert.Contains(t, patternsByIndex[18], "dark_cloud_cover")
}

func TestCandlestickPatternSets(t *testing.T) {
	t.Parallel()

	t.Run("all", func(t *testing.T) {
		config := (&CandlestickPatternConfig{}).WithPatternsAll()

		assert.Contains(t, config.EnabledPatterns, "doji")
		assert.Contains(t, config.EnabledPatterns, "hammer")
		assert.Len(t, config.EnabledPatterns, 14)
	})

	t.Run("core", func(t *testing.T) {
		config := (&CandlestickPatternConfig{}).WithPatternsCore()

		assert.Contains(t, config.EnabledPatterns, "engulfing_bull")
		assert.Contains(t, config.EnabledPatterns, "hammer")
		assert.Len(t, config.EnabledPatterns, 6)
	})

	t.Run("bullish", func(t *testing.T) {
		config := (&CandlestickPatternConfig{}).WithPatternsBullish()

		assert.Contains(t, config.EnabledPatterns, "hammer")
		assert.NotContains(t, config.EnabledPatterns, "shooting_star")
		assert.Len(t, config.EnabledPatterns, 7)
	})

	t.Run("bearish", func(t *testing.T) {
		config := (&CandlestickPatternConfig{}).WithPatternsBearish()

		assert.Contains(t, config.EnabledPatterns, "shooting_star")
		assert.NotContains(t, config.EnabledPatterns, "hammer")
		assert.Len(t, config.EnabledPatterns, 6)
	})

	t.Run("reversal", func(t *testing.T) {
		config := (&CandlestickPatternConfig{}).WithPatternsReversal()

		assert.Contains(t, config.EnabledPatterns, "hammer")
		assert.NotContains(t, config.EnabledPatterns, "marubozu_bull")
		assert.Len(t, config.EnabledPatterns, 10)
	})

	t.Run("trend", func(t *testing.T) {
		config := (&CandlestickPatternConfig{}).WithPatternsTrend()

		assert.Contains(t, config.EnabledPatterns, "marubozu_bull")
		assert.NotContains(t, config.EnabledPatterns, "hammer")
		assert.Len(t, config.EnabledPatterns, 2)
	})
}

func TestPatternFormatterCustom(t *testing.T) {
	t.Parallel()

	// Data with a clear Doji at index 1
	data := []OHLCData{
		{Open: 100, High: 110, Low: 95, Close: 105},
		{Open: 105, High: 107, Low: 103, Close: 105}, // Doji
		{Open: 105, High: 112, Low: 98, Close: 108},
	}

	// Custom formatter that prefixes with PF: and joins all detected pattern types
	customFormatter := func(patterns []PatternDetectionResult, seriesName string, value float64) (string, *LabelStyle) {
		if len(patterns) == 0 {
			return "", nil
		}
		names := make([]string, len(patterns))
		for i, p := range patterns {
			names[i] = p.PatternType
		}
		return "PF:" + strings.Join(names, "+"), &LabelStyle{FontStyle: FontStyle{FontColor: ColorGray}}
	}

	mkOpt := func(cfg CandlestickPatternConfig, userLabel bool) CandlestickChartOption {
		series := CandlestickSeries{
			Data: data,
			Label: SeriesLabel{
				Show: Ptr(userLabel),
				LabelFormatter: func(index int, name string, val float64) (string, *LabelStyle) {
					return "UserLabel", nil
				},
			},
			PatternConfig: &cfg,
		}
		labels := []string{"1", "2", "3"}
		return CandlestickChartOption{
			XAxis:      XAxisOption{Labels: labels},
			YAxis:      make([]YAxisOption, 1),
			SeriesList: CandlestickSeriesList{series},
			Padding:    NewBoxEqual(10),
		}
	}

	t.Run("pattern_priority_mode", func(t *testing.T) {
		cfg := CandlestickPatternConfig{
			PreferPatternLabels: true,
			EnabledPatterns:     []string{"doji"},
			DojiThreshold:       0.001,
			PatternFormatter:    customFormatter,
		}
		opt := mkOpt(cfg, true)
		p := NewPainter(PainterOptions{OutputFormat: ChartOutputSVG, Width: 800, Height: 600})
		require.NoError(t, p.CandlestickChart(opt))
		svg, err := p.Bytes()
		require.NoError(t, err)
		s := string(svg)
		// With pattern priority, pattern label should be shown at index 1 where Doji is detected
		assert.Contains(t, s, "PF:"+"doji")
		// User labels should still appear at indices 0 and 2 where no patterns are detected
		assert.Contains(t, s, "UserLabel")
	})

	t.Run("user_priority_mode", func(t *testing.T) {
		cfg := CandlestickPatternConfig{
			PreferPatternLabels: false,
			EnabledPatterns:     []string{"doji"},
			DojiThreshold:       0.001,
			PatternFormatter:    customFormatter,
		}
		opt := mkOpt(cfg, true)
		p := NewPainter(PainterOptions{OutputFormat: ChartOutputSVG, Width: 800, Height: 600})
		require.NoError(t, p.CandlestickChart(opt))
		svg, err := p.Bytes()
		require.NoError(t, err)
		s := string(svg)
		// With user priority, user labels take precedence everywhere they're provided
		assert.Contains(t, s, "UserLabel")
		assert.NotContains(t, s, "PF:")
	})

	t.Run("no_user_labels_shows_patterns", func(t *testing.T) {
		cfg := CandlestickPatternConfig{
			PreferPatternLabels: false,
			EnabledPatterns:     []string{"doji"},
			DojiThreshold:       0.001,
			PatternFormatter:    customFormatter,
		}
		opt := mkOpt(cfg, false) // user label disabled
		p := NewPainter(PainterOptions{OutputFormat: ChartOutputSVG, Width: 800, Height: 600})
		require.NoError(t, p.CandlestickChart(opt))
		svg, err := p.Bytes()
		require.NoError(t, err)
		s := string(svg)
		// When no user labels are provided, patterns should be shown
		assert.Contains(t, s, "PF:"+"doji")
		assert.NotContains(t, s, "UserLabel")
	})
}

func TestCandlestickChartPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		optGen func() CandlestickChartOption
		svg    string
		pngCRC uint32
	}{
		{
			name: "doji",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 105, High: 107, Low: 103, Close: 105}, // Pure Doji pattern - minimal body and minimal shadows
					{Open: 105, High: 112, Low: 98, Close: 108},  // Normal candle
				}
				return makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">111</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">101</text><text x=\"18\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">99</text><text x=\"18\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97</text><text x=\"18\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 46 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 46 569\nL 46 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 294 569\nL 294 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 542 569\nL 542 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"166\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"414\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"662\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path d=\"M 170 103\nL 170 257\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 170 411\nL 170 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 121 103\nL 219 103\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 121 564\nL 219 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 71 257\nL 269 257\nL 269 411\nL 71 411\nL 71 257\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 418 195\nL 418 257\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 418 257\nL 418 318\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 369 195\nL 467 195\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 369 318\nL 467 318\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 319 257\nL 517 257\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 666 41\nL 666 164\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 666 257\nL 666 472\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 617 41\nL 715 41\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 617 472\nL 715 472\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 567 164\nL 765 164\nL 765 257\nL 567 257\nL 567 164\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 423 244\nL 456 244\nL 456 244\nA 4 4 90.00 0 1 460 248\nL 460 261\nL 460 261\nA 4 4 90.00 0 1 456 265\nL 423 265\nL 423 265\nA 4 4 90.00 0 1 419 261\nL 419 248\nL 419 248\nA 4 4 90.00 0 1 423 244\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"423\" y=\"261\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text></svg>",
			pngCRC: 0x4c6fdc4d,
		},
		{
			name: "hammer",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 108, High: 109, Low: 98, Close: 107},  // Hammer pattern
					{Open: 107, High: 112, Low: 102, Close: 110}, // Normal candle
				}
				return makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">111</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">101</text><text x=\"18\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">99</text><text x=\"18\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97</text><text x=\"18\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 46 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 46 569\nL 46 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 294 569\nL 294 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 542 569\nL 542 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"166\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"414\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"662\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path d=\"M 170 103\nL 170 257\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 170 411\nL 170 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 121 103\nL 219 103\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 121 564\nL 219 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 71 257\nL 269 257\nL 269 411\nL 71 411\nL 71 257\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 418 134\nL 418 164\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 418 195\nL 418 472\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 369 134\nL 467 134\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 369 472\nL 467 472\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 319 164\nL 517 164\nL 517 195\nL 319 195\nL 319 164\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 666 41\nL 666 103\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 666 195\nL 666 349\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 617 41\nL 715 41\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 617 349\nL 715 349\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 567 103\nL 765 103\nL 765 195\nL 567 195\nL 567 103\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 423 182\nL 483 182\nL 483 182\nA 4 4 90.00 0 1 487 186\nL 487 199\nL 487 199\nA 4 4 90.00 0 1 483 203\nL 423 203\nL 423 203\nA 4 4 90.00 0 1 419 199\nL 419 186\nL 419 186\nA 4 4 90.00 0 1 423 182\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"423\" y=\"199\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text></svg>",
			pngCRC: 0x78c286f8,
		},
		{
			name: "inverted_hammer",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105}, // Normal candle
					{Open: 95, High: 107, Low: 94, Close: 96},   // Inverted hammer
					{Open: 96, High: 102, Low: 91, Close: 98},   // Normal candle
				}
				return makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.5</text><text x=\"22\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107.5</text><text x=\"22\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">102.5</text><text x=\"22\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97.5</text><text x=\"31\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">92.5</text><text x=\"31\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 55 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 59 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 59 569\nL 59 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 302 569\nL 302 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 546 569\nL 546 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"176\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"420\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"664\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path d=\"M 180 72\nL 180 195\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 180 318\nL 180 441\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 132 72\nL 228 72\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 132 441\nL 228 441\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 83 195\nL 277 195\nL 277 318\nL 83 318\nL 83 195\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 424 146\nL 424 417\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 424 441\nL 424 466\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 376 146\nL 472 146\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 376 466\nL 472 466\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 327 417\nL 521 417\nL 521 441\nL 327 441\nL 327 417\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 668 269\nL 668 368\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 668 417\nL 668 540\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 620 269\nL 716 269\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 620 540\nL 716 540\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 571 368\nL 765 368\nL 765 417\nL 571 417\nL 571 368\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 429 398\nL 521 398\nL 521 398\nA 4 4 90.00 0 1 525 402\nL 525 428\nL 525 428\nA 4 4 90.00 0 1 521 432\nL 429 432\nL 429 432\nA 4 4 90.00 0 1 425 428\nL 425 402\nL 425 402\nA 4 4 90.00 0 1 429 398\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"429\" y=\"415\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"433\" y=\"428\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text></svg>",
			pngCRC: 0xa77fd275,
		},
		{
			name: "shooting_star",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 107, High: 125, Low: 106, Close: 108}, // Shooting star - small body at bottom, long upper shadow
					{Open: 107, High: 112, Low: 102, Close: 109}, // Normal candle
				}
				return makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"85\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"154\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"223\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">115</text><text x=\"9\" y=\"292\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"361\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"430\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"499\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 79\nL 790 79\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 148\nL 790 148\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 217\nL 790 217\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 287\nL 790 287\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 356\nL 790 356\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 425\nL 790 425\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 494\nL 790 494\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 46 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 46 569\nL 46 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 294 569\nL 294 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 542 569\nL 542 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"166\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"414\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"662\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path d=\"M 170 287\nL 170 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 170 426\nL 170 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 121 287\nL 219 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 121 495\nL 219 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 71 357\nL 269 357\nL 269 426\nL 71 426\nL 71 357\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 418 80\nL 418 315\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 418 329\nL 418 343\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 369 80\nL 467 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 369 343\nL 467 343\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 319 315\nL 517 315\nL 517 329\nL 319 329\nL 319 315\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 666 260\nL 666 301\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 666 329\nL 666 398\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 617 260\nL 715 260\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 617 398\nL 715 398\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 567 301\nL 765 301\nL 765 329\nL 567 329\nL 567 301\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 423 296\nL 515 296\nL 515 296\nA 4 4 90.00 0 1 519 300\nL 519 326\nL 519 326\nA 4 4 90.00 0 1 515 330\nL 423 330\nL 423 330\nA 4 4 90.00 0 1 419 326\nL 419 300\nL 419 300\nA 4 4 90.00 0 1 423 296\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"423\" y=\"313\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"427\" y=\"326\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text></svg>",
			pngCRC: 0xdc0a2c43,
		},
		{
			name: "gravestone_doji",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 108, High: 125, Low: 108, Close: 108}, // Gravestone doji - minimal body at bottom, long upper shadow only
					{Open: 108, High: 115, Low: 103, Close: 110}, // Normal candle
				}
				return makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"85\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"154\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"223\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">115</text><text x=\"9\" y=\"292\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"361\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"430\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"499\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 79\nL 790 79\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 148\nL 790 148\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 217\nL 790 217\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 287\nL 790 287\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 356\nL 790 356\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 425\nL 790 425\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 494\nL 790 494\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 46 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 46 569\nL 46 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 294 569\nL 294 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 542 569\nL 542 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"166\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"414\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"662\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path d=\"M 170 287\nL 170 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 170 426\nL 170 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 121 287\nL 219 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 121 495\nL 219 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 71 357\nL 269 357\nL 269 426\nL 71 426\nL 71 357\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 418 80\nL 418 315\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 369 80\nL 467 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 369 315\nL 467 315\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 319 315\nL 517 315\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 666 218\nL 666 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 666 315\nL 666 384\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 617 218\nL 715 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 617 384\nL 715 384\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 567 287\nL 765 287\nL 765 315\nL 567 315\nL 567 287\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 423 283\nL 515 283\nL 515 283\nA 4 4 90.00 0 1 519 287\nL 519 339\nL 519 339\nA 4 4 90.00 0 1 515 343\nL 423 343\nL 423 343\nA 4 4 90.00 0 1 419 339\nL 419 287\nL 419 287\nA 4 4 90.00 0 1 423 283\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"423\" y=\"300\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"431\" y=\"313\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">† Gravestone</text><text x=\"452\" y=\"326\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"427\" y=\"339\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text></svg>",
			pngCRC: 0xb8d7f058,
		},
		{
			name: "dragonfly_doji",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 109, High: 110, Low: 90, Close: 109},  // Dragonfly doji
					{Open: 109, High: 115, Low: 104, Close: 112}, // Normal candle
				}
				return makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">117.5</text><text x=\"22\" y=\"66\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">115</text><text x=\"9\" y=\"116\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.5</text><text x=\"22\" y=\"166\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"216\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107.5</text><text x=\"22\" y=\"266\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"317\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">102.5</text><text x=\"22\" y=\"367\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"417\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97.5</text><text x=\"31\" y=\"467\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"517\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">92.5</text><text x=\"31\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 55 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 60\nL 790 60\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 110\nL 790 110\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 161\nL 790 161\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 211\nL 790 211\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 261\nL 790 261\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 312\nL 790 312\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 362\nL 790 362\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 412\nL 790 412\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 463\nL 790 463\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 513\nL 790 513\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 59 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 59 569\nL 59 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 302 569\nL 302 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 546 569\nL 546 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"176\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"420\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"664\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path d=\"M 180 162\nL 180 262\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 180 363\nL 180 464\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 132 162\nL 228 162\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 132 464\nL 228 464\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 83 262\nL 277 262\nL 277 363\nL 83 363\nL 83 262\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 424 162\nL 424 182\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 424 182\nL 424 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 376 162\nL 472 162\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 376 564\nL 472 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 327 182\nL 521 182\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 668 61\nL 668 121\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 668 182\nL 668 282\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 620 61\nL 716 61\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 620 282\nL 716 282\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 571 121\nL 765 121\nL 765 182\nL 571 182\nL 571 121\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 429 156\nL 497 156\nL 497 156\nA 4 4 90.00 0 1 501 160\nL 501 199\nL 501 199\nA 4 4 90.00 0 1 497 203\nL 429 203\nL 429 203\nA 4 4 90.00 0 1 425 199\nL 425 160\nL 425 160\nA 4 4 90.00 0 1 429 156\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"173\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"429\" y=\"186\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ψ Dragonfly</text><text x=\"446\" y=\"199\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text></svg>",
			pngCRC: 0x5432dcb8,
		},
		{
			name: "bullish_marubozu",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 100, High: 120, Low: 100, Close: 120}, // Bullish marubozu
					{Open: 120, High: 125, Low: 115, Close: 122}, // Normal candle
				}
				return makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"85\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"154\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"223\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">115</text><text x=\"9\" y=\"292\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"361\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"430\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"499\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 79\nL 790 79\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 148\nL 790 148\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 217\nL 790 217\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 287\nL 790 287\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 356\nL 790 356\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 425\nL 790 425\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 494\nL 790 494\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 46 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 46 569\nL 46 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 294 569\nL 294 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 542 569\nL 542 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"166\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"414\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"662\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path d=\"M 170 287\nL 170 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 170 426\nL 170 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 121 287\nL 219 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 121 495\nL 219 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 71 357\nL 269 357\nL 269 426\nL 71 426\nL 71 357\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 369 149\nL 467 149\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 369 426\nL 467 426\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 319 149\nL 517 149\nL 517 426\nL 319 426\nL 319 149\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 666 80\nL 666 121\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 666 149\nL 666 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 617 80\nL 715 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 617 218\nL 715 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 567 121\nL 765 121\nL 765 149\nL 567 149\nL 567 121\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 423 136\nL 515 136\nL 515 136\nA 4 4 90.00 0 1 519 140\nL 519 153\nL 519 153\nA 4 4 90.00 0 1 515 157\nL 423 157\nL 423 157\nA 4 4 90.00 0 1 419 153\nL 419 140\nL 419 140\nA 4 4 90.00 0 1 423 136\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"423\" y=\"153\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">^ Bull Marubozu</text></svg>",
			pngCRC: 0xc20c1671,
		},
		{
			name: "bearish_marubozu",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 120, High: 120, Low: 100, Close: 100}, // Bearish marubozu
					{Open: 100, High: 105, Low: 95, Close: 102},  // Normal candle
				}
				return makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"94\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"173\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">115</text><text x=\"9\" y=\"252\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"331\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"410\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"489\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 89\nL 790 89\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 168\nL 790 168\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 247\nL 790 247\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 326\nL 790 326\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 405\nL 790 405\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 484\nL 790 484\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 46 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 46 569\nL 46 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 294 569\nL 294 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 542 569\nL 542 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"166\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"414\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"662\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path d=\"M 170 248\nL 170 327\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 170 406\nL 170 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 121 248\nL 219 248\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 121 485\nL 219 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 71 327\nL 269 327\nL 269 406\nL 71 406\nL 71 327\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 369 90\nL 467 90\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 369 406\nL 467 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 319 90\nL 517 90\nL 517 406\nL 319 406\nL 319 90\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 666 327\nL 666 375\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 666 406\nL 666 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 617 327\nL 715 327\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 617 485\nL 715 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 567 375\nL 765 375\nL 765 406\nL 567 406\nL 567 375\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 423 393\nL 520 393\nL 520 393\nA 4 4 90.00 0 1 524 397\nL 524 410\nL 524 410\nA 4 4 90.00 0 1 520 414\nL 423 414\nL 423 414\nA 4 4 90.00 0 1 419 410\nL 419 397\nL 419 397\nA 4 4 90.00 0 1 423 393\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"423\" y=\"410\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">v Bear Marubozu</text></svg>",
			pngCRC: 0x145e4726,
		},
		{
			name: "bullish_engulfing",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 110, High: 112, Low: 105, Close: 106}, // Small bearish candle
					{Open: 104, High: 115, Low: 103, Close: 114}, // Bullish engulfing
					{Open: 114, High: 120, Low: 112, Close: 118}, // Normal candle
				}
				return makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"94\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"173\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">115</text><text x=\"9\" y=\"252\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"331\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"410\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"489\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 89\nL 790 89\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 168\nL 790 168\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 247\nL 790 247\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 326\nL 790 326\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 405\nL 790 405\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 484\nL 790 484\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 46 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 46 569\nL 46 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 232 569\nL 232 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 418 569\nL 418 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 604 569\nL 604 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"135\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"321\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"507\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"693\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><path d=\"M 139 248\nL 139 327\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 139 406\nL 139 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 102 248\nL 176 248\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 102 485\nL 176 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 65 327\nL 213 327\nL 213 406\nL 65 406\nL 65 327\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 325 216\nL 325 248\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 325 311\nL 325 327\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 288 216\nL 362 216\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 288 327\nL 362 327\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 251 248\nL 399 248\nL 399 311\nL 251 311\nL 251 248\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 511 169\nL 511 185\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 511 343\nL 511 359\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 474 169\nL 548 169\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 474 359\nL 548 359\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 437 185\nL 585 185\nL 585 343\nL 437 343\nL 437 185\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 697 90\nL 697 121\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 697 185\nL 697 216\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 660 90\nL 734 90\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 660 216\nL 734 216\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 623 121\nL 771 121\nL 771 185\nL 623 185\nL 623 121\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 516 172\nL 607 172\nL 607 172\nA 4 4 90.00 0 1 611 176\nL 611 189\nL 611 189\nA 4 4 90.00 0 1 607 193\nL 516 193\nL 516 193\nA 4 4 90.00 0 1 512 189\nL 512 176\nL 512 176\nA 4 4 90.00 0 1 516 172\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"516\" y=\"189\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text></svg>",
			pngCRC: 0xda831b91,
		},
		{
			name: "bearish_engulfing",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 106, High: 112, Low: 105, Close: 110}, // Small bullish candle
					{Open: 114, High: 115, Low: 103, Close: 104}, // Bearish engulfing
					{Open: 104, High: 108, Low: 100, Close: 102}, // Normal candle
				}
				return makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">117.5</text><text x=\"22\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">115</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.5</text><text x=\"22\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107.5</text><text x=\"22\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">102.5</text><text x=\"22\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97.5</text><text x=\"31\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><path d=\"M 55 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 55 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 59 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 59 569\nL 59 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 241 569\nL 241 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 424 569\nL 424 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 607 569\nL 607 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"146\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"328\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"511\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"694\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><path d=\"M 150 195\nL 150 318\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 150 441\nL 150 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 114 195\nL 186 195\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 114 564\nL 186 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 77 318\nL 223 318\nL 223 441\nL 77 441\nL 77 318\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 332 146\nL 332 195\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 332 294\nL 332 318\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 296 146\nL 368 146\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 296 318\nL 368 318\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 259 195\nL 405 195\nL 405 294\nL 259 294\nL 259 195\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 515 72\nL 515 97\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 515 343\nL 515 368\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 479 72\nL 551 72\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 479 368\nL 551 368\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 442 97\nL 588 97\nL 588 343\nL 442 343\nL 442 97\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 698 244\nL 698 343\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 698 392\nL 698 441\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 662 244\nL 734 244\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 662 441\nL 734 441\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 625 343\nL 771 343\nL 771 392\nL 625 392\nL 625 343\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 520 330\nL 615 330\nL 615 330\nA 4 4 90.00 0 1 619 334\nL 619 347\nL 619 347\nA 4 4 90.00 0 1 615 351\nL 520 351\nL 520 351\nA 4 4 90.00 0 1 516 347\nL 516 334\nL 516 334\nA 4 4 90.00 0 1 520 330\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"520\" y=\"347\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">V Bear Engulfing</text></svg>",
			pngCRC: 0xa679414f,
		},
		{
			name: "morning_star",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 120, High: 125, Low: 105, Close: 108}, // Large bearish
					{Open: 102, High: 104, Low: 100, Close: 103}, // Small body, gap down - overlaps are expected
					{Open: 108, High: 125, Low: 106, Close: 122}, // Large bullish, gap up
					{Open: 122, High: 128, Low: 120, Close: 125}, // Normal candle
				}
				return makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"85\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"154\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"223\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">115</text><text x=\"9\" y=\"292\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"361\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"430\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"499\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 79\nL 790 79\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 148\nL 790 148\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 217\nL 790 217\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 287\nL 790 287\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 356\nL 790 356\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 425\nL 790 425\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 494\nL 790 494\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 46 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 46 569\nL 46 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 194 569\nL 194 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 343 569\nL 343 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 492 569\nL 492 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 641 569\nL 641 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"116\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"264\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"413\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"562\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><text x=\"711\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">5</text><path d=\"M 120 287\nL 120 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 120 426\nL 120 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 91 287\nL 149 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 91 495\nL 149 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 61 357\nL 179 357\nL 179 426\nL 61 426\nL 61 357\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 268 80\nL 268 149\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 268 315\nL 268 357\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 239 80\nL 297 80\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 239 357\nL 297 357\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 209 149\nL 327 149\nL 327 315\nL 209 315\nL 209 149\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 417 371\nL 417 384\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 417 398\nL 417 426\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 388 371\nL 446 371\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 388 426\nL 446 426\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 358 384\nL 476 384\nL 476 398\nL 358 398\nL 358 384\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 566 80\nL 566 121\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 566 315\nL 566 343\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 537 80\nL 595 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 537 343\nL 595 343\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 507 121\nL 625 121\nL 625 315\nL 507 315\nL 507 121\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 715 38\nL 715 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 715 121\nL 715 149\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 686 38\nL 744 38\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 686 149\nL 744 149\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 656 80\nL 774 80\nL 774 121\nL 656 121\nL 656 80\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 571 108\nL 655 108\nL 655 108\nA 4 4 90.00 0 1 659 112\nL 659 125\nL 659 125\nA 4 4 90.00 0 1 655 129\nL 571 129\nL 571 129\nA 4 4 90.00 0 1 567 125\nL 567 112\nL 567 112\nA 4 4 90.00 0 1 571 108\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"571\" y=\"125\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">* Morning Star</text></svg>",
			pngCRC: 0xb5ace127,
		},
		{
			name: "evening_star",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 122, High: 140, Low: 120, Close: 138}, // Large bullish
					{Open: 142, High: 144, Low: 140, Close: 143}, // Small body, gap up - overlaps are expected
					{Open: 138, High: 140, Low: 115, Close: 118}, // Large bearish, gap down
					{Open: 118, High: 122, Low: 115, Close: 120}, // Normal candle - harami overlap is expected
				}
				return makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">150</text><text x=\"9\" y=\"108\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">140</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"292\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"476\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 102\nL 790 102\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 287\nL 790 287\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 471\nL 790 471\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 46 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 46 569\nL 46 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 194 569\nL 194 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 343 569\nL 343 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 492 569\nL 492 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 641 569\nL 641 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"116\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"264\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"413\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"562\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><text x=\"711\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">5</text><path d=\"M 120 380\nL 120 426\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 120 472\nL 120 518\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 91 380\nL 149 380\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 91 518\nL 149 518\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 61 426\nL 179 426\nL 179 472\nL 61 472\nL 61 426\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 268 103\nL 268 121\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 268 269\nL 268 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 239 103\nL 297 103\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 239 287\nL 297 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 209 121\nL 327 121\nL 327 269\nL 209 269\nL 209 121\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 417 66\nL 417 75\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 417 84\nL 417 103\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 388 66\nL 446 66\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 388 103\nL 446 103\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 358 75\nL 476 75\nL 476 84\nL 358 84\nL 358 75\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 566 103\nL 566 121\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 566 306\nL 566 334\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 537 103\nL 595 103\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 537 334\nL 595 334\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 507 121\nL 625 121\nL 625 306\nL 507 306\nL 507 121\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 715 269\nL 715 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 715 306\nL 715 334\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 686 269\nL 744 269\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 686 334\nL 744 334\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 656 287\nL 774 287\nL 774 306\nL 656 306\nL 656 287\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 571 293\nL 653 293\nL 653 293\nA 4 4 90.00 0 1 657 297\nL 657 310\nL 657 310\nA 4 4 90.00 0 1 653 314\nL 571 314\nL 571 314\nA 4 4 90.00 0 1 567 310\nL 567 297\nL 567 297\nA 4 4 90.00 0 1 571 293\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"571\" y=\"310\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">⁎ Evening Star</text></svg>",
			pngCRC: 0xf70cd1b,
		},
		{
			name: "piercing_line",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 120, High: 121, Low: 115, Close: 115}, // Bearish candle
					{Open: 112, High: 119, Low: 111, Close: 118}, // Piercing line (opens below prev low, closes above midpoint)
					{Open: 118, High: 125, Low: 116, Close: 122}, // Normal candle
				}
				return makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"85\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"154\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"223\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">115</text><text x=\"9\" y=\"292\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"361\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"430\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"499\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 79\nL 790 79\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 148\nL 790 148\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 217\nL 790 217\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 287\nL 790 287\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 356\nL 790 356\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 425\nL 790 425\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 494\nL 790 494\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 46 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 46 569\nL 46 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 232 569\nL 232 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 418 569\nL 418 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 604 569\nL 604 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"135\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"321\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"507\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"693\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><path d=\"M 139 287\nL 139 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 139 426\nL 139 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 102 287\nL 176 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 102 495\nL 176 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 65 357\nL 213 357\nL 213 426\nL 65 426\nL 65 357\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 325 135\nL 325 149\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 288 135\nL 362 135\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 288 218\nL 362 218\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 251 149\nL 399 149\nL 399 218\nL 251 218\nL 251 149\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 511 163\nL 511 177\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 511 260\nL 511 274\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 474 163\nL 548 163\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 474 274\nL 548 274\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 437 177\nL 585 177\nL 585 260\nL 437 260\nL 437 177\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 697 80\nL 697 121\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 697 177\nL 697 204\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 660 80\nL 734 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 660 204\nL 734 204\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 623 121\nL 771 121\nL 771 177\nL 623 177\nL 623 121\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 516 164\nL 597 164\nL 597 164\nA 4 4 90.00 0 1 601 168\nL 601 181\nL 601 181\nA 4 4 90.00 0 1 597 185\nL 516 185\nL 516 185\nA 4 4 90.00 0 1 512 181\nL 512 168\nL 512 168\nA 4 4 90.00 0 1 516 164\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"516\" y=\"181\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">| Piercing Line</text></svg>",
			pngCRC: 0x657d6b63,
		},
		{
			name: "dark_cloud_cover",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 118, High: 125, Low: 117, Close: 125}, // Bullish candle
					{Open: 127, High: 128, Low: 120, Close: 121}, // Dark cloud cover (opens above prev high, closes below midpoint)
					{Open: 121, High: 124, Low: 118, Close: 120}, // Normal candle
				}
				return makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"85\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"154\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"223\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">115</text><text x=\"9\" y=\"292\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"361\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"430\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"499\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 79\nL 790 79\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 148\nL 790 148\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 217\nL 790 217\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 287\nL 790 287\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 356\nL 790 356\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 425\nL 790 425\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 494\nL 790 494\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 46 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 46 569\nL 46 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 232 569\nL 232 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 418 569\nL 418 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 604 569\nL 604 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"135\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"321\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"507\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"693\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><path d=\"M 139 287\nL 139 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 139 426\nL 139 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 102 287\nL 176 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 102 495\nL 176 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 65 357\nL 213 357\nL 213 426\nL 65 426\nL 65 357\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 325 177\nL 325 191\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 288 80\nL 362 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 288 191\nL 362 191\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 251 80\nL 399 80\nL 399 177\nL 251 177\nL 251 80\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 511 38\nL 511 52\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 511 135\nL 511 149\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 474 38\nL 548 38\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 474 149\nL 548 149\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 437 52\nL 585 52\nL 585 135\nL 437 135\nL 437 52\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 697 94\nL 697 135\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 697 149\nL 697 177\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 660 94\nL 734 94\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 660 177\nL 734 177\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 623 135\nL 771 135\nL 771 149\nL 623 149\nL 623 135\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 516 122\nL 590 122\nL 590 122\nA 4 4 90.00 0 1 594 126\nL 594 139\nL 594 139\nA 4 4 90.00 0 1 590 143\nL 516 143\nL 516 143\nA 4 4 90.00 0 1 512 139\nL 512 126\nL 512 126\nA 4 4 90.00 0 1 516 122\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"516\" y=\"139\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ξ Dark Cloud</text></svg>",
			pngCRC: 0xeb8f9e89,
		},
		{
			name: "engulfing_and_stars",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 110, High: 112, Low: 105, Close: 106}, // Small bearish candle
					{Open: 104, High: 115, Low: 103, Close: 114}, // Bullish engulfing
					{Open: 120, High: 125, Low: 105, Close: 108}, // Large bearish (morning star setup)
					{Open: 102, High: 104, Low: 100, Close: 103}, // Small body, gap down
					{Open: 108, High: 125, Low: 106, Close: 122}, // Large bullish, gap up (morning star completion)
					{Open: 122, High: 128, Low: 120, Close: 125}, // Normal candle
				}
				opt := makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
				opt.XAxis = XAxisOption{Show: Ptr(false)}
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"88\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"160\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"232\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">115</text><text x=\"9\" y=\"305\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"377\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"449\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"521\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 82\nL 790 82\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 155\nL 790 155\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 227\nL 790 227\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 300\nL 790 300\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 372\nL 790 372\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 445\nL 790 445\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 517\nL 790 517\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 99 300\nL 99 373\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 99 445\nL 99 518\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 78 300\nL 120 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 78 518\nL 120 518\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 57 373\nL 141 373\nL 141 445\nL 57 445\nL 57 373\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 205 271\nL 205 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 205 358\nL 205 373\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 184 271\nL 226 271\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 184 373\nL 226 373\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 163 300\nL 247 300\nL 247 358\nL 163 358\nL 163 300\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 311 228\nL 311 242\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 311 387\nL 311 402\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 290 228\nL 332 228\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 290 402\nL 332 402\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 269 242\nL 353 242\nL 353 387\nL 269 387\nL 269 242\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 417 83\nL 417 155\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 417 329\nL 417 373\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 396 83\nL 438 83\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 396 373\nL 438 373\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 375 155\nL 459 155\nL 459 329\nL 375 329\nL 375 155\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 524 387\nL 524 402\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 524 416\nL 524 445\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 503 387\nL 545 387\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 503 445\nL 545 445\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 482 402\nL 566 402\nL 566 416\nL 482 416\nL 482 402\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 630 83\nL 630 126\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 630 329\nL 630 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 609 83\nL 651 83\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 609 358\nL 651 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 588 126\nL 672 126\nL 672 329\nL 588 329\nL 588 126\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 736 39\nL 736 83\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 736 126\nL 736 155\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 715 39\nL 757 39\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 715 155\nL 757 155\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 694 83\nL 778 83\nL 778 126\nL 694 126\nL 694 83\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 316 229\nL 407 229\nL 407 229\nA 4 4 90.00 0 1 411 233\nL 411 246\nL 411 246\nA 4 4 90.00 0 1 407 250\nL 316 250\nL 316 250\nA 4 4 90.00 0 1 312 246\nL 312 233\nL 312 233\nA 4 4 90.00 0 1 316 229\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"316\" y=\"246\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text><path d=\"M 422 316\nL 496 316\nL 496 316\nA 4 4 90.00 0 1 500 320\nL 500 333\nL 500 333\nA 4 4 90.00 0 1 496 337\nL 422 337\nL 422 337\nA 4 4 90.00 0 1 418 333\nL 418 320\nL 418 320\nA 4 4 90.00 0 1 422 316\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"422\" y=\"333\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ξ Dark Cloud</text><path d=\"M 635 113\nL 719 113\nL 719 113\nA 4 4 90.00 0 1 723 117\nL 723 130\nL 723 130\nA 4 4 90.00 0 1 719 134\nL 635 134\nL 635 134\nA 4 4 90.00 0 1 631 130\nL 631 117\nL 631 117\nA 4 4 90.00 0 1 635 113\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"635\" y=\"130\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">* Morning Star</text></svg>",
			pngCRC: 0xd08dfc0c,
		},
		{
			name: "combination_mixed",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},     // Normal candle
					{Open: 105, High: 108, Low: 102, Close: 105.05}, // Doji pattern
					{Open: 105, High: 107, Low: 95, Close: 106},     // Hammer pattern
					{Open: 110, High: 125, Low: 95, Close: 112},     // Spinning top pattern
					{Open: 100, High: 120, Low: 100, Close: 120},    // Bullish marubozu pattern
					{Open: 120, High: 120, Low: 100, Close: 100},    // Bearish marubozu pattern
					{Open: 110, High: 112, Low: 105, Close: 106},    // Small bearish candle
					{Open: 104, High: 115, Low: 103, Close: 114},    // Bullish engulfing
					{Open: 106, High: 125, Low: 105, Close: 107},    // Shooting star pattern
					{Open: 109, High: 110, Low: 90, Close: 108.9},   // Dragonfly doji pattern
					{Open: 108, High: 120, Low: 107, Close: 108.1},  // Gravestone doji pattern
					{Open: 108, High: 115, Low: 103, Close: 110},    // Normal candle
				}
				opt := makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
				opt.SeriesList[0].PatternConfig.EnabledPatterns = bulk.SliceFilterInPlace(func(pattern string) bool {
					// remove high volume patterns
					if pattern == "doji" {
						return false
					}
					return true
				}, opt.SeriesList[0].PatternConfig.EnabledPatterns)
				opt.XAxis = XAxisOption{Show: Ptr(false)}
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"88\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"160\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"232\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">115</text><text x=\"9\" y=\"305\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"377\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"449\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"521\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 82\nL 790 82\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 155\nL 790 155\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 227\nL 790 227\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 300\nL 790 300\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 372\nL 790 372\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 445\nL 790 445\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 517\nL 790 517\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 77 300\nL 77 373\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 77 445\nL 77 518\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 65 300\nL 89 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 65 518\nL 89 518\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 53 373\nL 101 373\nL 101 445\nL 53 445\nL 53 373\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 139 329\nL 139 372\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 139 373\nL 139 416\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 127 329\nL 151 329\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 127 416\nL 151 416\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 115 372\nL 163 372\nL 163 373\nL 115 373\nL 115 372\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 201 344\nL 201 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 201 373\nL 201 518\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 189 344\nL 213 344\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 189 518\nL 213 518\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 177 358\nL 225 358\nL 225 373\nL 177 373\nL 177 358\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 263 83\nL 263 271\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 263 300\nL 263 518\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 251 83\nL 275 83\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 251 518\nL 275 518\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 239 271\nL 287 271\nL 287 300\nL 239 300\nL 239 271\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 313 155\nL 337 155\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 313 445\nL 337 445\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 301 155\nL 349 155\nL 349 445\nL 301 445\nL 301 155\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 375 155\nL 399 155\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 375 445\nL 399 445\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 363 155\nL 411 155\nL 411 445\nL 363 445\nL 363 155\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 449 271\nL 449 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 449 358\nL 449 373\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 437 271\nL 461 271\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 437 373\nL 461 373\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 425 300\nL 473 300\nL 473 358\nL 425 358\nL 425 300\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 511 228\nL 511 242\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 511 387\nL 511 402\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 499 228\nL 523 228\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 499 402\nL 523 402\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 487 242\nL 535 242\nL 535 387\nL 487 387\nL 487 242\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 573 83\nL 573 344\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 573 358\nL 573 373\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 561 83\nL 585 83\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 561 373\nL 585 373\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 549 344\nL 597 344\nL 597 358\nL 549 358\nL 549 344\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 635 300\nL 635 315\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 635 316\nL 635 590\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 623 300\nL 647 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 623 590\nL 647 590\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 611 315\nL 659 315\nL 659 316\nL 611 316\nL 611 315\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 697 155\nL 697 328\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 697 329\nL 697 344\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 685 155\nL 709 155\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 685 344\nL 709 344\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 673 328\nL 721 328\nL 721 329\nL 673 329\nL 673 328\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 759 228\nL 759 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 759 329\nL 759 402\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 747 228\nL 771 228\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 747 402\nL 771 402\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 735 300\nL 783 300\nL 783 329\nL 735 329\nL 735 300\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 206 345\nL 266 345\nL 266 345\nA 4 4 90.00 0 1 270 349\nL 270 362\nL 270 362\nA 4 4 90.00 0 1 266 366\nL 206 366\nL 206 366\nA 4 4 90.00 0 1 202 362\nL 202 349\nL 202 349\nA 4 4 90.00 0 1 206 345\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"206\" y=\"362\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><path d=\"M 330 142\nL 422 142\nL 422 142\nA 4 4 90.00 0 1 426 146\nL 426 159\nL 426 159\nA 4 4 90.00 0 1 422 163\nL 330 163\nL 330 163\nA 4 4 90.00 0 1 326 159\nL 326 146\nL 326 146\nA 4 4 90.00 0 1 330 142\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"330\" y=\"159\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">^ Bull Marubozu</text><path d=\"M 392 432\nL 489 432\nL 489 432\nA 4 4 90.00 0 1 493 436\nL 493 449\nL 493 449\nA 4 4 90.00 0 1 489 453\nL 392 453\nL 392 453\nA 4 4 90.00 0 1 388 449\nL 388 436\nL 388 436\nA 4 4 90.00 0 1 392 432\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"392\" y=\"449\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">v Bear Marubozu</text><path d=\"M 516 229\nL 607 229\nL 607 229\nA 4 4 90.00 0 1 611 233\nL 611 246\nL 611 246\nA 4 4 90.00 0 1 607 250\nL 516 250\nL 516 250\nA 4 4 90.00 0 1 512 246\nL 512 233\nL 512 233\nA 4 4 90.00 0 1 516 229\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"516\" y=\"246\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text><path d=\"M 578 325\nL 670 325\nL 670 325\nA 4 4 90.00 0 1 674 329\nL 674 355\nL 674 355\nA 4 4 90.00 0 1 670 359\nL 578 359\nL 578 359\nA 4 4 90.00 0 1 574 355\nL 574 329\nL 574 329\nA 4 4 90.00 0 1 578 325\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"578\" y=\"342\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"582\" y=\"355\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><path d=\"M 640 297\nL 708 297\nL 708 297\nA 4 4 90.00 0 1 712 301\nL 712 327\nL 712 327\nA 4 4 90.00 0 1 708 331\nL 640 331\nL 640 331\nA 4 4 90.00 0 1 636 327\nL 636 301\nL 636 301\nA 4 4 90.00 0 1 640 297\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"644\" y=\"314\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"640\" y=\"327\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ψ Dragonfly</text><path d=\"M 702 302\nL 794 302\nL 794 302\nA 4 4 90.00 0 1 798 306\nL 798 345\nL 798 345\nA 4 4 90.00 0 1 794 349\nL 702 349\nL 702 349\nA 4 4 90.00 0 1 698 345\nL 698 306\nL 698 306\nA 4 4 90.00 0 1 702 302\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"702\" y=\"319\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"710\" y=\"332\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">† Gravestone</text><text x=\"706\" y=\"345\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text></svg>",
			pngCRC: 0x4e151df3,
		},
		{
			name: "combination_three_candle_patterns",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105}, // Normal candle
					// Morning star sequence
					{Open: 120, High: 125, Low: 105, Close: 108}, // Large bearish
					{Open: 102, High: 104, Low: 100, Close: 103}, // Small body, gap down
					{Open: 108, High: 125, Low: 106, Close: 122}, // Large bullish, gap up
					// Three white soldiers sequence
					{Open: 110, High: 115, Low: 109, Close: 114}, // First soldier
					{Open: 113, High: 118, Low: 112, Close: 117}, // Second soldier
					{Open: 116, High: 121, Low: 115, Close: 120}, // Third soldier
					// Evening star sequence
					{Open: 122, High: 140, Low: 120, Close: 138}, // Large bullish
					{Open: 142, High: 144, Low: 140, Close: 143}, // Small body, gap up
					{Open: 138, High: 140, Low: 115, Close: 118}, // Large bearish, gap down
					// Three black crows sequence
					{Open: 120, High: 121, Low: 115, Close: 116}, // Second crow
					{Open: 117, High: 118, Low: 112, Close: 113}, // Third crow
					{Open: 113, High: 132, Low: 106, Close: 128}, // Normal candle
				}
				opt := makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				opt.XAxis = XAxisOption{Show: Ptr(false)}
				opt.SeriesList[0].PatternConfig.EnabledPatterns = []string{
					"morning_star",
					"evening_star",
				}
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">150</text><text x=\"9\" y=\"112\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">140</text><text x=\"9\" y=\"208\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"305\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"401\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"497\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 106\nL 790 106\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 203\nL 790 203\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 300\nL 790 300\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 396\nL 790 396\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 493\nL 790 493\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 74 397\nL 74 445\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 74 494\nL 74 542\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 63 397\nL 85 397\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 63 542\nL 85 542\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 52 445\nL 96 445\nL 96 494\nL 52 494\nL 52 445\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 131 252\nL 131 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 131 416\nL 131 445\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 120 252\nL 142 252\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 120 445\nL 142 445\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 109 300\nL 153 300\nL 153 416\nL 109 416\nL 109 300\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 188 455\nL 188 465\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 188 474\nL 188 494\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 177 455\nL 199 455\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 177 494\nL 199 494\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 166 465\nL 210 465\nL 210 474\nL 166 474\nL 166 465\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 245 252\nL 245 281\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 245 416\nL 245 436\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 234 252\nL 256 252\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 234 436\nL 256 436\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 223 281\nL 267 281\nL 267 416\nL 223 416\nL 223 281\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 303 349\nL 303 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 303 397\nL 303 407\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 292 349\nL 314 349\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 292 407\nL 314 407\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 281 358\nL 325 358\nL 325 397\nL 281 397\nL 281 358\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 360 320\nL 360 329\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 360 368\nL 360 378\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 349 320\nL 371 320\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 349 378\nL 371 378\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 338 329\nL 382 329\nL 382 368\nL 338 368\nL 338 329\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 417 291\nL 417 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 417 339\nL 417 349\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 406 291\nL 428 291\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 406 349\nL 428 349\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 395 300\nL 439 300\nL 439 339\nL 395 339\nL 395 300\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 474 107\nL 474 126\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 474 281\nL 474 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 463 107\nL 485 107\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 463 300\nL 485 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 452 126\nL 496 126\nL 496 281\nL 452 281\nL 452 126\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 532 68\nL 532 78\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 532 88\nL 532 107\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 521 68\nL 543 68\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 521 107\nL 543 107\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 510 78\nL 554 78\nL 554 88\nL 510 88\nL 510 78\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 589 107\nL 589 126\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 589 320\nL 589 349\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 578 107\nL 600 107\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 578 349\nL 600 349\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 567 126\nL 611 126\nL 611 320\nL 567 320\nL 567 126\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 646 291\nL 646 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 646 339\nL 646 349\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 635 291\nL 657 291\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 635 349\nL 657 349\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 624 300\nL 668 300\nL 668 339\nL 624 339\nL 624 300\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 703 320\nL 703 329\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 703 368\nL 703 378\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 692 320\nL 714 320\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 692 378\nL 714 378\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 681 329\nL 725 329\nL 725 368\nL 681 368\nL 681 329\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 761 184\nL 761 223\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 761 368\nL 761 436\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 750 184\nL 772 184\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 750 436\nL 772 436\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 739 223\nL 783 223\nL 783 368\nL 739 368\nL 739 223\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 250 268\nL 334 268\nL 334 268\nA 4 4 90.00 0 1 338 272\nL 338 285\nL 338 285\nA 4 4 90.00 0 1 334 289\nL 250 289\nL 250 289\nA 4 4 90.00 0 1 246 285\nL 246 272\nL 246 272\nA 4 4 90.00 0 1 250 268\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"250\" y=\"285\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">* Morning Star</text><path d=\"M 594 307\nL 676 307\nL 676 307\nA 4 4 90.00 0 1 680 311\nL 680 324\nL 680 324\nA 4 4 90.00 0 1 676 328\nL 594 328\nL 594 328\nA 4 4 90.00 0 1 590 324\nL 590 311\nL 590 311\nA 4 4 90.00 0 1 594 307\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"594\" y=\"324\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">⁎ Evening Star</text></svg>",
			pngCRC: 0x69f6b541,
		},
		{
			name: "bullish_patterns",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 110, High: 112, Low: 105, Close: 106}, // Small bearish candle
					{Open: 104, High: 115, Low: 103, Close: 114}, // Bullish engulfing
					{Open: 108, High: 109, Low: 98, Close: 107},  // Hammer pattern
					{Open: 100, High: 120, Low: 100, Close: 120}, // Bullish belt hold / marubozu
					{Open: 120, High: 140, Low: 118, Close: 138}, // Large bullish
					{Open: 110, High: 119, Low: 110, Close: 118}, // Piercing line
					{Open: 118, High: 125, Low: 115, Close: 122}, // Normal candle
				}
				opt := makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
				opt.SeriesList[0].PatternConfig = (&CandlestickPatternConfig{}).WithPatternsBullish()
				opt.XAxis = XAxisOption{Show: Ptr(false)}
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">145</text><text x=\"9\" y=\"68\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">140</text><text x=\"9\" y=\"121\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">135</text><text x=\"9\" y=\"173\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"226\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"278\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"331\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">115</text><text x=\"9\" y=\"383\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"436\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"488\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"541\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 62\nL 790 62\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 115\nL 790 115\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 168\nL 790 168\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 220\nL 790 220\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 273\nL 790 273\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 326\nL 790 326\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 431\nL 790 431\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 484\nL 790 484\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 537\nL 790 537\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 92 380\nL 92 432\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 92 485\nL 92 538\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 74 380\nL 110 380\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 74 538\nL 110 538\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 55 432\nL 129 432\nL 129 485\nL 55 485\nL 55 432\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 185 358\nL 185 380\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 185 422\nL 185 432\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 167 358\nL 203 358\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 167 432\nL 203 432\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 148 380\nL 222 380\nL 222 422\nL 148 422\nL 148 380\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 278 327\nL 278 337\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 278 443\nL 278 453\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 260 327\nL 296 327\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 260 453\nL 296 453\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 241 337\nL 315 337\nL 315 443\nL 241 443\nL 241 337\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 371 390\nL 371 401\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 371 411\nL 371 506\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 353 390\nL 389 390\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 353 506\nL 389 506\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 334 401\nL 408 401\nL 408 411\nL 334 411\nL 334 401\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 446 274\nL 482 274\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 446 485\nL 482 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 427 274\nL 501 274\nL 501 485\nL 427 485\nL 427 274\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 557 63\nL 557 84\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 557 274\nL 557 295\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 539 63\nL 575 63\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 539 295\nL 575 295\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 520 84\nL 594 84\nL 594 274\nL 520 274\nL 520 84\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 650 285\nL 650 295\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 632 285\nL 668 285\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 632 380\nL 668 380\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 613 295\nL 687 295\nL 687 380\nL 613 380\nL 613 295\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 743 221\nL 743 253\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 743 295\nL 743 327\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 725 221\nL 761 221\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 725 327\nL 761 327\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 706 253\nL 780 253\nL 780 295\nL 706 295\nL 706 253\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 283 324\nL 374 324\nL 374 324\nA 4 4 90.00 0 1 378 328\nL 378 341\nL 378 341\nA 4 4 90.00 0 1 374 345\nL 283 345\nL 283 345\nA 4 4 90.00 0 1 279 341\nL 279 328\nL 279 328\nA 4 4 90.00 0 1 283 324\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"283\" y=\"341\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text><path d=\"M 376 398\nL 436 398\nL 436 398\nA 4 4 90.00 0 1 440 402\nL 440 415\nL 440 415\nA 4 4 90.00 0 1 436 419\nL 376 419\nL 376 419\nA 4 4 90.00 0 1 372 415\nL 372 402\nL 372 402\nA 4 4 90.00 0 1 376 398\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"376\" y=\"415\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><path d=\"M 469 255\nL 561 255\nL 561 255\nA 4 4 90.00 0 1 565 259\nL 565 285\nL 565 285\nA 4 4 90.00 0 1 561 289\nL 469 289\nL 469 289\nA 4 4 90.00 0 1 465 285\nL 465 259\nL 465 259\nA 4 4 90.00 0 1 469 255\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"469\" y=\"272\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">^ Bull Marubozu</text><text x=\"469\" y=\"285\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text></svg>",
			pngCRC: 0xae7ee66e,
		},
		{
			name: "bearish_patterns",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 106, High: 112, Low: 105, Close: 110}, // Small bullish candle
					{Open: 114, High: 115, Low: 103, Close: 104}, // Bearish engulfing
					{Open: 106, High: 125, Low: 105, Close: 107}, // Shooting star pattern
					{Open: 120, High: 120, Low: 100, Close: 100}, // Bearish belt hold / marubozu
					{Open: 118, High: 125, Low: 117, Close: 125}, // Bullish candle
					{Open: 127, High: 128, Low: 120, Close: 121}, // Dark cloud cover
					{Open: 121, High: 124, Low: 118, Close: 120}, // Normal candle
				}
				opt := makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
				opt.SeriesList[0].PatternConfig = (&CandlestickPatternConfig{}).WithPatternsBearish()
				opt.XAxis = XAxisOption{Show: Ptr(false)}
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"88\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"160\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"232\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">115</text><text x=\"9\" y=\"305\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"377\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"449\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"521\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 82\nL 790 82\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 155\nL 790 155\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 227\nL 790 227\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 300\nL 790 300\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 372\nL 790 372\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 445\nL 790 445\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 517\nL 790 517\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 92 300\nL 92 373\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 92 445\nL 92 518\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 74 300\nL 110 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 74 518\nL 110 518\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 55 373\nL 129 373\nL 129 445\nL 55 445\nL 55 373\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 185 271\nL 185 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 185 358\nL 185 373\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 167 271\nL 203 271\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 167 373\nL 203 373\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 148 300\nL 222 300\nL 222 358\nL 148 358\nL 148 300\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 278 228\nL 278 242\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 278 387\nL 278 402\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 260 228\nL 296 228\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 260 402\nL 296 402\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 241 242\nL 315 242\nL 315 387\nL 241 387\nL 241 242\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 371 83\nL 371 344\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 371 358\nL 371 373\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 353 83\nL 389 83\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 353 373\nL 389 373\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 334 344\nL 408 344\nL 408 358\nL 334 358\nL 334 344\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 446 155\nL 482 155\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 446 445\nL 482 445\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 427 155\nL 501 155\nL 501 445\nL 427 445\nL 427 155\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 557 184\nL 557 199\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 539 83\nL 575 83\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 539 199\nL 575 199\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 520 83\nL 594 83\nL 594 184\nL 520 184\nL 520 83\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 650 39\nL 650 54\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 650 141\nL 650 155\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 632 39\nL 668 39\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 632 155\nL 668 155\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 613 54\nL 687 54\nL 687 141\nL 613 141\nL 613 54\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 743 97\nL 743 141\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 743 155\nL 743 184\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 725 97\nL 761 97\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 725 184\nL 761 184\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 706 141\nL 780 141\nL 780 155\nL 706 155\nL 706 141\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 283 374\nL 378 374\nL 378 374\nA 4 4 90.00 0 1 382 378\nL 382 391\nL 382 391\nA 4 4 90.00 0 1 378 395\nL 283 395\nL 283 395\nA 4 4 90.00 0 1 279 391\nL 279 378\nL 279 378\nA 4 4 90.00 0 1 283 374\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"283\" y=\"391\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">V Bear Engulfing</text><path d=\"M 376 325\nL 468 325\nL 468 325\nA 4 4 90.00 0 1 472 329\nL 472 355\nL 472 355\nA 4 4 90.00 0 1 468 359\nL 376 359\nL 376 359\nA 4 4 90.00 0 1 372 355\nL 372 329\nL 372 329\nA 4 4 90.00 0 1 376 325\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"376\" y=\"342\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"384\" y=\"355\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">† Gravestone</text><path d=\"M 469 426\nL 566 426\nL 566 426\nA 4 4 90.00 0 1 570 430\nL 570 456\nL 570 456\nA 4 4 90.00 0 1 566 460\nL 469 460\nL 469 460\nA 4 4 90.00 0 1 465 456\nL 465 430\nL 465 430\nA 4 4 90.00 0 1 469 426\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"469\" y=\"443\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">v Bear Marubozu</text><text x=\"470\" y=\"456\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">V Bear Engulfing</text><path d=\"M 655 128\nL 729 128\nL 729 128\nA 4 4 90.00 0 1 733 132\nL 733 145\nL 733 145\nA 4 4 90.00 0 1 729 149\nL 655 149\nL 655 149\nA 4 4 90.00 0 1 651 145\nL 651 132\nL 651 132\nA 4 4 90.00 0 1 655 128\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"655\" y=\"145\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ξ Dark Cloud</text></svg>",
			pngCRC: 0xbd25c318,
		},
		{
			name: "reversal_patterns",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 120, High: 121, Low: 115, Close: 115}, // Bearish candle
					{Open: 112, High: 119, Low: 112, Close: 118}, // Piercing line (bullish reversal)
					{Open: 118, High: 125, Low: 118, Close: 125}, // Bullish candle
					{Open: 127, High: 127, Low: 120, Close: 121}, // Dark cloud cover (bearish reversal)
					{Open: 125, High: 126, Low: 100, Close: 102}, // Bearish with low at 100
					{Open: 102, High: 108, Low: 100, Close: 107}, // Tweezer bottom (bullish reversal)
					{Open: 107, High: 112, Low: 102, Close: 110}, // Normal candle
					// Additional reversal patterns
					{Open: 115, High: 117, Low: 95, Close: 114},    // Hammer pattern (bullish reversal)
					{Open: 112, High: 130, Low: 111, Close: 113},   // Shooting star pattern (bearish reversal)
					{Open: 108, High: 110, Low: 85, Close: 108.1},  // Dragonfly doji (bullish reversal)
					{Open: 105, High: 125, Low: 104, Close: 105.1}, // Gravestone doji (bearish reversal)
					{Open: 130, High: 135, Low: 110, Close: 115},   // Large bearish for engulfing setup
					{Open: 110, High: 140, Low: 108, Close: 138},   // Bullish engulfing (reversal)
					{Open: 140, High: 145, Low: 105, Close: 110},   // Bearish engulfing (reversal)
					// Three candle reversal patterns
					{Open: 125, High: 130, Low: 105, Close: 110}, // Large bearish for morning star
					{Open: 105, High: 108, Low: 102, Close: 106}, // Small body (morning star middle)
					{Open: 110, High: 135, Low: 108, Close: 130}, // Large bullish (morning star completion)
					{Open: 115, High: 140, Low: 113, Close: 135}, // Large bullish for evening star
					{Open: 138, High: 145, Low: 136, Close: 140}, // Small body (evening star middle)
					{Open: 135, High: 136, Low: 110, Close: 115}, // Large bearish (evening star completion)
				}
				opt := makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				opt.SeriesList[0].PatternConfig = (&CandlestickPatternConfig{}).WithPatternsReversal()
				opt.XAxis = XAxisOption{Show: Ptr(false)}
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">160</text><text x=\"9\" y=\"88\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">150</text><text x=\"9\" y=\"160\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">140</text><text x=\"9\" y=\"232\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"305\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"377\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"449\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"521\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><text x=\"18\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">80</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 82\nL 790 82\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 155\nL 790 155\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 227\nL 790 227\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 300\nL 790 300\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 372\nL 790 372\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 445\nL 790 445\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 517\nL 790 517\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 63 373\nL 63 409\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 63 445\nL 63 482\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 56 373\nL 70 373\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 56 482\nL 70 482\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 49 409\nL 77 409\nL 77 445\nL 49 445\nL 49 409\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 98 293\nL 98 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 91 293\nL 105 293\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 91 337\nL 105 337\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 84 300\nL 112 300\nL 112 337\nL 84 337\nL 84 300\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 134 308\nL 134 315\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 127 308\nL 141 308\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 127 358\nL 141 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 120 315\nL 148 315\nL 148 358\nL 120 358\nL 120 315\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 162 264\nL 176 264\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 162 315\nL 176 315\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 155 264\nL 183 264\nL 183 315\nL 155 315\nL 155 264\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 205 293\nL 205 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 198 250\nL 212 250\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 198 300\nL 212 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 191 250\nL 219 250\nL 219 293\nL 191 293\nL 191 250\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 240 257\nL 240 264\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 240 431\nL 240 445\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 233 257\nL 247 257\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 233 445\nL 247 445\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 226 264\nL 254 264\nL 254 431\nL 226 431\nL 226 264\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 276 387\nL 276 395\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 276 431\nL 276 445\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 269 387\nL 283 387\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 269 445\nL 283 445\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 262 395\nL 290 395\nL 290 431\nL 262 431\nL 262 395\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 311 358\nL 311 373\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 311 395\nL 311 431\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 304 358\nL 318 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 304 431\nL 318 431\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 297 373\nL 325 373\nL 325 395\nL 297 395\nL 297 373\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 346 322\nL 346 337\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 346 344\nL 346 482\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 339 322\nL 353 322\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 339 482\nL 353 482\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 332 337\nL 360 337\nL 360 344\nL 332 344\nL 332 337\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 382 228\nL 382 351\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 382 358\nL 382 366\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 375 228\nL 389 228\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 375 366\nL 389 366\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 368 351\nL 396 351\nL 396 358\nL 368 358\nL 368 351\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 417 373\nL 417 387\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 417 387\nL 417 554\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 410 373\nL 424 373\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 410 554\nL 424 554\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 403 387\nL 431 387\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 453 264\nL 453 409\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 453 409\nL 453 416\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 446 264\nL 460 264\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 446 416\nL 460 416\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 439 409\nL 467 409\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 488 192\nL 488 228\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 488 337\nL 488 373\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 481 192\nL 495 192\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 481 373\nL 495 373\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 474 228\nL 502 228\nL 502 337\nL 474 337\nL 474 228\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 524 155\nL 524 170\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 524 373\nL 524 387\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 517 155\nL 531 155\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 517 387\nL 531 387\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 510 170\nL 538 170\nL 538 373\nL 510 373\nL 510 170\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 559 119\nL 559 155\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 559 373\nL 559 409\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 552 119\nL 566 119\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 552 409\nL 566 409\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 545 155\nL 573 155\nL 573 373\nL 545 373\nL 545 155\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 594 228\nL 594 264\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 594 373\nL 594 409\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 587 228\nL 601 228\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 587 409\nL 601 409\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 580 264\nL 608 264\nL 608 373\nL 580 373\nL 580 264\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 630 387\nL 630 402\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 630 409\nL 630 431\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 623 387\nL 637 387\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 623 431\nL 637 431\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 616 402\nL 644 402\nL 644 409\nL 616 409\nL 616 402\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 665 192\nL 665 228\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 665 373\nL 665 387\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 658 192\nL 672 192\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 658 387\nL 672 387\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 651 228\nL 679 228\nL 679 373\nL 651 373\nL 651 228\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 701 155\nL 701 192\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 701 337\nL 701 351\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 694 155\nL 708 155\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 694 351\nL 708 351\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 687 192\nL 715 192\nL 715 337\nL 687 337\nL 687 192\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 736 119\nL 736 155\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 736 170\nL 736 184\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 729 119\nL 743 119\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 729 184\nL 743 184\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 722 155\nL 750 155\nL 750 170\nL 722 170\nL 722 155\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 772 184\nL 772 192\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 772 337\nL 772 373\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 765 184\nL 779 184\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 765 373\nL 779 373\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 758 192\nL 786 192\nL 786 337\nL 758 337\nL 758 192\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 139 302\nL 220 302\nL 220 302\nA 4 4 90.00 0 1 224 306\nL 224 319\nL 224 319\nA 4 4 90.00 0 1 220 323\nL 139 323\nL 139 323\nA 4 4 90.00 0 1 135 319\nL 135 306\nL 135 306\nA 4 4 90.00 0 1 139 302\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"139\" y=\"319\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">| Piercing Line</text><path d=\"M 210 280\nL 284 280\nL 284 280\nA 4 4 90.00 0 1 288 284\nL 288 297\nL 288 297\nA 4 4 90.00 0 1 284 301\nL 210 301\nL 210 301\nA 4 4 90.00 0 1 206 297\nL 206 284\nL 206 284\nA 4 4 90.00 0 1 210 280\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"210\" y=\"297\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ξ Dark Cloud</text><path d=\"M 351 325\nL 419 325\nL 419 325\nA 4 4 90.00 0 1 423 329\nL 423 355\nL 423 355\nA 4 4 90.00 0 1 419 359\nL 351 359\nL 351 359\nA 4 4 90.00 0 1 347 355\nL 347 329\nL 347 329\nA 4 4 90.00 0 1 351 325\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"355\" y=\"342\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"351\" y=\"355\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ψ Dragonfly</text><path d=\"M 387 338\nL 479 338\nL 479 338\nA 4 4 90.00 0 1 483 342\nL 483 355\nL 483 355\nA 4 4 90.00 0 1 479 359\nL 387 359\nL 387 359\nA 4 4 90.00 0 1 383 355\nL 383 342\nL 383 342\nA 4 4 90.00 0 1 387 338\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"387\" y=\"355\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><path d=\"M 422 368\nL 490 368\nL 490 368\nA 4 4 90.00 0 1 494 372\nL 494 398\nL 494 398\nA 4 4 90.00 0 1 490 402\nL 422 402\nL 422 402\nA 4 4 90.00 0 1 418 398\nL 418 372\nL 418 372\nA 4 4 90.00 0 1 422 368\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"426\" y=\"385\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"422\" y=\"398\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ψ Dragonfly</text><path d=\"M 458 390\nL 550 390\nL 550 390\nA 4 4 90.00 0 1 554 394\nL 554 420\nL 554 420\nA 4 4 90.00 0 1 550 424\nL 458 424\nL 458 424\nA 4 4 90.00 0 1 454 420\nL 454 394\nL 454 394\nA 4 4 90.00 0 1 458 390\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"458\" y=\"407\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"466\" y=\"420\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">† Gravestone</text><path d=\"M 529 157\nL 620 157\nL 620 157\nA 4 4 90.00 0 1 624 161\nL 624 174\nL 624 174\nA 4 4 90.00 0 1 620 178\nL 529 178\nL 529 178\nA 4 4 90.00 0 1 525 174\nL 525 161\nL 525 161\nA 4 4 90.00 0 1 529 157\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"529\" y=\"174\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text><path d=\"M 670 215\nL 754 215\nL 754 215\nA 4 4 90.00 0 1 758 219\nL 758 232\nL 758 232\nA 4 4 90.00 0 1 754 236\nL 670 236\nL 670 236\nA 4 4 90.00 0 1 666 232\nL 666 219\nL 666 219\nA 4 4 90.00 0 1 670 215\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"670\" y=\"232\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">* Morning Star</text><path d=\"M 718 324\nL 800 324\nL 800 324\nA 4 4 90.00 0 1 804 328\nL 804 341\nL 804 341\nA 4 4 90.00 0 1 800 345\nL 718 345\nL 718 345\nA 4 4 90.00 0 1 714 341\nL 714 328\nL 714 328\nA 4 4 90.00 0 1 718 324\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"718\" y=\"341\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">⁎ Evening Star</text></svg>",
			pngCRC: 0xeaf6f74e,
		},
		{
			name: "trend_patterns",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 110, High: 120, Low: 100, Close: 120}, // Marubozu bullish - trend continuation
					{Open: 125, High: 125, Low: 115, Close: 115}, // Marubozu bearish - trend continuation
					{Open: 120, High: 130, Low: 115, Close: 125}, // Large bullish for belt hold setup
					{Open: 120, High: 140, Low: 120, Close: 140}, // Belt hold bullish - trend continuation
					{Open: 135, High: 135, Low: 115, Close: 115}, // Belt hold bearish - trend continuation
					{Open: 118, High: 125, Low: 117, Close: 122}, // Normal candle
					{Open: 122, High: 130, Low: 120, Close: 128}, // Trend continuation candle
				}
				opt := makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
				opt.SeriesList[0].PatternConfig = (&CandlestickPatternConfig{}).WithPatternsTrend()
				opt.XAxis = XAxisOption{Show: Ptr(false)}
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">145</text><text x=\"9\" y=\"68\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">140</text><text x=\"9\" y=\"121\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">135</text><text x=\"9\" y=\"173\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"226\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"278\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"331\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">115</text><text x=\"9\" y=\"383\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"436\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"488\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"541\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 62\nL 790 62\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 115\nL 790 115\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 168\nL 790 168\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 220\nL 790 220\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 273\nL 790 273\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 326\nL 790 326\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 431\nL 790 431\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 484\nL 790 484\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 537\nL 790 537\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 92 380\nL 92 432\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 92 485\nL 92 538\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 74 380\nL 110 380\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 74 538\nL 110 538\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 55 432\nL 129 432\nL 129 485\nL 55 485\nL 55 432\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 185 380\nL 185 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 167 274\nL 203 274\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 167 485\nL 203 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 148 274\nL 222 274\nL 222 380\nL 148 380\nL 148 274\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 260 221\nL 296 221\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 260 327\nL 296 327\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 241 221\nL 315 221\nL 315 327\nL 241 327\nL 241 221\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 371 169\nL 371 221\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 371 274\nL 371 327\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 353 169\nL 389 169\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 353 327\nL 389 327\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 334 221\nL 408 221\nL 408 274\nL 334 274\nL 334 221\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 446 63\nL 482 63\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 446 274\nL 482 274\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 427 63\nL 501 63\nL 501 274\nL 427 274\nL 427 63\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 539 116\nL 575 116\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 539 327\nL 575 327\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 520 116\nL 594 116\nL 594 327\nL 520 327\nL 520 116\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 650 221\nL 650 253\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 650 295\nL 650 306\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 632 221\nL 668 221\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 632 306\nL 668 306\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 613 253\nL 687 253\nL 687 295\nL 613 295\nL 613 253\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 743 169\nL 743 190\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 743 253\nL 743 274\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 725 169\nL 761 169\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 725 274\nL 761 274\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 706 190\nL 780 190\nL 780 253\nL 706 253\nL 706 190\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 283 314\nL 380 314\nL 380 314\nA 4 4 90.00 0 1 384 318\nL 384 331\nL 384 331\nA 4 4 90.00 0 1 380 335\nL 283 335\nL 283 335\nA 4 4 90.00 0 1 279 331\nL 279 318\nL 279 318\nA 4 4 90.00 0 1 283 314\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"283\" y=\"331\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">v Bear Marubozu</text><path d=\"M 469 50\nL 561 50\nL 561 50\nA 4 4 90.00 0 1 565 54\nL 565 67\nL 565 67\nA 4 4 90.00 0 1 561 71\nL 469 71\nL 469 71\nA 4 4 90.00 0 1 465 67\nL 465 54\nL 465 54\nA 4 4 90.00 0 1 469 50\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"469\" y=\"67\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">^ Bull Marubozu</text><path d=\"M 562 314\nL 659 314\nL 659 314\nA 4 4 90.00 0 1 663 318\nL 663 331\nL 663 331\nA 4 4 90.00 0 1 659 335\nL 562 335\nL 562 335\nA 4 4 90.00 0 1 558 331\nL 558 318\nL 558 318\nA 4 4 90.00 0 1 562 314\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"562\" y=\"331\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">v Bear Marubozu</text></svg>",
			pngCRC: 0xef0217d,
		},
		{
			name: "all_patterns_showcase",
			optGen: func() CandlestickChartOption {
				// Comprehensive dataset showcasing all supported candlestick patterns
				data := []OHLCData{
					// 0: Setup - Normal candle
					{Open: 100, High: 110, Low: 95, Close: 105},
					// 1: Regular candle (reduce spinning top frequency)
					{Open: 105, High: 108, Low: 102, Close: 107},
					// 2: Hammer pattern
					{Open: 108, High: 109, Low: 98, Close: 107},
					// 3: Regular candle (was inverted hammer, reduce shooting star frequency)
					{Open: 95, High: 102, Low: 94, Close: 100},
					// 4: Regular candle (was shooting star, reduce frequency)
					{Open: 106, High: 115, Low: 105, Close: 112},
					// 5: Gravestone Doji pattern
					{Open: 108, High: 120, Low: 107, Close: 108.1},
					// 6: Hammer-like pattern (preserve dragonfly, reduce doji frequency)
					{Open: 109, High: 111, Low: 90, Close: 108},
					// 7: Bullish Marubozu pattern
					{Open: 100, High: 120, Low: 100, Close: 120},
					// 8: Bearish Marubozu pattern
					{Open: 120, High: 120, Low: 100, Close: 100},
					// 9: Regular candle (break harami pattern, reduce spinning top)
					{Open: 110, High: 120, Low: 107, Close: 118},
					// Setup for two-candle patterns - Large bearish candle
					{Open: 130, High: 135, Low: 110, Close: 115},
					// 11: Bullish Engulfing pattern
					{Open: 110, High: 140, Low: 108, Close: 138},
					// Setup for Bearish Engulfing - Large bullish candle
					{Open: 110, High: 140, Low: 108, Close: 138},
					// 13: Bearish Engulfing pattern (fixed to properly engulf)
					{Open: 140, High: 142, Low: 105, Close: 107},
					// Setup for Harami - Large bearish candle
					{Open: 130, High: 135, Low: 100, Close: 105},
					// 15: Regular candle (break harami by extending body)
					{Open: 110, High: 125, Low: 95, Close: 120},
					// Setup for Bearish Harami - Large bullish candle
					{Open: 100, High: 135, Low: 98, Close: 130},
					// 17: Bearish Harami pattern
					{Open: 125, High: 128, Low: 120, Close: 122},
					// Setup for Piercing Line - Bearish candle
					{Open: 120, High: 125, Low: 110, Close: 112},
					// 19: Piercing Line pattern
					{Open: 108, High: 125, Low: 107, Close: 118},
					// Setup for Dark Cloud Cover - Bullish candle
					{Open: 110, High: 125, Low: 108, Close: 123},
					// 21: Dark Cloud Cover pattern (fixed to gap up and close below midpoint)
					{Open: 128, High: 130, Low: 112, Close: 115},
					// Setup for Tweezer Top - Two candles with same high
					{Open: 110, High: 130, Low: 108, Close: 125},
					// 23: Tweezer Top pattern
					{Open: 123, High: 130, Low: 115, Close: 118},
					// Setup for Tweezer Bottom - Two candles with same low
					{Open: 120, High: 125, Low: 100, Close: 105},
					// 25: Tweezer Bottom pattern
					{Open: 108, High: 115, Low: 100, Close: 112},
					// Setup for Morning Star - Large bearish candle
					{Open: 130, High: 135, Low: 110, Close: 115},
					// 27: Morning Star middle - Small body with gap down (reduce spinning top)
					{Open: 108, High: 112, Low: 107, Close: 110},
					// 28: Morning Star completion - Large bullish candle
					{Open: 115, High: 140, Low: 113, Close: 135},
					// Setup for Evening Star - Large bullish candle
					{Open: 110, High: 140, Low: 108, Close: 135},
					// 30: Evening Star middle - Small body with proper gap up (fixed)
					{Open: 137, High: 145, Low: 136, Close: 140},
					// 31: Evening Star completion - Large bearish candle (fixed)
					{Open: 135, High: 136, Low: 115, Close: 120},
					// Setup for Three White Soldiers - Start with bearish sentiment
					{Open: 120, High: 125, Low: 110, Close: 115},
					// 33: Three White Soldiers - First soldier
					{Open: 118, High: 128, Low: 116, Close: 125},
					// 34: Three White Soldiers - Second soldier
					{Open: 127, High: 135, Low: 125, Close: 132},
					// 35: Three White Soldiers - Third soldier
					{Open: 134, High: 142, Low: 132, Close: 140},
					// Setup for Three Black Crows - Start with bullish sentiment
					{Open: 130, High: 145, Low: 128, Close: 142},
					// 37: Three Black Crows - First crow (fixed to open within previous body)
					{Open: 138, High: 140, Low: 128, Close: 132},
					// 38: Three Black Crows - Second crow (fixed to open within previous body)
					{Open: 130, High: 132, Low: 120, Close: 125},
					// 39: Three Black Crows - Third crow (fixed to open within previous body)
					{Open: 124, High: 127, Low: 115, Close: 118},
					// 40: Regular candle (reduce spinning top frequency)
					{Open: 115, High: 120, Low: 114, Close: 118},
					// 41: Regular candle (was spinning top, reduce frequency)
					{Open: 118, High: 125, Low: 115, Close: 122},
					// 42: Setup for Shooting Star - rising trend
					{Open: 120, High: 125, Low: 118, Close: 124},
					// 43: Shooting Star pattern - long upper shadow, small body near low
					{Open: 123, High: 140, Low: 122, Close: 125},
					// 44: Setup for Gravestone Doji - uptrend
					{Open: 125, High: 130, Low: 123, Close: 128},
					// 45: Gravestone Doji pattern - doji with long upper shadow
					{Open: 128, High: 145, Low: 127, Close: 128.05},
					// 46: Setup for Dragonfly Doji - downtrend
					{Open: 128, High: 130, Low: 125, Close: 126},
					// 47: Dragonfly Doji pattern - doji with long lower shadow
					{Open: 125, High: 126, Low: 110, Close: 125.05},
					// 48: Setup for Tweezer Bottom - bearish candle
					{Open: 125, High: 127, Low: 115, Close: 118},
					// 49: Tweezer Bottom pattern - same low as previous, bullish reversal
					{Open: 120, High: 125, Low: 115, Close: 123},
					// 50: Setup for Three Black Crows - high bullish candle
					{Open: 120, High: 135, Low: 118, Close: 133},
					// 51: Three Black Crows - First crow (bearish, substantial body)
					{Open: 132, High: 133, Low: 125, Close: 126},
					// 52: Three Black Crows - Second crow (bearish, opens within prev body, closes lower)
					{Open: 130, High: 131, Low: 121, Close: 122},
					// 53: Three Black Crows - Third crow (bearish, opens within prev body, closes lower)
					{Open: 125, High: 126, Low: 115, Close: 116},
					// 54: Long-Legged Doji pattern - very long shadows on both sides, small body
					{Open: 118, High: 135, Low: 95, Close: 118.1},
				}

				opt := makePatternChartOption(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
				opt.XAxis = XAxisOption{Show: Ptr(false)}
				return opt
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">160</text><text x=\"9\" y=\"98\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">150</text><text x=\"9\" y=\"181\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">140</text><text x=\"9\" y=\"263\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"346\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"428\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"511\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 92\nL 790 92\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 175\nL 790 175\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 258\nL 790 258\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 341\nL 790 341\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 424\nL 790 424\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 42 507\nL 790 507\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path d=\"M 52 425\nL 52 466\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 52 508\nL 52 549\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 50 425\nL 54 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 50 549\nL 54 549\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 47 466\nL 57 466\nL 57 508\nL 47 508\nL 47 466\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 66 441\nL 66 450\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 66 466\nL 66 491\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 64 441\nL 68 441\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 64 491\nL 68 491\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 61 450\nL 71 450\nL 71 466\nL 61 466\nL 61 450\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 79 433\nL 79 441\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 79 450\nL 79 524\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 77 433\nL 81 433\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 77 524\nL 81 524\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 74 441\nL 84 441\nL 84 450\nL 74 450\nL 74 441\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 93 491\nL 93 508\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 93 549\nL 93 557\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 91 491\nL 95 491\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 91 557\nL 95 557\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 88 508\nL 98 508\nL 98 549\nL 88 549\nL 88 508\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 106 383\nL 106 408\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 106 458\nL 106 466\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 104 383\nL 108 383\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 104 466\nL 108 466\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 101 408\nL 111 408\nL 111 458\nL 101 458\nL 101 408\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 120 342\nL 120 441\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 120 441\nL 120 450\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 118 342\nL 122 342\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 118 450\nL 122 450\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 115 441\nL 125 441\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 133 416\nL 133 433\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 133 441\nL 133 590\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 131 416\nL 135 416\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 131 590\nL 135 590\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 128 433\nL 138 433\nL 138 441\nL 128 441\nL 128 433\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 145 342\nL 149 342\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 145 508\nL 149 508\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 142 342\nL 152 342\nL 152 508\nL 142 508\nL 142 342\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 158 342\nL 162 342\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 158 508\nL 162 508\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 155 342\nL 165 342\nL 165 508\nL 155 508\nL 155 342\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 174 342\nL 174 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 174 425\nL 174 450\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 172 342\nL 176 342\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 172 450\nL 176 450\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 169 358\nL 179 358\nL 179 425\nL 169 425\nL 169 358\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 187 218\nL 187 259\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 187 383\nL 187 425\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 185 218\nL 189 218\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 185 425\nL 189 425\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 182 259\nL 192 259\nL 192 383\nL 182 383\nL 182 259\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 201 176\nL 201 193\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 201 425\nL 201 441\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 199 176\nL 203 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 199 441\nL 203 441\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 196 193\nL 206 193\nL 206 425\nL 196 425\nL 196 193\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 214 176\nL 214 193\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 214 425\nL 214 441\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 212 176\nL 216 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 212 441\nL 216 441\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 209 193\nL 219 193\nL 219 425\nL 209 425\nL 209 193\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 228 160\nL 228 176\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 228 450\nL 228 466\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 226 160\nL 230 160\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 226 466\nL 230 466\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 223 176\nL 233 176\nL 233 450\nL 223 450\nL 223 176\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 241 218\nL 241 259\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 241 466\nL 241 508\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 239 218\nL 243 218\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 239 508\nL 243 508\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 236 259\nL 246 259\nL 246 466\nL 236 466\nL 236 259\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 255 300\nL 255 342\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 255 425\nL 255 549\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 253 300\nL 257 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 253 549\nL 257 549\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 250 342\nL 260 342\nL 260 425\nL 250 425\nL 250 342\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 268 218\nL 268 259\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 268 508\nL 268 524\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 266 218\nL 270 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 266 524\nL 270 524\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 263 259\nL 273 259\nL 273 508\nL 263 508\nL 263 259\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 282 276\nL 282 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 282 325\nL 282 342\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 280 276\nL 284 276\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 280 342\nL 284 342\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 277 300\nL 287 300\nL 287 325\nL 277 325\nL 277 300\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 296 300\nL 296 342\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 296 408\nL 296 425\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 294 300\nL 298 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 294 425\nL 298 425\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 291 342\nL 301 342\nL 301 408\nL 291 408\nL 291 342\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 309 300\nL 309 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 309 441\nL 309 450\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 307 300\nL 311 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 307 450\nL 311 450\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 304 358\nL 314 358\nL 314 441\nL 304 441\nL 304 358\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 323 300\nL 323 317\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 323 425\nL 323 441\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 321 300\nL 325 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 321 441\nL 325 441\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 318 317\nL 328 317\nL 328 425\nL 318 425\nL 318 317\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 336 259\nL 336 276\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 336 383\nL 336 408\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 334 259\nL 338 259\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 334 408\nL 338 408\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 331 276\nL 341 276\nL 341 383\nL 331 383\nL 331 276\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 350 259\nL 350 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 350 425\nL 350 441\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 348 259\nL 352 259\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 348 441\nL 352 441\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 345 300\nL 355 300\nL 355 425\nL 345 425\nL 345 300\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 363 259\nL 363 317\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 363 358\nL 363 383\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 361 259\nL 365 259\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 361 383\nL 365 383\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 358 317\nL 368 317\nL 368 358\nL 358 358\nL 358 317\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 377 300\nL 377 342\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 377 466\nL 377 508\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 375 300\nL 379 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 375 508\nL 379 508\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 372 342\nL 382 342\nL 382 466\nL 372 466\nL 372 342\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 390 383\nL 390 408\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 390 441\nL 390 508\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 388 383\nL 392 383\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 388 508\nL 392 508\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 385 408\nL 395 408\nL 395 441\nL 385 441\nL 385 408\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 404 218\nL 404 259\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 404 383\nL 404 425\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 402 218\nL 406 218\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 402 425\nL 406 425\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 399 259\nL 409 259\nL 409 383\nL 399 383\nL 399 259\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 417 408\nL 417 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 417 441\nL 417 450\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 415 408\nL 419 408\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 415 450\nL 419 450\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 412 425\nL 422 425\nL 422 441\nL 412 441\nL 412 425\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 431 176\nL 431 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 431 383\nL 431 400\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 429 176\nL 433 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 429 400\nL 433 400\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 426 218\nL 436 218\nL 436 383\nL 426 383\nL 426 218\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 444 176\nL 444 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 444 425\nL 444 441\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 442 176\nL 446 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 442 441\nL 446 441\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 439 218\nL 449 218\nL 449 425\nL 439 425\nL 439 218\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 458 135\nL 458 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 458 201\nL 458 209\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 456 135\nL 460 135\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 456 209\nL 460 209\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 453 176\nL 463 176\nL 463 201\nL 453 201\nL 453 176\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 471 209\nL 471 218\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 471 342\nL 471 383\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 469 209\nL 473 209\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 469 383\nL 473 383\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 466 218\nL 476 218\nL 476 342\nL 466 342\nL 466 218\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 485 300\nL 485 342\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 485 383\nL 485 425\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 483 300\nL 487 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 483 425\nL 487 425\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 480 342\nL 490 342\nL 490 383\nL 480 383\nL 480 342\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 498 276\nL 498 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 498 358\nL 498 375\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 496 276\nL 500 276\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 496 375\nL 500 375\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 493 300\nL 503 300\nL 503 358\nL 493 358\nL 493 300\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 512 218\nL 512 242\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 512 284\nL 512 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 510 218\nL 514 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 510 300\nL 514 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 507 242\nL 517 242\nL 517 284\nL 507 284\nL 507 242\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 525 160\nL 525 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 525 226\nL 525 242\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 523 160\nL 527 160\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 523 242\nL 527 242\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 520 176\nL 530 176\nL 530 226\nL 520 226\nL 520 176\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 539 135\nL 539 160\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 539 259\nL 539 276\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 537 135\nL 541 135\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 537 276\nL 541 276\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 534 160\nL 544 160\nL 544 259\nL 534 259\nL 534 160\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 553 176\nL 553 193\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 553 242\nL 553 276\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 551 176\nL 555 176\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 551 276\nL 555 276\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 548 193\nL 558 193\nL 558 242\nL 548 242\nL 548 193\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 566 242\nL 566 259\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 566 300\nL 566 342\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 564 242\nL 568 242\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 564 342\nL 568 342\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 561 259\nL 571 259\nL 571 300\nL 561 300\nL 561 259\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 580 284\nL 580 309\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 580 358\nL 580 383\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 578 284\nL 582 284\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 578 383\nL 582 383\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 575 309\nL 585 309\nL 585 358\nL 575 358\nL 575 309\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 593 342\nL 593 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 593 383\nL 593 392\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 591 342\nL 595 342\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 591 392\nL 595 392\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 588 358\nL 598 358\nL 598 383\nL 588 383\nL 588 358\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 607 300\nL 607 325\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 607 358\nL 607 383\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 605 300\nL 609 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 605 383\nL 609 383\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 602 325\nL 612 325\nL 612 358\nL 602 358\nL 602 325\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 620 300\nL 620 309\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 620 342\nL 620 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 618 300\nL 622 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 618 358\nL 622 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 615 309\nL 625 309\nL 625 342\nL 615 342\nL 615 309\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 634 176\nL 634 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 634 317\nL 634 325\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 632 176\nL 636 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 632 325\nL 636 325\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 629 300\nL 639 300\nL 639 317\nL 629 317\nL 629 300\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 647 259\nL 647 276\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 647 300\nL 647 317\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 645 259\nL 649 259\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 645 317\nL 649 317\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 642 276\nL 652 276\nL 652 300\nL 642 300\nL 642 276\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 661 135\nL 661 275\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 661 276\nL 661 284\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 659 135\nL 663 135\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 659 284\nL 663 284\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 656 275\nL 666 275\nL 666 276\nL 656 276\nL 656 275\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 674 259\nL 674 276\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 674 292\nL 674 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 672 259\nL 676 259\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 672 300\nL 676 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 669 276\nL 679 276\nL 679 292\nL 669 292\nL 669 276\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 688 292\nL 688 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 688 300\nL 688 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 686 292\nL 690 292\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 686 425\nL 690 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 683 300\nL 693 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 701 284\nL 701 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 701 358\nL 701 383\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 699 284\nL 703 284\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 699 383\nL 703 383\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 696 300\nL 706 300\nL 706 358\nL 696 358\nL 696 300\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 715 300\nL 715 317\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 715 342\nL 715 383\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 713 300\nL 717 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 713 383\nL 717 383\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 710 317\nL 720 317\nL 720 342\nL 710 342\nL 710 317\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 728 218\nL 728 234\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 728 342\nL 728 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 726 218\nL 730 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 726 358\nL 730 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 723 234\nL 733 234\nL 733 342\nL 723 342\nL 723 234\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path d=\"M 742 234\nL 742 242\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 742 292\nL 742 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 740 234\nL 744 234\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 740 300\nL 744 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 737 242\nL 747 242\nL 747 292\nL 737 292\nL 737 242\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 755 251\nL 755 259\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 755 325\nL 755 334\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 753 251\nL 757 251\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 753 334\nL 757 334\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 750 259\nL 760 259\nL 760 325\nL 750 325\nL 750 259\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 769 292\nL 769 300\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 769 375\nL 769 383\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 767 292\nL 771 292\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 767 383\nL 771 383\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path d=\"M 764 300\nL 774 300\nL 774 375\nL 764 375\nL 764 300\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path d=\"M 783 218\nL 783 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 783 358\nL 783 549\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 781 218\nL 785 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 781 549\nL 785 549\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 778 358\nL 788 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path d=\"M 84 437\nL 144 437\nL 144 437\nA 4 4 90.00 0 1 148 441\nL 148 454\nL 148 454\nA 4 4 90.00 0 1 144 458\nL 84 458\nL 84 458\nA 4 4 90.00 0 1 80 454\nL 80 441\nL 80 441\nA 4 4 90.00 0 1 84 437\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"84\" y=\"454\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><path d=\"M 125 409\nL 217 409\nL 217 409\nA 4 4 90.00 0 1 221 413\nL 221 465\nL 221 465\nA 4 4 90.00 0 1 217 469\nL 125 469\nL 125 469\nA 4 4 90.00 0 1 121 465\nL 121 413\nL 121 413\nA 4 4 90.00 0 1 125 409\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"125\" y=\"426\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"133\" y=\"439\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">† Gravestone</text><text x=\"154\" y=\"452\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"129\" y=\"465\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><path d=\"M 138 428\nL 198 428\nL 198 428\nA 4 4 90.00 0 1 202 432\nL 202 445\nL 202 445\nA 4 4 90.00 0 1 198 449\nL 138 449\nL 138 449\nA 4 4 90.00 0 1 134 445\nL 134 432\nL 134 432\nA 4 4 90.00 0 1 138 428\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"138\" y=\"445\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><path d=\"M 152 323\nL 244 323\nL 244 323\nA 4 4 90.00 0 1 248 327\nL 248 353\nL 248 353\nA 4 4 90.00 0 1 244 357\nL 152 357\nL 152 357\nA 4 4 90.00 0 1 148 353\nL 148 327\nL 148 327\nA 4 4 90.00 0 1 152 323\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"152\" y=\"340\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text><text x=\"152\" y=\"353\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">^ Bull Marubozu</text><path d=\"M 165 495\nL 262 495\nL 262 495\nA 4 4 90.00 0 1 266 499\nL 266 512\nL 266 512\nA 4 4 90.00 0 1 262 516\nL 165 516\nL 165 516\nA 4 4 90.00 0 1 161 512\nL 161 499\nL 161 499\nA 4 4 90.00 0 1 165 495\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"165\" y=\"512\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">v Bear Marubozu</text><path d=\"M 206 180\nL 297 180\nL 297 180\nA 4 4 90.00 0 1 301 184\nL 301 197\nL 301 197\nA 4 4 90.00 0 1 297 201\nL 206 201\nL 206 201\nA 4 4 90.00 0 1 202 197\nL 202 184\nL 202 184\nA 4 4 90.00 0 1 206 180\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"206\" y=\"197\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text><path d=\"M 233 437\nL 328 437\nL 328 437\nA 4 4 90.00 0 1 332 441\nL 332 454\nL 332 454\nA 4 4 90.00 0 1 328 458\nL 233 458\nL 233 458\nA 4 4 90.00 0 1 229 454\nL 229 441\nL 229 441\nA 4 4 90.00 0 1 233 437\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"233\" y=\"454\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">V Bear Engulfing</text><path d=\"M 314 345\nL 395 345\nL 395 345\nA 4 4 90.00 0 1 399 349\nL 399 362\nL 399 362\nA 4 4 90.00 0 1 395 366\nL 314 366\nL 314 366\nA 4 4 90.00 0 1 310 362\nL 310 349\nL 310 349\nA 4 4 90.00 0 1 314 345\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"314\" y=\"362\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">| Piercing Line</text><path d=\"M 341 370\nL 415 370\nL 415 370\nA 4 4 90.00 0 1 419 374\nL 419 387\nL 419 387\nA 4 4 90.00 0 1 415 391\nL 341 391\nL 341 391\nA 4 4 90.00 0 1 337 387\nL 337 374\nL 337 374\nA 4 4 90.00 0 1 341 370\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"341\" y=\"387\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ξ Dark Cloud</text><path d=\"M 355 287\nL 436 287\nL 436 287\nA 4 4 90.00 0 1 440 291\nL 440 304\nL 440 304\nA 4 4 90.00 0 1 436 308\nL 355 308\nL 355 308\nA 4 4 90.00 0 1 351 304\nL 351 291\nL 351 291\nA 4 4 90.00 0 1 355 287\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"355\" y=\"304\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">| Piercing Line</text><path d=\"M 436 205\nL 520 205\nL 520 205\nA 4 4 90.00 0 1 524 209\nL 524 222\nL 524 222\nA 4 4 90.00 0 1 520 226\nL 436 226\nL 436 226\nA 4 4 90.00 0 1 432 222\nL 432 209\nL 432 209\nA 4 4 90.00 0 1 436 205\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"436\" y=\"222\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">* Morning Star</text><path d=\"M 476 329\nL 558 329\nL 558 329\nA 4 4 90.00 0 1 562 333\nL 562 346\nL 562 346\nA 4 4 90.00 0 1 558 350\nL 476 350\nL 476 350\nA 4 4 90.00 0 1 472 346\nL 472 333\nL 472 333\nA 4 4 90.00 0 1 476 329\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"476\" y=\"346\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">⁎ Evening Star</text><path d=\"M 639 281\nL 731 281\nL 731 281\nA 4 4 90.00 0 1 735 285\nL 735 311\nL 735 311\nA 4 4 90.00 0 1 731 315\nL 639 315\nL 639 315\nA 4 4 90.00 0 1 635 311\nL 635 285\nL 635 285\nA 4 4 90.00 0 1 639 281\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"639\" y=\"298\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"643\" y=\"311\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><path d=\"M 666 243\nL 758 243\nL 758 243\nA 4 4 90.00 0 1 762 247\nL 762 299\nL 762 299\nA 4 4 90.00 0 1 758 303\nL 666 303\nL 666 303\nA 4 4 90.00 0 1 662 299\nL 662 247\nL 662 247\nA 4 4 90.00 0 1 666 243\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"666\" y=\"260\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"674\" y=\"273\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">† Gravestone</text><text x=\"695\" y=\"286\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"670\" y=\"299\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><path d=\"M 693 274\nL 761 274\nL 761 274\nA 4 4 90.00 0 1 765 278\nL 765 317\nL 765 317\nA 4 4 90.00 0 1 761 321\nL 693 321\nL 693 321\nA 4 4 90.00 0 1 689 317\nL 689 278\nL 689 278\nA 4 4 90.00 0 1 693 274\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"697\" y=\"291\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"693\" y=\"304\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ψ Dragonfly</text><text x=\"710\" y=\"317\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><path d=\"M 767 345\nL 800 345\nL 800 345\nA 4 4 90.00 0 1 804 349\nL 804 362\nL 804 362\nA 4 4 90.00 0 1 800 366\nL 767 366\nL 767 366\nA 4 4 90.00 0 1 763 362\nL 763 349\nL 763 349\nA 4 4 90.00 0 1 767 345\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"767\" y=\"362\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text></svg>",
			pngCRC: 0xece44fd5,
		},
	}

	for i, tc := range tests {
		t.Run(strconv.Itoa(i)+"-"+tc.name, func(t *testing.T) {
			p := NewPainter(PainterOptions{
				OutputFormat: ChartOutputSVG,
				Width:        800,
				Height:       600,
			})
			r := NewPainter(PainterOptions{
				OutputFormat: ChartOutputPNG,
				Width:        800,
				Height:       600,
			})

			opt := tc.optGen()
			opt.Theme = GetTheme(ThemeVividLight)

			validateCandlestickChartRender(t, p, r, opt, tc.svg, tc.pngCRC)
		})
	}
}

func TestCandlestickPatternConfigMergePatterns(t *testing.T) {
	t.Parallel()

	t.Run("merge_two_configs", func(t *testing.T) {
		config1 := &CandlestickPatternConfig{
			PreferPatternLabels: true,
			EnabledPatterns:     []string{"doji", "hammer"},
			DojiThreshold:       0.01,
		}
		config2 := &CandlestickPatternConfig{
			PreferPatternLabels: false,
			EnabledPatterns:     []string{"shooting_star", "doji"}, // Doji is duplicate
			DojiThreshold:       0.02,
		}

		merged := config1.MergePatterns(config2)

		// Should preserve config1's settings
		assert.True(t, merged.PreferPatternLabels)
		assert.InDelta(t, 0.01, merged.DojiThreshold, 0)

		// Should have union of patterns without duplicates, preserving order
		assert.Len(t, merged.EnabledPatterns, 3)
		assert.Equal(t, "doji", merged.EnabledPatterns[0])
		assert.Equal(t, "hammer", merged.EnabledPatterns[1])
		assert.Equal(t, "shooting_star", merged.EnabledPatterns[2])
	})

	t.Run("merge_with_nil", func(t *testing.T) {
		config := &CandlestickPatternConfig{
			PreferPatternLabels: true,
			EnabledPatterns:     []string{"doji", "hammer"},
		}

		// Merge nil with config
		var nilConfig *CandlestickPatternConfig
		merged1 := nilConfig.MergePatterns(config)
		assert.NotNil(t, merged1)
		assert.True(t, merged1.PreferPatternLabels)
		assert.Len(t, merged1.EnabledPatterns, 2)

		// Merge config with nil
		merged2 := config.MergePatterns(nil)
		assert.NotNil(t, merged2)
		assert.True(t, merged2.PreferPatternLabels)
		assert.Len(t, merged2.EnabledPatterns, 2)

		// Merge nil with nil
		merged3 := nilConfig.MergePatterns(nil)
		assert.Nil(t, merged3)
	})

	t.Run("merge_identical_patterns", func(t *testing.T) {
		config1 := &CandlestickPatternConfig{
			EnabledPatterns: []string{"doji", "hammer", "shooting_star"},
		}
		config2 := &CandlestickPatternConfig{
			EnabledPatterns: []string{"doji", "hammer", "shooting_star"},
		}

		merged := config1.MergePatterns(config2)
		assert.Len(t, merged.EnabledPatterns, 3) // No duplicates
		assert.Equal(t, "doji", merged.EnabledPatterns[0])
		assert.Equal(t, "hammer", merged.EnabledPatterns[1])
		assert.Equal(t, "shooting_star", merged.EnabledPatterns[2])
	})

	t.Run("merge_empty_patterns", func(t *testing.T) {
		config1 := &CandlestickPatternConfig{
			PreferPatternLabels: true,
			EnabledPatterns:     []string{},
		}
		config2 := &CandlestickPatternConfig{
			EnabledPatterns: []string{"doji", "hammer"},
		}

		merged := config1.MergePatterns(config2)
		assert.True(t, merged.PreferPatternLabels)
		assert.Len(t, merged.EnabledPatterns, 2)
		assert.Equal(t, "doji", merged.EnabledPatterns[0])
		assert.Equal(t, "hammer", merged.EnabledPatterns[1])
	})

	t.Run("merge_predefined_configs", func(t *testing.T) {
		core := (&CandlestickPatternConfig{}).WithPatternsCore()
		trend := (&CandlestickPatternConfig{}).WithPatternsTrend()

		merged := core.MergePatterns(trend)

		// Should have all patterns from both configs
		assert.Len(t, merged.EnabledPatterns, len(core.EnabledPatterns)+len(trend.EnabledPatterns))

		// Should preserve core config's settings
		assert.Equal(t, core.PreferPatternLabels, merged.PreferPatternLabels)

		// Should contain patterns from both
		assert.Contains(t, merged.EnabledPatterns, "engulfing_bull") // From core
		assert.Contains(t, merged.EnabledPatterns, "marubozu_bull")  // From Trend
	})
}
