package drawing

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// PolylineBresenham draws a polyline to an image. Segments which can not be rounded into int
// coordinates, or which fall entirely outside the image bounds, are skipped.
func PolylineBresenham(img draw.Image, c color.Color, s ...float64) {
	b := img.Bounds()
	for i := 2; i+1 < len(s); i += 2 {
		seg := [4]float64{s[i-2], s[i-1], s[i], s[i+1]}
		if hasOutOfIntRange(seg[:]) {
			continue // an implementation defined conversion would walk the whole int space
		}
		x0, y0 := int(math.Round(seg[0])), int(math.Round(seg[1]))
		x1, y1 := int(math.Round(seg[2])), int(math.Round(seg[3]))
		if max(x0, x1) < b.Min.X || min(x0, x1) >= b.Max.X ||
			max(y0, y1) < b.Min.Y || min(y0, y1) >= b.Max.Y {
			continue // every pixel the segment could paint is out of bounds
		}
		Bresenham(img, c, x0, y0, x1, y1)
	}
}

// maxBresenhamSteps is the longest segment walked as given, past it the segment is clipped. It sits
// far above any image dimension because clipping shifts the pixels a partially visible line paints.
const maxBresenhamSteps = 1 << 20

// Bresenham draws a line between (x0, y0) and (x1, y1). Segments over a million pixels long are
// clipped to the image bounds first.
func Bresenham(img draw.Image, color color.Color, x0, y0, x1, y1 int) {
	// subtracted as float64, ends near the int limits overflow the deltas below
	if max(math.Abs(float64(x1)-float64(x0)), math.Abs(float64(y1)-float64(y0))) > maxBresenhamSteps {
		fx0, fy0, fx1, fy1, ok := clipSegment(img.Bounds(),
			float64(x0), float64(y0), float64(x1), float64(y1))
		if !ok {
			return // no pixel the segment could paint is in bounds
		}
		x0, y0 = int(math.Round(fx0)), int(math.Round(fy0))
		x1, y1 = int(math.Round(fx1)), int(math.Round(fy1))
	}

	dx := absInt(x1 - x0)
	dy := absInt(y1 - y0)
	var sx, sy int
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}
	err := dx - dy

	var e2 int
	for {
		img.Set(x0, y0, color)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 = 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// clipSegment clips the segment to the area the rectangle covers, returning the clipped endpoints
// and whether any part of the segment falls within it.
func clipSegment(b image.Rectangle, x0, y0, x1, y1 float64) (float64, float64, float64, float64, bool) {
	if b.Empty() {
		return 0, 0, 0, 0, false
	}
	dx, dy := x1-x0, y1-y0
	// half a pixel of margin, a rasterized line can round into a pixel the exact segment misses
	lim := [4]float64{float64(b.Min.X) - 0.5, float64(b.Max.X) - 0.5,
		float64(b.Min.Y) - 0.5, float64(b.Max.Y) - 0.5}
	edges := [4][2]float64{ // parameter scale, signed distance to the edge
		{-dx, x0 - lim[0]}, {dx, lim[1] - x0},
		{-dy, y0 - lim[2]}, {dy, lim[3] - y0},
	}
	t0, t1 := 0.0, 1.0
	e0, e1 := -1, -1 // edge each end was clipped against
	for i, e := range edges {
		p, q := e[0], e[1]
		if p == 0 {
			if q < 0 {
				return 0, 0, 0, 0, false // parallel to the edge and outside it
			}
			continue
		}
		r := q / p
		if p < 0 {
			if r > t1 {
				return 0, 0, 0, 0, false
			} else if r > t0 {
				t0, e0 = r, i
			}
		} else if r < t0 {
			return 0, 0, 0, 0, false
		} else if r < t1 {
			t1, e1 = r, i
		}
	}
	cx0, cy0 := clipEnd(b, x0+t0*dx, y0+t0*dy, e0, lim)
	cx1, cy1 := clipEnd(b, x0+t1*dx, y0+t1*dy, e1, lim)
	return cx0, cy0, cx1, cy1, true
}

// clipEnd snaps a clipped end onto the edge it met, then holds it on a pixel of the rectangle.
func clipEnd(b image.Rectangle, x, y float64, edge int, lim [4]float64) (float64, float64) {
	switch edge {
	case 0, 1:
		x = lim[edge] // the parameter loses the edge to a rounding step at large magnitudes
	case 2, 3:
		y = lim[edge]
	}
	return min(max(x, float64(b.Min.X)), float64(b.Max.X-1)),
		min(max(y, float64(b.Min.Y)), float64(b.Max.Y-1))
}
