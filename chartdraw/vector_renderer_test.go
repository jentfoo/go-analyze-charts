package chartdraw

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-analyze/charts/chartdraw/drawing"
)

func TestVectorRendererPath(t *testing.T) {
	t.Parallel()

	vr := SVG(100, 100)

	typed, isTyped := vr.(*vectorRenderer)
	assert.True(t, isTyped)

	typed.MoveTo(0, 0)
	typed.LineTo(100, 100)
	typed.LineTo(0, 100)
	typed.Close()
	typed.FillStroke()

	buffer := bytes.NewBuffer([]byte{})
	require.NoError(t, typed.Save(buffer))

	raw := buffer.String()
	assert.True(t, strings.HasPrefix(raw, "<svg"))
	assert.True(t, strings.HasSuffix(raw, "</svg>"))
}

func TestVectorRendererSaveTwice(t *testing.T) {
	t.Parallel()

	vr := SVG(100, 100).(*vectorRenderer)
	vr.SetFont(GetDefaultFont())
	vr.SetFontSize(10)
	vr.MoveTo(0, 0)
	vr.LineTo(50, 60)
	vr.Close()
	vr.FillStroke()
	vr.Text("hello", 5, 20)

	first := bytes.Buffer{}
	require.NoError(t, vr.Save(&first))
	assert.Equal(t, 1, strings.Count(first.String(), "</svg>"))

	second := bytes.Buffer{}
	require.NoError(t, vr.Save(&second))
	assert.Equal(t, first.Bytes(), second.Bytes())
	assert.Equal(t, 1, strings.Count(second.String(), "</svg>"))

	dec := xml.NewDecoder(&second)
	for {
		if _, err := dec.Token(); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			t.Fatal(err)
		}
	}
}

func TestVectorRendererSaveReflectsModifications(t *testing.T) {
	t.Parallel()

	vr := SVG(100, 50).(*vectorRenderer)
	vr.SetFont(GetDefaultFont())
	vr.SetFontSize(10)

	first := bytes.Buffer{}
	require.NoError(t, vr.Save(&first))
	assert.NotContains(t, first.String(), "<text")

	// draw after the first save; must appear in later snapshots
	vr.Text("late", 5, 20)
	second := bytes.Buffer{}
	require.NoError(t, vr.Save(&second))

	out := second.String()
	assert.Contains(t, out, "<text")
	assert.Equal(t, 1, strings.Count(out, "</svg>"))

	dec := xml.NewDecoder(&second)
	for {
		if _, err := dec.Token(); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			t.Fatal(err)
		}
	}
}

func TestVectorRendererMeasureText(t *testing.T) {
	t.Parallel()

	vr := SVG(100, 100)

	vr.SetDPI(defaultDPI)
	vr.SetFont(GetDefaultFont())
	vr.SetFontSize(12.0)

	tb := vr.MeasureText("Ljp")
	assert.Equal(t, 21, tb.Width())
	assert.Equal(t, 16, tb.Height())
}

func TestCanvasStyleSVG(t *testing.T) {
	t.Parallel()

	set := Style{
		StrokeColor: drawing.ColorWhite,
		StrokeWidth: 5.0,
		FillColor:   drawing.ColorWhite,
		FontStyle: FontStyle{
			FontColor: drawing.ColorWhite,
			Font:      GetDefaultFont(),
			FontSize:  12,
		},
		Padding: DefaultBackgroundPadding,
	}

	var bb bytes.Buffer
	styleAsSVG(&bb, set, defaultDPI, false)
	svgString := bb.String()
	assert.NotEmpty(t, svgString)
	assert.True(t, strings.HasPrefix(svgString, "style=\""))
	assert.Contains(t, svgString, "stroke:white")
	assert.Contains(t, svgString, "stroke-width:5")
	assert.Contains(t, svgString, "fill:white")
	assert.NotContains(t, svgString, "font-size")
	assert.NotContains(t, svgString, "font-family")
	assert.True(t, strings.HasSuffix(svgString, "\""))

	bb.Reset()
	styleAsSVG(&bb, set, defaultDPI, true)
	svgString = bb.String()
	assert.True(t, strings.HasPrefix(svgString, "style=\""))
	assert.Contains(t, svgString, "stroke:white")
	assert.Contains(t, svgString, "stroke-width:5")
	assert.Contains(t, svgString, "fill:white")
	assert.Contains(t, svgString, "font-size")
	assert.Contains(t, svgString, "font-family")
	assert.True(t, strings.HasSuffix(svgString, "\""))

	set.StrokeWidth = 0.25
	bb.Reset()
	styleAsSVG(&bb, set, defaultDPI, false)
	assert.Contains(t, bb.String(), "stroke-width:0.25")
}

