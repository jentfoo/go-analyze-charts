package chartdraw

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/golang/freetype/truetype"

	"github.com/go-analyze/charts/chartdraw/drawing"
)

// svgPrecision is the decimal precision used for geometry and stroke values. Sub-pixel accuracy
// matters because SVG derives the arc center from the endpoints and radii.
const svgPrecision = 2

// arcSplitGap is the sweep left uncovered below which an arc is emitted as two segments. Shorter
// remainders serialize to a chord too short to fix the arc center against svgPrecision rounding.
const arcSplitGap = 0.02

// escapes XML-special chars in SVG text content
var svgTextEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")

// SVG returns a new svg vector renderer.
func SVG(width, height int) Renderer {
	buffer := bytes.NewBuffer([]byte{})
	canvas := newCanvas(buffer)
	canvas.Start(width, height)
	return &vectorRenderer{
		b: buffer,
		c: canvas,
		s: &Style{},
		p: []string{},
	}
}

// SVGWithCSS returns a svg vector renderer constructor with the attached custom CSS.
// The optional nonce argument sets a CSP nonce.
func SVGWithCSS(css string, nonce string) func(width, height int) Renderer {
	return func(width, height int) Renderer {
		buffer := bytes.NewBuffer([]byte{})
		canvas := newCanvas(buffer)
		canvas.css = css
		canvas.nonce = nonce
		canvas.Start(width, height)
		return &vectorRenderer{
			b: buffer,
			c: canvas,
			s: &Style{},
			p: []string{},
		}
	}
}

// fontFaceKey is the key for caching font faces
type fontFaceKey struct {
	font *truetype.Font
	dpi  float64
	size float64
}

// vectorRenderer renders chart commands to a bitmap.
type vectorRenderer struct {
	b         *bytes.Buffer
	c         *canvas
	s         *Style
	p         []string
	faceCache map[fontFaceKey]font.Face
}

// measureStringWithFallback is a custom MeasureString that provides estimated sizes for missing glyphs.
func (vr *vectorRenderer) measureStringWithFallback(face font.Face, s string, fontSize, dpi float64) fixed.Int26_6 {
	var advance fixed.Int26_6
	prevC := rune(-1)
	for _, c := range s {
		if prevC >= 0 {
			advance += face.Kern(prevC, c)
		}

		glyphAdvance, ok := face.GlyphAdvance(c)

		// emoji's may be filled in by other fonts, if glyph is missing or has very small advance, try fallback fonts
		if (!ok || glyphAdvance/64 <= fixed.Int26_6(fontSize*0.75)) && isEmojiOrSymbol(c) {
			var foundInFallback bool
			var symbolFontFallback bool // true if glyph only exists in a symbol-specific font
			for _, fallbackName := range drawing.FallbackFonts {
				if fallbackFont := drawing.GetFont(fallbackName); fallbackFont != nil {
					// Check if fallback font has this character
					if fallbackIndex := fallbackFont.Index(c); fallbackIndex != 0 {
						fallbackFace := vr.cachedFontFace(fallbackFont, dpi, fontSize)
						if fallbackAdvance, fallbackOk := fallbackFace.GlyphAdvance(c); fallbackOk {
							if fallbackAdvance > glyphAdvance || !ok {
								glyphAdvance = fallbackAdvance
							}
							foundInFallback = true
							symbolFontFallback = strings.Contains(fallbackName, "symbol")
							break
						}
					}
				}
			}

			if !foundInFallback { // If no fallback font has the character, estimate symbol width
				glyphAdvance = fixed.Int26_6(fontSize * 64)
			} else if symbolFontFallback {
				// Symbol fonts may have narrow advance widths that don't match browser rendering,
				// since SVG text is rendered by the browser using system fonts, not our embedded fonts.
				// Ensure a minimum advance to prevent undersized label backgrounds.
				if minAdvance := fixed.Int26_6(fontSize * 64); glyphAdvance < minAdvance {
					glyphAdvance = minAdvance
				}
			}
		}

		advance += glyphAdvance
		prevC = c
	}

	return advance
}

