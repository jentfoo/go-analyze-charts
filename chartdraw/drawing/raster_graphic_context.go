package drawing

import (
	"errors"
	"image"
	"image/color"
	"math"

	"github.com/golang/freetype/raster"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// DefaultDPI is the default image DPI used when no resolution is set.
const DefaultDPI = 96.0

// NewRasterGraphicContext creates a new Graphic context from an image.
func NewRasterGraphicContext(img *image.RGBA) *RasterGraphicContext {
	painter := raster.NewRGBAPainter(img)
	return NewRasterGraphicContextWithPainter(img, painter)
}

// NewRasterGraphicContextWithPainter creates a new Graphic context from an image and a Painter (see
// Freetype-go). Path rendering requires the image bounds maximum to be positive.
func NewRasterGraphicContextWithPainter(img draw.Image, painter Painter) *RasterGraphicContext {
	// rasterizers clip spans in absolute image coordinates, so they must reach the bounds maximum
	bounds := img.Bounds()
	return &RasterGraphicContext{
		NewStackGraphicContext(),
		img,
		painter,
		raster.NewRasterizer(bounds.Max.X, bounds.Max.Y),
		raster.NewRasterizer(bounds.Max.X, bounds.Max.Y),
		&truetype.GlyphBuf{},
		DefaultDPI,
	}
}

// RasterGraphicContext is the implementation of GraphicContext for a raster image.
type RasterGraphicContext struct {
	*StackGraphicContext
	img              draw.Image
	painter          Painter
	fillRasterizer   *raster.Rasterizer
	strokeRasterizer *raster.Rasterizer
	glyphBuf         *truetype.GlyphBuf
	dpi              float64
}

// SetDPI sets the screen resolution in dots per inch.
func (rgc *RasterGraphicContext) SetDPI(dpi float64) {
	rgc.dpi = dpi
	rgc.recalc()
}

// GetDPI returns the resolution of the Image GraphicContext.
func (rgc *RasterGraphicContext) GetDPI() float64 {
	return rgc.dpi
}

// Clear fills the canvas with a transparent color.
func (rgc *RasterGraphicContext) Clear() {
	draw.Draw(rgc.img, rgc.img.Bounds(), image.Transparent, image.Point{}, draw.Src)
}

// FillRect draws a filled rectangle with the provided coordinates and the current set FillColor.
func (rgc *RasterGraphicContext) FillRect(x1, y1, x2, y2 int) {
	imageColor := image.NewUniform(rgc.current.FillColor)
	draw.Draw(rgc.img, image.Rect(x1, y1, x2, y2), imageColor, image.Point{}, draw.Over)
}

// DrawImage draws the raster image in the current canvas.
func (rgc *RasterGraphicContext) DrawImage(img image.Image) {
	DrawImage(img, rgc.img, rgc.current.Tr, draw.Over, BilinearFilter)
}

func (rgc *RasterGraphicContext) drawGlyph(glyph truetype.Index, dx, dy float64) error {
	if err := rgc.glyphBuf.Load(rgc.current.Font, fixed.Int26_6(rgc.current.Scale), glyph, font.HintingNone); err != nil {
		return err
	}
	var e0 int
	for _, e1 := range rgc.glyphBuf.Ends {
		DrawContour(rgc, rgc.glyphBuf.Points[e0:e1], dx, dy)
		e0 = e1
	}
	return nil
}

// findFontForRune finds the best font for rendering a rune, trying fallback fonts if needed.
// Returns the font to use and the glyph index within that font.
func findFontForRune(primaryFont *truetype.Font, r rune) (*truetype.Font, truetype.Index) {
	if index := primaryFont.Index(r); index != 0 {
		return primaryFont, index
	}

	// Try fallback fonts for special symbols
	for _, fallbackName := range FallbackFonts {
		fallbackFont := GetFont(fallbackName)
		if fallbackFont == nil {
			continue // fallback not loaded
		} else if primaryFont.Name(0) == fallbackFont.Name(0) {
			continue // Skip if it's the same as primary font
		}

		fallbackIndex := fallbackFont.Index(r)
		if fallbackIndex != 0 {
			return fallbackFont, fallbackIndex
		}
	}

	// No font has this character, return primary font with 0 index
	return primaryFont, 0
}

// CreateStringPath creates a path from the string s at x, y, and returns the string width.
// The text is placed so that the left edge of the em square of the first character of s
// and the baseline intersect at x, y. The majority of the affected pixels will be
// above and to the right of the point, but some may be below or to the left.
// For example, drawing a string that starts with a 'J' in an italic font may
// affect pixels below and left of the point.
func (rgc *RasterGraphicContext) CreateStringPath(s string, x, y float64) (cursor float64, err error) {
	f := rgc.GetFont()
	if f == nil {
		err = errors.New("no font loaded, cannot continue")
		return
	}
	rgc.recalc()

	startx := x
	var prevFont *truetype.Font
	var prevIndex truetype.Index
	for _, rc := range s {
		currentFont, index := findFontForRune(f, rc)

		if prevFont != nil { // Apply kerning from whichever font provided the previous character
			nextIndex := index
			if prevFont != currentFont {
				nextIndex = prevFont.Index(rc)
			}
			x += fUnitsToFloat64(prevFont.Kern(fixed.Int26_6(rgc.current.Scale), prevIndex, nextIndex))
		}

		if currentFont != f {
			rgc.SetFont(currentFont) // Temporarily switch to fallback font for this glyph
			err = rgc.drawGlyph(index, x, y)
			rgc.SetFont(f)
		} else {
			err = rgc.drawGlyph(index, x, y)
		}
		if err != nil {
			cursor = x - startx
			return
		}

		x += fUnitsToFloat64(currentFont.HMetric(fixed.Int26_6(rgc.current.Scale), index).AdvanceWidth)
		prevFont, prevIndex = currentFont, index
	}
	cursor = x - startx
	return
}

// GetStringBounds returns the approximate pixel bounds of a string, where the right bound includes
// the trailing advance so trailing whitespace is measured.
func (rgc *RasterGraphicContext) GetStringBounds(s string) (left, top, right, bottom float64, err error) {
	f := rgc.GetFont()
	if f == nil {
		err = errors.New("no font loaded, cannot continue")
		return
	}
	rgc.recalc()

	left = math.MaxFloat64
	top = math.MaxFloat64

	var cursor float64
	var prevFont *truetype.Font
	var prevIndex truetype.Index
	for _, rc := range s {
		currentFont, index := findFontForRune(f, rc)

		if prevFont != nil { // Apply kerning from whichever font provided the previous character
			nextIndex := index
			if prevFont != currentFont {
				nextIndex = prevFont.Index(rc)
			}
			cursor += fUnitsToFloat64(prevFont.Kern(fixed.Int26_6(rgc.current.Scale), prevIndex, nextIndex))
		}

		if err = rgc.glyphBuf.Load(currentFont, fixed.Int26_6(rgc.current.Scale), index, font.HintingNone); err != nil {
			return
		}

		var e0 int
		for _, e1 := range rgc.glyphBuf.Ends {
			ps := rgc.glyphBuf.Points[e0:e1]
			for _, p := range ps {
				x, y := pointToF64Point(p)
				top = min(top, y)
				bottom = max(bottom, y)
				left = min(left, x+cursor)
				right = max(right, x+cursor)
			}
			e0 = e1
		}
		cursor += fUnitsToFloat64(currentFont.HMetric(fixed.Int26_6(rgc.current.Scale), index).AdvanceWidth)
		prevFont, prevIndex = currentFont, index
	}
	right = max(right, cursor)   // include trailing advance, matching the vector renderer
	if left == math.MaxFloat64 { // no outlines, empty or whitespace only
		left, top = 0, 0
	}
	return
}

// recalc updates the 26.6 fixed point em scale from the current font size and DPI.
func (rgc *RasterGraphicContext) recalc() {
	// rounded to match the freetype face scale used for SVG measurement
	rgc.current.Scale = math.Round(PointsToPixels(rgc.dpi, rgc.current.FontSizePoints) * 64)
}

// SetFont sets the font used to draw text.
func (rgc *RasterGraphicContext) SetFont(font *truetype.Font) {
	rgc.current.Font = font
}

// GetFont returns the font used to draw text.
func (rgc *RasterGraphicContext) GetFont() *truetype.Font {
	return rgc.current.Font
}

// SetFontSize sets the font size in points (as in “a 12 point font”).
func (rgc *RasterGraphicContext) SetFontSize(fontSizePoints float64) {
	rgc.current.FontSizePoints = fontSizePoints
	rgc.recalc()
}

func (rgc *RasterGraphicContext) paint(rasterizer *raster.Rasterizer, color color.Color) {
	rgc.painter.SetColor(color)
	rasterizer.Rasterize(rgc.painter)
	rasterizer.Clear()
	rgc.current.Path.Clear()
}

// Stroke strokes the paths with the color specified by SetStrokeColor
func (rgc *RasterGraphicContext) Stroke(paths ...*Path) {
	if rgc.current.LineWidth == 0 {
		rgc.current.Path.Clear()
		return
	}
	paths = append(paths, rgc.current.Path)

	rgc.strokeRasterizer.UseNonZeroWinding = true

	stroker := NewLineStroker(Transformer{Tr: rgc.current.Tr, Flattener: FtLineBuilder{Adder: rgc.strokeRasterizer}})
	stroker.HalfLineWidth = rgc.current.LineWidth / 2

	var liner Flattener
	if dashable(rgc.current.Dash, rgc.current.DashOffset) {
		liner = NewDashVertexConverter(rgc.current.Dash, rgc.current.DashOffset, stroker)
	} else {
		liner = stroker
	}
	for _, p := range paths {
		Flatten(p, liner, rgc.current.Tr.GetScale())
	}

	rgc.paint(rgc.strokeRasterizer, rgc.current.StrokeColor)
}

func isRectanglePath(path *Path) bool {
	if len(path.Components) != 5 {
		return false
	} else if path.Components[0] != MoveToComponent {
		return false
	}
	for i := 1; i < 4; i++ {
		if path.Components[i] != LineToComponent {
			return false
		}
	}
	x1, y1 := path.Points[0], path.Points[1]
	x2, y2 := path.Points[2], path.Points[3]
	x3, y3 := path.Points[4], path.Points[5]
	x4, y4 := path.Points[6], path.Points[7]
	switch path.Components[4] {
	case LineToComponent:
		if path.Points[8] != x1 || path.Points[9] != y1 {
			return false // fifth segment must return to the start
		}
	case CloseComponent:
	default:
		return false
	}

	// Check if opposite sides are equal
	return (x1 == x4 && x2 == x3 && y1 == y2 && y3 == y4) || (x1 == x2 && x3 == x4 && y1 == y4 && y2 == y3)
}

// rectFastPathBounds returns the pixel bounds of an axis-aligned rectangle path, and false when the
// path is not such a rectangle or any edge is not on a representable pixel boundary.
func rectFastPathBounds(path *Path) (int, int, int, int, bool) {
	if !isRectanglePath(path) {
		return 0, 0, 0, 0, false
	}
	x1, y1 := path.Points[0], path.Points[1]
	x2, y2 := path.Points[4], path.Points[5]
	if x2 < x1 {
		x1, x2 = x2, x1
	}
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	for _, v := range [...]float64{x1, y1, x2, y2} {
		// NaN fails the equality, infinities fail the range
		if v != math.Trunc(v) || v < math.MinInt32 || v > math.MaxInt32 {
			return 0, 0, 0, 0, false
		}
	}
	return int(x1), int(y1), int(x2), int(y2), true
}

// Fill fills the paths with the color specified by SetFillColor.
func (rgc *RasterGraphicContext) Fill(paths ...*Path) {
	paths = append(paths, rgc.current.Path)
	if len(paths) == 1 && rgc.current.Tr.IsIdentity() {
		if x1, y1, x2, y2, ok := rectFastPathBounds(paths[0]); ok {
			// pixel aligned rectangles of a uniform color draw more efficiently
			rgc.FillRect(x1, y1, x2, y2)
			rgc.current.Path.Clear() // draw complete
			return
		}
	}

	rgc.fillRasterizer.UseNonZeroWinding = rgc.current.FillRule == FillRuleWinding

	flattener := Transformer{Tr: rgc.current.Tr, Flattener: FtLineBuilder{Adder: rgc.fillRasterizer}}
	for _, p := range paths {
		Flatten(p, flattener, rgc.current.Tr.GetScale())
	}

	rgc.paint(rgc.fillRasterizer, rgc.current.FillColor)
}

// FillStroke first fills the paths and then strokes them.
func (rgc *RasterGraphicContext) FillStroke(paths ...*Path) {
	paths = append(paths, rgc.current.Path)
	if len(paths) == 1 && rgc.current.Tr.IsIdentity() {
		if x1, y1, x2, y2, ok := rectFastPathBounds(paths[0]); ok {
			// pixel aligned rectangles of a uniform color draw more efficiently, stroke the line after
			rgc.FillRect(x1, y1, x2, y2)
			rgc.Stroke() // draw path for stroke
			return
		}
	}

	rgc.fillRasterizer.UseNonZeroWinding = rgc.current.FillRule == FillRuleWinding
	rgc.strokeRasterizer.UseNonZeroWinding = true

	flattener := Transformer{Tr: rgc.current.Tr, Flattener: FtLineBuilder{Adder: rgc.fillRasterizer}}

	stroker := NewLineStroker(Transformer{Tr: rgc.current.Tr, Flattener: FtLineBuilder{Adder: rgc.strokeRasterizer}})
	stroker.HalfLineWidth = rgc.current.LineWidth / 2

	var liner Flattener
	if dashable(rgc.current.Dash, rgc.current.DashOffset) {
		liner = NewDashVertexConverter(rgc.current.Dash, rgc.current.DashOffset, stroker)
	} else {
		liner = stroker
	}

	demux := DemuxFlattener{Flatteners: []Flattener{flattener, liner}}
	for _, p := range paths {
		Flatten(p, demux, rgc.current.Tr.GetScale())
	}

	// Fill
	rgc.paint(rgc.fillRasterizer, rgc.current.FillColor)
	// Stroke
	rgc.paint(rgc.strokeRasterizer, rgc.current.StrokeColor)
}
