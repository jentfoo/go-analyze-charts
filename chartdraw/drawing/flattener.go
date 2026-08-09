package drawing

// Liner receive segment definition.
type Liner interface {
	// LineTo Draw a line from the current position to the point (x, y).
	LineTo(x, y float64)
}

// Flattener receive segment definition.
type Flattener interface {
	// MoveTo Start a New line from the point (x, y).
	MoveTo(x, y float64)
	// LineTo Draw a line from the current position to the point (x, y).
	LineTo(x, y float64)
	// End mark the current line as finished.
	End()
}

// Flatten convert curves into straight segments keeping join segments info.
func Flatten(path *Path, flattener Flattener, scale float64) {
	var startX, startY float64 // moveTo point starting a path
	var x, y float64           // current point
	var i int
	for _, cmp := range path.Components {
		switch cmp {
		case MoveToComponent:
			x, y = path.Points[i], path.Points[i+1]
			startX, startY = x, y
			if i != 0 {
				flattener.End()
			}
			flattener.MoveTo(x, y)
			i += 2
		case LineToComponent:
			x, y = path.Points[i], path.Points[i+1]
			flattener.LineTo(x, y)
			i += 2
		case QuadCurveToComponent:
			// prepend current point as the curve start
			quad := [6]float64{x, y}
			copy(quad[2:], path.Points[i:i+4])
			TraceQuad(flattener, quad[:], defaultFlatteningThreshold)
			end := [2]float64{path.Points[i+2], path.Points[i+3]}
			if !hasNonFinite(end[:]) { // only reachable on a hand built path, the builder rejects these
				x, y = end[0], end[1]
				flattener.LineTo(x, y)
			}
			i += 4
		case CubicCurveToComponent:
			cubic := [8]float64{x, y}
			copy(cubic[2:], path.Points[i:i+6])
			TraceCubic(flattener, cubic[:], defaultFlatteningThreshold)
			end := [2]float64{path.Points[i+4], path.Points[i+5]}
			if !hasNonFinite(end[:]) { // only reachable on a hand built path, the builder rejects these
				x, y = end[0], end[1]
				flattener.LineTo(x, y)
			}
			i += 6
		case ArcToComponent:
			ax, ay := TraceArc(flattener, path.Points[i], path.Points[i+1], path.Points[i+2], path.Points[i+3], path.Points[i+4], path.Points[i+5], scale)
			end := [2]float64{ax, ay}
			if !hasNonFinite(end[:]) { // a suppressed arc must not move the pen to NaN
				x, y = ax, ay
				flattener.LineTo(x, y)
			}
			i += 6
		case CloseComponent:
			if x != startX || y != startY {
				flattener.LineTo(startX, startY)
			}
			x, y = startX, startY // pen is at the subpath start
		}
	}
	flattener.End()
}

// SegmentedPath is a path of disparate point sections.
type SegmentedPath struct {
	Points []float64
}

// MoveTo records the first point of a new segment (for PathBuilder interface).
func (p *SegmentedPath) MoveTo(x, y float64) {
	p.Points = append(p.Points, x, y)
}

// LineTo appends a point to the current path segment (for PathBuilder interface).
func (p *SegmentedPath) LineTo(x, y float64) {
	p.Points = append(p.Points, x, y)
}

// End completes the current path segment (for PathBuilder interface).
func (p *SegmentedPath) End() {
	// Nothing to do
}
