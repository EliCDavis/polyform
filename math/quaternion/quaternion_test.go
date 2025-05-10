package quaternion_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestToFromEulerAngles(t *testing.T) {
	tests := map[string]struct {
		rotation quaternion.Quaternion
		euler    vector3.Float64
	}{
		"no rotation": {
			rotation: quaternion.Identity(),
			euler:    vector3.Zero[float64](),
		},
		"rotate around y": {
			rotation: quaternion.FromTheta(math.Pi, vector3.New(0., 1., 0.)),
			euler:    vector3.New(float64(math.Pi), 0, float64(math.Pi)),
		},
		"rotate around x": {
			rotation: quaternion.FromTheta(math.Pi, vector3.New(1., 0., 0.)),
			euler:    vector3.New(float64(math.Pi), 0, 0),
		},
		"rotate around z": {
			rotation: quaternion.FromTheta(math.Pi, vector3.New(0., 0., 1.)),
			euler:    vector3.New(0, 0, float64(math.Pi)),
		},
	}

	epsilon := 0.000000000000001
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			eulerAngles := tc.rotation.ToEulerAngles()
			assert.InDelta(t, tc.euler.X(), eulerAngles.X(), epsilon)
			assert.InDelta(t, tc.euler.Y(), eulerAngles.Y(), epsilon)
			assert.InDelta(t, tc.euler.Z(), eulerAngles.Z(), epsilon)

			back := quaternion.FromEulerAngle(eulerAngles)
			assert.InDelta(t, tc.rotation.Dir().X(), back.Dir().X(), epsilon)
			assert.InDelta(t, tc.rotation.Dir().Y(), back.Dir().Y(), epsilon)
			assert.InDelta(t, tc.rotation.Dir().Z(), back.Dir().Z(), epsilon)
			assert.InDelta(t, tc.rotation.W(), back.W(), epsilon)
		})
	}
}

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
			rot := quaternion.FromTheta(tc.theta, tc.dir)
			rotated := rot.Rotate(tc.v)
			assert.InDelta(t, tc.want.X(), rotated.X(), 0.00000001)
			assert.InDelta(t, tc.want.Y(), rotated.Y(), 0.00000001)
			assert.InDelta(t, tc.want.Z(), rotated.Z(), 0.00000001)
		})
	}
}

func TestConstructor_FromMatrix(t *testing.T) {

	tests := map[string]struct {
		matrix mat.Matrix4x4
		want   quaternion.Quaternion
	}{
		"identity": {
			matrix: mat.Matrix4x4{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				0, 0, 0, 1,
			},
			want: quaternion.Identity(),
		},

		"x": {
			matrix: mat.Matrix4x4{
				1, 0, 0, 0,
				0, 0, -1, 0,
				0, 1, 0, 0,
				0, 0, 0, 1,
			},
			want: quaternion.FromEulerAngle(
				vector3.New(math.Pi/2, 0., 0.),
			),
		},

		"y": {
			matrix: mat.Matrix4x4{
				0, 0, 1, 0,
				0, 1, 0, 0,
				-1, 0, 0, 0,
				0, 0, 0, 1,
			},
			want: (quaternion.FromEulerAngle(
				vector3.New(0, math.Pi/2, 0.),
			)),
		},

		"z": {
			matrix: mat.Matrix4x4{
				0, -1, 0, 0,
				1, 0, 0, 0,
				0, 0, 1, 0,
				0, 0, 0, 1,
			},
			want: quaternion.FromEulerAngle(
				vector3.New(0, 0, math.Pi/2),
			),
		},

		// TODO: I think something is wrong FromEulerAngle
		// "rotation-yz": {
		// 	matrix: mat.Matrix4x4{
		// 		0, 0, 1, 0,
		// 		1, 0, 0, 0,
		// 		0, 1, 0, 0,
		// 		0, 0, 0, 1,
		// 	},
		// 	want: trs.Rotation(quaternion.FromEulerAngle(
		// 		vector3.New(0, math.Pi/2, math.Pi/2),
		// 	)),
		// },

	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, quaternion.FromMatrix(tc.matrix))
		})
	}
}
