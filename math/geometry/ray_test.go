package geometry_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestTriangleRayHit(t *testing.T) {
	tri := geometry.Triangle{vector3.New(0., 0., 0.), vector3.New(2., 0., 0.), vector3.New(0., 2., 0.)}
	down := vector3.New(0., 0., -1.)
	cast := func(origin, direction vector3.Float64) (geometry.TriangleHit, bool) {
		return tri.RayHit(geometry.NewRay(origin, direction))
	}

	hit, ok := cast(vector3.New(0.5, 0.5, 3.), down)
	assert.True(t, ok)
	assert.InDelta(t, 3, hit.Distance, 1e-12)
	assert.InDelta(t, 0.25, hit.U, 1e-12)
	assert.InDelta(t, 0.25, hit.V, 1e-12)
	assert.True(t, hit.Inside(0))
	assert.InDelta(t, 0.25, hit.Margin(), 1e-12)

	outside, ok := cast(vector3.New(3., 3., 1.), down)
	assert.True(t, ok, "the plane is still hit")
	assert.False(t, outside.Inside(0))
	assert.Negative(t, outside.Margin())

	grazing, ok := cast(vector3.New(-1e-12, 1., 1.), down)
	assert.True(t, ok)
	assert.False(t, grazing.Inside(0))
	assert.True(t, grazing.Inside(1e-9))

	behind, ok := cast(vector3.New(0.5, 0.5, -3.), down)
	assert.True(t, ok)
	assert.Negative(t, behind.Distance)

	_, ok = cast(vector3.New(0., 0., 1.), vector3.New(1., 0., 0.))
	assert.False(t, ok, "parallel to the plane")
}

func TestRay(t *testing.T) {
	ray := geometry.NewRay(vector3.Zero[float64](), vector3.Up[float64]())

	assert.Equal(t, vector3.Zero[float64](), ray.Origin())
	assert.Equal(t, vector3.Up[float64](), ray.Direction())
	assert.Equal(t, vector3.Up[float64]().Scale(2), ray.At(2))
}
