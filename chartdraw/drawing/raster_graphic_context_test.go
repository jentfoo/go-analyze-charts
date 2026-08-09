package drawing

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/golang/freetype/raster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/draw"
	"golang.org/x/image/math/fixed"
)

func TestRasterGraphicContext(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		t.Parallel()

		img := image.NewRGBA(image.Rect(0, 0, 3, 3))
		rgc := NewRasterGraphicContext(img)
		assert.InDelta(t, DefaultDPI, rgc.GetDPI(), 0.0)
		rgc.SetDPI(72)
		assert.InDelta(t, 72.0, rgc.GetDPI(), 0.0)
	})

	t.Run("matrix_ops", func(t *testing.T) {
		t.Parallel()

		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		ctx := NewRasterGraphicContext(img)

		originalMatrix := ctx.GetMatrixTransform()
		assert.NotNil(t, originalMatrix)

		identityMatrix := NewIdentityMatrix()
		ctx.SetMatrixTransform(identityMatrix)
		currentMatrix := ctx.GetMatrixTransform()
		assert.Equal(t, identityMatrix, currentMatrix)

		ctx.Translate(10, 20)
		ctx.Scale(2, 3)
		ctx.Rotate(0.5)

		// The matrix should have changed
		transformedMatrix := ctx.GetMatrixTransform()
		assert.NotEqual(t, identityMatrix, transformedMatrix)
	})

	t.Run("font_ops", func(t *testing.T) {
		t.Parallel()

		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		ctx := NewRasterGraphicContext(img)

		ctx.SetFontSize(12)
		fontSize := ctx.GetFontSize()
		assert.InDelta(t, 12.0, fontSize, 0.001)

		ctx.SetFontSize(24)
		fontSize = ctx.GetFontSize()
		assert.InDelta(t, 24.0, fontSize, 0.001)

		originalFont := ctx.GetFont()
		ctx.SetFont(nil)
		currentFont := ctx.GetFont()
		assert.Nil(t, currentFont)

		// Restore original font if it existed
		if originalFont != nil {
			ctx.SetFont(originalFont)
			restoredFont := ctx.GetFont()
			assert.Equal(t, originalFont, restoredFont)
		}
	})

	t.Run("dpi_ops", func(t *testing.T) {
		t.Parallel()

		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		ctx := NewRasterGraphicContext(img)

		ctx.SetDPI(72.0)
		dpi := ctx.GetDPI()
		assert.InDelta(t, 72.0, dpi, 0.001)

		ctx.SetDPI(300.0)
		dpi = ctx.GetDPI()
		assert.InDelta(t, 300.0, dpi, 0.001)
	})

	t.Run("save_restore", func(t *testing.T) {
		t.Parallel()

		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		ctx := NewRasterGraphicContext(img)

		// Set initial state
		ctx.SetLineWidth(5)
		ctx.SetFontSize(16)
		ctx.Translate(10, 20)

		// Capture the state before saving
		expectedMatrix := ctx.GetMatrixTransform()
		expectedFontSize := ctx.GetFontSize()

		ctx.Save()

		// Modify state
		ctx.SetLineWidth(10)
		ctx.SetFontSize(32)
		ctx.Translate(30, 40)

		// Verify state was modified
		modifiedMatrix := ctx.GetMatrixTransform()
		modifiedFontSize := ctx.GetFontSize()
		assert.NotEqual(t, expectedMatrix, modifiedMatrix)
		assert.NotEqual(t, expectedFontSize, modifiedFontSize)

		// Restore the state
		ctx.Restore()

		// Validate the restore state - should exactly match the saved state
		restoredMatrix := ctx.GetMatrixTransform()
		restoredFontSize := ctx.GetFontSize()
		assert.Equal(t, expectedMatrix, restoredMatrix)
		assert.InDelta(t, expectedFontSize, restoredFontSize, 0.001)
	})

	t.Run("text_ops", func(t *testing.T) {
		t.Parallel()

		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		ctx := NewRasterGraphicContext(img)
		font := getTestFont(t)
		ctx.SetFont(font)
		ctx.SetFontSize(12)

		left, top, right, bottom, err := ctx.GetStringBounds("Hello")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, right, left)
		assert.GreaterOrEqual(t, bottom, top)

		cursor, err := ctx.CreateStringPath("Test", 10, 20)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, cursor, 0.0)

		cursor, err = ctx.CreateStringPath("World", 30, 40)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, cursor, 0.0)
		ctx.Fill()

		cursor, err = ctx.CreateStringPath("StrokeAt", 50, 60)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, cursor, 0.0)
		ctx.Stroke()
	})

	t.Run("fill_rect_clear", func(t *testing.T) {
		t.Parallel()

		img := image.NewRGBA(image.Rect(0, 0, 2, 2))
		rgc := NewRasterGraphicContext(img)
		rgc.SetFillColor(color.RGBA{255, 0, 0, 255})
		rgc.FillRect(0, 0, 2, 2)
		_, _, _, a := img.At(1, 1).RGBA()
		assert.Equal(t, uint32(0xffff), a)

		rgc.Clear()
		_, _, _, a = img.At(1, 1).RGBA()
		assert.Equal(t, uint32(0), a)
	})

	t.Run("clear_offset_bounds", func(t *testing.T) {
		t.Parallel()

		img := image.NewRGBA(image.Rect(10, 10, 12, 12))
		rgc := NewRasterGraphicContext(img)
		rgc.SetFillColor(color.RGBA{255, 0, 0, 255})
		rgc.FillRect(10, 10, 12, 12)
		_, _, _, a := img.At(11, 11).RGBA()
		assert.Equal(t, uint32(0xffff), a)

		rgc.Clear()
		_, _, _, a = img.At(11, 11).RGBA()
		assert.Equal(t, uint32(0), a)
	})

	t.Run("text_offset_bounds", func(t *testing.T) {
		t.Parallel()

		img := image.NewRGBA(image.Rect(20, 20, 80, 80))
		rgc := NewRasterGraphicContext(img)
		rgc.SetFont(getTestFont(t))
		rgc.SetFontSize(10)
		rgc.SetFillColor(color.White)

		// placed past the image width and height so only absolute span coordinates reach it
		_, err := rgc.CreateStringPath("A", 60, 75)
		require.NoError(t, err)
		rgc.Fill()

		var found bool
		for y := 20; y < 80 && !found; y++ {
			for x := 20; x < 80 && !found; x++ {
				found = img.RGBAAt(x, y).A != 0
			}
		}
		assert.True(t, found)
	})

	t.Run("draw_image_offset_bounds", func(t *testing.T) {
		t.Parallel()

		src := image.NewRGBA(image.Rect(0, 0, 2, 2))
		draw.Draw(src, src.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
		dst := image.NewRGBA(image.Rect(10, 10, 30, 30))
		rgc := NewRasterGraphicContext(dst)
		rgc.Translate(15, 15)
		rgc.DrawImage(src)

		assert.NotZero(t, dst.RGBAAt(15, 15).A)
		assert.Zero(t, dst.RGBAAt(10, 10).A)
	})

	t.Run("fill_after_clear", func(t *testing.T) {
		t.Parallel()

		img := image.NewRGBA(image.Rect(0, 0, 3, 3))
		rgc := NewRasterGraphicContext(img)
		rgc.SetFillColor(color.RGBA{0, 0, 255, 255})
		rgc.Clear()

		p := &Path{}
		p.MoveTo(0, 0)
		p.LineTo(2, 0)
		p.LineTo(2, 2)
		p.LineTo(0, 2)
		p.LineTo(0, 0)
		rgc.Fill(p)

		assert.Equal(t, color.RGBA{0, 0, 255, 255}, img.At(1, 1))
	})

	t.Run("fill_rectangle", func(t *testing.T) {
		t.Parallel()

		img := image.NewRGBA(image.Rect(0, 0, 3, 3))
		rgc := NewRasterGraphicContext(img)
		rgc.SetFillColor(color.RGBA{0, 255, 0, 255})
		p := &Path{}
		p.MoveTo(0, 0)
		p.LineTo(2, 0)
		p.LineTo(2, 2)
		p.LineTo(0, 2)
		p.LineTo(0, 0)
		rgc.Fill(p)
		_, _, _, a := img.At(1, 1).RGBA()
		assert.Equal(t, uint32(0xffff), a)
	})

	t.Run("draw_image", func(t *testing.T) {
		t.Parallel()

		src := image.NewRGBA(image.Rect(0, 0, 1, 1))
		src.Set(0, 0, color.White)
		dst := image.NewRGBA(image.Rect(0, 0, 3, 3))
		rgc := NewRasterGraphicContext(dst)
		rgc.DrawImage(src)
		_, _, _, a := dst.At(0, 0).RGBA()
		assert.Equal(t, uint32(0xffff), a)
	})

	t.Run("stroke_fill", func(t *testing.T) {
		t.Parallel()

		img := image.NewRGBA(image.Rect(0, 0, 3, 3))
		rgc := NewRasterGraphicContext(img)
		rgc.SetLineWidth(1)
		rgc.SetStrokeColor(color.Black)
		rgc.SetFillColor(color.RGBA{0, 0, 255, 255})

		p := &Path{}
		p.MoveTo(0, 0)
		p.LineTo(2, 0)
		p.LineTo(2, 2)
		p.LineTo(0, 2)
		p.Close()

		rgc.FillStroke(p)
		_, _, _, a := img.At(1, 1).RGBA()
		assert.Equal(t, uint32(0xffff), a) // fill
		_, _, _, a = img.At(0, 0).RGBA()
		assert.Equal(t, uint32(0xffff), a) // stroke

		img2 := image.NewRGBA(image.Rect(0, 0, 3, 3))
		rgc = NewRasterGraphicContext(img2)
		rgc.SetLineWidth(1)
		rgc.SetStrokeColor(color.Black)
		p = &Path{}
		p.MoveTo(0, 0)
		p.LineTo(2, 0)
		rgc.Stroke(p)
		_, _, _, a = img2.At(1, 0).RGBA()
		assert.Equal(t, uint32(0xffff), a)
	})

	t.Run("font_funcs", func(t *testing.T) {
		t.Parallel()

		img := image.NewRGBA(image.Rect(0, 0, 20, 20))
		rgc := NewRasterGraphicContext(img)

		f := getTestFont(t)
		rgc.SetFont(f)
		assert.Equal(t, f, rgc.GetFont())

		rgc.SetFontSize(12)
		assert.InDelta(t, 12.0, rgc.GetFontSize(), 0.0)

		rgc.SetFontSize(8)
		wSmall, err := rgc.CreateStringPath("A", 0, 0)
		require.NoError(t, err)

		rgc.SetFontSize(16)
		wLarge, err := rgc.CreateStringPath("A", 0, 0)
		require.NoError(t, err)
		assert.Greater(t, wLarge, wSmall)
	})
}

