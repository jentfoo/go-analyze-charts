package chartdraw

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"

	"github.com/golang/freetype/truetype"

	"github.com/go-analyze/charts/chartdraw/drawing"
)

// PNG returns a new png raster renderer.
func PNG(width, height int) Renderer {
	i := image.NewRGBA(image.Rect(0, 0, width, height))
	return &rasterRenderer{
		i:          i,
		gc:         drawing.NewRasterGraphicContext(i),
		encodeFunc: png.Encode,
	}
}

// JPG returns a new jpg raster renderer.
func JPG(width, height int) Renderer {
	i := image.NewRGBA(image.Rect(0, 0, width, height))
	return &rasterRenderer{
		i:  i,
		gc: drawing.NewRasterGraphicContext(i),
		encodeFunc: func(w io.Writer, i image.Image) error {
			return jpeg.Encode(w, i, &jpeg.Options{Quality: 90})
		},
	}
}

// rasterRenderer renders chart commands to a bitmap.
type rasterRenderer struct {
	i          *image.RGBA
	gc         *drawing.RasterGraphicContext
	encodeFunc func(w io.Writer, i image.Image) error
	renderErrs []error

	rotateRadians *float64

	s Style
}

func (rr *rasterRenderer) ResetStyle() {
	rr.s = Style{
		FontStyle: FontStyle{
			Font: rr.s.Font,
		},
	}
	rr.ClearTextRotation()
}

// GetDPI returns the dpi.
func (rr *rasterRenderer) GetDPI() float64 {
	return rr.gc.GetDPI()
}

// SetDPI sets the rendering DPI (for Renderer interface).
func (rr *rasterRenderer) SetDPI(dpi float64) {
	rr.gc.SetDPI(dpi)
}

// SetClassName is ignored because raster images have no class names (for Renderer interface).
func (rr *rasterRenderer) SetClassName(_ string) {}

// SetStrokeColor sets the stroke color for future paths (for Renderer interface).
func (rr *rasterRenderer) SetStrokeColor(c drawing.Color) {
	rr.s.StrokeColor = c
}

// SetStrokeWidth sets the width of drawn lines (for Renderer interface).
func (rr *rasterRenderer) SetStrokeWidth(width float64) {
	rr.s.StrokeWidth = width
}

// SetStrokeDashArray sets the stroke dash array.
func (rr *rasterRenderer) SetStrokeDashArray(dashArray []float64) {
	rr.s.StrokeDashArray = dashArray
}

// SetFillColor sets the fill color for future paths (for Renderer interface).
func (rr *rasterRenderer) SetFillColor(c drawing.Color) {
	rr.s.FillColor = c
}

// MoveTo moves the drawing cursor to the given position (for PathBuilder interface).
func (rr *rasterRenderer) MoveTo(x, y int) {
	rr.gc.MoveTo(float64(x), float64(y))
}

// LineTo adds a line to the current path (for PathBuilder interface).
func (rr *rasterRenderer) LineTo(x, y int) {
	rr.gc.LineTo(float64(x), float64(y))
}

// QuadCurveTo adds a quadratic curve to the current path (for PathBuilder interface).
func (rr *rasterRenderer) QuadCurveTo(cx, cy, x, y int) {
	rr.gc.QuadCurveTo(float64(cx), float64(cy), float64(x), float64(y))
}

// ArcTo appends an elliptical arc to the current path (for PathBuilder interface).
func (rr *rasterRenderer) ArcTo(cx, cy int, rx, ry, startAngle, delta float64) {
	rr.gc.ArcTo(float64(cx), float64(cy), rx, ry, startAngle, delta)
}

// Close closes the current path (for PathBuilder interface).
func (rr *rasterRenderer) Close() {
	rr.gc.Close()
}

// Stroke renders the path outline without filling it (for PathBuilder interface).
func (rr *rasterRenderer) Stroke() {
	rr.gc.SetStrokeColor(rr.s.StrokeColor)
	rr.gc.SetLineWidth(rr.s.StrokeWidth)
	rr.gc.SetLineDash(rr.s.StrokeDashArray, 0)
	rr.gc.Stroke()
}

// Fill renders the path fill without stroking it (for PathBuilder interface).
func (rr *rasterRenderer) Fill() {
	rr.gc.SetFillColor(rr.s.FillColor)
	rr.gc.Fill()
}

