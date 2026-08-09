package drawing

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/draw"
)

func TestDrawImageTransform(t *testing.T) {
	t.Parallel()

	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.Set(0, 0, color.White)
	dst := image.NewRGBA(image.Rect(0, 0, 3, 3))
	DrawImage(src, dst, NewTranslationMatrix(1, 1), draw.Over, LinearFilter)
	_, _, _, a := dst.At(1, 1).RGBA()
	assert.Equal(t, uint32(0xffff), a)
}

func TestDrawImageScale(t *testing.T) {
	t.Parallel()

	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.Set(0, 0, color.White)
	dst := image.NewRGBA(image.Rect(0, 0, 2, 2))
	DrawImage(src, dst, NewScaleMatrix(2, 2), draw.Over, LinearFilter)
	_, _, _, a := dst.At(1, 1).RGBA()
	assert.Equal(t, uint32(0xffff), a)
}

func TestDrawImageRotation(t *testing.T) {
	t.Parallel()

	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	src.Set(1, 0, color.White)
	tr := NewTranslationMatrix(2, 2)
	tr.Rotate(math.Pi / 2)
	x, y := tr.TransformPoint(1, 0)
	require.InDelta(t, 2.0, x, 1e-9)
	require.InDelta(t, 3.0, y, 1e-9)

	dst := image.NewRGBA(image.Rect(0, 0, 8, 8))
	DrawImage(src, dst, tr, draw.Over, LinearFilter)

	// marker must land where TransformPoint says, not mirrored by a transposed matrix
	_, _, _, a := dst.At(1, 3).RGBA()
	assert.Equal(t, uint32(0xffff), a)
	_, _, _, a = dst.At(2, 0).RGBA()
	assert.Equal(t, uint32(0), a)
}

func TestDrawImageUnknownFilter(t *testing.T) {
	t.Parallel()

	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.Set(0, 0, color.White)
	dst := image.NewRGBA(image.Rect(0, 0, 3, 3))
	assert.NotPanics(t, func() {
		DrawImage(src, dst, NewTranslationMatrix(1, 1), draw.Over, ImageFilter(99))
	})
	_, _, _, a := dst.At(1, 1).RGBA()
	assert.Equal(t, uint32(0xffff), a)
}

func TestDrawImageFilters(t *testing.T) {
	t.Parallel()

	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.Set(0, 0, color.White)

	dst1 := image.NewRGBA(image.Rect(0, 0, 1, 1))
	DrawImage(src, dst1, NewIdentityMatrix(), draw.Over, BilinearFilter)
	_, _, _, a := dst1.At(0, 0).RGBA()
	assert.Equal(t, uint32(0xffff), a)

	dst2 := image.NewRGBA(image.Rect(0, 0, 1, 1))
	DrawImage(src, dst2, NewIdentityMatrix(), draw.Over, BicubicFilter)
	_, _, _, a = dst2.At(0, 0).RGBA()
	assert.Equal(t, uint32(0xffff), a)
}
