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

	t.Run("non_finite_point", func(t *testing.T) {
		rec := &recordFlattenerEnd{}
		ls := NewLineStroker(rec)
		ls.MoveTo(0, 0)
		ls.LineTo(math.NaN(), 0)
		ls.End()

		assert.Equal(t, []string{"E"}, rec.moves)
	})
}