func TestRasterRecalcScale(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	rgc := NewRasterGraphicContext(img)

	for _, dpi := range []float64{72, 92, 96, 300} {
		for _, size := range []float64{8, 10, 11, 12, 14, 28} {
			rgc.SetDPI(dpi)
			rgc.SetFontSize(size)

			// rounded so the truncating fixed.Int26_6 conversions match the freetype face scale
			assert.InDelta(t, math.Round(PointsToPixels(dpi, size)*64), rgc.current.Scale, 0)
		}
	}
}

func TestRasterCreateStringPathAndBounds(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	rgc := NewRasterGraphicContext(img)
	f := getTestFont(t)
	rgc.SetFont(f)
	rgc.SetFontSize(10)

	idx := f.Index('A')
	expected := fUnitsToFloat64(f.HMetric(fixed.Int26_6(rgc.current.Scale), idx).AdvanceWidth)
	cursor, err := rgc.CreateStringPath("A", 0, 0)
	require.NoError(t, err)
	assert.InDelta(t, expected, cursor, 0.001)
	assert.False(t, rgc.current.Path.IsEmpty())

	left, top, right, bottom, err := rgc.GetStringBounds("A")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, cursor, right-left)

	pbLeft, pbTop, pbRight, pbBottom := pathBounds(rgc.current.Path)
	assert.InDelta(t, left, pbLeft, 0.001)
	assert.InDelta(t, top, pbTop, 0.001)
	assert.InDelta(t, bottom, pbBottom, 0.001)
	assert.GreaterOrEqual(t, right, pbRight) // bounds include the advance, the path only the ink
}

