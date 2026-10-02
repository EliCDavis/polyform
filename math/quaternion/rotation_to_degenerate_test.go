package quaternion_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRotationToHandlesAnExactReversal(t *testing.T) {
	axes := map[string]vector3.Float64{
		"X": vector3.New(1., 0., 0.),
		"Y": vector3.New(0., 1., 0.),
		"Z": vector3.New(0., 0., 1.),
	}

	for name, from := range axes {
		for _, sign := range []float64{1, -1} {
			from := from.Scale(sign)
			t.Run(name, func(t *testing.T) {
				q := quaternion.RotationTo(from, from.Scale(-1))

				d := q.Dir()
				require.False(t,
					math.IsNaN(d.X()) || math.IsNaN(d.Y()) || math.IsNaN(d.Z()) || math.IsNaN(q.W()),
					"a reversal has to produce a real quaternion")

				got := q.Rotate(from)
				want := from.Scale(-1)
				assert.InDelta(t, want.X(), got.X(), 1e-9)
				assert.InDelta(t, want.Y(), got.Y(), 1e-9)
				assert.InDelta(t, want.Z(), got.Z(), 1e-9)
			})
		}
	}
}
