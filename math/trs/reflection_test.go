package trs_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromMatrixKeepsAReflection(t *testing.T) {
	for name, scale := range map[string]vector3.Float64{
		"mirror x":         vector3.New(-1., 1., 1.),
		"mirror y":         vector3.New(1., -1., 1.),
		"mirror z":         vector3.New(1., 1., -1.),
		"mirror x, scaled": vector3.New(-2., 1., 0.5),
	} {
		t.Run(name, func(t *testing.T) {
			mirrored := trs.New(vector3.New(1., 2., 3.), rot(math.Pi/3), scale)

			decomposed, err := trs.FromMatrix(mirrored.Matrix())
			require.NoError(t, err)
			assert.NoError(t, decomposed.WithinDelta(mirrored, 1e-9))
			assert.Less(t, decomposed.Matrix().Determinant(), 0.)
		})
	}
}

func TestFromMatrixKeepsAReflectionComposedWithAnInstance(t *testing.T) {
	mirror := trs.Scale(vector3.New(-1., 1., 1.))
	instance := trs.New(vector3.New(0., 0., 4.), rot(math.Pi/5), vector3.One[float64]())

	composed, err := trs.FromMatrix(mirror.Multiply(instance))
	require.NoError(t, err)
	assert.Less(t, composed.Scale().X(), 0.)
	for _, p := range []vector3.Float64{vector3.New(1., 0., 0.), vector3.New(1., -2., 3.)} {
		want := mirror.Transform(instance.Transform(p))
		assert.InDelta(t, 0, composed.Transform(p).Sub(want).Length(), 1e-9)
	}
}

func TestTransformNormalOfAPureRotationIsTheRotation(t *testing.T) {
	x := trs.New(vector3.New(1., 2., 3.), rot(math.Pi/7), vector3.One[float64]())
	n := vector3.New(1., 2., 3.).Normalized()

	assert.Equal(t, x.Rotation().Rotate(n), x.TransformNormal(n))
}

func TestTransformNormalStaysPerpendicularUnderNonUniformScale(t *testing.T) {
	x := trs.New(vector3.New(1., 2., 3.), rot(math.Pi/5), vector3.New(2., 0.2, 1.))
	tangent := vector3.New(1., 1., 0.)
	normal := vector3.New(1., -1., 0.).Normalized()

	movedTangent := x.Transform(tangent).Sub(x.Transform(vector3.Zero[float64]()))
	movedNormal := x.TransformNormal(normal)

	assert.InDelta(t, 0, movedTangent.Dot(movedNormal), 1e-9)
	assert.InDelta(t, 1, movedNormal.Length(), 1e-9)
}

func TestTransformNormalOfACollapsedAxisFacesAlongIt(t *testing.T) {
	flat := trs.Scale(vector3.New(1., 0., 1.))

	n := flat.TransformNormal(vector3.New(0.6, 0.8, 0.))
	assert.InDelta(t, 0, n.Sub(vector3.Up[float64]()).Length(), 1e-9)
}
