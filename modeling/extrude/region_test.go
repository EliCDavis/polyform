package extrude_test

import (
	"testing"

	"github.com/EliCDavis/polyform/arrays"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/csg"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/sketch"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func square(size float64, center vector2.Float64) geometry.Shape {
	return sketch.RoundedRectangle(vector2.New(size, size), 0, center, 1)
}

func TestRegionWithDepthIsAClosedSolid(t *testing.T) {
	region, err := triangulation.FillDifference(
		[]geometry.Shape{square(10, vector2.New(0., 0.))},
		[]geometry.Shape{square(4, vector2.New(-1., 0.)), square(4, vector2.New(1., 0.))},
	)
	require.NoError(t, err)

	solid := extrude.RegionZ(region, 1)
	require.NoError(t, csg.CheckClosed(solid))
	assert.InDelta(t, 100.-6*4, signedVolume(solid), 1e-9)
}

func box(minX, minY, maxX, maxY float64) geometry.Shape {
	return geometry.Shape{
		vector2.New(minX, minY), vector2.New(maxX, minY),
		vector2.New(maxX, maxY), vector2.New(minX, maxY),
	}
}

func TestRegionStaysClosedWhenOutlineEdgesNearlyCoincide(t *testing.T) {
	for _, offset := range []float64{1e-10, 1e-9, 2.5e-9, 1e-8, 2.5e-8, 1e-7, 1e-6} {
		nearlyFlush := []geometry.Shape{box(0, 0, 2, 1), box(1, -0.5, 3, 1+offset), box(2.5, 1-offset, 3.5, 2)}

		filled, err := triangulation.FillDifference(nearlyFlush, nil)
		require.NoError(t, err, "outlines off by %g", offset)
		assert.NoError(t, csg.CheckClosed(extrude.RegionZ(filled, 0.3)), "outlines off by %g", offset)

		cut, err := triangulation.FillDifference([]geometry.Shape{box(-2, -2, 6, 4)}, nearlyFlush)
		require.NoError(t, err, "holes off by %g", offset)
		assert.NoError(t, csg.CheckClosed(extrude.RegionZ(cut, 0.3)), "holes off by %g", offset)
	}
}

func TestRegionWithoutDepthIsOneFaceForward(t *testing.T) {
	region, err := triangulation.Fill(square(2, vector2.New(0., 0.)))
	require.NoError(t, err)

	face := extrude.RegionZ(region, 0)
	normals := face.Float3Attribute(modeling.NormalAttribute)
	for i := 0; i < normals.Len(); i++ {
		assert.Equal(t, 1., normals.At(i).Z())
	}
	assert.Equal(t, 2, face.PrimitiveCount())
}

func TestAreaNodeExtrudesAFilledAreaWithASpreadListOfHoles(t *testing.T) {
	pads := nodes.GetNodeOutputPort[[][]vector2.Float64](&nodes.Struct[sketch.RoundedRectangleNode]{
		Data: sketch.RoundedRectangleNode{
			Size:   nodes.ConstOutput[vector2.Float64]{Val: vector2.New(1., 1.)},
			Center: nodes.ConstOutput[[]vector2.Float64]{Val: []vector2.Float64{vector2.New(-2., 0.), vector2.New(2., 0.)}},
		},
	}, "Out")
	spread := &nodes.Struct[arrays.SpreadNode]{}
	require.NoError(t, spread.Inputs()["Array"].(nodes.SingleValueInputPort).Set(pads))
	holes, ok := spread.Outputs()["Out"].(nodes.Output[[]vector2.Float64])
	require.True(t, ok, "a spread of outlines is typed as one outline")

	board := nodes.GetNodeOutputPort[sketch.Area](&nodes.Struct[sketch.FillNode]{
		Data: sketch.FillNode{
			Outlines: []nodes.LiftedPort[[]vector2.Float64]{nodes.ConstOutput[[]vector2.Float64]{Val: square(10, vector2.New(0., 0.))}},
			Holes:    []nodes.LiftedPort[[]vector2.Float64]{holes},
		},
	}, "Out")
	solid := nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[extrude.AreaNode]{
		Data: extrude.AreaNode{Area: board, Depth: nodes.ConstOutput[float64]{Val: 1}},
	}, "Z").Value()

	require.NoError(t, csg.CheckClosed(solid))
	assert.InDelta(t, 100.-2, signedVolume(solid), 1e-9)
}

func TestAreaNodeExtrudesAStrokedRingAsARing(t *testing.T) {
	path := []vector2.Float64{vector2.New(0., 0.), vector2.New(10., 0.), vector2.New(10., 10.), vector2.New(0., 10.)}
	ring, err := sketch.Stroke(path, 2, 16, true)
	require.NoError(t, err)

	solid := nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[extrude.AreaNode]{
		Data: extrude.AreaNode{
			Area:  nodes.ConstOutput[sketch.Area]{Val: ring},
			Depth: nodes.ConstOutput[float64]{Val: 1},
		},
	}, "Z").Value()

	require.NoError(t, csg.CheckClosed(solid))
	assert.InDelta(t, ring.Size(), signedVolume(solid), 1e-6)
}

