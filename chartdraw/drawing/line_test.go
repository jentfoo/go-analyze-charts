package drawing

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBresenhamDiagonal(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 5, 5))
	Bresenham(img, color.White, 0, 0, 4, 4)

	for i := 0; i <= 4; i++ {
		r, g, b, a := img.At(i, i).RGBA()
		assert.Equal(t, uint32(0xffff), r)
		assert.Equal(t, uint32(0xffff), g)
		assert.Equal(t, uint32(0xffff), b)
		assert.Equal(t, uint32(0xffff), a)
	}

	_, _, _, a := img.At(0, 1).RGBA()
	assert.Equal(t, uint32(0), a)
}

func TestPolylineBresenham(t *testing.T) {
	t.Parallel()

	t.Run("positive_coords", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 5, 5))
		PolylineBresenham(img, color.White, 0, 0, 2, 0, 2, 2)

		expected := [][2]int{{0, 0}, {1, 0}, {2, 0}, {2, 1}, {2, 2}}
		for _, p := range expected {
			_, _, _, a := img.At(p[0], p[1]).RGBA()
			assert.Equal(t, uint32(0xffff), a)
		}

		_, _, _, a := img.At(1, 1).RGBA()
		assert.Equal(t, uint32(0), a)
	})

	t.Run("negative_fractional_start", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 6, 6))
		PolylineBresenham(img, color.White, -1.6, 0.4, 4, 2)

		// -1.6 rounds to -2, truncation toward zero would paint (0,0) instead of (0,1)
		_, _, _, a := img.At(0, 1).RGBA()
		assert.Equal(t, uint32(0xffff), a)
		_, _, _, a = img.At(0, 0).RGBA()
		assert.Equal(t, uint32(0), a)
	})

	t.Run("non_finite_skipped", func(t *testing.T) {
		// converting a non-finite value to an int yields MinInt, which would walk the int space
		coords := []float64{math.NaN(), math.Inf(1), math.Inf(-1)}

		for _, c := range coords {
			img := image.NewRGBA(image.Rect(0, 0, 5, 5))
			PolylineBresenham(img, color.White, 0, 0, c, 2, 4, 4)

			assert.True(t, imageEmpty(img))
		}
	})

	t.Run("out_of_int_range_skipped", func(t *testing.T) {
		// 1e300 converts to MinInt, straddling the image so a bounds check can not reject it
		img := image.NewRGBA(image.Rect(0, 0, 5, 5))
		PolylineBresenham(img, color.White, 0, 0, 1e300, 1e300)

		assert.True(t, imageEmpty(img))
	})

	t.Run("outside_bounds_skipped", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 5, 5))
		PolylineBresenham(img, color.White, 1e18, 1e18, 2e18, 2e18)

		assert.True(t, imageEmpty(img))
	})

	t.Run("partially_visible_drawn", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 5, 5))
		PolylineBresenham(img, color.White, -10, 0, 10, 0)

		for x := 0; x < 5; x++ {
			_, _, _, a := img.At(x, 0).RGBA()
			assert.Equal(t, uint32(0xffff), a)
		}
	})

	t.Run("extreme_finite_clipped", func(t *testing.T) {
		// walking this unclipped would step 1e15 times, painting five pixels
		img := image.NewRGBA(image.Rect(0, 0, 5, 5))
		PolylineBresenham(img, color.White, 0, 0, 1e15, 0)

		for x := 0; x < 5; x++ {
			_, _, _, a := img.At(x, 0).RGBA()
			assert.Equal(t, uint32(0xffff), a)
		}
		_, _, _, a := img.At(0, 1).RGBA()
		assert.Equal(t, uint32(0), a)
	})
}

