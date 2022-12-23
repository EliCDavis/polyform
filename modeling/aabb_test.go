package modeling_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
	"github.com/stretchr/testify/assert"
)

func TestAABBContains(t *testing.T) {
	tests := map[string]struct {
		center vector.Vector3
		size   vector.Vector3
		point  vector.Vector3
		result bool
	}{
		"simple": {
			center: vector.Vector3Zero(),
			size:   vector.Vector3One(),
			point:  vector.Vector3Zero(),
			result: true,
		},
		"on corner": {
			center: vector.Vector3Zero(),
			size:   vector.Vector3One(),
			point:  vector.NewVector3(0.5, 0.5, 0.5),
			result: true,
		},
		"on edge": {
			center: vector.Vector3Zero(),
			size:   vector.Vector3One(),
			point:  vector.NewVector3(0, 0.5, 0.5),
			result: true,
		},
		"on face": {
			center: vector.Vector3Zero(),
			size:   vector.Vector3One(),
			point:  vector.NewVector3(0, 0, 0.5),
			result: true,
		},
		"0 volume": {
			center: vector.Vector3Zero(),
			size:   vector.Vector3Zero(),
			point:  vector.NewVector3(0, 0, 0),
			result: true,
		},
		"outside corner": {
			center: vector.Vector3Zero(),
			size:   vector.Vector3One(),
			point:  vector.NewVector3(0.500001, 0.500001, 0.500001),
			result: false,
		},
		"outside edge": {
			center: vector.Vector3Zero(),
			size:   vector.Vector3One(),
			point:  vector.NewVector3(0, 0.500001, 0.500001),
			result: false,
		},
		"outside face": {
			center: vector.Vector3Zero(),
			size:   vector.Vector3One(),
			point:  vector.NewVector3(0, 0, 0.500001),
			result: false,
		},

		"outside corner 2": {
			center: vector.Vector3Zero(),
			size:   vector.Vector3One(),
			point:  vector.NewVector3(-0.500001, -0.500001, -0.500001),
			result: false,
		},
		"outside edge 2": {
			center: vector.Vector3Zero(),
			size:   vector.Vector3One(),
			point:  vector.NewVector3(0, -0.500001, -0.500001),
			result: false,
		},
		"outside face 2": {
			center: vector.Vector3Zero(),
			size:   vector.Vector3One(),
			point:  vector.NewVector3(0, 0, -0.500001),
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
	pts := []vector.Vector3{
		vector.NewVector3(0.1, 0, 0),
		vector.NewVector3(0.1, 0.1, 0),
		vector.NewVector3(0.1, 0.1, 0.1),
		vector.NewVector3(-0.1, 0, 0),
		vector.NewVector3(0, -0.1, 0),
		vector.NewVector3(0.1, -0.1, -0.1),
		vector.NewVector3(0, 1, 0),
		vector.NewVector3(0, -1, 0),
	}
	aabb := modeling.NewAABB(vector.Vector3Zero(), vector.Vector3Zero())

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
	aabb := modeling.NewAABB(vector.Vector3Zero(), vector.Vector3One())

	// ACT ====================================================================
	aabb.Expand(2)

	// ASSERT =================================================================
	assert.Equal(t, vector.NewVector3(3, 3, 3), aabb.Size())
}
