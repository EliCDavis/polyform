package sketch_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling/sketch"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func square() []vector2.Float64 {
	return sketch.RoundedRectangle(vector2.New(2., 2.), 0, vector2.New(0., 0.), 1)
}

func TestBevelChamfersEveryCornerOfAClosedOutline(t *testing.T) {
	cut := sketch.Corners{Distance: 0.5, Closed: true}.Apply(square())

	assert.Len(t, cut, 8)
	assert.InDelta(t, 4-4*0.125, geometry.Shape(cut).SignedArea(), 1e-9)
}

func TestBevelRoundsARightAngleWithItsDistanceAsTheRadius(t *testing.T) {
	cut := sketch.Corners{Distance: 1, Round: true, Segments: 64, Closed: true}.Apply(square())

	for i, p := range cut {
		assert.InDeltaf(t, 1, p.Length(), 1e-9, "point %d", i)
	}
	assert.InDelta(t, math.Pi, geometry.Shape(cut).SignedArea(), 0.01)
}

func TestBevelLeavesTheEndsOfAnOpenPath(t *testing.T) {
	path := []vector2.Float64{vector2.New(0., 0.), vector2.New(10., 0.), vector2.New(10., 10.)}
	cut := sketch.Corners{Distance: 2}.Apply(path)

	assert.Equal(t, []vector2.Float64{vector2.New(0., 0.), vector2.New(8., 0.), vector2.New(10., 2.), vector2.New(10., 10.)}, cut)
}

func TestBevelStopsAtTheShorterSide(t *testing.T) {
	path := []vector2.Float64{vector2.New(0., 0.), vector2.New(1., 0.), vector2.New(1., 10.)}
	cut := sketch.Corners{Distance: 5}.Apply(path)

	assert.Equal(t, []vector2.Float64{vector2.New(0., 0.), vector2.New(1., 1.), vector2.New(1., 10.)}, cut)
}

func TestBevelPathNodeCutsEachPathOfAList(t *testing.T) {
	paths := nodes.GetNodeOutputPort[[][]vector2.Float64](&nodes.Struct[sketch.BevelPathNode]{Data: sketch.BevelPathNode{
		Outline: nodes.ConstOutput[[][]vector2.Float64]{Val: [][]vector2.Float64{
			{vector2.New(0., 0.), vector2.New(4., 0.), vector2.New(4., 4.)},
			{vector2.New(0., 0.), vector2.New(0., 4.), vector2.New(4., 4.), vector2.New(4., 8.)},
		}},
		Distance: nodes.ConstOutput[float64]{Val: 1},
	}}, "Out").Value()

	require.Len(t, paths, 2)
	assert.Len(t, paths[0], 4)
	assert.Len(t, paths[1], 6)
}

func TestBevelTreatsFloatNoiseAsARepeatedPoint(t *testing.T) {
	noisy := []vector2.Float64{
		vector2.New(0., 0.), vector2.New(1e-15, 0.),
		vector2.New(2., 0.), vector2.New(2., 2.), vector2.New(0., 2.), vector2.New(0., 1e-15),
	}
	cut := sketch.Corners{Distance: 0.5, Closed: true}.Apply(noisy)

	assert.Len(t, cut, 8)
	assert.InDelta(t, 4-4*0.125, geometry.Shape(cut).SignedArea(), 1e-9)
}

func TestBevelAreaCutsOutlinesAndHoles(t *testing.T) {
	frame, err := sketch.Fill(
		[]geometry.Shape{sketch.RoundedRectangle(vector2.New(4., 4.), 0, vector2.New(0., 0.), 1)},
		[]geometry.Shape{sketch.RoundedRectangle(vector2.New(2., 2.), 0, vector2.New(0., 0.), 1)},
	)
	require.NoError(t, err)

	cut, err := sketch.BevelArea(frame, sketch.Corners{Distance: 0.5})
	require.NoError(t, err)

	require.Len(t, cut.Polygons, 1)
	require.Len(t, cut.Polygons[0].Holes, 1)
	assert.Len(t, cut.Polygons[0].Outline, 8)
	assert.Len(t, cut.Polygons[0].Holes[0], 8)
	assert.InDelta(t, 12, cut.Size(), 1e-9)
}

func TestBevelNodeRoundsEveryPolygon(t *testing.T) {
	two, err := sketch.Fill([]geometry.Shape{
		sketch.RoundedRectangle(vector2.New(2., 2.), 0, vector2.New(0., 0.), 1),
		sketch.RoundedRectangle(vector2.New(2., 2.), 0, vector2.New(5., 0.), 1),
	}, nil)
	require.NoError(t, err)

	area := nodes.GetNodeOutputPort[sketch.Area](&nodes.Struct[sketch.BevelNode]{Data: sketch.BevelNode{
		Area:     nodes.ConstOutput[sketch.Area]{Val: two},
		Distance: nodes.ConstOutput[float64]{Val: 1},
		Round:    nodes.ConstOutput[bool]{Val: true},
		Segments: nodes.ConstOutput[int]{Val: 64},
	}}, "Out").Value()

	require.Len(t, area.Polygons, 2)
	for i, polygon := range area.Polygons {
		assert.Greaterf(t, len(polygon.Outline), 60, "polygon %d", i)
	}
	assert.InDelta(t, 2*math.Pi, area.Size(), 0.01)
}
