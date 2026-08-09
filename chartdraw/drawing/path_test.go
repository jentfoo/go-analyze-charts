package drawing

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

type recordFlattener struct {
	moves []string
}

func (r *recordFlattener) MoveTo(x, y float64) {
	r.moves = append(r.moves, fmt.Sprintf("M%.1f,%.1f", x, y))
}

func (r *recordFlattener) LineTo(x, y float64) {
	r.moves = append(r.moves, fmt.Sprintf("L%.1f,%.1f", x, y))
}

func (r *recordFlattener) End() {}

func TestPathBasicOps(t *testing.T) {
	t.Parallel()

	p := &Path{}
	p.LineTo(1, 2)
	p.LineTo(3, 4)
	assert.Equal(t, []PathComponent{MoveToComponent, LineToComponent, LineToComponent}, p.Components)
	assert.InDeltaSlice(t, []float64{0, 0, 1, 2, 3, 4}, p.Points, 0.0001)
}

func TestPathArcTo(t *testing.T) {
	t.Parallel()

	p := &Path{}
	p.ArcTo(0, 0, 1, 1, 0, math.Pi/2)

	expectX := 0.0
	expectY := 1.0
	assert.InDelta(t, expectX, p.x, 0.0001)
	assert.InDelta(t, expectY, p.y, 0.0001)
	assert.InDeltaSlice(t, []float64{1, 0, 0, 0, 1, 1, 0, math.Pi / 2}, p.Points, 0.0001)
	assert.Equal(t, MoveToComponent, p.Components[0])
	assert.Equal(t, ArcToComponent, p.Components[1])
}

func TestTransformer(t *testing.T) {
	t.Parallel()

	rec := &recordFlattener{}
	tr := Transformer{Tr: NewTranslationMatrix(2, 3), Flattener: rec}
	tr.MoveTo(1, 1)
	tr.LineTo(2, 2)
	tr.End()

	assert.Equal(t, []string{"M3.0,4.0", "L4.0,5.0"}, rec.moves)
}

func TestPathCurveAndClose(t *testing.T) {
	t.Parallel()

	p := &Path{}
	p.QuadCurveTo(1, 1, 2, 2)
	p.CubicCurveTo(3, 3, 4, 4, 5, 5)
	p.Close()

	expComp := []PathComponent{MoveToComponent, QuadCurveToComponent, CubicCurveToComponent, CloseComponent}
	assert.Equal(t, expComp, p.Components)
	assert.InDeltaSlice(t, []float64{0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5}, p.Points, 0.0001)
	// close returns the pen to the implicit MoveTo(0,0) subpath start
	assert.InDelta(t, 0.0, p.x, 0.0001)
	assert.InDelta(t, 0.0, p.y, 0.0001)
}

func TestPathCopyClearIsEmpty(t *testing.T) {
	t.Parallel()

	p := &Path{}
	p.LineTo(1, 1)
	copyP := p.Copy()

	p.Clear()
	assert.True(t, p.IsEmpty())
	assert.False(t, copyP.IsEmpty())

	p2 := &Path{}
	assert.True(t, p2.IsEmpty())
}

func TestPathString(t *testing.T) {
	t.Parallel()

	p := &Path{}
	p.MoveTo(0, 0)
	p.LineTo(1, 1)
	p.QuadCurveTo(2, 2, 3, 3)
	p.CubicCurveTo(4, 4, 5, 5, 6, 6)
	p.Close()

	got := p.String()
	expect := "" +
		"MoveTo: 0.000000, 0.000000\n" +
		"LineTo: 1.000000, 1.000000\n" +
		"QuadCurveTo: 2.000000, 2.000000, 3.000000, 3.000000\n" +
		"CubicCurveTo: 4.000000, 4.000000, 5.000000, 5.000000, 6.000000, 6.000000\n" +
		"Close\n"
	assert.Equal(t, expect, got)
}

func TestPathLastPointAndArc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		ops   func(*Path)
		x, y  float64
		delta float64
	}{
		{
			name: "line and quad",
			ops: func(p *Path) {
				p.MoveTo(1, 1)
				p.LineTo(2, 2)
				p.QuadCurveTo(3, 3, 4, 4)
			},
			x: 4, y: 4, delta: 0,
		},
		{
			name: "arc positive",
			ops: func(p *Path) {
				p.ArcTo(0, 0, 1, 1, 0, math.Pi/2)
			},
			x: 0, y: 1, delta: 0.0001,
		},
		{
			name: "arc negative",
			ops: func(p *Path) {
				p.ArcTo(0, 0, 1, 1, math.Pi/2, -math.Pi/2)
			},
			x: 1, y: 0, delta: 0.0001,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &Path{}
			tc.ops(p)
			x, y := p.LastPoint()
			assert.InDelta(t, tc.x, x, tc.delta)
			assert.InDelta(t, tc.y, y, tc.delta)
		})
	}
}

