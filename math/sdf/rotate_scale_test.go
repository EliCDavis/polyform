package sdf_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func boxAt(half vector3.Float64) func(vector3.Float64) float64 {
	return sdf.Box(vector3.Zero[float64](), half.Scale(2))
}

func TestRotateMovesTheSurface(t *testing.T) {
	field := sdf.Rotate(
		boxAt(vector3.New(1., 0.2, 0.2)),
		quaternion.FromEulerAngle(vector3.New(0., math.Pi/2, 0.)),
	)

	assert.InDelta(t, 0.2, surfaceAlong(field, vector3.New(1., 0., 0.)), 1e-6)
	assert.InDelta(t, 1.0, surfaceAlong(field, vector3.New(0., 0., 1.)), 1e-6)
}

func TestRotatePreservesDistance(t *testing.T) {
	box := boxAt(vector3.New(1., 0.2, 0.2))
	q := quaternion.FromEulerAngle(vector3.New(0.3, 0.7, -0.2))
	rotated := sdf.Rotate(box, q)

	assert.LessOrEqual(t, worstGradient(rotated), 1.0001)

	for _, p := range []vector3.Float64{
		vector3.New(2., 0., 0.), vector3.New(0., 3., 0.),
		vector3.New(1., 1., 1.), vector3.New(0.1, -0.4, 0.9),
	} {
		assert.InDelta(t, box(p), rotated(q.Rotate(p)), 1e-9)
	}
}

func TestRotateThenBackIsIdentity(t *testing.T) {
	box := boxAt(vector3.New(1., 0.4, 0.7))
	q := quaternion.FromEulerAngle(vector3.New(0.4, -1.1, 0.25))
	inverse := quaternion.New(q.Normalize().Dir().Scale(-1), q.Normalize().W())

	round := sdf.Rotate(sdf.Rotate(box, q), inverse)
	for _, p := range []vector3.Float64{
		vector3.New(0.5, 0., 0.), vector3.New(0., 0.9, 0.3), vector3.New(-1.2, 0.4, 2.),
	} {
		assert.InDelta(t, box(p), round(p), 1e-9)
	}
}

func TestScaleUniformIsExact(t *testing.T) {
	field, err := sdf.Scale(sdf.Sphere(vector3.Zero[float64](), 1), vector3.Fill(0.25))
	require.NoError(t, err)

	assert.InDelta(t, -0.25, field(vector3.Zero[float64]()), 1e-9)
	assert.InDelta(t, 0.25, field(vector3.New(0.5, 0., 0.)), 1e-9)
	assert.LessOrEqual(t, worstGradient(field), 1.0001)
}

func TestScaleNonUniformStaysMarchable(t *testing.T) {
	for _, scale := range []vector3.Float64{
		vector3.New(1., 0.25, 1.),
		vector3.New(0.2, 1., 0.35),
		vector3.New(3., 1., 1.),
	} {
		field, err := sdf.Scale(sdf.Sphere(vector3.Zero[float64](), 1), scale)
		require.NoError(t, err)

		assert.LessOrEqual(t, worstGradient(field), 1.0001, "scale %v", scale)
		assert.InDelta(t, scale.X(), surfaceAlong(field, vector3.New(1., 0., 0.)), 1e-6)
		assert.InDelta(t, scale.Y(), surfaceAlong(field, vector3.New(0., 1., 0.)), 1e-6)
		assert.InDelta(t, scale.Z(), surfaceAlong(field, vector3.New(0., 0., 1.)), 1e-6)
	}
}

func TestScaleZeroComponentIsAnError(t *testing.T) {
	field, err := sdf.Scale(sdf.Sphere(vector3.Zero[float64](), 1), vector3.New(1., 0., 1.))

	require.Error(t, err)
	assert.Nil(t, field)
	assert.Contains(t, err.Error(), "collapses an axis")
}

func TestScaleThenRotateKeepsTheSquash(t *testing.T) {
	const angle = math.Pi / 4
	squashed, err := sdf.Scale(sdf.Sphere(vector3.Zero[float64](), 1), vector3.New(0.25, 1., 1.))
	require.NoError(t, err)
	field := sdf.Rotate(squashed, quaternion.FromEulerAngle(vector3.New(0., angle, 0.)))

	thin := vector3.New(math.Cos(angle), 0., -math.Sin(angle))
	fat := vector3.New(math.Sin(angle), 0., math.Cos(angle))

	assert.InDelta(t, 0.25, surfaceAlong(field, thin), 1e-6)
	assert.InDelta(t, 1.0, surfaceAlong(field, fat), 1e-6)
	assert.LessOrEqual(t, worstGradient(field), 1.0001)
}