func TestCanvasClassSVG(t *testing.T) {
	t.Parallel()

	set := Style{
		ClassName: "test-class",
	}

	var bb bytes.Buffer
	styleAsSVG(&bb, set, defaultDPI, false)
	assert.Equal(t, "class=\"test-class\"", bb.String())
}

func TestCanvasCustomInlineStylesheet(t *testing.T) {
	t.Parallel()

	b := strings.Builder{}

	canvas := &canvas{
		w:   &b,
		bb:  bytes.NewBuffer(make([]byte, 0, 80)),
		css: ".background { fill: red }",
	}

	canvas.Start(200, 200)

	assert.Contains(t, b.String(), fmt.Sprintf(`<style type="text/css"><![CDATA[%s]]></style>`, canvas.css))
}

func TestCanvasCustomInlineStylesheetWithNonce(t *testing.T) {
	t.Parallel()

	b := strings.Builder{}

	canvas := &canvas{
		w:     &b,
		bb:    bytes.NewBuffer(make([]byte, 0, 80)),
		css:   ".background { fill: red }",
		nonce: "RAND0MSTRING",
	}

	canvas.Start(200, 200)

	assert.Contains(t, b.String(), fmt.Sprintf(`<style type="text/css" nonce="%s"><![CDATA[%s]]></style>`, canvas.nonce, canvas.css))
}

func TestSVGWithCSS(t *testing.T) {
	t.Parallel()

	maker := SVGWithCSS(".cls{fill:red}", "nonce")
	r := maker(10, 10)
	vr, ok := r.(*vectorRenderer)
	require.True(t, ok)

	b := bytes.Buffer{}
	require.NoError(t, vr.Save(&b))
	out := b.String()
	assert.Contains(t, out, "nonce=\"nonce\"")
	assert.Contains(t, out, ".cls{fill:red}")
}

func TestCanvasBasicElements(t *testing.T) {
	t.Parallel()

	b := strings.Builder{}
	c := &canvas{w: &b, bb: bytes.NewBuffer(make([]byte, 0, 80))}
	c.Start(50, 50)

	c.Path([]string{"M 0 0", "L 10 10"}, Style{StrokeDashArray: []float64{1, 2}, StrokeWidth: 2, StrokeColor: drawing.ColorBlack})
	c.Text(5, 5, "hi", Style{FontStyle: FontStyle{Font: GetDefaultFont(), FontSize: 10}})
	c.Circle(5, 5, 3, Style{FillColor: drawing.ColorRed})

	out := b.String()
	assert.Contains(t, out, "stroke-dasharray=\"1, 2\"")
	assert.Contains(t, out, "<text")
	assert.Contains(t, out, "<circle")
}

func TestCanvasPathDashArray(t *testing.T) {
	t.Parallel()

	pathDash := func(dash []float64) string {
		b := strings.Builder{}
		c := &canvas{w: &b, bb: bytes.NewBuffer(make([]byte, 0, 80))}
		c.Path([]string{"M 0 0", "L 10 10"}, Style{StrokeDashArray: dash, StrokeWidth: 2, StrokeColor: drawing.ColorBlack})
		return b.String()
	}

	t.Run("valid_dash", func(t *testing.T) {
		assert.Contains(t, pathDash([]float64{1, 2}), "stroke-dasharray=\"1, 2\"")
	})

	t.Run("fractional_dash", func(t *testing.T) {
		assert.Contains(t, pathDash([]float64{0.25, 2}), "stroke-dasharray=\"0.25, 2\"")
	})

	t.Run("degenerate_dash_omitted", func(t *testing.T) {
		for _, dash := range [][]float64{{0, 0}, {-5, 5}, {5, -5}} {
			assert.NotContains(t, pathDash(dash), "stroke-dasharray", dash)
		}
	})
}