func TestRasterGetStringBounds(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	rgc := NewRasterGraphicContext(img)
	rgc.SetFont(getTestFont(t))
	rgc.SetFontSize(10)

	t.Run("empty_string", func(t *testing.T) {
		left, top, right, bottom, err := rgc.GetStringBounds("")
		require.NoError(t, err)
		assert.Zero(t, left)
		assert.Zero(t, top)
		assert.Zero(t, right)
		assert.Zero(t, bottom)
	})

	t.Run("space_only", func(t *testing.T) {
		left, top, right, bottom, err := rgc.GetStringBounds("   ")
		require.NoError(t, err)
		assert.Zero(t, left)
		assert.Zero(t, top)
		assert.Zero(t, bottom)
		assert.Positive(t, right) // advance width of the spaces
	})

	t.Run("ordinary_string", func(t *testing.T) {
		left, top, right, bottom, err := rgc.GetStringBounds("Ay")
		require.NoError(t, err)
		assert.Less(t, left, right)
		assert.Less(t, top, bottom)
		assert.Less(t, right, 100.0)
	})

	t.Run("trailing_space", func(t *testing.T) {
		_, _, right, _, err := rgc.GetStringBounds("hi")
		require.NoError(t, err)
		_, _, spacedRight, _, err := rgc.GetStringBounds("hi ")
		require.NoError(t, err)
		assert.Greater(t, spacedRight, right)
	})
}

