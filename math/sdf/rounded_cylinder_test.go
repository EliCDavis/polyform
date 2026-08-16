package sdf_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestRoundedCylinder(t *testing.T) {
	cylinder := sdf.RoundedCylinder(vector3.Zero[float64](), 1, 0.25, 1)

	tests := map[string]struct {
		pos  vector3.Float64
		want float64
	}{
		"center":                    {pos: vector3.Zero[float64](), want: -1.25},
		"north pole, on the fillet": {pos: vector3.New(0., 1.25, 0.), want: 0.},
		"far outside":               {pos: vector3.New(0., 5., 0.), want: 3.75},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.InDelta(t, tc.want, cylinder(tc.pos), 1e-9)
		})
	}
}