func TestCanvasTextEscaping(t *testing.T) {
	t.Parallel()

	t.Run("escapes_special_chars", func(t *testing.T) {
		b := strings.Builder{}
		c := &canvas{w: &b, bb: bytes.NewBuffer(make([]byte, 0, 80))}
		c.Text(5, 5, `a<b & c"'`, Style{FontStyle: FontStyle{Font: GetDefaultFont(), FontSize: 10}})

		out := b.String()
		assert.Contains(t, out, `a&lt;b &amp; c"'`)
		assert.NotContains(t, out, "a<b & c")
	})

	t.Run("valid_xml", func(t *testing.T) {
		vr := SVG(100, 100).(*vectorRenderer)
		vr.SetFont(GetDefaultFont())
		vr.SetFontSize(10)
		vr.Text("P&G x<10 A>B", 5, 5)

		buf := bytes.Buffer{}
		require.NoError(t, vr.Save(&buf))

		dec := xml.NewDecoder(&buf)
		for {
			_, err := dec.Token()
			if errors.Is(err, io.EOF) {
				break
			}
			require.NoError(t, err)
		}
	})
}

func TestFormatFloatMinimized(t *testing.T) {
	t.Parallel()

	t.Run("trailing_zeros_trimmed", func(t *testing.T) {
		assert.Equal(t, "1", formatFloatMinimized(1, 2))
		assert.Equal(t, "1.2", formatFloatMinimized(1.20, 2))
		assert.Equal(t, "2.5", formatFloatMinimized(2.50, 2))
		assert.Equal(t, "1000", formatFloatMinimized(1000.0000001, 2))
	})

	t.Run("precision_respected", func(t *testing.T) {
		assert.Equal(t, "0.25", formatFloatMinimized(0.25, 2))
		assert.Equal(t, "0.2", formatFloatMinimized(0.25, 1))
		assert.Equal(t, "0", formatFloatMinimized(0.25, 0))
	})

	t.Run("negative_zero", func(t *testing.T) {
		assert.Equal(t, "0", formatFloatMinimized(-1.2e-14, 2))
		assert.Equal(t, "0", formatFloatMinimized(-1.2e-14, 0))
	})
}

