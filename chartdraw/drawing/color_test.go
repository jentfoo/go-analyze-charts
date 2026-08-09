package drawing

import (
	"image/color"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorFromHex(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected Color
	}{
		{name: "white", input: "FFFFFF", expected: ColorWhite},
		{name: "short_white", input: "FFF", expected: ColorWhite},
		{name: "black", input: "000000", expected: ColorBlack},
		{name: "short_black", input: "000", expected: ColorBlack},
		{name: "red", input: "FF0000", expected: ColorRed},
		{name: "short_red", input: "F00", expected: ColorRed},
		{name: "green", input: "008000", expected: ColorGreen},
		{name: "blue", input: "0000FF", expected: ColorBlue},
		{name: "short_blue", input: "00F", expected: ColorBlue},
		{name: "with_hash", input: "#FF0000", expected: ColorRed},
		{name: "empty_string", input: "", expected: Color{}},
		{name: "hash_only", input: "#", expected: Color{}},
		{name: "one_digit", input: "F", expected: Color{}},
		{name: "two_digits", input: "FF", expected: Color{}},
		{name: "five_digits", input: "FFFFF", expected: Color{}},
		{name: "seven_digits", input: "FFFFFFF", expected: Color{}},
		{name: "four_digit_rgba", input: "F00A", expected: Color{R: 255, A: 0xAA}},
		{name: "four_digit_opaque", input: "F00F", expected: ColorRed},
		{name: "eight_digit_rgba", input: "#FF000080", expected: Color{R: 255, A: 0x80}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ColorFromHex(tc.input))
		})
	}
}

func TestColorFromAlphaMixedRGBA(t *testing.T) {
	t.Parallel()

	t.Run("known_colors", func(t *testing.T) {
		black := ColorFromAlphaMixedRGBA(color.Black.RGBA())
		assert.True(t, black.Equals(ColorBlack), black.String())

		white := ColorFromAlphaMixedRGBA(color.White.RGBA())
		assert.True(t, white.Equals(ColorWhite), white.String())
	})
	t.Run("alpha16_midpoint", func(t *testing.T) {
		c := ColorFromAlphaMixedRGBA(color.Alpha16{A: 0x7FFF}.RGBA())
		assert.Equal(t, uint8(127), c.A)
	})
	t.Run("transparent_input", func(t *testing.T) {
		assert.Equal(t, Color{}, ColorFromAlphaMixedRGBA(color.Transparent.RGBA()))
		assert.Equal(t, Color{}, ColorFromAlphaMixedRGBA(0, 0, 0, 0))
	})
	t.Run("round_trip", func(t *testing.T) {
		for _, expected := range []Color{
			{R: 10, G: 20, B: 30, A: 255}, {R: 84, G: 112, B: 198, A: 255},
			{R: 10, G: 20, B: 30, A: 200}, {R: 84, G: 112, B: 198, A: 204},
			{R: 255, G: 255, B: 255, A: 180}, {R: 200, G: 100, B: 50, A: 128},
			{R: 1, G: 2, B: 3, A: 64},
		} {
			assert.Equal(t, expected, ColorFromAlphaMixedRGBA(expected.RGBA()))
		}
	})
}

func Test_ColorFromRGBA(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected Color
	}{
		{name: "spaced_alpha_one", input: "rgba(192, 192, 192, 1.0)", expected: ColorSilver},
		{name: "alpha_one", input: "rgba(192,192,192,1.0)", expected: ColorSilver},
		{name: "alpha_integer", input: "rgba(192,192,192,255)", expected: ColorSilver},
		{name: "alpha_fraction", input: "rgba(192,192,192,.981)", expected: ColorSilver.WithAlpha(250)},
		{name: "alpha_half", input: "rgba(192,192,192,0.5)", expected: ColorSilver.WithAlpha(128)},
		{name: "channel_above_range", input: "rgb(999,0,0)", expected: ColorRed},
		{name: "channel_way_above_range", input: "rgb(40000,0,0)", expected: ColorRed},
		{name: "negative_channel", input: "rgb(-5,0,0)", expected: ColorBlack},
		{name: "negative_channel_way_below_range", input: "rgb(-40000,0,0)", expected: ColorBlack},
		{name: "alpha_above_255", input: "rgba(0,0,0,300)", expected: ColorBlack},
		{name: "negative_alpha", input: "rgba(0,0,0,-0.5)", expected: Color{}},
		{name: "nan_alpha", input: "rgba(0,0,0,NaN)", expected: Color{}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ColorFromRGBA(tc.input))
		})
	}
}

func TestParseColor(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		Input    string
		Expected Color
	}{
		{"", Color{}},
		{"unknown", ColorBlack},
		{"white", ColorWhite},
		{"WHITE", ColorWhite}, // caps!
		{"black", ColorBlack},
		{"red", ColorRed},
		{"gray", ColorGray},
		{"grey", ColorGray},
		{"green", ColorGreen},
		{"blue", ColorBlue},
		{"silver", ColorSilver},
		{"maroon", ColorMaroon},
		{"purple", ColorPurple},
		{"fuchsia", ColorFuchsia},
		{"lime", ColorLime},
		{"olive", ColorOlive},
		{"yellow", ColorYellow},
		{"navy", ColorNavy},
		{"teal", ColorTeal},
		{"aqua", ColorAqua},

		{"rgba(192, 192, 192, 1.0)", ColorSilver},
		{"rgba(192,192,192,1.0)", ColorSilver},
		{"rgb(192, 192, 192)", ColorSilver},
		{"rgb(192,192,192)", ColorSilver},

		{"#FF0000", ColorRed},
		{"#008000", ColorGreen},
		{"#0000FF", ColorBlue},
		{"#F00", ColorRed},
		{"#080", Color{0, 136, 0, 255}},
		{"#00F", ColorBlue},
		{"#ab", Color{}},
		{"#FF000080", Color{R: 255, A: 128}},
	}

	for index, tc := range testCases {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			actual := ParseColor(tc.Input)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}
