package animation_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/polyform/modeling/animation"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestNewSkeleton_SingleJoint(t *testing.T) {
	// ACT ====================================================================
	skeleton := animation.NewSkeleton(animation.NewJoint(
		"Head",
		vector3.Zero[float64](),
		vector3.Up[float64](),
		vector3.Forward[float64](),
	))

	// ASSERT =================================================================
	assert.Len(t, skeleton.Children(0), 0)
	assert.Equal(t, 0, skeleton.Lookup("Head"))
	assert.Equal(t, 1, skeleton.JointCount())
	assert.Equal(t, mat.Identity(), skeleton.RelativeMatrix(0))
}

func TestNewSkeleton_2Levels(t *testing.T) {
	// ACT ====================================================================
	skeleton := animation.NewSkeleton(animation.NewJoint(
		"Head",
		vector3.Zero[float64](),
		vector3.Up[float64](),
		vector3.Forward[float64](),
		animation.NewJoint(
			"Left Hand",
			vector3.Left[float64](),
			vector3.Up[float64](),
			vector3.Forward[float64](),
		),
		animation.NewJoint(
			"Right Hand",
			vector3.Right[float64](),
			vector3.Up[float64](),
			vector3.Forward[float64](),
		),
	))

	// ASSERT =================================================================
	assert.Equal(t, 3, skeleton.JointCount())

	assert.Len(t, skeleton.Children(0), 2)
	assert.Len(t, skeleton.Children(1), 0)
	assert.Len(t, skeleton.Children(2), 0)

	assert.Equal(t, 0, skeleton.Lookup("Head"))
	assert.Equal(t, 1, skeleton.Lookup("Head/Left Hand"))
	assert.Equal(t, 2, skeleton.Lookup("Head/Right Hand"))

	assert.Equal(t, mat.Identity(), skeleton.RelativeMatrix(0))
	assert.Equal(
		t,
		mat.Matrix4x4{
			X00: 1, X01: 0, X02: 0, X03: -1,
			X10: 0, X11: 1, X12: 0, X13: 0,
			X20: 0, X21: 0, X22: 1, X23: 0,
			X30: 0, X31: 0, X32: 0, X33: 1,
		},
		skeleton.RelativeMatrix(1),
	)
	assert.Equal(
		t,
		mat.Matrix4x4{
			X00: 1, X01: 0, X02: 0, X03: 1,
			X10: 0, X11: 1, X12: 0, X13: 0,
			X20: 0, X21: 0, X22: 1, X23: 0,
			X30: 0, X31: 0, X32: 0, X33: 1,
		},
		skeleton.RelativeMatrix(2),
	)
}
