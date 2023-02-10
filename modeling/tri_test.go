package modeling_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestTri_PointInside(t *testing.T) {
	mesh := modeling.NewMesh(
		[]int{0, 1, 2},
		map[string][]vector3.Vector[float64]{
			modeling.PositionAttribute: []vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 0., 0.),
			},
		},
		nil,
		nil,
		nil,
	)

	tri := mesh.Tri(0)

	assert.True(t, tri.PointInSide(vector3.New(.25, .25, 0.)))
	assert.False(t, tri.PointInSide(vector3.New(-.25, .25, 0.)))
	assert.False(t, tri.PointInSide(vector3.New(-.25, .25, .25)))
}

func TestTri_LineIntersects(t *testing.T) {
	mesh := modeling.NewMesh(
		[]int{0, 1, 2},
		map[string][]vector3.Vector[float64]{
			modeling.PositionAttribute: []vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 0., 0.),
			},
		},
		nil,
		nil,
		nil,
	)

	tri := mesh.Tri(0)

	line := geometry.NewLine3D(
		vector3.New(.25, .25, -.25),
		vector3.New(.25, .25, .25),
	)

	// ACT ====================================================================
	intersection, intersects := tri.LineIntersects(line)

	// ASSERT =================================================================
	assert.True(t, intersects)
	assert.InDelta(t, .25, intersection.X(), 0.0000001)
	assert.InDelta(t, .25, intersection.Y(), 0.0000001)
	assert.InDelta(t, 0, intersection.Z(), 0.0000001)
}
