package drawing

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidDash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		dash     []float64
		expected bool
	}{
		{name: "normal", dash: []float64{5, 5}, expected: true},
		{name: "zero_gap", dash: []float64{5, 0}, expected: true},
		{name: "zero_dash", dash: []float64{0, 5}, expected: true},
		{name: "single_entry", dash: []float64{5}, expected: true},
		{name: "all_zero", dash: []float64{0, 0}, expected: false},
		{name: "leading_negative", dash: []float64{-5, 5}, expected: false},
		{name: "trailing_negative", dash: []float64{5, -5}, expected: false},
		{name: "nan_entry", dash: []float64{5, math.NaN()}, expected: false},
		{name: "inf_entry", dash: []float64{5, math.Inf(1)}, expected: false},
		{name: "empty", dash: []float64{}, expected: false},
		{name: "nil", dash: nil, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ValidDash(tt.dash))
		})
	}
}

func TestDashable(t *testing.T) {
	t.Parallel()

	assert.True(t, dashable([]float64{5, 5}, 2))
	assert.False(t, dashable([]float64{0, 0}, 2))
	assert.False(t, dashable([]float64{5, 5}, math.Inf(1)))
	assert.False(t, dashable([]float64{5, 5}, math.NaN()))
}

type recordFlattenerEnd struct {
	moves []string
}

func (r *recordFlattenerEnd) MoveTo(x, y float64) {
	r.moves = append(r.moves, fmt.Sprintf("M%.1f,%.1f", x, y))
}

func (r *recordFlattenerEnd) LineTo(x, y float64) {
	r.moves = append(r.moves, fmt.Sprintf("L%.1f,%.1f", x, y))
}

func (r *recordFlattenerEnd) End() {
	r.moves = append(r.moves, "E")
}

func TestDashVertexConverterLineTo(t *testing.T) {
	t.Parallel()

	t.Run("single_segment", func(t *testing.T) {
		rec := &recordFlattenerEnd{}
		d := NewDashVertexConverter([]float64{2, 2}, 0, rec)
		d.MoveTo(0, 0)
		d.LineTo(5, 0)
		d.End()

		expect := []string{"M0.0,0.0", "L2.0,0.0", "E", "M4.0,0.0", "L5.0,0.0", "E"}
		assert.Equal(t, expect, rec.moves)
	})

	t.Run("short_segments_accumulate", func(t *testing.T) {
		rec := &recordFlattenerEnd{}
		d := NewDashVertexConverter([]float64{10, 10}, 0, rec)
		d.MoveTo(0, 0)
		for x := 3.0; x <= 60; x += 3 {
			d.LineTo(x, 0)
		}
		d.End()

		// phase advances across short segments: dash ends at x=10, gaps, resumes at x=20
		assert.Contains(t, rec.moves, "L10.0,0.0")
		assert.Contains(t, rec.moves, "M20.0,0.0")
		assert.Contains(t, rec.moves, "E")
	})

	t.Run("zero_gap_exact_landing", func(t *testing.T) {
		rec := &recordFlattenerEnd{}
		d := NewDashVertexConverter([]float64{5, 0}, 0, rec)
		d.MoveTo(0, 0)
		d.LineTo(5, 0) // lands exactly on the zero length gap
		d.LineTo(10, 0)
		d.End()

		expect := []string{"M0.0,0.0", "L5.0,0.0", "E", "M5.0,0.0", "L10.0,0.0", "E", "M10.0,0.0", "E"}
		assert.Equal(t, expect, rec.moves)
	})

	t.Run("zero_length_dash_entry", func(t *testing.T) {
		rec := &recordFlattenerEnd{}
		d := NewDashVertexConverter([]float64{0, 5}, 0, rec)
		d.MoveTo(0, 0)
		d.LineTo(0, 0) // zero length segment on a zero length dash entry
		d.LineTo(10, 0)
		d.End()

		expect := []string{"M0.0,0.0", "L0.0,0.0", "E", "M5.0,0.0", "L5.0,0.0", "E", "M10.0,0.0", "L10.0,0.0", "E"}
		assert.Equal(t, expect, rec.moves)
	})

	t.Run("repeated_point_keeps_phase", func(t *testing.T) {
		rec := &recordFlattenerEnd{}
		d := NewDashVertexConverter([]float64{4, 4}, 0, rec)
		d.MoveTo(0, 0)
		d.LineTo(2, 0)
		d.LineTo(2, 0) // duplicate point only repeats the vertex, the dash phase is unchanged
		d.LineTo(8, 0)
		d.End()

		expect := []string{"M0.0,0.0", "L2.0,0.0", "L2.0,0.0", "L4.0,0.0", "E", "M8.0,0.0", "L8.0,0.0", "E"}
		assert.Equal(t, expect, rec.moves)
	})

	t.Run("no_nan_vertices", func(t *testing.T) {
		for _, dash := range [][]float64{{5, 0}, {0, 5}, {5}, {0, 5, 0}, {5, 0, 5}} {
			rec := &recordFlattenerEnd{}
			d := NewDashVertexConverter(dash, 0, rec)
			d.MoveTo(0, 0)
			d.LineTo(5, 0)
			d.LineTo(5, 0)
			d.LineTo(10, 5)
			d.End()

			for _, m := range rec.moves {
				assert.NotContains(t, m, "NaN", dash)
			}
		}
	})
}

