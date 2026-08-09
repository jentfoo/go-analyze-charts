package drawing

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/image/math/fixed"
)

type mockAdder struct {
	starts []fixed.Point26_6
	adds   []fixed.Point26_6
}

func (m *mockAdder) Start(p fixed.Point26_6)      { m.starts = append(m.starts, p) }
func (m *mockAdder) Add1(p fixed.Point26_6)       { m.adds = append(m.adds, p) }
func (m *mockAdder) Add2(b, c fixed.Point26_6)    {}
func (m *mockAdder) Add3(b, c, d fixed.Point26_6) {}

func TestFtLineBuilderMoveToLineTo(t *testing.T) {
	t.Parallel()

	ad := &mockAdder{}
	ft := FtLineBuilder{Adder: ad}
	ft.MoveTo(1, 2)
	ft.LineTo(3, 4)
	ft.End()

	if assert.Len(t, ad.starts, 1) {
		assert.Equal(t, fixed.Int26_6(64), ad.starts[0].X)
		assert.Equal(t, fixed.Int26_6(128), ad.starts[0].Y)
	}
	if assert.Len(t, ad.adds, 1) {
		assert.Equal(t, fixed.Int26_6(192), ad.adds[0].X)
		assert.Equal(t, fixed.Int26_6(256), ad.adds[0].Y)
	}
}

func TestToFixed26_6(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   float64
		want fixed.Int26_6
	}{
		{"zero", 0, 0},
		{"whole", 3, 192},
		{"exact_sixty_fourth", 0.015625, 1},
		{"half_rounds_away", 0.0234375, 2},            // 1.5 in 26.6
		{"half_not_to_even", 0.0703125, 5},            // 4.5 in 26.6, away from zero not to even
		{"half_rounds_away_negative", -0.0234375, -2}, // -1.5 in 26.6, away from zero not up
		{"negative", -2.5, -160},
		{"at_bound", maxRasterPixel, maxRasterPixel * 64},
		{"within_bound", maxRasterPixel - 1, (maxRasterPixel - 1) * 64},
		{"overflow_positive", 1e9, maxRasterPixel * 64},
		{"overflow_negative", -1e9, -maxRasterPixel * 64},
		{"positive_inf", math.Inf(1), maxRasterPixel * 64},
		{"negative_inf", math.Inf(-1), -maxRasterPixel * 64},
		{"nan", math.NaN(), -maxRasterPixel * 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, toFixed26_6(tt.in))
		})
	}
}

func TestFtLineBuilderClamps(t *testing.T) {
	t.Parallel()

	// an unclamped conversion wraps huge and NaN coordinates to a large negative position
	ad := &mockAdder{}
	ft := FtLineBuilder{Adder: ad}
	ft.MoveTo(1e12, math.NaN())
	ft.LineTo(-1e12, math.Inf(1))

	if assert.Len(t, ad.starts, 1) {
		assert.Equal(t, fixed.Int26_6(maxRasterPixel*64), ad.starts[0].X)
		assert.Equal(t, fixed.Int26_6(-maxRasterPixel*64), ad.starts[0].Y)
	}
	if assert.Len(t, ad.adds, 1) {
		assert.Equal(t, fixed.Int26_6(-maxRasterPixel*64), ad.adds[0].X)
		assert.Equal(t, fixed.Int26_6(maxRasterPixel*64), ad.adds[0].Y)
	}
}
