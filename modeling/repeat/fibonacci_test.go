package repeat_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func finite(v vector3.Float64) bool {
	return !math.IsNaN(v.X()) && !math.IsNaN(v.Y()) && !math.IsNaN(v.Z()) &&
		!math.IsInf(v.X(), 0) && !math.IsInf(v.Y(), 0) && !math.IsInf(v.Z(), 0)
}

func TestFibonacciSpherePointsCountOneIsThePole(t *testing.T) {
	points := repeat.FibonacciSpherePoints(1, 2)

	require.Len(t, points, 1)
	assert.InDelta(t, 0, points[0].Sub(vector3.New(0., 2., 0.)).Length(), 1e-9)
}

func TestFibonacciSpherePointsCountTwoAreThePoles(t *testing.T) {
	points := repeat.FibonacciSpherePoints(2, 1)

	require.Len(t, points, 2)
	assert.InDelta(t, 1, points[0].Y(), 1e-9)
	assert.InDelta(t, -1, points[1].Y(), 1e-9)
	for _, p := range points {
		assert.True(t, finite(p), "%v", p)
	}
}

func TestFibonacciSpiralCountOneIsTheCentre(t *testing.T) {
	points := repeat.FibonacciSpiral(1, 3)

	require.Len(t, points, 1)
	assert.InDelta(t, 0, points[0].Position().Length(), 1e-9)
	assert.True(t, finite(points[0].Rotation().Dir()))
}

func TestFibonacciSpiralCountTwoReachesTheRadius(t *testing.T) {
	points := repeat.FibonacciSpiral(2, 3)

	require.Len(t, points, 2)
	assert.InDelta(t, 0, points[0].Position().Length(), 1e-9)
	assert.InDelta(t, 3, points[1].Position().Length(), 1e-9)
}

func TestFibonacciSphereCountOneHasAFiniteRotation(t *testing.T) {
	transforms := repeat.FibonacciSphere(1, 1)

	require.Len(t, transforms, 1)
	assert.True(t, finite(transforms[0].Position()))
	assert.True(t, finite(transforms[0].Rotation().Dir()))
	assert.False(t, math.IsNaN(transforms[0].Rotation().W()))
}
