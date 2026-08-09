package drawing

import (
	"math"

	"github.com/golang/freetype/raster"
	"golang.org/x/image/math/fixed"
)

// FtLineBuilder is a builder for freetype raster glyphs.
type FtLineBuilder struct {
	Adder raster.Adder
}

// MoveTo starts a new segment at the given point (for PathBuilder interface).
func (liner FtLineBuilder) MoveTo(x, y float64) {
	liner.Adder.Start(fixed.Point26_6{X: toFixed26_6(x), Y: toFixed26_6(y)})
}

// LineTo adds a line segment from the current point to the specified point (for PathBuilder interface).
func (liner FtLineBuilder) LineTo(x, y float64) {
	liner.Adder.Add1(fixed.Point26_6{X: toFixed26_6(x), Y: toFixed26_6(y)})
}

// End finalizes the current path (for PathBuilder interface).
func (liner FtLineBuilder) End() {}

// maxRasterPixel bounds a coordinate handed to the rasterizer, capping the range an image can
// render rather than the ±2^25 the 26.6 format holds. freetype scales edge deltas by 64 in int32,
// and edges measurably wrap between 2^19 and 2^20 pixels, so this keeps a factor of four margin.
// Clamping per axis projects an out of range end onto the bound rather than clipping the edge,
// which keeps fill winding intact.
const maxRasterPixel = 1 << 17

// toFixed26_6 rounds a pixel coordinate to 26.6 fixed point, clamping values out of range or not
// finite to ±maxRasterPixel.
func toFixed26_6(v float64) fixed.Int26_6 {
	const bound = maxRasterPixel * 64
	scaled := math.Round(v * 64)
	if !(scaled > -bound) { // also catches NaN
		return -bound
	} else if scaled > bound {
		return bound
	}
	return fixed.Int26_6(scaled)
}
