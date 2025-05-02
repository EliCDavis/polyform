package trs_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestConstructor_Position(t *testing.T) {

	// ARRANGE ================================================================
	transform := trs.Position(vector3.New(0., 1., 2.))

	// ACT ====================================================================
	position := transform.Position()
	rotation := transform.Rotation()
	scale := transform.Scale()

	// ASSERT =================================================================
	assert.Equal(t, vector3.New(0., 1., 2.), position)
	assert.Equal(t, quaternion.Identity(), rotation)
	assert.Equal(t, vector3.New(1., 1., 1.), scale)

}

func TestConstructor_Rotation(t *testing.T) {

	// ARRANGE ================================================================
	rot := quaternion.FromTheta(math.Pi, vector3.Up[float64]())
	transform := trs.Rotation(rot)

	// ACT ====================================================================
	position := transform.Position()
	rotation := transform.Rotation()
	scale := transform.Scale()

	// ASSERT =================================================================
	assert.Equal(t, vector3.New(0., 0., 0.), position)
	assert.Equal(t, rot, rotation)
	assert.Equal(t, vector3.New(1., 1., 1.), scale)

}

func TestConstructor_Scale(t *testing.T) {

	// ARRANGE ================================================================
	transform := trs.Scale(vector3.New(0., 1., 2.))

	// ACT ====================================================================
	position := transform.Position()
	rotation := transform.Rotation()
	scale := transform.Scale()

	// ASSERT =================================================================
	assert.Equal(t, vector3.New(0., 0., 0.), position)
	assert.Equal(t, quaternion.Identity(), rotation)
	assert.Equal(t, vector3.New(0., 1., 2.), scale)

}

func TestConstructor_New(t *testing.T) {

	// ARRANGE ================================================================
	rot := quaternion.FromTheta(math.Pi, vector3.Up[float64]())
	transform := trs.New(vector3.New(1., 2., 3.), rot, vector3.New(4., 5., 6.))

	// ACT ====================================================================
	position := transform.Position()
	rotation := transform.Rotation()
	scale := transform.Scale()

	// ASSERT =================================================================
	assert.Equal(t, vector3.New(1., 2., 3.), position)
	assert.Equal(t, rot, rotation)
	assert.Equal(t, vector3.New(4., 5., 6.), scale)

}

func TestConstructor_FromMatrix(t *testing.T) {

	tests := map[string]struct {
		matrix mat.Matrix4x4
		want   trs.TRS
	}{
		"identity": {
			matrix: mat.Matrix4x4{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				0, 0, 0, 1,
			},
			want: trs.Identity(),
		},

		"position": {
			matrix: mat.Matrix4x4{
				1, 0, 0, 1,
				0, 1, 0, 2,
				0, 0, 1, 3,
				0, 0, 0, 1,
			},
			want: trs.Position(vector3.New(1., 2., 3.)),
		},

		"scale": {
			matrix: mat.Matrix4x4{
				2, 0, 0, 0,
				0, 2, 0, 0,
				0, 0, 2, 0,
				0, 0, 0, 1,
			},
			want: trs.Scale(vector3.New(2., 2., 2.)),
		},

		"rotation-x": {
			matrix: mat.Matrix4x4{
				1, 0, 0, 0,
				0, 0, -1, 0,
				0, 1, 0, 0,
				0, 0, 0, 1,
			},
			want: trs.Rotation(quaternion.FromEulerAngle(
				vector3.New(math.Pi/2, 0., 0.),
			)),
		},

		"rotation-y": {
			matrix: mat.Matrix4x4{
				0, 0, 1, 0,
				0, 1, 0, 0,
				-1, 0, 0, 0,
				0, 0, 0, 1,
			},
			want: trs.Rotation(quaternion.FromEulerAngle(
				vector3.New(0, math.Pi/2, 0.),
			)),
		},

		"rotation-z": {
			matrix: mat.Matrix4x4{
				0, -1, 0, 0,
				1, 0, 0, 0,
				0, 0, 1, 0,
				0, 0, 0, 1,
			},
			want: trs.Rotation(quaternion.FromEulerAngle(
				vector3.New(0, 0, math.Pi/2),
			)),
		},

		"scale and position": {
			matrix: mat.Matrix4x4{
				2, 0, 0, 3,
				0, 2, 0, 4,
				0, 0, 2, 5,
				0, 0, 0, 1,
			},
			want: trs.New(
				vector3.New(3., 4., 5.),
				quaternion.Identity(),
				vector3.New(2., 2., 2.),
			),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			transform := trs.FromMatrix(tc.matrix)
			assert.Equal(t, tc.want.Position(), transform.Position())
			assert.Equal(t, tc.want.Rotation(), transform.Rotation())
			assert.Equal(t, tc.want.Scale(), transform.Scale())
		})
	}
}

