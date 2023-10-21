package modeling_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestQuaternion_Rotate(t *testing.T) {

	// ARRANGE
	tests := map[string]struct {
		theta float64
		dir   vector3.Float64
		v     vector3.Float64
		want  vector3.Float64
	}{
		"simple": {
			theta: math.Pi / 2,
			dir:   vector3.Up[float64](),
			v:     vector3.New(0., 0., 1.),
			want:  vector3.New(1., 0., 0.),
		},
	}

	// ACT / ASSERT
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rot := modeling.UnitQuaternionFromTheta(tc.theta, tc.dir)
			rotated := rot.Rotate(tc.v)
			assert.InDelta(t, tc.want.X(), rotated.X(), 0.00000001)
			assert.InDelta(t, tc.want.Y(), rotated.Y(), 0.00000001)
			assert.InDelta(t, tc.want.Z(), rotated.Z(), 0.00000001)
		})
	}
}
