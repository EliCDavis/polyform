package sdf_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestTorus(t *testing.T) {
	torus := sdf.Torus(vector3.Zero[float64](), 1, 0.25)

	tests := map[string]struct {
		pos  vector3.Float64
		want float64
	}{
		"center of tube cross-section": {pos: vector3.New(1., 0., 0.), want: -0.25},
		"on the tube's outer surface":  {pos: vector3.New(1.25, 0., 0.), want: 0.},
		"in the donut hole":            {pos: vector3.Zero[float64](), want: 0.75},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.InDelta(t, tc.want, torus(tc.pos), 1e-9)
		})
	}
}
