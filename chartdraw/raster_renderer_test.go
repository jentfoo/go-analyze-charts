package chartdraw

import (
	"bytes"
	"errors"
	"hash/crc32"
	"image"
	"image/png"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-analyze/charts/chartdraw/drawing"
)

func hashImage(t *testing.T, r *rasterRenderer) uint32 {
	t.Helper()

	iw := &ImageWriter{}
	require.NoError(t, r.Save(iw))
	img, err := iw.Image()
	require.NoError(t, err)
	rgba := img.(*image.RGBA)
	return crc32.ChecksumIEEE(rgba.Pix)
}

func TestRasterRendererRotationNonAccumulating(t *testing.T) {
	t.Parallel()

	texts := []struct {
		body string
		x, y int
	}{
		{"hi", 4, 6},
		{"yo", 12, 14},
	}

	// one rotation setup for both draws, the pattern that used to compound the transform
	single := PNG(50, 60).(*rasterRenderer)
	single.SetFont(GetDefaultFont())
	single.SetFontSize(10)
	single.SetFontColor(drawing.ColorBlack)
	single.SetTextRotation(math.Pi / 2)
	for _, tt := range texts {
		single.Text(tt.body, tt.x, tt.y)
	}

	// each draw bracketed by its own rotation setup, always correct
	bracketed := PNG(50, 60).(*rasterRenderer)
	for _, tt := range texts {
		bracketed.SetFont(GetDefaultFont())
		bracketed.SetFontSize(10)
		bracketed.SetFontColor(drawing.ColorBlack)
		bracketed.SetTextRotation(math.Pi / 2)
		bracketed.Text(tt.body, tt.x, tt.y)
		bracketed.ClearTextRotation()
	}

	blank := hashImage(t, PNG(50, 60).(*rasterRenderer))
	assert.NotEqual(t, blank, hashImage(t, bracketed)) // guard against comparing empty canvases
	assert.Equal(t, hashImage(t, single), hashImage(t, bracketed))
}

func TestRasterRendererRotationRestoresTransform(t *testing.T) {
	t.Parallel()

	rr := PNG(50, 60).(*rasterRenderer)
	rr.SetFont(GetDefaultFont())
	rr.SetFontSize(10)
	rr.SetFontColor(drawing.ColorBlack)
	rr.SetTextRotation(math.Pi / 2)
	rr.Text("hi", 4, 6)

	assert.Equal(t, drawing.NewIdentityMatrix(), rr.gc.GetMatrixTransform())
}

func TestRasterRendererRotationLinesUnaffected(t *testing.T) {
	t.Parallel()

	rr := PNG(50, 60).(*rasterRenderer)
	rr.SetFont(GetDefaultFont())
	rr.SetFontSize(10)
	rr.SetFontColor(drawing.ColorBlack)
	rr.SetTextRotation(math.Pi / 2)
	rr.Text("hi", 4, 6)

	// a line drawn after rotated text must land at its literal coordinates
	rr.SetStrokeColor(drawing.ColorBlack)
	rr.SetStrokeWidth(2)
	rr.MoveTo(0, 55)
	rr.LineTo(49, 55)
	rr.Stroke()

	assert.Equal(t, drawing.ColorBlack, at(rr.i, 24, 54))
	assert.Equal(t, drawing.ColorBlack, at(rr.i, 24, 55))
	assert.Zero(t, at(rr.i, 24, 53).A)
}

