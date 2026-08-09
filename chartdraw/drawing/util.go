package drawing

import (
	"math"
	"slices"

	"golang.org/x/image/math/fixed"

	"github.com/golang/freetype/truetype"
)

// PointsToPixels returns the pixels for a given number of points at a DPI.
func PointsToPixels(dpi, points float64) (pixels float64) {
	return (points * dpi) / 72.0
}

// isNonFinite reports whether v is NaN or infinite.
func isNonFinite(v float64) bool {
	return math.IsNaN(v) || math.IsInf(v, 0)
}

// hasNonFinite reports whether any value is NaN or infinite.
func hasNonFinite(vals []float64) bool {
	return slices.ContainsFunc(vals, isNonFinite)
}

func absInt(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func distance(x1, y1, x2, y2 float64) float64 {
	return vectorDistance(x2-x1, y2-y1)
}

// vectorDistance returns the length of the vector, finite for any finite input
func vectorDistance(dx, dy float64) float64 {
	return math.Hypot(dx, dy) // squaring would overflow to +Inf past ~1.3e154
}

func pointToF64Point(p truetype.Point) (x, y float64) {
	return fUnitsToFloat64(p.X), -fUnitsToFloat64(p.Y)
}

func fUnitsToFloat64(x fixed.Int26_6) float64 {
	scaled := x << 2
	return float64(scaled/256) + float64(scaled%256)/256.0
}