func TestAreaNodeKeepsAnIslandInsideAHole(t *testing.T) {
	frame, err := sketch.Fill([]geometry.Shape{square(10, vector2.New(0., 0.))}, []geometry.Shape{square(6, vector2.New(0., 0.))})
	require.NoError(t, err)
	island, err := sketch.Fill([]geometry.Shape{square(2, vector2.New(0., 0.))}, nil)
	require.NoError(t, err)
	both, err := sketch.Union(frame, island)
	require.NoError(t, err)

	solid, err := extrude.AreaZ(both, 1)
	require.NoError(t, err)

	require.NoError(t, csg.CheckClosed(solid))
	assert.InDelta(t, 100.-36+4, signedVolume(solid), 1e-9)
}

func TestAreaNodeGivesOneSolidPerArea(t *testing.T) {
	small, err := sketch.Fill([]geometry.Shape{square(2, vector2.New(0., 0.))}, nil)
	require.NoError(t, err)
	large, err := sketch.Fill([]geometry.Shape{square(4, vector2.New(0., 0.))}, nil)
	require.NoError(t, err)

	solids := nodes.GetNodeOutputPort[[]modeling.Mesh](&nodes.Struct[extrude.AreaNode]{
		Data: extrude.AreaNode{
			Area:  nodes.ConstOutput[[]sketch.Area]{Val: []sketch.Area{small, large}},
			Depth: nodes.ConstOutput[float64]{Val: 1},
		},
	}, "Z").Value()

	require.Len(t, solids, 2)
	assert.InDelta(t, 4, signedVolume(solids[0]), 1e-9)
	assert.InDelta(t, 16, signedVolume(solids[1]), 1e-9)
}

func TestAreaNodeExtrudesAlongEachAxis(t *testing.T) {
	plate, err := sketch.Fill([]geometry.Shape{box(1, 2, 4, 6)}, nil)
	require.NoError(t, err)
	node := &nodes.Struct[extrude.AreaNode]{
		Data: extrude.AreaNode{
			Area:  nodes.ConstOutput[sketch.Area]{Val: plate},
			Depth: nodes.ConstOutput[float64]{Val: 0.5},
		},
	}

	for name, tc := range map[string]struct {
		min vector3.Float64
		max vector3.Float64
	}{
		"X": {vector3.New(0., 1., 2.), vector3.New(0.5, 4., 6.)},
		"Y": {vector3.New(1., 0., 2.), vector3.New(4., 0.5, 6.)},
		"Z": {vector3.New(1., 2., 0.), vector3.New(4., 6., 0.5)},
	} {
		solid := nodes.GetNodeOutputPort[modeling.Mesh](node, name).Value()
		require.NoError(t, csg.CheckClosed(solid), name)
		assert.InDelta(t, 6, signedVolume(solid), 1e-9, name)

		bounds := solid.BoundingBox(modeling.PositionAttribute)
		assert.InDelta(t, 0, bounds.Min().Distance(tc.min), 1e-9, name)
		assert.InDelta(t, 0, bounds.Max().Distance(tc.max), 1e-9, name)
	}
}

func TestAreaNodeNegativeDepthExtrudesDownEachAxis(t *testing.T) {
	plate, err := sketch.Fill([]geometry.Shape{box(1, 2, 4, 6)}, nil)
	require.NoError(t, err)
	node := &nodes.Struct[extrude.AreaNode]{
		Data: extrude.AreaNode{
			Area:  nodes.ConstOutput[sketch.Area]{Val: plate},
			Depth: nodes.ConstOutput[float64]{Val: -0.5},
		},
	}

	for name, tc := range map[string]struct {
		min vector3.Float64
		max vector3.Float64
	}{
		"X": {vector3.New(-0.5, 1., 2.), vector3.New(0., 4., 6.)},
		"Y": {vector3.New(1., -0.5, 2.), vector3.New(4., 0., 6.)},
		"Z": {vector3.New(1., 2., -0.5), vector3.New(4., 6., 0.)},
	} {
		solid := nodes.GetNodeOutputPort[modeling.Mesh](node, name).Value()
		require.NoError(t, csg.CheckClosed(solid), name)
		assert.InDelta(t, 6, signedVolume(solid), 1e-9, name)

		bounds := solid.BoundingBox(modeling.PositionAttribute)
		assert.InDelta(t, 0, bounds.Min().Distance(tc.min), 1e-9, name)
		assert.InDelta(t, 0, bounds.Max().Distance(tc.max), 1e-9, name)
	}
}