func TestRasterFillAndStrokeString(t *testing.T) {
	t.Parallel()

	f := getTestFont(t)

	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	rgc := NewRasterGraphicContext(img)
	rgc.SetFont(f)
	rgc.SetFontSize(10)
	rgc.SetFillColor(color.White)

	left, top, right, bottom, err := rgc.GetStringBounds("A")
	require.NoError(t, err)
	x, y := 10.0, 30.0
	_, err = rgc.CreateStringPath("A", x, y)
	require.NoError(t, err)
	rgc.Fill()

	x1 := int(math.Floor(left + x))
	y1 := int(math.Floor(top + y))
	found := false
	for yy := y1; yy < int(math.Ceil(bottom+y)) && !found; yy++ {
		for xx := x1; xx < int(math.Ceil(right+x)) && !found; xx++ {
			_, _, _, a := img.At(xx, yy).RGBA()
			if a != 0 {
				found = true
			}
		}
	}
	assert.True(t, found, "filled text not drawn")

	img2 := image.NewRGBA(image.Rect(0, 0, 50, 50))
	rgc2 := NewRasterGraphicContext(img2)
	rgc2.SetFont(f)
	rgc2.SetFontSize(10)
	rgc2.SetStrokeColor(color.White)

	_, err = rgc2.CreateStringPath("A", x, y)
	require.NoError(t, err)
	rgc2.Stroke()
	found = false
	for yy := y1; yy < int(math.Ceil(bottom+y)) && !found; yy++ {
		for xx := x1; xx < int(math.Ceil(right+x)) && !found; xx++ {
			_, _, _, a := img2.At(xx, yy).RGBA()
			if a != 0 {
				found = true
			}
		}
	}
	assert.True(t, found, "stroked text not drawn")
}

