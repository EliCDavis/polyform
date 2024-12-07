package trs_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestTransformArray(t *testing.T) {

	// ARRANGE ================================================================
	transform := trs.New(
		vector3.New(1., 2., 3.),
		quaternion.FromTheta(math.Pi, vector3.Forward[float64]()),
		vector3.New(4., 5., 6.),
	)
	points := []vector3.Float64{
		vector3.Zero[float64](),
		vector3.One[float64](),
	}

	// ACT ====================================================================
	out := transform.TransformArray(points)

	// ASSERT =================================================================
	assert.Len(t, out, len(points))
	assert.Equal(t, vector3.New(1., 2., 3.), out[0])
	assert.InDelta(t, -3., out[1].X(), 0.0000001)
	assert.InDelta(t, -3., out[1].Y(), 0.0000001)
	assert.InDelta(t, 9., out[1].Z(), 0.0000001)

}

func TestTransformInPlace(t *testing.T) {

	// ARRANGE ================================================================
	transform := trs.New(
		vector3.New(1., 2., 3.),
		quaternion.FromTheta(math.Pi, vector3.Forward[float64]()),
		vector3.New(4., 5., 6.),
	)
	points := []vector3.Float64{
		vector3.Zero[float64](),
		vector3.One[float64](),
	}

	// ACT ====================================================================
	transform.TransformInPlace(points)

	// ASSERT =================================================================
	assert.Equal(t, vector3.New(1., 2., 3.), points[0])
	assert.InDelta(t, -3., points[1].X(), 0.0000001)
	assert.InDelta(t, -3., points[1].Y(), 0.0000001)
	assert.InDelta(t, 9., points[1].Z(), 0.0000001)

}