func TestVectorRendererArcTo(t *testing.T) {
	t.Parallel()

	arcPath := func(moveFirst bool, rx, ry, startAngle, delta float64) []string {
		vr := SVG(100, 100).(*vectorRenderer)
		if moveFirst {
			vr.MoveTo(50, 50)
		}
		vr.ArcTo(50, 50, rx, ry, startAngle, delta)
		return vr.p
	}

	t.Run("no_x_axis_rotation", func(t *testing.T) {
		p := arcPath(false, 20, 20, 0, _pi2)
		require.Len(t, p, 2)
		assert.Equal(t, "M 70 50", p[0])
		assert.Equal(t, "A 20 20 0 0 1 50 70", p[1])
	})

	t.Run("elliptical_arc", func(t *testing.T) {
		p := arcPath(false, 30, 10, 0, _pi2)
		require.Len(t, p, 2)
		assert.Equal(t, "M 80 50", p[0])
		assert.Equal(t, "A 30 10 0 0 1 50 60", p[1])
	})

	t.Run("line_to_start_when_path_open", func(t *testing.T) {
		p := arcPath(true, 20, 20, 0, _pi2)
		require.Len(t, p, 3)
		assert.Equal(t, "M 50 50", p[0])
		assert.Equal(t, "L 70 50", p[1])
	})

	t.Run("negative_delta", func(t *testing.T) {
		p := arcPath(false, 20, 20, 0, -_pi2)
		require.Len(t, p, 2)
		assert.Equal(t, "A 20 20 0 0 0 50 30", p[1])
	})

	t.Run("large_arc_clockwise", func(t *testing.T) {
		p := arcPath(false, 20, 20, 0, _3pi2)
		require.Len(t, p, 2)
		assert.Equal(t, "A 20 20 0 1 1 50 30", p[1])
	})

	t.Run("large_arc_counter_clockwise", func(t *testing.T) {
		p := arcPath(false, 20, 20, 0, -_3pi2)
		require.Len(t, p, 2)
		assert.Equal(t, "A 20 20 0 1 0 50 70", p[1])
	})

	t.Run("semicircle_flags", func(t *testing.T) {
		p := arcPath(false, 20, 20, 0, _pi)
		require.Len(t, p, 2)
		assert.Equal(t, "A 20 20 0 0 1 30 50", p[1])
	})

	t.Run("full_circle_delta", func(t *testing.T) {
		p := arcPath(false, 20, 20, 0, _2pi)
		require.Len(t, p, 3)
		assert.Equal(t, "M 70 50", p[0])
		assert.Equal(t, "A 20 20 0 0 1 30 50", p[1])
		assert.Equal(t, "A 20 20 0 0 1 70 50", p[2]) // returns to the start point
	})

	t.Run("full_circle_negative", func(t *testing.T) {
		p := arcPath(false, 20, 20, 0, -_2pi)
		require.Len(t, p, 3)
		assert.Equal(t, "A 20 20 0 0 0 30 50", p[1])
		assert.Equal(t, "A 20 20 0 0 0 70 50", p[2])
	})

	t.Run("full_circle_from_percent", func(t *testing.T) {
		assert.Len(t, arcPath(false, 20, 20, 0, PercentToRadians(1.0)), 3)
	})

	t.Run("over_full_circle_clamped", func(t *testing.T) {
		assert.Equal(t, arcPath(false, 20, 20, 0, _2pi), arcPath(false, 20, 20, 0, 3*_pi))
	})

	t.Run("fractional_geometry", func(t *testing.T) {
		p := arcPath(true, 20.25, 20.25, 0, _pi2)
		require.Len(t, p, 3)
		assert.Equal(t, "L 70.25 50", p[1])
		assert.Equal(t, "A 20.25 20.25 0 0 1 50 70.25", p[2])
	})

	t.Run("zero_delta", func(t *testing.T) {
		p := arcPath(false, 20, 20, 0, 0)
		require.Len(t, p, 2)
		assert.Equal(t, "M 70 50", p[0])
		assert.Equal(t, "A 20 20 0 0 0 70 50", p[1])
	})

	t.Run("near_full_circle_split", func(t *testing.T) {
		// endpoints collide at svgPrecision, so a single arc would render as nothing
		p := arcPath(false, 80, 80, 0, _2pi-1e-6)
		require.Len(t, p, 3)
		assert.Equal(t, "M 130 50", p[0])
		assert.Equal(t, "A 80 80 0 0 1 -30 50", p[1])
		assert.Equal(t, "A 80 80 0 0 1 130 50", p[2])
	})

	t.Run("short_chord_split", func(t *testing.T) {
		// endpoints are distinct but too close to pin the arc center after rounding
		assert.Len(t, arcPath(false, 80, 80, 0, _2pi-arcSplitGap/2), 3)
		assert.Len(t, arcPath(false, 80, 80, 0, -(_2pi-arcSplitGap/2)), 3)
	})

	t.Run("wide_gap_single_segment", func(t *testing.T) {
		assert.Len(t, arcPath(false, 80, 80, 0, _2pi-2*arcSplitGap), 2)
	})

	t.Run("non_finite_params", func(t *testing.T) {
		nan, posInf, negInf := math.NaN(), math.Inf(1), math.Inf(-1)
		for _, v := range []float64{nan, posInf, negInf} {
			assert.Empty(t, arcPath(false, v, 20, 0, _pi2))
			assert.Empty(t, arcPath(false, 20, v, 0, _pi2))
			assert.Empty(t, arcPath(false, 20, 20, v, _pi2))
			assert.Empty(t, arcPath(false, 20, 20, 0, v))
		}
	})
}

func TestVectorRendererCircle(t *testing.T) {
	t.Parallel()

	circleSVG := func(radius float64) string {
		vr := SVG(20, 20).(*vectorRenderer)
		vr.SetFillColor(drawing.ColorRed)
		vr.Circle(radius, 10, 10)

		buf := bytes.Buffer{}
		require.NoError(t, vr.Save(&buf))
		return buf.String()
	}

	t.Run("integral_radius", func(t *testing.T) {
		assert.Contains(t, circleSVG(3), `r="3"`)
	})

	t.Run("fractional_radius", func(t *testing.T) {
		assert.Contains(t, circleSVG(2.4), `r="2.4"`)
	})

	t.Run("sub_pixel_radius", func(t *testing.T) {
		out := circleSVG(0.4)
		assert.Contains(t, out, `r="0.4"`)
		assert.NotContains(t, out, `r="0"`)
	})
}

func TestVectorRendererTextRotation(t *testing.T) {
	t.Parallel()

	vr := SVG(20, 20).(*vectorRenderer)
	vr.SetClassName("cls")
	vr.SetStrokeColor(drawing.ColorBlack)
	vr.SetFillColor(drawing.ColorRed)
	vr.SetTextRotation(math.Pi / 2)
	vr.Text("A", 10, 10)
	vr.ClearTextRotation()
	vr.Text("B", 10, 15)

	buf := bytes.Buffer{}
	require.NoError(t, vr.Save(&buf))
	out := buf.String()
	assert.Contains(t, out, "class=\"cls\"")
	assert.Contains(t, out, "rotate(90.00")
	assert.Contains(t, out, "B</text>")
}
