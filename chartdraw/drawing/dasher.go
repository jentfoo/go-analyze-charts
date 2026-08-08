package drawing

import (
	"math"
	"slices"
)

// ValidDash reports whether the dash pattern can be rendered, invalid patterns should be
// rendered as a solid line.
func ValidDash(dash []float64) bool {
	if len(dash) == 0 || hasNonFinite(dash) {
		return false
	}
	var sum float64
	for _, d := range dash {
		if d < 0 {
			return false
		}
		sum += d
	}
	return sum > 0
}

// dashable reports whether the dash state can be walked without stalling the dasher.
func dashable(dash []float64, dashOffset float64) bool {
	return ValidDash(dash) && !math.IsNaN(dashOffset) && !math.IsInf(dashOffset, 0)
}

// NewDashVertexConverter creates a new dash converter. Odd length patterns are repeated so line
// and gap alternate per the SVG stroke-dasharray spec, and dashOffset is wrapped into one cycle.
func NewDashVertexConverter(dash []float64, dashOffset float64, flattener Flattener) *DashVertexConverter {
	var dasher DashVertexConverter
	if len(dash)%2 == 1 {
		dasher.dash = append(slices.Clone(dash), dash...) // clone, the caller keeps its slice
	} else {
		dasher.dash = dash
	}
	var sum float64
	for _, d := range dasher.dash {
		sum += d
	}
	if sum > 0 {
		dashOffset = math.Mod(dashOffset, sum)
		if dashOffset < 0 {
			dashOffset += sum
		}
	}
	dasher.dashOffset = dashOffset
	dasher.next = flattener
	return &dasher
}

// DashVertexConverter is a converter for dash vertexes.
type DashVertexConverter struct {
	next           Flattener
	x, y, distance float64
	dash           []float64
	currentDash    int
	dashOffset     float64
}

// LineTo adds a dashed line segment to the path (for PathBuilder interface).
func (dasher *DashVertexConverter) LineTo(x, y float64) {
	dasher.lineTo(x, y)
}

// MoveTo sets the starting point for the dashed path (for PathBuilder interface).
func (dasher *DashVertexConverter) MoveTo(x, y float64) {
	dasher.next.MoveTo(x, y)
	dasher.x, dasher.y = x, y
	dasher.distance = dasher.dashOffset
	dasher.currentDash = 0
}

// End forwards the completed path to the next converter (for PathBuilder interface).
func (dasher *DashVertexConverter) End() {
	dasher.next.End()
}

func (dasher *DashVertexConverter) lineTo(x, y float64) {
	rest := dasher.dash[dasher.currentDash] - dasher.distance
	for rest < 0 {
		dasher.distance -= dasher.dash[dasher.currentDash]
		dasher.currentDash = (dasher.currentDash + 1) % len(dasher.dash)
		rest = dasher.dash[dasher.currentDash] - dasher.distance
	}
	d := distance(dasher.x, dasher.y, x, y)
	for d > 0 && d >= rest { // d > 0 avoids a 0/0 split on a zero length dash entry
		k := rest / d
		lx := dasher.x + k*(x-dasher.x)
		ly := dasher.y + k*(y-dasher.y)
		dasher.emit(lx, ly)
		d -= rest
		dasher.x, dasher.y = lx, ly
		dasher.currentDash = (dasher.currentDash + 1) % len(dasher.dash)
		rest = dasher.dash[dasher.currentDash]
	}
	dasher.distance = dasher.dash[dasher.currentDash] - rest + d
	dasher.emit(x, y)
	if dasher.distance >= dasher.dash[dasher.currentDash] {
		// ended on the boundary, consume the exhausted entry
		dasher.distance -= dasher.dash[dasher.currentDash]
		dasher.currentDash = (dasher.currentDash + 1) % len(dasher.dash)
	}
	dasher.x, dasher.y = x, y
}

// emit draws to the point when the current dash entry is on, otherwise breaks the subpath.
func (dasher *DashVertexConverter) emit(x, y float64) {
	if dasher.currentDash%2 == 0 {
		dasher.next.LineTo(x, y)
	} else { // gap
		dasher.next.End()
		dasher.next.MoveTo(x, y)
	}
}