func pathBounds(p *Path) (left, top, right, bottom float64) {
	if len(p.Points) == 0 {
		return
	}
	left, top = p.Points[0], p.Points[1]
	right, bottom = left, top
	for i := 2; i < len(p.Points); i += 2 {
		x, y := p.Points[i], p.Points[i+1]
		if x < left {
			left = x
		}
		if y < top {
			top = y
		}
		if x > right {
			right = x
		}
		if y > bottom {
			bottom = y
		}
	}
	return
}

func TestNewRasterGraphicContextWithPainter(t *testing.T) {
	t.Parallel()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	p := raster.NewRGBAPainter(img)
	rgc := NewRasterGraphicContextWithPainter(img, p)
	if rgc.painter != p {
		t.Fatalf("painter not set")
	}
}

func TestIsRectanglePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		build    func(p *Path)
		expected bool
	}{
		{
			name: "closed_rectangle",
			build: func(p *Path) {
				p.MoveTo(0, 0)
				p.LineTo(10, 0)
				p.LineTo(10, 10)
				p.LineTo(0, 10)
				p.Close()
			},
			expected: true,
		},
		{
			name: "five_lineto_rectangle",
			build: func(p *Path) {
				p.MoveTo(0, 0)
				p.LineTo(10, 0)
				p.LineTo(10, 10)
				p.LineTo(0, 10)
				p.LineTo(0, 0)
			},
			expected: true,
		},
		{
			name: "open_zigzag",
			build: func(p *Path) {
				p.MoveTo(0, 0)
				p.LineTo(10, 0)
				p.LineTo(10, 10)
				p.LineTo(0, 10)
				p.LineTo(35, 35)
			},
			expected: false,
		},
		{
			name: "contains_curve",
			build: func(p *Path) {
				p.MoveTo(0, 0)
				p.LineTo(10, 0)
				p.LineTo(10, 10)
				p.QuadCurveTo(0, 10, 35, 35)
				p.LineTo(0, 0)
			},
			expected: false,
		},
		{
			name: "unclosed_five_lineto",
			build: func(p *Path) {
				p.MoveTo(0, 0)
				p.LineTo(10, 0)
				p.LineTo(10, 10)
				p.LineTo(0, 10)
				p.LineTo(0, 5)
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Path{}
			tt.build(p)
			assert.Equal(t, tt.expected, isRectanglePath(p))
		})
	}
}

func TestRectFastPathBounds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		build    func(p *Path)
		expected [4]int
		ok       bool
	}{
		{
			name: "integer_rectangle",
			build: func(p *Path) {
				p.MoveTo(0, 0)
				p.LineTo(2, 0)
				p.LineTo(2, 2)
				p.LineTo(0, 2)
				p.LineTo(0, 0)
			},
			expected: [4]int{0, 0, 2, 2},
			ok:       true,
		},
		{
			name: "reversed_corners",
			build: func(p *Path) {
				p.MoveTo(12, 8)
				p.LineTo(12, 3)
				p.LineTo(4, 3)
				p.LineTo(4, 8)
				p.Close()
			},
			expected: [4]int{4, 3, 12, 8},
			ok:       true,
		},
		{
			name: "fractional_edges",
			build: func(p *Path) {
				p.MoveTo(0.5, 0)
				p.LineTo(2.5, 0)
				p.LineTo(2.5, 2)
				p.LineTo(0.5, 2)
				p.Close()
			},
		},
		{
			name: "infinite_edge",
			build: func(p *Path) {
				p.MoveTo(0, 0)
				p.LineTo(math.Inf(1), 0)
				p.LineTo(math.Inf(1), 2)
				p.LineTo(0, 2)
				p.Close()
			},
		},
		{
			name: "nan_edge",
			build: func(p *Path) {
				p.MoveTo(0, 0)
				p.LineTo(math.NaN(), 0)
				p.LineTo(math.NaN(), 2)
				p.LineTo(0, 2)
				p.Close()
			},
		},
		{
			name: "non_rectangle",
			build: func(p *Path) {
				p.MoveTo(0, 0)
				p.LineTo(1, 1)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Path{}
			tt.build(p)
			x1, y1, x2, y2, ok := rectFastPathBounds(p)
			assert.Equal(t, tt.ok, ok)
			if tt.ok {
				assert.Equal(t, tt.expected, [4]int{x1, y1, x2, y2})
			}
		})
	}
}

