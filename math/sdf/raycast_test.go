package sdf_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func unitSphere() sample.Vec3ToFloat {
	return sdf.Sphere(vector3.Zero[float64](), 1.5)
}

func TestRaycastFromOutside(t *testing.T) {
	distance, hit := sdf.Raycast(unitSphere(), vector3.New(5., 0., 0.), vector3.New(-1., 0., 0.), 10)

	require.True(t, hit)
	assert.InDelta(t, 3.5, distance, 0.001)
}

func TestRaycastFromInside(t *testing.T) {
	distance, hit := sdf.Raycast(unitSphere(), vector3.Zero[float64](), vector3.New(0., 1., 0.), 10)

	require.True(t, hit)
	assert.InDelta(t, 1.5, distance, 0.001)
}

func TestRaycastMissesWhenPointedAway(t *testing.T) {
	_, hit := sdf.Raycast(unitSphere(), vector3.New(5., 0., 0.), vector3.New(1., 0., 0.), 10)
	assert.False(t, hit)
}

func TestRaycastRespectsMaxDistance(t *testing.T) {
	_, hit := sdf.Raycast(unitSphere(), vector3.New(5., 0., 0.), vector3.New(-1., 0., 0.), 1)
	assert.False(t, hit, "the surface is 3.5 away")
}

func TestRaycastRefusesADegenerateRay(t *testing.T) {
	_, hit := sdf.Raycast(unitSphere(), vector3.New(5., 0., 0.), vector3.Zero[float64](), 10)
	assert.False(t, hit)
}

func TestNormalPointsOutOfTheShape(t *testing.T) {
	normal := sdf.Normal(unitSphere(), vector3.New(1.5, 0., 0.))

	assert.InDelta(t, 1, normal.X(), 0.01)
	assert.InDelta(t, 0, normal.Y(), 0.01)
	assert.InDelta(t, 0, normal.Z(), 0.01)
}

func TestNormalOfAFlatFieldIsZeroRatherThanNaN(t *testing.T) {
	flat := func(vector3.Float64) float64 { return 1 }
	assert.Equal(t, vector3.Zero[float64](), sdf.Normal(flat, vector3.Zero[float64]()))
}

func TestRaycastNodeReportsThePointAndNormal(t *testing.T) {
	node := &nodes.Struct[sdf.RaycastNode]{
		Data: sdf.RaycastNode{
			Field:     nodes.ConstOutput[sample.Vec3ToFloat]{Val: unitSphere()},
			Origin:    nodes.ConstOutput[vector3.Float64]{Val: vector3.New(5., 0., 0.)},
			Direction: nodes.ConstOutput[vector3.Float64]{Val: vector3.New(-1., 0., 0.)},
		},
	}

	point := nodes.GetNodeOutputPort[vector3.Float64](node, "Point").Value()
	assert.InDelta(t, 1.5, point.X(), 0.001)

	normal := nodes.GetNodeOutputPort[vector3.Float64](node, "Normal").Value()
	assert.InDelta(t, 1, normal.X(), 0.01)

	assert.InDelta(t, 3.5, nodes.GetNodeOutputPort[float64](node, "Distance").Value(), 0.001)
	assert.True(t, nodes.GetNodeOutputPort[bool](node, "Hit").Value())
}

func TestRaycastNodeMarchesAnArrayOfOrigins(t *testing.T) {
	node := &nodes.Struct[sdf.RaycastNode]{
		Data: sdf.RaycastNode{
			Field: nodes.ConstOutput[sample.Vec3ToFloat]{Val: unitSphere()},
			Origin: nodes.ConstOutput[[]vector3.Float64]{Val: []vector3.Float64{
				vector3.New(5., 0., 0.),
				vector3.New(0., 5., 0.),
			}},
			Direction: nodes.ConstOutput[[]vector3.Float64]{Val: []vector3.Float64{
				vector3.New(-1., 0., 0.),
				vector3.New(0., -1., 0.),
			}},
		},
	}

	points := nodes.GetNodeOutputPort[[]vector3.Float64](node, "Point").Value()
	require.Len(t, points, 2)
	assert.InDelta(t, 1.5, points[0].X(), 0.001)
	assert.InDelta(t, 1.5, points[1].Y(), 0.001)
}