func TestBresenhamClipped(t *testing.T) {
	t.Parallel()

	t.Run("int_range_endpoints", func(t *testing.T) {
		// x1-x0 overflows int, corrupting the deltas the loop exits on
		img := image.NewRGBA(image.Rect(0, 0, 5, 5))
		Bresenham(img, color.White, math.MinInt+1, 0, math.MaxInt-1, 0)

		for x := 0; x < 5; x++ {
			_, _, _, a := img.At(x, 0).RGBA()
			assert.Equal(t, uint32(0xffff), a)
		}
		_, _, _, a := img.At(0, 1).RGBA()
		assert.Equal(t, uint32(0), a)
	})

	t.Run("long_miss_skipped", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 5, 5))
		Bresenham(img, color.White, 10, -1e18, 10, 1e18)

		assert.True(t, imageEmpty(img))
	})

	t.Run("long_diagonal_entering", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 5, 5))
		Bresenham(img, color.White, -(1 << 21), -(1 << 21), 4, 4)

		for i := 0; i <= 4; i++ {
			_, _, _, a := img.At(i, i).RGBA()
			assert.Equal(t, uint32(0xffff), a)
		}
		_, _, _, a := img.At(0, 1).RGBA()
		assert.Equal(t, uint32(0), a)
	})

	t.Run("above_bound_in_bounds_intact", func(t *testing.T) {
		// clipping a segment already in bounds is a no-op, however long it is
		const w = 1<<20 + 8
		img := image.NewRGBA(image.Rect(0, 0, w, 1))
		Bresenham(img, color.White, 0, 0, w-1, 0)

		for _, x := range []int{0, 1, w / 2, w - 2, w - 1} {
			_, _, _, a := img.At(x, 0).RGBA()
			assert.Equal(t, uint32(0xffff), a)
		}
	})

	t.Run("below_step_bound_walked", func(t *testing.T) {
		// walked as given rather than clipped, holding the pixels a partial line has always painted
		img := image.NewRGBA(image.Rect(0, 0, 5, 5))
		Bresenham(img, color.White, 4-maxBresenhamSteps, 0, 4, 0)

		for x := 0; x < 5; x++ {
			_, _, _, a := img.At(x, 0).RGBA()
			assert.Equal(t, uint32(0xffff), a)
		}
	})

	t.Run("offset_bounds_respected", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(2, 2, 5, 5))
		Bresenham(img, color.White, -1e18, 3, 1e18, 3)

		for x := 2; x < 5; x++ {
			_, _, _, a := img.At(x, 3).RGBA()
			assert.Equal(t, uint32(0xffff), a)
		}
	})
}

func TestClipSegment(t *testing.T) {
	t.Parallel()

	b := image.Rect(0, 0, 5, 5)
	tests := []struct {
		name string
		in   [4]float64
		want [4]float64
		ok   bool
	}{
		{"fully_inside", [4]float64{1, 1, 3, 3}, [4]float64{1, 1, 3, 3}, true},
		{"straddling_both_ends", [4]float64{-10, 0, 10, 0}, [4]float64{0, 0, 4, 0}, true},
		{"extreme_finite_end", [4]float64{0, 0, 1e15, 0}, [4]float64{0, 0, 4, 0}, true},
		{"diagonal_entering", [4]float64{-4, -4, 4, 4}, [4]float64{0, 0, 4, 4}, true},
		{"vertical_inside", [4]float64{2, -3, 2, 9}, [4]float64{2, 0, 2, 4}, true},
		{"vertical_outside", [4]float64{9, -3, 9, 9}, [4]float64{}, false},
		{"horizontal_outside", [4]float64{-3, 9, 9, 9}, [4]float64{}, false},
		{"fully_outside", [4]float64{6, 6, 9, 9}, [4]float64{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x0, y0, x1, y1, ok := clipSegment(b, tt.in[0], tt.in[1], tt.in[2], tt.in[3])
			got := []float64{x0, y0, x1, y1}

			assert.Equal(t, tt.ok, ok)
			assert.InDeltaSlice(t, tt.want[:], got, 0.0001)
		})
	}

	t.Run("empty_rectangle", func(t *testing.T) {
		_, _, _, _, ok := clipSegment(image.Rect(0, 0, 0, 0), -1, 0, 1, 0)

		assert.False(t, ok)
	})
}

func imageEmpty(img *image.RGBA) bool {
	b := img.Bounds()
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			if _, _, _, a := img.At(x, y).RGBA(); a != 0 {
				return false
			}
		}
	}
	return true
}
