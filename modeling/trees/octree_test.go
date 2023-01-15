package trees_test

import (
	"math/rand"
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/trees"
	"github.com/EliCDavis/vector"
	"github.com/stretchr/testify/assert"
)

func TestOctreeSingleTri(t *testing.T) {
	// ARRANGE ================================================================
	mesh := modeling.NewMesh(
		[]int{0, 1, 2},
		map[string][]vector.Vector3{
			modeling.PositionAttribute: {
				vector.NewVector3(0, 0, 0),
				vector.NewVector3(0, 1, 0),
				vector.NewVector3(1, 1, 0),
			},
		},
		nil,
		nil,
		nil,
	)
	tree := trees.FromMesh(mesh)

	// ACT ====================================================================
	p := tree.ClosestPoint(vector.Vector3Zero())

	// ASSERT =================================================================
	assert.Equal(t, p, vector.Vector3Zero())
}

func TestOctreeTwoTris(t *testing.T) {
	// ARRANGE ================================================================
	mesh := modeling.NewMesh(
		[]int{0, 1, 2, 0, 2, 3},
		map[string][]vector.Vector3{
			modeling.PositionAttribute: {
				vector.NewVector3(0, 0, 0),
				vector.NewVector3(0, 1, 0),
				vector.NewVector3(1, 1, 0),
				vector.NewVector3(1, 0, 0),
			},
		},
		nil,
		nil,
		nil,
	)
	tree := trees.FromMesh(mesh)

	// ACT ====================================================================
	p := tree.ClosestPoint(vector.Vector3Zero())

	// ASSERT =================================================================
	assert.Equal(t, p, vector.Vector3Zero())
}

func TestOctreeSphere(t *testing.T) {
	// ARRANGE ================================================================
	mesh := primitives.UVSphere(1, 100, 100)
	tree := trees.FromMesh(mesh)

	testPointCount := 1000
	testPoints := make([]vector.Vector3, testPointCount)
	for i := 0; i < testPointCount; i++ {
		testPoints[i] = vector.NewVector3(
			-1+(rand.Float64()*2),
			-1+(rand.Float64()*2),
			-1+(rand.Float64()*2),
		).Normalized()
	}

	// ACT / ASSERT ===========================================================
	for i := 0; i < testPointCount; i++ {
		p := tree.ClosestPoint(testPoints[i])
		assert.InDelta(t, testPoints[i].X(), p.X(), 0.05)
		assert.InDelta(t, testPoints[i].Y(), p.Y(), 0.05)
		assert.InDelta(t, testPoints[i].Z(), p.Z(), 0.05)

	}

}
