package sdf_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// worstGradient walks the field and returns the steepest rate of change it
// finds. A signed distance field must never exceed 1: at 4, marching cubes
// interpolates the surface to the wrong place between voxel corners and a
// smooth union blends over a distance that isn't the one it was given.
func worstGradient(field func(vector3.Float64) float64) float64 {
	const step = 0.005
	worst := 0.
	for x := -1.5; x <= 1.5; x += 0.1 {
		for y := -1.5; y <= 1.5; y += 0.1 {
			for z := -1.5; z <= 1.5; z += 0.1 {
				at := vector3.New(x, y, z)
				for _, dir := range []vector3.Float64{
					vector3.New(step, 0., 0.),
					vector3.New(0., step, 0.),
					vector3.New(0., 0., step),
				} {
					g := math.Abs(field(at.Add(dir))-field(at)) / step
					if g > worst {
						worst = g
					}
				}
			}
		}
	}
	return worst
}

func scaledSphere(s vector3.Float64) func(vector3.Float64) float64 {
	return sdf.Transform(
		sdf.Sphere(vector3.Zero[float64](), 1),
		trs.New(vector3.Zero[float64](), trs.Identity().Rotation(), s),
	)
}

func TestTransformKeepsFieldValidUnderScale(t *testing.T) {
	for _, tc := range []struct {
		name  string
		scale vector3.Float64
	}{
		{"uniform shrink", vector3.New(0.25, 0.25, 0.25)},
		{"uniform grow", vector3.New(3., 3., 3.)},
		{"squashed on one axis", vector3.New(1., 0.25, 1.)},
		{"squashed on two axes", vector3.New(0.2, 1., 0.35)},
		{"identity", vector3.New(1., 1., 1.)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.LessOrEqual(t, worstGradient(scaledSphere(tc.scale)), 1.0001,
				"a distance field's gradient must not exceed 1")
		})
	}
}

func TestTransformKeepsSurfaceInPlace(t *testing.T) {
	field := scaledSphere(vector3.New(1., 0.25, 1.))

	assert.InDelta(t, 0, field(vector3.New(0., 0.25, 0.)), 1e-9, "top of the squashed sphere")
	assert.InDelta(t, 0, field(vector3.New(1., 0., 0.)), 1e-9, "unscaled axis is untouched")
	assert.Negative(t, field(vector3.Zero[float64]()), "center is inside")
	assert.Positive(t, field(vector3.New(0., 0.5, 0.)), "above the squashed top is outside")
}

func TestTransformReportsTrueDistanceWhenUniform(t *testing.T) {
	field := scaledSphere(vector3.New(0.25, 0.25, 0.25))

	// A radius-1 sphere scaled to 0.25 is a radius-0.25 sphere.
	assert.InDelta(t, -0.25, field(vector3.Zero[float64]()), 1e-9)
	assert.InDelta(t, 0.25, field(vector3.New(0.5, 0., 0.)), 1e-9)
	assert.InDelta(t, 0.75, field(vector3.New(0., 0., 1.)), 1e-9)
}

func TestTransformedFieldIsCleanToMarch(t *testing.T) {
	scale := vector3.New(1., 0.35, 1.)
	field := scaledSphere(scale)

	inside := func(p vector3.Float64) bool {
		q := vector3.New(p.X()/scale.X(), p.Y()/scale.Y(), p.Z()/scale.Z())
		return q.Length() < 1
	}

	for x := -1.4; x <= 1.4; x += 0.07 {
		for y := -1.4; y <= 1.4; y += 0.07 {
			at := vector3.New(x, y, 0.)
			v := field(at)
			require.False(t, math.IsNaN(v) || math.IsInf(v, 0), "field is non-finite at %v", at)

			if math.Abs(v) < 0.02 {
				continue // too close to the surface to call
			}
			assert.Equal(t, inside(at), v < 0, "sign disagrees with the ellipsoid at %v", at)
		}
	}
}
