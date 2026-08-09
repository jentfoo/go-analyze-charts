package drawing

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineStrokerLine(t *testing.T) {
	t.Parallel()

	t.Run("horizontal", func(t *testing.T) {
		rec := &recordFlattenerEnd{}
		ls := NewLineStroker(rec)
		ls.MoveTo(0, 0)
		ls.LineTo(2, 0)
		ls.End()

		expect := []string{"M0.0,-0.5", "L2.0,-0.5", "L2.0,0.5", "L0.0,0.5", "L0.0,-0.5", "E"}
		assert.Equal(t, expect, rec.moves)
	})

	t.Run("closed_square_seam", func(t *testing.T) {
		rec := &recordFlattenerEnd{}
		ls := NewLineStroker(rec)
		ls.MoveTo(0, 0)
		ls.LineTo(2, 0)
		ls.LineTo(2, 2)
		ls.LineTo(0, 2)
		ls.LineTo(0, 0)
		ls.End()

		// seam pair "L0.0,-0.5" / trailing "L0.0,0.5" joins the close like interior vertices
		expect := []string{
			"M0.0,-0.5", "L2.0,-0.5", "L2.5,0.0", "L2.5,2.0", "L2.0,2.5", "L0.0,2.5",
			"L-0.5,2.0", "L-0.5,0.0", "L0.0,-0.5",
			"L0.0,0.5", "L0.5,0.0", "L0.5,2.0", "L0.0,1.5", "L2.0,1.5", "L1.5,2.0",
			"L1.5,0.0", "L2.0,0.5", "L0.0,0.5",
			"L0.0,-0.5", "E",
		}
		assert.Equal(t, expect, rec.moves)
	})

	t.Run("open_path_no_seam", func(t *testing.T) {
		rec := &recordFlattenerEnd{}
		ls := NewLineStroker(rec)
		ls.MoveTo(0, 0)
		ls.LineTo(2, 0)
		ls.LineTo(2, 2)
		ls.End()

		expect := []string{
			"M0.0,-0.5", "L2.0,-0.5", "L2.5,0.0", "L2.5,2.0",
			"L1.5,2.0", "L1.5,0.0", "L2.0,0.5", "L0.0,0.5",
			"L0.0,-0.5", "E",
		}
		assert.Equal(t, expect, rec.moves)
	})

	t.Run("non_finite_point", func(t *testing.T) {
		// an infinite length passes a bare d > 0 guard and normalizes to NaN offsets
		for _, x := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
			rec := &recordFlattenerEnd{}
			ls := NewLineStroker(rec)
			ls.MoveTo(0, 0)
			ls.LineTo(x, 0)
			ls.End()

			assert.Equal(t, []string{"E"}, rec.moves)
		}
	})
}