func TestStroke(t *testing.T) {
	t.Parallel()

	strokeDashed := func(dash []float64, dashOffset float64) *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, 40, 40))
		rgc := NewRasterGraphicContext(img)
		rgc.SetStrokeColor(color.RGBA{R: 255, A: 255})
		rgc.SetLineWidth(2)
		rgc.SetLineDash(dash, dashOffset)
		rgc.MoveTo(2, 20)
		rgc.LineTo(38, 20)
		rgc.Stroke()
		return img
	}
	solid := strokeDashed(nil, 0)

	t.Run("valid_dash", func(t *testing.T) {
		img := strokeDashed([]float64{4, 4}, 0)

		assert.NotZero(t, img.RGBAAt(3, 20).A) // within first dash
		assert.Zero(t, img.RGBAAt(8, 20).A)    // within first gap
	})

	t.Run("degenerate_dash_solid", func(t *testing.T) {
		for _, dash := range [][]float64{{0, 0}, {-5, 5}, {5, -5}, {math.NaN(), 5}, {math.Inf(1), 5}} {
			assert.Equal(t, solid.Pix, strokeDashed(dash, 0).Pix, dash)
		}
	})

	t.Run("non_finite_offset_solid", func(t *testing.T) {
		assert.Equal(t, solid.Pix, strokeDashed([]float64{5, 5}, math.Inf(1)).Pix)
		assert.Equal(t, solid.Pix, strokeDashed([]float64{5, 5}, math.NaN()).Pix)
	})

	t.Run("odd_dash_doubled", func(t *testing.T) {
		img := strokeDashed([]float64{6}, 0)

		assert.NotZero(t, img.RGBAAt(3, 20).A) // within first dash
		assert.Zero(t, img.RGBAAt(11, 20).A)   // within first gap
	})

	t.Run("offset_bounds", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(10, 10, 30, 30))
		rgc := NewRasterGraphicContext(img)
		rgc.SetStrokeColor(color.RGBA{R: 255, A: 255})
		rgc.SetLineWidth(2)
		rgc.MoveTo(12, 12)
		rgc.LineTo(28, 28)
		rgc.Stroke()

		assert.NotZero(t, img.RGBAAt(26, 26).A) // unreachable when rasterizers are sized by width/height
		assert.NotZero(t, img.RGBAAt(14, 14).A)
	})

	t.Run("zero_gap_dash_solid", func(t *testing.T) {
		img := strokeDashed([]float64{6, 0}, 0) // zero length gaps draw a solid line

		for x := 3; x < 37; x++ {
			assert.NotZero(t, img.RGBAAt(x, 20).A, x)
		}
	})

	t.Run("extreme_coordinates", func(t *testing.T) {
		// each case must terminate, an unbounded dash walk hangs the test binary
		dashTo := func(x float64) *image.RGBA {
			img := image.NewRGBA(image.Rect(0, 0, 40, 40))
			rgc := NewRasterGraphicContext(img)
			rgc.SetStrokeColor(color.RGBA{R: 255, A: 255})
			rgc.SetLineWidth(2)
			rgc.SetLineDash([]float64{4, 4}, 0)
			rgc.MoveTo(20, 20)
			rgc.LineTo(x, 20)
			rgc.Stroke()
			return img
		}

		for _, x := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
			assert.Zero(t, dashTo(x).RGBAAt(21, 20).A, x) // the segment is dropped entirely
			assert.Zero(t, dashTo(x).RGBAAt(19, 20).A, x)
		}
		assert.NotZero(t, dashTo(1e15).RGBAAt(21, 20).A) // drawn solid past the dash budget
		assert.NotZero(t, dashTo(-1e15).RGBAAt(19, 20).A)
		assert.NotZero(t, dashTo(1e200).RGBAAt(21, 20).A) // a squared length would overflow here
		assert.NotZero(t, dashTo(-1e200).RGBAAt(19, 20).A)
	})
}