// isEmojiOrSymbol checks if a rune is likely a symbol, emoji, or special character.
func isEmojiOrSymbol(r rune) bool {
	return (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
		(r >= 0x1F300 && r <= 0x1F5FF) || // Misc Symbols and Pictographs
		(r >= 0x1F680 && r <= 0x1F6FF) || // Transport and Map Symbols
		(r >= 0x1F700 && r <= 0x1F77F) || // Alchemical Symbols
		(r >= 0x1F780 && r <= 0x1F7FF) || // Geometric Shapes Extended
		(r >= 0x1F800 && r <= 0x1F8FF) || // Supplemental Arrows-C
		(r >= 0x1F900 && r <= 0x1F9FF) || // Supplemental Symbols and Pictographs
		(r >= 0x1FA00 && r <= 0x1FA6F) || // Chess Symbols
		(r >= 0x1FA70 && r <= 0x1FAFF) || // Symbols and Pictographs Extended-A
		(r >= 0x2600 && r <= 0x26FF) || // Miscellaneous Symbols
		(r >= 0x2700 && r <= 0x27BF) || // Dingbats
		(r >= 0xFE00 && r <= 0xFE0F) || // Variation Selectors
		(r >= 0x1F000 && r <= 0x1F02F) || // Mahjong Tiles
		(r >= 0x1F030 && r <= 0x1F09F) || // Domino Tiles
		(r >= 0x1F0A0 && r <= 0x1F0FF) || // Playing Cards
		(r >= 0x23E9 && r <= 0x23EC) || // Play/Pause buttons
		(r >= 0x23F0 && r <= 0x23F3) || // Alarm Clock
		(r >= 0x25A0 && r <= 0x25FF) || // Geometric Shapes
		(r >= 0x2934 && r <= 0x2935) || // Arrow symbols
		(r >= 0x2B05 && r <= 0x2B07) || // Arrow symbols
		(r >= 0x2B1B && r <= 0x2B1C) || // Square symbols
		(r >= 0x2B50 && r <= 0x2B55) || // Star symbols
		(r == 0x3030) || (r == 0x303D) || // Wave dash, Part alternation mark
		(r >= 0x3297 && r <= 0x3299) || // Circled ideographs
		// Mathematical operators
		(r >= 0x2200 && r <= 0x22FF) || // Mathematical Operators
		(r >= 0x2A00 && r <= 0x2AFF) || // Supplemental Mathematical Operators
		(r >= 0x27C0 && r <= 0x27EF) || // Miscellaneous Mathematical Symbols-A
		(r >= 0x2980 && r <= 0x29FF) || // Miscellaneous Mathematical Symbols-B
		(r >= 0x2100 && r <= 0x214F) || // Letterlike Symbols (includes √, ∞, etc.)
		// Currency symbols
		(r >= 0x20A0 && r <= 0x20CF) || // Currency Symbols
		// Box drawing and block elements
		(r >= 0x2500 && r <= 0x257F) || // Box Drawing
		(r >= 0x2580 && r <= 0x259F) || // Block Elements
		// Arrows
		(r >= 0x2190 && r <= 0x21FF) || // Arrows
		(r >= 0x27F0 && r <= 0x27FF) || // Supplemental Arrows-A
		(r >= 0x2900 && r <= 0x297F) || // Supplemental Arrows-B
		// Additional useful ranges
		(r >= 0x2000 && r <= 0x206F) // General Punctuation (includes em dash, etc.)
}

// cachedFontFace gets a cached font face or creates and caches a new one.
func (vr *vectorRenderer) cachedFontFace(f *truetype.Font, dpi, size float64) font.Face {
	if vr.faceCache == nil {
		vr.faceCache = make(map[fontFaceKey]font.Face)
	}

	key := fontFaceKey{font: f, dpi: dpi, size: size}
	if face, exists := vr.faceCache[key]; exists {
		return face
	}

	face := truetype.NewFace(f, &truetype.Options{
		DPI:  dpi,
		Size: size,
	})
	vr.faceCache[key] = face
	return face
}

func (vr *vectorRenderer) ResetStyle() {
	vr.s = &Style{
		FontStyle: FontStyle{
			Font: vr.s.Font,
		},
	}
}

// GetDPI returns the dpi.
func (vr *vectorRenderer) GetDPI() float64 {
	return vr.c.dpi
}

// SetDPI sets the rendering DPI (for Renderer interface).
func (vr *vectorRenderer) SetDPI(dpi float64) {
	vr.c.dpi = dpi
}

// SetClassName sets the CSS class name for the next drawing operations (for Renderer interface).
func (vr *vectorRenderer) SetClassName(classname string) {
	vr.s.ClassName = classname
}

// SetStrokeColor changes the stroke color for subsequent paths (for Renderer interface).
func (vr *vectorRenderer) SetStrokeColor(c drawing.Color) {
	vr.s.StrokeColor = c
}

// SetFillColor changes the fill color for subsequent paths (for Renderer interface).
func (vr *vectorRenderer) SetFillColor(c drawing.Color) {
	vr.s.FillColor = c
}

// SetStrokeWidth sets the width of drawn lines (for Renderer interface).
func (vr *vectorRenderer) SetStrokeWidth(width float64) {
	vr.s.StrokeWidth = width
}

// SetStrokeDashArray sets the stroke dash array.
func (vr *vectorRenderer) SetStrokeDashArray(dashArray []float64) {
	vr.s.StrokeDashArray = dashArray
}

// MoveTo starts a new path at the specified coordinates (for PathBuilder interface).
func (vr *vectorRenderer) MoveTo(x, y int) {
	vr.p = append(vr.p, "M "+strconv.Itoa(x)+" "+strconv.Itoa(y))
}

// LineTo adds a line segment to the current path (for PathBuilder interface).
func (vr *vectorRenderer) LineTo(x, y int) {
	vr.p = append(vr.p, "L "+strconv.Itoa(x)+" "+strconv.Itoa(y))
}

// QuadCurveTo draws a quad curve.
func (vr *vectorRenderer) QuadCurveTo(cx, cy, x, y int) {
	vr.p = append(vr.p, "Q"+strconv.Itoa(cx)+","+strconv.Itoa(cy)+" "+strconv.Itoa(x)+","+strconv.Itoa(y))
}

// ArcTo appends an elliptical arc to the current path (for PathBuilder interface).
// A negative delta sweeps counter-clockwise, and the sweep is limited to a single revolution.
// Parameters which are not finite leave the path unchanged.
func (vr *vectorRenderer) ArcTo(cx, cy int, rx, ry, startAngle, delta float64) {
	if hasNonFinite(rx, ry, startAngle, delta) {
		return // checked before the clamp, which would fold ±Inf into a full revolution
	}
	delta = min(max(delta, -_2pi), _2pi)
	cxf, cyf := float64(cx), float64(cy)

	startX := formatFloatMinimized(cxf+rx*math.Cos(startAngle), svgPrecision)
	startY := formatFloatMinimized(cyf+ry*math.Sin(startAngle), svgPrecision)
	endX := formatFloatMinimized(cxf+rx*math.Cos(startAngle+delta), svgPrecision)
	endY := formatFloatMinimized(cyf+ry*math.Sin(startAngle+delta), svgPrecision)

	startCmd := "M "
	if len(vr.p) > 0 {
		startCmd = "L "
	}
	vr.p = append(vr.p, startCmd+startX+" "+startY)

	absDelta := math.Abs(delta)
	segments := 1
	// a near-full sweep leaves the arc center ill-conditioned, and coincident serialized
	// endpoints are omitted by SVG entirely
	if _2pi-absDelta < arcSplitGap || (absDelta > _pi && startX == endX && startY == endY) {
		segments = 2
	}
	segDelta := delta / float64(segments)

	var sweepFlag int
	if delta > 0 {
		sweepFlag = 1 // sweep towards increasing angle, clockwise on screen
	}
	var largeArcFlag int
	if absDelta/float64(segments) > _pi {
		largeArcFlag = 1
	}

	rxStr, ryStr := formatFloatMinimized(rx, svgPrecision), formatFloatMinimized(ry, svgPrecision)
	for i := 1; i <= segments; i++ {
		angle := startAngle + segDelta*float64(i)
		vr.p = append(vr.p, fmt.Sprintf("A %s %s 0 %d %d %s %s", rxStr, ryStr, largeArcFlag, sweepFlag,
			formatFloatMinimized(cxf+rx*math.Cos(angle), svgPrecision),
			formatFloatMinimized(cyf+ry*math.Sin(angle), svgPrecision)))
	}
}

// Close closes a shape.
func (vr *vectorRenderer) Close() {
	vr.p = append(vr.p, "Z")
}

// Stroke draws the path with no fill.
func (vr *vectorRenderer) Stroke() {
	vr.drawPath()
}

// Fill draws the path with no stroke.
func (vr *vectorRenderer) Fill() {
	vr.drawPath()
}

// FillStroke draws the path with both fill and stroke.
func (vr *vectorRenderer) FillStroke() {
	vr.drawPath()
}

// drawPath draws the path set into the p slice.
func (vr *vectorRenderer) drawPath() {
	vr.c.Path(vr.p, vr.s.GetFillAndStrokeOptions())
	vr.p = vr.p[:0] // clear the path
}

// Circle draws a circle with the current style (for PathBuilder interface).
func (vr *vectorRenderer) Circle(radius float64, x, y int) {
	vr.c.Circle(x, y, radius, vr.s.GetFillAndStrokeOptions())
}

// SetFont specifies the font used for text operations (for Renderer interface).
func (vr *vectorRenderer) SetFont(f *truetype.Font) {
	vr.s.Font = f
}

// SetFontColor sets the color used to draw text (for Renderer interface).
func (vr *vectorRenderer) SetFontColor(c drawing.Color) {
	vr.s.FontColor = c
}

// SetFontSize sets the size of the font in points (for Renderer interface).
func (vr *vectorRenderer) SetFontSize(size float64) {
	vr.s.FontSize = size
}

// Text draws a text blob (for Renderer interface).
func (vr *vectorRenderer) Text(body string, x, y int) {
	vr.c.Text(x, y, body, vr.s.GetTextOptions())
}

// MeasureText uses the truetype font drawer to measure the width of text.
func (vr *vectorRenderer) MeasureText(body string) (box Box) {
	textFont := vr.s.GetFont()
	if textFont != nil {
		face := vr.cachedFontFace(textFont, vr.c.dpi, vr.s.FontSize)
		box.Right = vr.measureStringWithFallback(face, body, vr.s.FontSize, vr.c.dpi).Ceil()
		box.Bottom = int(math.Ceil(drawing.PointsToPixels(vr.c.dpi, vr.s.FontSize)))
		box.IsSet = true
		if vr.c.textTheta == nil {
			return
		}
		box = box.Corners().Rotate(RadiansToDegrees(*vr.c.textTheta)).Box()
	}
	return
}

// SetTextRotation sets the text rotation.
func (vr *vectorRenderer) SetTextRotation(radians float64) {
	if radians == 0 {
		vr.c.textTheta = nil
	} else {
		vr.c.textTheta = &radians
	}
}

// ClearTextRotation clears the text rotation.
func (vr *vectorRenderer) ClearTextRotation() {
	vr.c.textTheta = nil
}

// Save writes a complete SVG snapshot of everything drawn so far to w. It may be called any number
// of times, and reflects any drawing done since an earlier call.
func (vr *vectorRenderer) Save(w io.Writer) error {
	if _, err := w.Write(vr.b.Bytes()); err != nil {
		return err
	}
	_, err := w.Write([]byte("</svg>"))
	return err
}

func newCanvas(w io.Writer) *canvas {
	return &canvas{
		w:   w,
		bb:  bytes.NewBuffer(make([]byte, 0, 200)),
		dpi: defaultDPI,
	}
}

type canvas struct {
	w         io.Writer
	bb        *bytes.Buffer
	dpi       float64
	textTheta *float64
	width     int
	height    int
	css       string
	nonce     string
}

func (c *canvas) Start(width, height int) {
	c.width = width
	c.height = height
	_, _ = c.w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 ` + strconv.Itoa(c.width) + ` ` + strconv.Itoa(c.height) + `">`))
	if c.css != "" {
		_, _ = c.w.Write([]byte(`<style type="text/css"`))
		if c.nonce != "" {
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy
			_, _ = c.w.Write([]byte(` nonce="` + c.nonce + `"`))
		}
		// To avoid compatibility issues between XML and CSS (f.e. with child selectors) we should encapsulate the CSS with CDATA.
		_, _ = c.w.Write([]byte(`><![CDATA[` + c.css + `]]></style>`))
	}
}

func (c *canvas) Path(parts []string, style Style) {
	if len(parts) == 0 {
		return
	}
	bb := c.bb
	defer c.bb.Reset()

	bb.WriteString(`<path`)
	if drawing.ValidDash(style.StrokeDashArray) {
		bb.WriteString(" stroke-dasharray=\"")
		for i, v := range style.StrokeDashArray {
			if i > 0 {
				bb.WriteString(", ")
			}
			bb.WriteString(formatFloatMinimized(v, svgPrecision))
		}
		bb.WriteString("\"")
	}
	bb.WriteString(` d="`)
	for i, p := range parts {
		if i > 0 {
			bb.WriteRune('\n')
		}
		bb.WriteString(p)
	}
	bb.WriteString(`" `)
	styleAsSVG(bb, style, c.dpi, false)
	bb.WriteString(`/>`)

	_, _ = c.w.Write(bb.Bytes())
}

func (c *canvas) Text(x, y int, body string, style Style) {
	if body == "" {
		return
	}
	bb := c.bb
	defer c.bb.Reset()

	bb.WriteString(`<text x="`)
	bb.WriteString(strconv.Itoa(x))
	bb.WriteString(`" y="`)
	bb.WriteString(strconv.Itoa(y))
	bb.WriteString(`" `)
	styleAsSVG(bb, style, c.dpi, true)
	if c.textTheta != nil {
		_, _ = fmt.Fprintf(bb, ` transform="rotate(%0.2f,%d,%d)"`, RadiansToDegrees(*c.textTheta), x, y)
	}
	bb.WriteRune('>')
	_, _ = svgTextEscaper.WriteString(bb, body)
	bb.WriteString("</text>")

	_, _ = c.w.Write(bb.Bytes())
}

func (c *canvas) Circle(x, y int, r float64, style Style) {
	bb := c.bb
	defer c.bb.Reset()

	bb.WriteString(`<circle cx="`)
	bb.WriteString(strconv.Itoa(x))
	bb.WriteString(`" cy="`)
	bb.WriteString(strconv.Itoa(y))
	bb.WriteString(`" r="`)
	bb.WriteString(formatFloatMinimized(r, svgPrecision))
	bb.WriteString(`" `)
	styleAsSVG(bb, style, c.dpi, true)
	bb.WriteString(`/>`)

	_, _ = c.w.Write(bb.Bytes())
}

// nonFinite reports whether v is NaN or infinite.
func nonFinite(v float64) bool {
	return math.IsNaN(v) || math.IsInf(v, 0)
}

// hasNonFinite reports whether any value is NaN or infinite.
func hasNonFinite(vals ...float64) bool {
	return slices.ContainsFunc(vals, nonFinite)
}

// styleAsSVG returns the style as a svg style or class string.
func styleAsSVG(bb *bytes.Buffer, s Style, dpi float64, applyText bool) {
	sw := s.StrokeWidth
	sc := s.StrokeColor
	fc := s.FillColor
	f := s.Font
	fs := s.FontSize
	fnc := s.FontColor

	if s.ClassName != "" {
		bb.WriteString("class=\"")
		bb.WriteString(s.ClassName)
		if !sc.IsZero() {
			bb.WriteString(" stroke")
		}
		if !fc.IsZero() {
			bb.WriteString(" fill")
		}
		if applyText && (fs != 0 || f != nil) {
			bb.WriteString(" text")
		}
		bb.WriteString("\"")
		return
	}

	bb.WriteString("style=\"")

	if sw != 0 && !sc.IsTransparent() {
		bb.WriteString("stroke-width:")
		bb.WriteString(formatFloatMinimized(sw, svgPrecision))
		bb.WriteString(";stroke:")
		bb.WriteString(sc.String())
	} else {
		bb.WriteString("stroke:none")
	}

	if applyText && !fnc.IsTransparent() {
		bb.WriteString(";fill:")
		bb.WriteString(fnc.String())
	} else if !fc.IsTransparent() {
		bb.WriteString(";fill:")
		bb.WriteString(fc.String())
	} else {
		bb.WriteString(";fill:none")
	}

	if applyText {
		if fs != 0 {
			bb.WriteString(";font-size:")
			bb.WriteString(formatFloatMinimized(drawing.PointsToPixels(dpi, fs), 1))
			bb.WriteString("px")
		}
		if f != nil {
			if name := f.Name(truetype.NameIDFontFamily); name != "" {
				bb.WriteString(";font-family:'")
				bb.WriteString(name)
				bb.WriteString(`',sans-serif`)
			} else {
				bb.WriteString(";font-family:sans-serif")
			}
		}
	}

	bb.WriteRune('"')
}

// formatFloatMinimized formats a float to at most precision decimal places, trimming trailing
// zeros so the result is as short as possible.
func formatFloatMinimized(val float64, precision int) string {
	if nonFinite(val) {
		return "0" // non-finite values would serialize as invalid attribute text
	}
	str := strconv.FormatFloat(val, 'f', precision, 64)
	if precision > 0 {
		str = strings.TrimRight(str, "0") // e.g. "1.20" -> "1.2", "20.00" -> "20."
		str = strings.TrimRight(str, ".")
	}
	if str == "-0" {
		return "0" // trig residue near zero must not serialize as negative zero
	}
	return str
}