func TestPathMoveToSubpathFusion(t *testing.T) {
	t.Parallel()

	t.Run("moveto_after_lineto_kept", func(t *testing.T) {
		p := &Path{}
		p.MoveTo(0, 0)
		p.LineTo(5, 5)
		p.MoveTo(5, 5) // starts a new subpath where the previous ended
		p.LineTo(10, 0)

		expected := []PathComponent{
			MoveToComponent,
			LineToComponent,
			MoveToComponent,
			LineToComponent,
		}
		assert.Equal(t, expected, p.Components)
	})

	t.Run("duplicate_moveto_skipped", func(t *testing.T) {
		p := &Path{}
		p.MoveTo(1, 1)
		p.MoveTo(1, 1)

		assert.Equal(t, []PathComponent{MoveToComponent}, p.Components)
	})
}

func TestPathClosePenPosition(t *testing.T) {
	t.Parallel()

	t.Run("line_after_close_kept", func(t *testing.T) {
		p := &Path{}
		p.MoveTo(0, 0)
		p.LineTo(10, 10)
		p.Close()
		p.LineTo(10, 10) // real segment from the subpath start, deduped away when the pen was stale

		expected := []PathComponent{
			MoveToComponent,
			LineToComponent,
			CloseComponent,
			LineToComponent,
		}
		assert.Equal(t, expected, p.Components)
	})

	t.Run("last_point_after_close", func(t *testing.T) {
		p := &Path{}
		p.MoveTo(3, 4)
		p.LineTo(10, 10)
		p.Close()

		x, y := p.LastPoint()
		assert.InDelta(t, 3.0, x, 0.0001)
		assert.InDelta(t, 4.0, y, 0.0001)
	})

	t.Run("arc_after_close_connects", func(t *testing.T) {
		// the arc starts at (1,0), the stale pre-close pen, so its connecting line was dropped
		p := &Path{}
		p.MoveTo(0, 0)
		p.LineTo(1, 0)
		p.Close()
		p.ArcTo(0, 0, 1, 1, 0, math.Pi/2)

		expected := []PathComponent{
			MoveToComponent,
			LineToComponent,
			CloseComponent,
			LineToComponent,
			ArcToComponent,
		}
		assert.Equal(t, expected, p.Components)
	})

	t.Run("second_subpath_start", func(t *testing.T) {
		p := &Path{}
		p.MoveTo(0, 0)
		p.LineTo(5, 5)
		p.MoveTo(20, 20)
		p.LineTo(25, 25)
		p.Close()

		x, y := p.LastPoint()
		assert.InDelta(t, 20.0, x, 0.0001)
		assert.InDelta(t, 20.0, y, 0.0001)
	})

	t.Run("copy_preserves_start", func(t *testing.T) {
		p := &Path{}
		p.MoveTo(3, 4)
		p.LineTo(10, 10)
		copyP := p.Copy()
		copyP.Close()

		x, y := copyP.LastPoint()
		assert.InDelta(t, 3.0, x, 0.0001)
		assert.InDelta(t, 4.0, y, 0.0001)
	})

	t.Run("clear_resets_pen", func(t *testing.T) {
		p := &Path{}
		p.MoveTo(3, 4)
		p.LineTo(10, 10)
		p.Clear()

		x, y := p.LastPoint()
		assert.Zero(t, x)
		assert.Zero(t, y)
	})
}

func TestPathArcToNonFinite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                              string
		cx, cy, rx, ry, startAngle, delta float64
	}{
		{"nan_center_x", math.NaN(), 0, 1, 1, 0, math.Pi},
		{"nan_center_y", 0, math.NaN(), 1, 1, 0, math.Pi},
		{"nan_radius_x", 0, 0, math.NaN(), 1, 0, math.Pi},
		{"nan_radius_y", 0, 0, 1, math.NaN(), 0, math.Pi},
		{"nan_start_angle", 0, 0, 1, 1, math.NaN(), math.Pi},
		{"nan_delta", 0, 0, 1, 1, 0, math.NaN()},
		{"inf_radius_x", 0, 0, math.Inf(1), 1, 0, math.Pi},
		{"inf_center_y", 0, math.Inf(-1), 1, 1, 0, math.Pi},
		{"inf_start_angle", 0, 0, 1, 1, math.Inf(1), math.Pi},
		{"neg_inf_start_angle", 0, 0, 1, 1, math.Inf(-1), math.Pi},
		{"angle_sum_overflow", 0, 0, 1, 1, math.MaxFloat64, math.MaxFloat64},
		{"angle_sum_underflow", 0, 0, 1, 1, -math.MaxFloat64, -math.MaxFloat64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Path{}
			p.MoveTo(2, 3)
			p.ArcTo(tt.cx, tt.cy, tt.rx, tt.ry, tt.startAngle, tt.delta)

			// neither the arc nor its connecting line may be recorded
			assert.Equal(t, []PathComponent{MoveToComponent}, p.Components)
			x, y := p.LastPoint()
			assert.InDelta(t, 2.0, x, 0.0001)
			assert.InDelta(t, 3.0, y, 0.0001)
		})
	}
}