func TestFill(t *testing.T) {
	t.Parallel()

	t.Run("extreme_coordinates", func(t *testing.T) {
		for _, x := range []float64{math.NaN(), math.Inf(1), 1e12} {
			img := image.NewRGBA(image.Rect(0, 0, 40, 40))
			rgc := NewRasterGraphicContext(img)
			rgc.SetFillColor(color.RGBA{R: 255, A: 255})
			rgc.MoveTo(10, 10)
			rgc.LineTo(x, 20)
			rgc.LineTo(30, 30)
			rgc.Close()
			rgc.Fill()

			assert.Zero(t, img.RGBAAt(2, 2).A, x) // nothing outside the triangle
		}
	})

	t.Run("rect_translated", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 40, 40))
		rgc := NewRasterGraphicContext(img)
		rgc.SetFillColor(color.RGBA{R: 255, A: 255})
		rgc.Translate(10, 10)
		rgc.MoveTo(0, 0)
		rgc.LineTo(5, 0)
		rgc.LineTo(5, 5)
		rgc.LineTo(0, 5)
		rgc.Close()
		rgc.Fill()

		assert.NotZero(t, img.RGBAAt(12, 12).A) // filled at translated position
		assert.Zero(t, img.RGBAAt(2, 2).A)      // not filled at raw coordinates
	})

	t.Run("open_zigzag_not_dropped", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 40, 40))
		rgc := NewRasterGraphicContext(img)
		rgc.SetFillColor(color.RGBA{R: 255, A: 255})
		rgc.MoveTo(0, 0)
		rgc.LineTo(10, 0)
		rgc.LineTo(10, 10)
		rgc.LineTo(0, 10)
		rgc.LineTo(35, 35)
		rgc.Fill()

		assert.NotZero(t, img.RGBAAt(15, 20).A) // dropped diagonal region outside the 10x10 rect
	})

	t.Run("integer_rect_fast_path", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 40, 40))
		rgc := NewRasterGraphicContext(img)
		rgc.SetFillColor(color.RGBA{R: 255, A: 255})
		rgc.MoveTo(10, 10)
		rgc.LineTo(20, 10)
		rgc.LineTo(20, 20)
		rgc.LineTo(10, 20)
		rgc.Close()
		rgc.Fill()

		for _, x := range []int{10, 15, 19} {
			assert.Equal(t, uint8(255), img.RGBAAt(x, 15).A, x)
			assert.Equal(t, uint8(255), img.RGBAAt(15, x).A, x)
		}
		for _, x := range []int{9, 20} { // no antialiased fringe outside the exact bounds
			assert.Zero(t, img.RGBAAt(x, 15).A, x)
			assert.Zero(t, img.RGBAAt(15, x).A, x)
		}
	})

	t.Run("fractional_rect_antialiased", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 40, 40))
		rgc := NewRasterGraphicContext(img)
		rgc.SetFillColor(color.RGBA{R: 255, A: 255})
		rgc.MoveTo(10.25, 10.25)
		rgc.LineTo(20.75, 10.25)
		rgc.LineTo(20.75, 20.75)
		rgc.LineTo(10.25, 20.75)
		rgc.Close()
		rgc.Fill()

		assert.Equal(t, uint8(255), img.RGBAAt(15, 15).A) // interior fully covered
		for _, edge := range [][2]int{{10, 15}, {20, 15}, {15, 10}, {15, 20}} {
			a := img.RGBAAt(edge[0], edge[1]).A
			assert.Positive(t, a, edge)
			assert.Less(t, a, uint8(255), edge)
		}
	})

	t.Run("fractional_rect_translation_stable", func(t *testing.T) {
		fillFractionalRect := func(offset float64) *image.RGBA {
			img := image.NewRGBA(image.Rect(0, 0, 40, 40))
			rgc := NewRasterGraphicContext(img)
			rgc.SetFillColor(color.RGBA{R: 255, A: 255})
			rgc.Translate(offset, offset)
			rgc.MoveTo(10.25, 10.25)
			rgc.LineTo(20.75, 10.25)
			rgc.LineTo(20.75, 20.75)
			rgc.LineTo(10.25, 20.75)
			rgc.Close()
			rgc.Fill()
			return img
		}
		identity, translated := fillFractionalRect(0), fillFractionalRect(5)

		for y := 9; y <= 21; y++ {
			for x := 9; x <= 21; x++ {
				assert.Equal(t, identity.RGBAAt(x, y).A, translated.RGBAAt(x+5, y+5).A, [2]int{x, y})
			}
		}
	})

	t.Run("offset_bounds", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(10, 10, 30, 30))
		fillTriangle(NewRasterGraphicContext(img), 12, 28)

		assert.NotZero(t, img.RGBAAt(27, 25).A) // unreachable when rasterizers are sized by width/height
		assert.NotZero(t, img.RGBAAt(20, 14).A)
		assert.Zero(t, img.RGBAAt(14, 26).A) // outside the triangle
	})

	t.Run("subimage_bounds", func(t *testing.T) {
		base := image.NewRGBA(image.Rect(0, 0, 40, 40))
		sub, ok := base.SubImage(image.Rect(20, 20, 40, 40)).(*image.RGBA)
		require.True(t, ok)
		fillTriangle(NewRasterGraphicContext(sub), 22, 38)

		assert.NotZero(t, sub.RGBAAt(37, 35).A)
		assert.NotZero(t, base.RGBAAt(37, 35).A)
		assert.Zero(t, base.RGBAAt(10, 10).A) // outside the sub-image
	})
}