func TestColorHelperMethods(t *testing.T) {
	t.Parallel()

	chTests := []struct {
		f      float64
		expect uint8
	}{
		{-0.1, 0},
		{0.5, 128},
		{1.5, 255},
	}
	for i, tc := range chTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tc.expect, ColorChannelFromFloat(tc.f))
		})
	}

	c := Color{R: 10, G: 20, B: 30, A: 255}
	r, g, b, a := c.RGBA()
	assert.Equal(t, uint32(2570), r)
	assert.Equal(t, uint32(5140), g)
	assert.Equal(t, uint32(7710), b)
	assert.Equal(t, uint32(65535), a)

	zero := Color{}
	assert.True(t, zero.IsZero())
	assert.True(t, zero.IsTransparent())
	assert.False(t, c.IsZero())

	withAlpha := c.WithAlpha(128)
	assert.Equal(t, uint8(128), withAlpha.A)

	avg := ColorRed.AverageWith(ColorBlue)
	assert.Equal(t, Color{R: 127, G: 0, B: 127, A: 255}, avg)
	assert.Equal(t, ColorWhite, ColorWhite.AverageWith(ColorWhite))
	assert.Equal(t, uint8(150), Color{R: 200}.AverageWith(Color{R: 100}).R)

	assert.Equal(t, "rgb(10,20,30)", c.StringRGB())
	assert.Equal(t, "rgba(10,20,30,0.502)", c.WithAlpha(128).StringRGBA())
}

func TestColorRGBA(t *testing.T) {
	t.Parallel()

	for _, c := range []Color{
		{R: 10, G: 20, B: 30, A: 255}, {R: 10, G: 20, B: 30, A: 200},
		{R: 200, G: 100, B: 50, A: 128}, {R: 255, G: 255, B: 255, A: 1},
		{R: 84, G: 112, B: 198, A: 0},
	} {
		r, g, b, a := c.RGBA()
		er, eg, eb, ea := color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A}.RGBA()
		assert.Equal(t, []uint32{er, eg, eb, ea}, []uint32{r, g, b, a})
	}
}

func TestColorStringRGBA(t *testing.T) {
	t.Parallel()

	c := Color{R: 10, G: 20, B: 30}

	t.Run("low_alpha_visible", func(t *testing.T) {
		assert.Equal(t, "rgba(10,20,30,0.039)", c.WithAlpha(10).StringRGBA())
	})
	t.Run("trimmed_output", func(t *testing.T) {
		assert.Equal(t, "rgba(10,20,30,0.8)", c.WithAlpha(204).StringRGBA())
	})
	t.Run("zero_alpha", func(t *testing.T) {
		assert.Equal(t, "rgba(10,20,30,0)", c.StringRGBA())
	})
	t.Run("opaque_alpha", func(t *testing.T) {
		opaque := c.WithAlpha(255)
		assert.Equal(t, "rgba(10,20,30,1)", opaque.StringRGBA())
		assert.Equal(t, "rgb(10,20,30)", opaque.String())
	})
	t.Run("distinct_alphas", func(t *testing.T) {
		seen := make(map[string]struct{}, 256)
		for a := 0; a < 256; a++ {
			seen[c.WithAlpha(uint8(a)).StringRGBA()] = struct{}{}
		}
		assert.Len(t, seen, 256)
	})
}

func TestColorHSLRoundTrip(t *testing.T) {
	t.Parallel()

	// rounded channel conversion makes RGB -> HSL -> RGB an exact identity
	for r := 0; r < 256; r += 5 {
		for g := 0; g < 256; g += 7 {
			for b := 0; b < 256; b += 11 {
				c := Color{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
				assert.Equal(t, c, c.WithAdjustHSL(0, 0, 0))
			}
		}
	}
}

func TestColorHSLConversions(t *testing.T) {
	t.Parallel()

	h, s, l := ColorRed.HSL()
	assert.InDelta(t, 0.0, h, 0.001)
	assert.InDelta(t, 1.0, s, 0.001)
	assert.InDelta(t, 0.5, l, 0.001)

	r, g, b := hslToRGB(h, s, l)
	assert.Equal(t, ColorRed.R, r)
	assert.Equal(t, ColorRed.G, g)
	assert.Equal(t, ColorRed.B, b)

	adjusted := ColorRed.WithAdjustHSL(120, 0, 0)
	assert.Equal(t, ColorLime.R, adjusted.R)
	assert.Equal(t, ColorLime.G, adjusted.G)
	assert.Equal(t, ColorLime.B, adjusted.B)

	assert.InDelta(t, 0.25, clamp(0.25, 0, 1), 0.0)
	assert.InDelta(t, 0.0, clamp(-0.5, 0, 1), 0.0)
	assert.InDelta(t, 1.0, clamp(2, 0, 1), 0.0)

	assert.InDelta(t, 0.0, hue2rgb(0, 1, 0), 0.0001)
	assert.InDelta(t, 1.0, hue2rgb(0, 1, 1.0/6.0), 0.0001)
}