func TestPathNonFinite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		build func(p *Path)
	}{
		{"lineto_nan_x", func(p *Path) { p.LineTo(math.NaN(), 5) }},
		{"lineto_inf_y", func(p *Path) { p.LineTo(5, math.Inf(1)) }},
		{"quad_nan_control", func(p *Path) { p.QuadCurveTo(math.NaN(), 1, 5, 5) }},
		{"quad_inf_end", func(p *Path) { p.QuadCurveTo(1, 1, math.Inf(-1), 5) }},
		{"cubic_nan_control", func(p *Path) { p.CubicCurveTo(1, 1, 2, math.NaN(), 5, 5) }},
		{"cubic_inf_end", func(p *Path) { p.CubicCurveTo(1, 1, 2, 2, 5, math.Inf(1)) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Path{}
			p.MoveTo(2, 3)
			tt.build(p)

			assert.Equal(t, []PathComponent{MoveToComponent}, p.Components)
			assert.Equal(t, []float64{2, 3}, p.Points)
			x, y := p.LastPoint()
			assert.InDelta(t, 2.0, x, 0)
			assert.InDelta(t, 3.0, y, 0)
		})
	}

	t.Run("empty_path_not_started", func(t *testing.T) {
		p := &Path{} // the implicit MoveTo(0, 0) must not run either
		p.LineTo(math.NaN(), 5)

		assert.True(t, p.IsEmpty())
	})
}

func TestPathMoveToNonFinite(t *testing.T) {
	t.Parallel()

	t.Run("subpath_suppressed", func(t *testing.T) {
		p := &Path{}
		p.MoveTo(0, 0)
		p.LineTo(10, 0)
		p.MoveTo(math.NaN(), 5)
		// the suppressed subpath must not splice onto the previous one
		p.LineTo(20, 20)
		p.QuadCurveTo(1, 1, 2, 2)
		p.CubicCurveTo(1, 1, 2, 2, 3, 3)
		p.ArcTo(0, 0, 1, 1, 0, math.Pi)
		p.Close()

		assert.Equal(t, []PathComponent{MoveToComponent, LineToComponent}, p.Components)
		assert.Equal(t, []float64{0, 0, 10, 0}, p.Points)
	})

	t.Run("valid_moveto_recovers", func(t *testing.T) {
		p := &Path{}
		p.MoveTo(math.Inf(1), 0)
		p.LineTo(5, 5)
		p.MoveTo(1, 1)
		p.LineTo(5, 5)

		assert.Equal(t, []PathComponent{MoveToComponent, LineToComponent}, p.Components)
		assert.Equal(t, []float64{1, 1, 5, 5}, p.Points)
	})

	t.Run("copy_and_clear", func(t *testing.T) {
		p := &Path{}
		p.MoveTo(math.NaN(), 0)

		c := p.Copy()
		c.LineTo(5, 5)
		assert.True(t, c.IsEmpty())

		p.Clear()
		p.LineTo(5, 5)
		assert.Equal(t, []PathComponent{MoveToComponent, LineToComponent}, p.Components)
	})
}

func TestPathArcToAngleNormalization(t *testing.T) {
	t.Parallel()

	// the normalize loops must not spin, endAngle is derived from delta so neither can run
	tests := []struct {
		name              string
		startAngle, delta float64
	}{
		{"huge_positive_delta", 1e300, 1e300},
		{"huge_negative_delta", 1e300, -1e300},
		{"huge_negative_start", -1e300, 1e300},
		{"zero_delta", 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Path{}
			p.ArcTo(0, 0, 1, 1, tt.startAngle, tt.delta)

			x, y := p.LastPoint()
			assert.False(t, isNonFinite(x) || isNonFinite(y))
		})
	}
}