// FillStroke fills and then strokes the current path (for PathBuilder interface).
func (rr *rasterRenderer) FillStroke() {
	rr.gc.SetFillColor(rr.s.FillColor)
	rr.gc.SetStrokeColor(rr.s.StrokeColor)
	rr.gc.SetLineWidth(rr.s.StrokeWidth)
	rr.gc.SetLineDash(rr.s.StrokeDashArray, 0)
	rr.gc.FillStroke()
}

// Circle adds a circle to the current path; it is only painted by a completion call
// (Stroke/Fill/FillStroke), and renders nothing without one.
func (rr *rasterRenderer) Circle(radius float64, x, y int) {
	if nonFinite(radius) || radius <= 0 {
		// keep current point consistent with valid-radius path
		rr.gc.MoveTo(float64(x), float64(y))
		return
	}
	xf, yf := float64(x), float64(y)
	rr.gc.MoveTo(xf+radius, yf) // explicit MoveTo at arc start to avoid LineTo from prior path, see issue #78
	rr.gc.ArcTo(xf, yf, radius, radius, 0, _2pi)
}

// SetFont sets the font used for text drawing (for Renderer interface).
func (rr *rasterRenderer) SetFont(f *truetype.Font) {
	rr.s.Font = f
}

// SetFontSize sets the font size in points (for Renderer interface).
func (rr *rasterRenderer) SetFontSize(size float64) {
	rr.s.FontSize = size
}

// SetFontColor sets the color used for text drawing (for Renderer interface).
func (rr *rasterRenderer) SetFontColor(c drawing.Color) {
	rr.s.FontColor = c
}

// Text draws the provided string at the given coordinates (for Renderer interface).
func (rr *rasterRenderer) Text(body string, x, y int) {
	if body == "" {
		return
	}
	rr.gc.SetFont(rr.s.Font)
	rr.gc.SetFontSize(rr.s.FontSize)
	rr.gc.SetFillColor(rr.s.FontColor)

	xf, yf := float64(x), float64(y)
	if rr.rotateRadians != nil {
		// rotate about the anchor for this draw only, matching the vector renderer
		tr := rr.gc.GetMatrixTransform()
		defer rr.gc.SetMatrixTransform(tr)
		rr.gc.Translate(xf, yf)
		rr.gc.Rotate(*rr.rotateRadians)
		xf, yf = 0, 0
	}
	if _, err := rr.gc.CreateStringPath(body, xf, yf); err != nil {
		rr.renderErrs = append(rr.renderErrs, err)
	}
	rr.gc.Fill()
}

// MeasureText returns the width and em height in pixels of a string, where the width covers the
// advance and any ink extending past it.
func (rr *rasterRenderer) MeasureText(body string) Box {
	rr.gc.SetFont(rr.s.Font)
	rr.gc.SetFontSize(rr.s.FontSize)
	rr.gc.SetFillColor(rr.s.FontColor)
	// left ink overhang is ignored to match the vector renderer, padding the right edge never covered it
	_, _, r, _, err := rr.gc.GetStringBounds(body)
	if err != nil {
		return Box{}
	}

	textBox := Box{
		Right: int(math.Ceil(r)),
		// em height rather than ink height, matching the vector renderer
		Bottom: int(math.Ceil(drawing.PointsToPixels(rr.gc.GetDPI(), rr.s.FontSize))),
		IsSet:  true,
	}
	if rr.rotateRadians == nil {
		return textBox
	}

	return textBox.Corners().Rotate(RadiansToDegrees(*rr.rotateRadians)).Box()
}

// SetTextRotation sets a text rotation.
func (rr *rasterRenderer) SetTextRotation(radians float64) {
	rr.rotateRadians = &radians
}

// ClearTextRotation clears text rotation.
func (rr *rasterRenderer) ClearTextRotation() {
	rr.rotateRadians = nil
}

// Save writes the rendered image to the provided writer (for Renderer interface).
func (rr *rasterRenderer) Save(w io.Writer) error {
	if len(rr.renderErrs) > 0 {
		return fmt.Errorf("queued rendering errors: %v", rr.renderErrs)
	} else if typed, isTyped := w.(RGBACollector); isTyped {
		typed.SetRGBA(rr.i)
		return nil
	} else if rr.encodeFunc != nil {
		return rr.encodeFunc(w, rr.i)
	}
	return png.Encode(w, rr.i)
}
