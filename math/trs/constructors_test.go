package trs_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			AssertMatrixInDelta(t, tc.matrix, matrix, delta)
		})
	}
}

// func TestConstructor_ToFromMatrix_rotationFuzz(t *testing.T) {

// 	inc := 50
// 	incF := float64(inc)
// 	for x := range inc {
// 		for y := range inc {
// 			for z := range inc {
// 				dir := vector3.New(x-25, y-25, z-25).ToFloat64().Normalized()
// 				if dir.ContainsNaN() {
// 					continue
// 				}
// 				for w := range inc {
// 					q := quaternion.FromTheta(float64(w)/incF, dir)
// 					transformation := trs.FromMatrix(trs.Rotation(q).Matrix())
// 					AssertRotationInDelta(t, q, transformation.Rotation(), 0.000000001)
// 				}
// 			}
// 		}
// 	}

// }

func FuzzToFromMatrix_RotationAndScale(f *testing.F) {
	f.Add(1., 1., 1., 1., 2., 3., 4.)

	f.Fuzz(func(t *testing.T, sx float64, sy float64, sz float64, rx float64, ry float64, rz float64, w float64) {
		scale := vector3.New(sx, sy, sz).Abs()
		if scale.X() == 0 {
			scale = scale.SetX(1)
		}
		if scale.Y() == 0 {
			scale = scale.SetY(1)
		}
		if scale.Z() == 0 {
			scale = scale.SetZ(1)
		}
		rotation := quaternion.FromTheta(w, vector3.New(rx, ry, rz).Normalized())

		delta := 0.000000001
		back := trs.FromMatrix(trs.New(vector3.Float64{}, rotation, scale).Matrix())

		require.InDelta(t, scale.X(), back.Scale().X(), delta, "Scale-X")
		require.InDelta(t, scale.Y(), back.Scale().Y(), delta, "Scale-Y")
		require.InDelta(t, scale.Z(), back.Scale().Z(), delta, "Scale-Z")

		assert.InDelta(t, 1., math.Abs(rotation.Dot(back.Rotation())), delta, "Rotation")
	})
}

// func FuzzToFromMatrix_Scale(f *testing.F) {
// 	f.Add(1., 1., 1.)

// 	f.Fuzz(func(t *testing.T, sx float64, sy float64, sz float64) {
// 		scale := vector3.New(sx, sy, sz).Abs()

// 		delta := 0.000000000001
// 		back := trs.FromMatrix(trs.Scale(scale).Matrix())

// 		require.InDelta(t, scale.X(), back.Scale().X(), delta, "X")
// 		require.InDelta(t, scale.Y(), back.Scale().Y(), delta, "Y")
// 		require.InDelta(t, scale.Z(), back.Scale().Z(), delta, "Z")

// 	})
// }

func AssertRotationInDelta(t *testing.T, expected, actual quaternion.Quaternion, delta float64) {
	assert.InDelta(t, expected.Dir().X(), actual.Dir().X(), delta, "X")
	assert.InDelta(t, expected.Dir().Y(), actual.Dir().Y(), delta, "Y")
	assert.InDelta(t, expected.Dir().Z(), actual.Dir().Z(), delta, "Z")
	assert.InDelta(t, expected.W(), actual.W(), delta, "W")
}

func AssertMatrixInDelta(t *testing.T, expected, actual mat.Matrix4x4, delta float64) {
	require.InDelta(t, expected.X00, actual.X00, delta, "X00")
	require.InDelta(t, expected.X01, actual.X01, delta, "X01")
	require.InDelta(t, expected.X02, actual.X02, delta, "X02")
	require.InDelta(t, expected.X03, actual.X03, delta, "X03")

	require.InDelta(t, expected.X10, actual.X10, delta, "X10")
	require.InDelta(t, expected.X11, actual.X11, delta, "X11")
	require.InDelta(t, expected.X12, actual.X12, delta, "X12")
	require.InDelta(t, expected.X13, actual.X13, delta, "X13")

	require.InDelta(t, expected.X20, actual.X20, delta, "X20")
	require.InDelta(t, expected.X21, actual.X21, delta, "X21")
	require.InDelta(t, expected.X22, actual.X22, delta, "X22")
	require.InDelta(t, expected.X23, actual.X23, delta, "X23")

	require.InDelta(t, expected.X30, actual.X30, delta, "X30")
	require.InDelta(t, expected.X31, actual.X31, delta, "X31")
	require.InDelta(t, expected.X32, actual.X32, delta, "X32")
	require.InDelta(t, expected.X33, actual.X33, delta, "X33")
}
