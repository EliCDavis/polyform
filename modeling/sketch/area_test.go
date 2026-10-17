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

func squareArea(t *testing.T, size float64, center vector2.Float64) sketch.Area {
	t.Helper()
	area, err := sketch.Fill([]geometry.Shape{sketch.RoundedRectangle(vector2.New(size, size), 0, center, 1)}, nil)
	require.NoError(t, err)
	return area
}

func requireWellFormed(t *testing.T, area sketch.Area) {
	t.Helper()
	for i, polygon := range area.Polygons {
		assert.Positive(t, geometry.Shape(polygon.Outline).SignedArea(), "polygon %d outline runs counter-clockwise", i)
		for j, hole := range polygon.Holes {
			assert.Negative(t, geometry.Shape(hole).SignedArea(), "polygon %d hole %d runs clockwise", i, j)
		}
	}
}

func squarePath(size float64) []vector2.Float64 {
	return []vector2.Float64{vector2.New(0., 0.), vector2.New(size, 0.), vector2.New(size, size), vector2.New(0., size)}
}

func TestFillCutsHolesOutOfOutlines(t *testing.T) {
	area, err := sketch.Fill(
		[]geometry.Shape{sketch.RoundedRectangle(vector2.New(10., 10.), 0, vector2.New(0., 0.), 1)},
		[]geometry.Shape{sketch.RoundedRectangle(vector2.New(4., 4.), 0, vector2.New(0., 0.), 1)},
	)
	require.NoError(t, err)

	require.Len(t, area.Polygons, 1)
	assert.Len(t, area.Polygons[0].Holes, 1)
	assert.InDelta(t, 84, area.Size(), 1e-9)
	requireWellFormed(t, area)
}

func TestStrokeOfAClosedPathIsARing(t *testing.T) {
	ring, err := sketch.Stroke(squarePath(10), 2, 16, true)
	require.NoError(t, err)

	require.Len(t, ring.Polygons, 1)
	assert.Len(t, ring.Polygons[0].Holes, 1)
	assert.InDelta(t, 12*12-(4-math.Pi)-8*8, ring.Size(), 0.05, "rounded outside corners, sharp inside ones")
	requireWellFormed(t, ring)
}

func TestStrokeClosedMatchesRepeatingTheFirstPoint(t *testing.T) {
	closed, err := sketch.Stroke(squarePath(10), 2, 16, true)
	require.NoError(t, err)
	repeated, err := sketch.Stroke(append(squarePath(10), vector2.New(0., 0.)), 2, 16, false)
	require.NoError(t, err)

	assert.InDelta(t, repeated.Size(), closed.Size(), 1e-9)
}

func TestStrokeOfAPathEnclosingTwoGapsHasTwoHoles(t *testing.T) {
	path := []vector2.Float64{
		vector2.New(0., 0.), vector2.New(20., 0.), vector2.New(20., 10.), vector2.New(0., 10.),
		vector2.New(0., 0.), vector2.New(10., 0.), vector2.New(10., 10.),
	}

	area, err := sketch.Stroke(path, 2, 16, false)
	require.NoError(t, err)

	require.Len(t, area.Polygons, 1)
	assert.Len(t, area.Polygons[0].Holes, 2)
	requireWellFormed(t, area)
}

func TestUnionMergesOverlappingAreas(t *testing.T) {
	merged, err := sketch.Union(squareArea(t, 4, vector2.New(0., 0.)), squareArea(t, 4, vector2.New(2., 0.)))
	require.NoError(t, err)

	require.Len(t, merged.Polygons, 1)
	assert.InDelta(t, 24, merged.Size(), 1e-9)
}

func TestIntersectKeepsOnlyTheSharedArea(t *testing.T) {
	shared, err := sketch.Intersect(squareArea(t, 4, vector2.New(0., 0.)), squareArea(t, 4, vector2.New(2., 0.)))
	require.NoError(t, err)

	assert.InDelta(t, 8, shared.Size(), 1e-9)
}

func TestSubtractingARingLeavesAFrameAndAnIsland(t *testing.T) {
	ring, err := sketch.Stroke(
		[]vector2.Float64{vector2.New(-5., -5.), vector2.New(5., -5.), vector2.New(5., 5.), vector2.New(-5., 5.)},
		2, 16, true,
	)
	require.NoError(t, err)

	cut, err := sketch.Subtract(squareArea(t, 20, vector2.New(0., 0.)), ring)
	require.NoError(t, err)

	require.Len(t, cut.Polygons, 2, "the frame outside the ring and the island inside it")
	holes := len(cut.Polygons[0].Holes) + len(cut.Polygons[1].Holes)
	assert.Equal(t, 1, holes, "only the frame has a hole; the island sits inside it as its own polygon")
	assert.InDelta(t, 400-ring.Size(), cut.Size(), 1e-6)
	requireWellFormed(t, cut)
}

func TestMirroringAnAreaKeepsItsHolesAsHoles(t *testing.T) {
	ring, err := sketch.Stroke(squarePath(10), 2, 16, true)
	require.NoError(t, err)

	mirrored := ring.Transform(vector2.New(3., 0.), 30, vector2.New(-1., 2.))

	require.Len(t, mirrored.Polygons, 1)
	assert.Len(t, mirrored.Polygons[0].Holes, 1)
	assert.InDelta(t, 2*ring.Size(), mirrored.Size(), 1e-6)
	requireWellFormed(t, mirrored)
}

func TestAreaContainsItsSolidButNotItsHoles(t *testing.T) {
	ring, err := sketch.Stroke(squarePath(10), 2, 16, true)
	require.NoError(t, err)

	assert.True(t, ring.Contains(vector2.New(0., 5.)))
	assert.False(t, ring.Contains(vector2.New(5., 5.)))
	assert.False(t, ring.Contains(vector2.New(20., 5.)))
}

func TestStrokeNodeGivesOneAreaPerPath(t *testing.T) {
	areas := nodes.GetNodeOutputPort[[]sketch.Area](&nodes.Struct[sketch.StrokeNode]{Data: sketch.StrokeNode{
		Path:   nodes.ConstOutput[[][]vector2.Float64]{Val: [][]vector2.Float64{squarePath(4), squarePath(6)}},
		Width:  nodes.ConstOutput[float64]{Val: 0.5},
		Closed: nodes.ConstOutput[bool]{Val: true},
	}}, "Out").Value()

	require.Len(t, areas, 2)
	for _, area := range areas {
		require.Len(t, area.Polygons, 1)
		assert.Len(t, area.Polygons[0].Holes, 1)
	}
}

func TestSubtractNodeCutsEveryRemovedArea(t *testing.T) {
	cut := nodes.GetNodeOutputPort[sketch.Area](&nodes.Struct[sketch.SubtractNode]{Data: sketch.SubtractNode{
		Base: nodes.ConstOutput[sketch.Area]{Val: squareArea(t, 10, vector2.New(0., 0.))},
		Remove: []nodes.Output[sketch.Area]{
			nodes.ConstOutput[sketch.Area]{Val: squareArea(t, 2, vector2.New(-2., 0.))},
			nodes.ConstOutput[sketch.Area]{Val: squareArea(t, 2, vector2.New(2., 0.))},
		},
	}}, "Out").Value()

	require.Len(t, cut.Polygons, 1)
	assert.Len(t, cut.Polygons[0].Holes, 2)
	assert.InDelta(t, 92, cut.Size(), 1e-9)
}