func TestRasterRendererSavePNG(t *testing.T) {
	t.Parallel()

	rr := PNG(10, 10).(*rasterRenderer)
	buf := bytes.Buffer{}
	require.NoError(t, rr.Save(&buf))
	img, err := png.Decode(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	assert.Equal(t, 10, img.Bounds().Dx())
}

func TestRasterRendererCircleHash(t *testing.T) {
	t.Parallel()

	rr := PNG(20, 20).(*rasterRenderer)
	rr.SetFillColor(drawing.ColorWhite)
	rr.SetStrokeColor(drawing.ColorRed)
	rr.MoveTo(3, 3)
	rr.LineTo(4, 4)
	rr.Circle(5, 10, 10)
	rr.FillStroke()

	h := hashImage(t, rr)
	assert.Equal(t, uint32(0x59e4b5dd), h)
}

func TestRasterRendererRectangleHash(t *testing.T) {
	t.Parallel()

	rr := PNG(20, 20).(*rasterRenderer)
	rr.SetFillColor(drawing.ColorWhite)
	rr.SetStrokeColor(drawing.ColorRed)
	rr.MoveTo(2, 2)
	rr.LineTo(18, 2)
	rr.LineTo(18, 18)
	rr.LineTo(2, 18)
	rr.Close()
	rr.FillStroke()

	h := hashImage(t, rr)
	assert.Equal(t, uint32(0xcb26bf6d), h)
}

func TestRasterRendererArcHash(t *testing.T) {
	t.Parallel()

	rr := PNG(20, 20).(*rasterRenderer)
	rr.SetFillColor(drawing.ColorWhite)
	rr.SetStrokeColor(drawing.ColorRed)
	rr.MoveTo(10, 10)
	rr.ArcTo(10, 10, 8, 8, 0, math.Pi)
	rr.FillStroke()

	h := hashImage(t, rr)
	assert.Equal(t, uint32(0xa5291dba), h)
}

func TestRasterRendererQuadHash(t *testing.T) {
	t.Parallel()

	rr := PNG(20, 20).(*rasterRenderer)
	rr.SetStrokeColor(drawing.ColorBlue)
	rr.SetStrokeWidth(1)
	rr.MoveTo(2, 18)
	rr.QuadCurveTo(10, 0, 18, 18)
	rr.Stroke()

	h := hashImage(t, rr)
	assert.Equal(t, uint32(0xb5ef51e3), h)
}

func TestRasterRendererTextHash(t *testing.T) {
	t.Parallel()

	rr := PNG(50, 20).(*rasterRenderer)
	rr.SetFont(GetDefaultFont())
	rr.SetFontSize(10)
	rr.SetFontColor(drawing.ColorBlack)
	rr.Text("hi", 2, 12)

	h := hashImage(t, rr)
	assert.Equal(t, uint32(0x51c2f68c), h)
}

func TestRasterRendererDefaultDPI(t *testing.T) {
	t.Parallel()

	assert.InDelta(t, drawing.DefaultDPI, PNG(10, 10).GetDPI(), 0)
	assert.InDelta(t, drawing.DefaultDPI, JPG(10, 10).GetDPI(), 0)
	assert.InDelta(t, drawing.DefaultDPI, SVG(10, 10).GetDPI(), 0)
}

func TestRasterRendererMeasureText(t *testing.T) {
	t.Parallel()

	rr := PNG(200, 50).(*rasterRenderer)
	rr.SetFont(GetDefaultFont())
	rr.SetFontSize(10)

	emHeight := int(math.Ceil(drawing.PointsToPixels(rr.GetDPI(), 10)))

	empty := rr.MeasureText("")
	assert.Zero(t, empty.Width())
	assert.Equal(t, emHeight, empty.Height())

	space := rr.MeasureText(" ")
	assert.Positive(t, space.Width()) // advance width, no ink
	assert.Equal(t, emHeight, space.Height())

	text := rr.MeasureText("hi")
	assert.Greater(t, text.Width(), space.Width())
	assert.Equal(t, emHeight, text.Height())

	// height is the em box, independent of the ink of the string
	assert.Equal(t, emHeight, rr.MeasureText("Ayg").Height())

	assert.Greater(t, rr.MeasureText("hi ").Width(), text.Width())

	vr := SVG(200, 50)
	vr.SetFont(GetDefaultFont())
	vr.SetFontSize(10)
	// glyphs with a negative left side bearing measure the same in both renderers
	assert.Equal(t, vr.MeasureText("jjjj").Width(), rr.MeasureText("jjjj").Width())
}

func BenchmarkRaterCircle(b *testing.B) {
	testRadius := []float64{400, 200, 128, 64, 16, 8, 2}
	bb := &bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		png := PNG(800, 800)
		jpg := JPG(800, 800)

		var flip bool
		for _, r := range testRadius {
			color := drawing.ColorNavy
			if flip {
				color = drawing.ColorThistle
				flip = false
			} else {
				flip = true
			}

			png.SetFillColor(color)
			png.Circle(r, 400, 400)
			png.Fill()

			jpg.SetFillColor(color)
			jpg.Circle(r, 400, 400)
			jpg.Fill()
		}

		bb.Reset()
		_ = png.Save(bb)
		bb.Reset()
		_ = jpg.Save(bb)
	}
}

func TestRasterRendererCircleInvalidRadiusDrawsNothing(t *testing.T) {
	t.Parallel()

	blank := hashImage(t, PNG(40, 50).(*rasterRenderer))

	for _, radius := range []float64{0, -1, -5, math.NaN(), math.Inf(1), math.Inf(-1)} {
		rr := PNG(40, 50).(*rasterRenderer)
		rr.SetFillColor(drawing.ColorBlack)
		rr.Circle(radius, 20, 25)
		rr.Fill()

		// non-positive or non-finite radius draws nothing, matching SVG's invalid-value behavior
		assert.Equal(t, blank, hashImage(t, rr), "radius %v", radius)
	}
}

func TestRasterRendererSaveWithQueuedError(t *testing.T) {
	t.Parallel()

	rr := PNG(10, 10).(*rasterRenderer)
	rr.renderErrs = append(rr.renderErrs, errors.New("boom"))

	err := rr.Save(&bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}