func TestDashVertexConverterPattern(t *testing.T) {
	t.Parallel()

	run := func(dash []float64, dashOffset float64) []string {
		rec := &recordFlattenerEnd{}
		d := NewDashVertexConverter(dash, dashOffset, rec)
		d.MoveTo(0, 0)
		d.LineTo(20, 0)
		d.End()
		return rec.moves
	}

	t.Run("odd_single_entry", func(t *testing.T) {
		expect := []string{"M0.0,0.0", "L5.0,0.0", "E", "M10.0,0.0", "L15.0,0.0", "E", "M20.0,0.0", "L20.0,0.0", "E"}
		assert.Equal(t, expect, run([]float64{5}, 0))
		assert.Equal(t, run([]float64{5, 5}, 0), run([]float64{5}, 0))
	})

	t.Run("odd_three_entries", func(t *testing.T) {
		// {4,2,1} alternates as on 4, off 2, on 1, off 4, on 2, off 1
		expect := []string{"M0.0,0.0", "L4.0,0.0", "E", "M6.0,0.0", "L7.0,0.0", "E", "M11.0,0.0",
			"L13.0,0.0", "E", "M14.0,0.0", "L18.0,0.0", "E", "M20.0,0.0", "L20.0,0.0", "E"}
		assert.Equal(t, expect, run([]float64{4, 2, 1}, 0))
		assert.Equal(t, run([]float64{4, 2, 1, 4, 2, 1}, 0), run([]float64{4, 2, 1}, 0))
	})

	t.Run("even_pattern_unchanged", func(t *testing.T) {
		expect := []string{"M0.0,0.0", "L2.0,0.0", "E", "M4.0,0.0", "L6.0,0.0", "E", "M8.0,0.0",
			"L10.0,0.0", "E", "M12.0,0.0", "L14.0,0.0", "E", "M16.0,0.0", "L18.0,0.0", "E", "M20.0,0.0", "L20.0,0.0", "E"}
		assert.Equal(t, expect, run([]float64{2, 2}, 0))
	})

	t.Run("offset_wraps_cycle", func(t *testing.T) {
		assert.Equal(t, run([]float64{2, 2}, 2), run([]float64{2, 2}, 1e12+2))
	})

	t.Run("offset_wraps_doubled_cycle", func(t *testing.T) {
		assert.Equal(t, run([]float64{5}, 3), run([]float64{5}, 13))
	})

	t.Run("negative_offset", func(t *testing.T) {
		assert.Equal(t, run([]float64{2, 2}, 3), run([]float64{2, 2}, -1))
	})
}
