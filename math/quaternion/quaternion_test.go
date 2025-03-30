package quaternion_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestToEulerAngles(t *testing.T) {
	tests := map[string]struct {
		input quaternion.Quaternion
		want  vector3.Float64
	}{
		"no rotation": {
			input: quaternion.FromTheta(0, vector3.New(0., 1., 0.)),
		},
		"rotate around y": {
			input: quaternion.FromTheta(math.Pi, vector3.New(0., 1., 0.)),
			want:  vector3.New(float64(math.Pi), 0, float64(math.Pi)),
		},
		"rotate around x": {
			input: quaternion.FromTheta(math.Pi, vector3.New(1., 0., 0.)),
			want:  vector3.New(float64(math.Pi), 0, 0),
		},
		"rotate around z": {
			input: quaternion.FromTheta(math.Pi, vector3.New(0., 0., 1.)),
			want:  vector3.New(0, 0, float64(math.Pi)),
		},
	}

	epsilon := 0.000000000000001
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			eulerAngles := tc.input.ToEulerAngles()
			assert.InDelta(t, tc.want.X(), eulerAngles.X(), epsilon)
			assert.InDelta(t, tc.want.Y(), eulerAngles.Y(), epsilon)
			assert.InDelta(t, tc.want.Z(), eulerAngles.Z(), epsilon)

			back := quaternion.FromEulerAngle(eulerAngles)
			assert.InDelta(t, tc.input.Dir().X(), back.Dir().X(), epsilon)
			assert.InDelta(t, tc.input.Dir().Y(), back.Dir().Y(), epsilon)
			assert.InDelta(t, tc.input.Dir().Z(), back.Dir().Z(), epsilon)
			assert.InDelta(t, tc.input.W(), back.W(), epsilon)
		})
	}
}
