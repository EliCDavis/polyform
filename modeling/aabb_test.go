package modeling_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestAABBContains(t *testing.T) {
	tests := map[string]struct {
		center vector3.Float64
		size   vector3.Float64
		point  vector3.Float64
		result bool
	}{
		"simple": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.Zero[float64](),
			result: true,
		},
		"on corner": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(0.5, 0.5, 0.5),
			result: true,
		},
		"on edge": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(0., 0.5, 0.5),
			result: true,
		},
		"on face": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(0., 0., 0.5),
			result: true,
		},
		"0 volume": {
			center: vector3.Zero[float64](),
			size:   vector3.Zero[float64](),
			point:  vector3.New(0., 0., 0.),
			result: true,
		},
		"outside corner": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(0.500001, 0.500001, 0.500001),
			result: false,
		},
		"outside edge": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(0., 0.500001, 0.500001),
			result: false,
		},
		"outside face": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(0., 0., 0.500001),
			result: false,
		},

		"outside corner 2": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(-0.500001, -0.500001, -0.500001),
			result: false,
		},
		"outside edge 2": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(0., -0.500001, -0.500001),
			result: false,
		},
		"outside face 2": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(0., 0., -0.500001),
			result: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			aabb := modeling.NewAABB(tc.center, tc.size)
			assert.Equal(t, tc.result, aabb.Contains(tc.point))
		})
	}
}

func TestAABBEncapsulate(t *testing.T) {
	// ARRANGE ================================================================
	pts := []vector3.Float64{
		vector3.New[float64](0.1, 0, 0),
		vector3.New[float64](0.1, 0.1, 0),
		vector3.New[float64](0.1, 0.1, 0.1),
		vector3.New[float64](-0.1, 0, 0),
		vector3.New[float64](0, -0.1, 0),
		vector3.New[float64](0.1, -0.1, -0.1),
		vector3.New[float64](0, 1, 0),
		vector3.New[float64](0, -1, 0),
	}
	aabb := modeling.NewAABB(vector3.Zero[float64](), vector3.Zero[float64]())

	// ACT/ASSERT =============================================================
	for _, pt := range pts {
		assert.False(t, aabb.Contains(pt))
		aabb.EncapsulatePoint(pt)
		assert.True(t, aabb.Contains(pt))
	}

	// Make sure we now encapsulate everything
	for _, pt := range pts {
		assert.True(t, aabb.Contains(pt))
	}
}

func TestAABBExpand(t *testing.T) {
	// ARRANGE ================================================================
	aabb := modeling.NewAABB(vector3.Zero[float64](), vector3.One[float64]())

	// ACT ====================================================================
	aabb.Expand(2)

	// ASSERT =================================================================
	assert.Equal(t, vector3.New(3., 3., 3.), aabb.Size())
}
