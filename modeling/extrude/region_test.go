package extrude_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/csg"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func square(size float64, center vector2.Float64) geometry.Shape {
	return geometry.RoundedRectangle(vector2.New(size, size), 0, center, 1)
}

func TestRegionWithDepthIsAClosedSolid(t *testing.T) {
	region, err := triangulation.FillDifference(
		[]geometry.Shape{square(10, vector2.New(0., 0.))},
		[]geometry.Shape{square(4, vector2.New(-1., 0.)), square(4, vector2.New(1., 0.))},
	)
	require.NoError(t, err)

	solid := extrude.Region(region, 1)
	require.NoError(t, csg.CheckClosed(solid))
	assert.InDelta(t, 100.-6*4, signedVolume(solid), 1e-9)
}

func TestRegionWithoutDepthIsOneFaceForward(t *testing.T) {
	region, err := triangulation.Fill(square(2, vector2.New(0., 0.)))
	require.NoError(t, err)

	face := extrude.Region(region, 0)
	normals := face.Float3Attribute(modeling.NormalAttribute)
	for i := 0; i < normals.Len(); i++ {
		assert.Equal(t, 1., normals.At(i).Z())
	}
	assert.Equal(t, 2, face.PrimitiveCount())
}

func TestRegionNodeTakesAListOfHolesOnOneConnection(t *testing.T) {
	pads := nodes.GetNodeOutputPort[[][]vector2.Float64](&nodes.Struct[geometry.RoundedRectangleNode]{
		Data: geometry.RoundedRectangleNode{
			Size:   nodes.ConstOutput[vector2.Float64]{Val: vector2.New(1., 1.)},
			Center: nodes.ConstOutput[[]vector2.Float64]{Val: []vector2.Float64{vector2.New(-2., 0.), vector2.New(2., 0.)}},
		},
	}, "Out")

	solid := nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[extrude.RegionNode]{
		Data: extrude.RegionNode{
			Outlines: []nodes.LiftedPort[[]vector2.Float64]{nodes.ConstOutput[[]vector2.Float64]{Val: square(10, vector2.New(0., 0.))}},
			Holes:    []nodes.LiftedPort[[]vector2.Float64]{pads},
			Depth:    nodes.ConstOutput[float64]{Val: 1},
		},
	}, "Out").Value()

	assert.InDelta(t, 100.-2, signedVolume(solid), 1e-9)
}