func TestConstructor_ToFromMatrix(t *testing.T) {

	tests := map[string]struct {
		matrix mat.Matrix4x4
	}{
		"identity": {
			matrix: mat.Matrix4x4{
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 1, 0,
				0, 0, 0, 1,
			},
		},

		"position": {
			matrix: mat.Matrix4x4{
				1, 0, 0, 1,
				0, 1, 0, 2,
				0, 0, 1, 3,
				0, 0, 0, 1,
			},
		},
		"scale": {
			matrix: mat.Matrix4x4{
				2, 0, 0, 0,
				0, 2, 0, 0,
				0, 0, 2, 0,
				0, 0, 0, 1,
			},
		},
		"scale and position": {
			matrix: mat.Matrix4x4{
				2, 0, 0, 3,
				0, 2, 0, 4,
				0, 0, 2, 5,
				0, 0, 0, 1,
			},
		},

		"x-rotation": {
			matrix: mat.Matrix4x4{
				1, 0, 0, 0,
				0, 0, -1, 0,
				0, 1, 0, 0,
				0, 0, 0, 1,
			},
		},

		"y-rotation": {
			matrix: mat.Matrix4x4{
				0, 0, 1, 0,
				0, 1, 0, 0,
				-1, 0, 0, 0,
				0, 0, 0, 1,
			},
		},

		"z-rotation": {
			matrix: mat.Matrix4x4{
				0, -1, 0, 0,
				1, 0, 0, 0,
				0, 0, 1, 0,
				0, 0, 0, 1,
			},
		},

		"yz-rotation": {
			matrix: mat.Matrix4x4{
				0, 0, 1, 0,
				1, 0, 0, 0,
				0, 1, 0, 0,
				0, 0, 0, 1,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			delta := 0.000000000001
			matrix := trs.FromMatrix(tc.matrix).Matrix()
			assert.InDelta(t, tc.matrix.X00, matrix.X00, delta, "X00")
			assert.InDelta(t, tc.matrix.X01, matrix.X01, delta, "X01")
			assert.InDelta(t, tc.matrix.X02, matrix.X02, delta, "X02")
			assert.InDelta(t, tc.matrix.X03, matrix.X03, delta, "X03")

			assert.InDelta(t, tc.matrix.X10, matrix.X10, delta, "X10")
			assert.InDelta(t, tc.matrix.X11, matrix.X11, delta, "X11")
			assert.InDelta(t, tc.matrix.X12, matrix.X12, delta, "X12")
			assert.InDelta(t, tc.matrix.X13, matrix.X13, delta, "X13")

			assert.InDelta(t, tc.matrix.X20, matrix.X20, delta, "X20")
			assert.InDelta(t, tc.matrix.X21, matrix.X21, delta, "X21")
			assert.InDelta(t, tc.matrix.X22, matrix.X22, delta, "X22")
			assert.InDelta(t, tc.matrix.X23, matrix.X23, delta, "X23")

			assert.InDelta(t, tc.matrix.X30, matrix.X30, delta, "X30")
			assert.InDelta(t, tc.matrix.X31, matrix.X31, delta, "X31")
			assert.InDelta(t, tc.matrix.X32, matrix.X32, delta, "X32")
			assert.InDelta(t, tc.matrix.X33, matrix.X33, delta, "X33")
		})
	}

}
