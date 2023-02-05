package geometry_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestAABB(t *testing.T) {
	center := vector3.New(2., 3., 4.)
	aabb := geometry.NewAABB(center, vector3.One[float64]())

	assert.Equal(t, center, aabb.Center())
	assert.True(t, aabb.Contains(center))
	assert.False(t, aabb.Contains(vector3.One[float64]()))
	assert.False(t, aabb.Contains(vector3.Zero[float64]()))
	assert.False(t, aabb.Contains(vector3.Right[float64]()))
}

func TestAABBClosesetPoint(t *testing.T) {
	center := vector3.New(2., 3., 4.)
	aabb := geometry.NewAABB(center, vector3.One[float64]().Scale(2))

	tests := map[string]struct {
		input vector3.Float64
		want  vector3.Float64
	}{
		"center": {input: vector3.New(2., 3., 4.), want: vector3.New(2., 3., 4.)},
		"left":   {input: vector3.New(0., 3., 4.), want: vector3.New(1., 3., 4.)},
		"right":  {input: vector3.New(4., 3., 4.), want: vector3.New(3., 3., 4.)},
		"up":     {input: vector3.New(2., 5., 4.), want: vector3.New(2., 4., 4.)},
		"down":   {input: vector3.New(2., 1., 4.), want: vector3.New(2., 2., 4.)},
		"fwd":    {input: vector3.New(2., 3., 6.), want: vector3.New(2., 3., 5.)},
		"back":   {input: vector3.New(2., 3., 2.), want: vector3.New(2., 3., 3.)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := aabb.ClosestPoint(tc.input)
			assert.InDelta(t, tc.want.X(), got.X(), 0.00000001)
			assert.InDelta(t, tc.want.Y(), got.Y(), 0.00000001)
			assert.InDelta(t, tc.want.Z(), got.Z(), 0.00000001)
		})
	}
}

func TestAABBIntersectsRayInRange(t *testing.T) {
	center := vector3.New(2., 3., 4.)
	aabb := geometry.NewAABB(center, vector3.One[float64]().Scale(2))

	tests := map[string]struct {
		ray      geometry.Ray
		min, max float64
		want     bool
	}{
		"up at origin": {
			ray:  geometry.NewRay(vector3.New(0., 0., 0.), vector3.New(0., 1., 0.)),
			min:  0,
			max:  10,
			want: false,
		},
		"centered under aabb": {
			ray:  geometry.NewRay(vector3.New(2., 0., 4.), vector3.New(0., 1., 0.)),
			min:  0,
			max:  10,
			want: true,
		},
		"centered under aabb stop short": {
			ray:  geometry.NewRay(vector3.New(2., 0., 4.), vector3.New(0., 1., 0.)),
			min:  0,
			max:  0.5,
			want: false,
		},
		"centered under aabb skips over": {
			ray:  geometry.NewRay(vector3.New(2., 0., 4.), vector3.New(0., 1., 0.)),
			min:  5,
			max:  10,
			want: false,
		},

		"down at origin": {
			ray:  geometry.NewRay(vector3.New(0., 0., 0.), vector3.New(0., -1., 0.)),
			min:  0,
			max:  10,
			want: false,
		},
		"centered over aabb": {
			ray:  geometry.NewRay(vector3.New(2., 10., 4.), vector3.New(0., -1., 0.)),
			min:  0,
			max:  10,
			want: true,
		},
		"centered over aabb stop short": {
			ray:  geometry.NewRay(vector3.New(2., 10., 4.), vector3.New(0., -1., 0.)),
			min:  0,
			max:  0.5,
			want: false,
		},
		"centered over aabb skips over": {
			ray:  geometry.NewRay(vector3.New(2., 0., 4.), vector3.New(0., -1., 0.)),
			min:  9,
			max:  10,
			want: false,
		},

		"right at origin": {
			ray:  geometry.NewRay(vector3.New(0., 0., 0.), vector3.New(1., 0., 0.)),
			min:  0,
			max:  10,
			want: false,
		},
		"centered left of origin": {
			ray:  geometry.NewRay(vector3.New(0., 3., 4.), vector3.New(1., 0., 0.)),
			min:  0,
			max:  10,
			want: true,
		},
		"centered left of origin stop short": {
			ray:  geometry.NewRay(vector3.New(0., 3., 4.), vector3.New(1., 0., 0.)),
			min:  0,
			max:  0.5,
			want: false,
		},
		"centered left of origin skips over": {
			ray:  geometry.NewRay(vector3.New(0., 3., 4.), vector3.New(1., 0., 0.)),
			min:  5,
			max:  10,
			want: false,
		},

		"forward at origin": {
			ray:  geometry.NewRay(vector3.New(0., 0., 0.), vector3.New(0., 0., 1.)),
			min:  0,
			max:  10,
			want: false,
		},
		"centered forward of origin": {
			ray:  geometry.NewRay(vector3.New(2., 3., 0.), vector3.New(0., 0., 1.)),
			min:  0,
			max:  10,
			want: true,
		},
		"centered forward of origin stop short": {
			ray:  geometry.NewRay(vector3.New(2., 3., 0.), vector3.New(0., 0., 1.)),
			min:  0,
			max:  0.5,
			want: false,
		},
		"centered forward of origin skips over": {
			ray:  geometry.NewRay(vector3.New(2., 3., 0.), vector3.New(0., 0., 1.)),
			min:  5,
			max:  10,
			want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := aabb.IntersectsRayInRange(tc.ray, tc.min, tc.max)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAABBEncapsulatePoints(t *testing.T) {
	aabb := geometry.NewEmptyAABB()

	aabb.EncapsulatePoint(vector3.New(-0.5, 0., 0.))
	aabb.EncapsulatePoint(vector3.New(0.5, 0., 0.))

	aabb.EncapsulatePoint(vector3.New(0., 0.5, 0.))
	aabb.EncapsulatePoint(vector3.New(0., -0.5, 0.))

	aabb.EncapsulatePoint(vector3.New(0., 0., 0.5))
	aabb.EncapsulatePoint(vector3.New(0., 0., -0.5))

	assert.Equal(t, vector3.Zero[float64](), aabb.Center())
	assert.Equal(t, vector3.One[float64](), aabb.Size())
}

func TestAABBFromPoints(t *testing.T) {
	aabb := geometry.NewAABBFromPoints(
		vector3.New(-0.5, 0., 0.),
		vector3.New(0.5, 0., 0.),
		vector3.New(0., 0.5, 0.),
		vector3.New(0., -0.5, 0.),
		vector3.New(0., 0., 0.5),
		vector3.New(0., 0., -0.5),
	)

	assert.Equal(t, vector3.Zero[float64](), aabb.Center())
	assert.Equal(t, vector3.One[float64](), aabb.Size())
}