// fillTriangle fills a right triangle with corners at (lo, lo), (hi, lo) and (hi, hi).
func fillTriangle(rgc *RasterGraphicContext, lo, hi float64) {
	rgc.SetFillColor(color.RGBA{R: 255, A: 255})
	rgc.MoveTo(lo, lo)
	rgc.LineTo(hi, lo)
	rgc.LineTo(hi, hi)
	rgc.Close()
	rgc.Fill()
}

func TestFillStroke(t *testing.T) {
	t.Parallel()

	t.Run("rect_translated", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 40, 40))
		rgc := NewRasterGraphicContext(img)
		rgc.SetFillColor(color.RGBA{R: 255, A: 255})
		rgc.SetStrokeColor(color.RGBA{R: 255, A: 255})
		rgc.Translate(10, 10)
		rgc.MoveTo(0, 0)
		rgc.LineTo(5, 0)
		rgc.LineTo(5, 5)
		rgc.LineTo(0, 5)
		rgc.Close()
		rgc.FillStroke()

		assert.NotZero(t, img.RGBAAt(12, 12).A) // filled at translated position
		assert.Zero(t, img.RGBAAt(2, 2).A)      // not filled at raw coordinates
	})

	t.Run("degenerate_dash_solid", func(t *testing.T) {
		// triangle avoids the rectangle fast path, exercising the FillStroke dasher
		fillStrokeDashed := func(dash []float64) *image.RGBA {
			img := image.NewRGBA(image.Rect(0, 0, 40, 40))
			rgc := NewRasterGraphicContext(img)
			rgc.SetFillColor(color.RGBA{G: 255, A: 255})
			rgc.SetStrokeColor(color.RGBA{R: 255, A: 255})
			rgc.SetLineWidth(2)
			rgc.SetLineDash(dash, 0)
			rgc.MoveTo(5, 5)
			rgc.LineTo(35, 10)
			rgc.LineTo(20, 35)
			rgc.Close()
			rgc.FillStroke()
			return img
		}
		solid := fillStrokeDashed(nil)

		for _, dash := range [][]float64{{0, 0}, {-5, 5}, {5, -5}} {
			assert.Equal(t, solid.Pix, fillStrokeDashed(dash).Pix, dash)
		}
	})
}
