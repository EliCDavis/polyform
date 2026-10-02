package sdf_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestGrowMovesTheSurfaceOutwardAndInward(t *testing.T) {
	sphere := sdf.Sphere(vector3.Zero[float64](), 1)
	p := vector3.New(1.25, 0., 0.)
	assert.InDelta(t, 0, sdf.Grow(sphere, 0.25)(p), 1e-12)
	assert.InDelta(t, 0.5, sdf.Grow(sphere, -0.25)(p), 1e-12)
}

func TestGrowByZeroHandsBackTheSameField(t *testing.T) {
	sphere := sdf.Sphere(vector3.Zero[float64](), 1)
	grown := sdf.Grow(sphere, 0)

	for _, p := range []vector3.Float64{
		vector3.Zero[float64](),
		vector3.New(0.5, -0.25, 1.75),
		vector3.New(-3., 2., 0.),
	} {
		assert.Equal(t, sphere(p), grown(p), "at %v", p)
	}
}

func TestShellIsHollowAndCenteredOnTheSurface(t *testing.T) {
	shell := sdf.Shell(sdf.Sphere(vector3.Zero[float64](), 1), 0.2)
	assert.InDelta(t, -0.1, shell(vector3.New(1., 0., 0.)), 1e-12)
	assert.InDelta(t, 0, shell(vector3.New(0.9, 0., 0.)), 1e-12)
	assert.InDelta(t, 0, shell(vector3.New(1.1, 0., 0.)), 1e-12)
	assert.Greater(t, shell(vector3.Zero[float64]()), 0.0)
}
