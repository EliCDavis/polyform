package sketch_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling/sketch"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func squareAt(size float64, center vector2.Float64) []vector2.Float64 {
	return sketch.RoundedRectangle(vector2.New(size, size), 0, center, 1)
}

func outlines(loops ...[]vector2.Float64) nodes.ConstOutput[[][]vector2.Float64] {
	return nodes.ConstOutput[[][]vector2.Float64]{Val: loops}
}

func outline(loop []vector2.Float64) nodes.ConstOutput[[]vector2.Float64] {
	return nodes.ConstOutput[[]vector2.Float64]{Val: loop}
}

func TestFillNodeMergesOutlinesFromSeparateConnections(t *testing.T) {
	area := nodes.GetNodeOutputPort[sketch.Area](&nodes.Struct[sketch.FillNode]{Data: sketch.FillNode{
		Outlines: []nodes.LiftedPort[[]vector2.Float64]{
			outline(squareAt(2, vector2.New(0., 0.))),
			outline(squareAt(2, vector2.New(1., 0.))),
		},
	}}, "Out").Value()

	assert.Len(t, area.Polygons, 1)
	assert.InDelta(t, 6., area.Size(), 1e-9)
}

func TestFillNodeGivesOneAreaPerOutlineInAList(t *testing.T) {
	areas := nodes.GetNodeOutputPort[[]sketch.Area](&nodes.Struct[sketch.FillNode]{Data: sketch.FillNode{
		Outlines: []nodes.LiftedPort[[]vector2.Float64]{
			outlines(squareAt(2, vector2.New(0., 0.)), squareAt(2, vector2.New(1., 0.)), squareAt(4, vector2.New(10., 0.))),
		},
	}}, "Out").Value()

	require.Len(t, areas, 3)
	assert.InDelta(t, 4., areas[0].Size(), 1e-9)
	assert.InDelta(t, 4., areas[1].Size(), 1e-9)
	assert.InDelta(t, 16., areas[2].Size(), 1e-9)
}

func TestFillNodeCutsASingleHoleFromEveryAreaInAList(t *testing.T) {
	areas := nodes.GetNodeOutputPort[[]sketch.Area](&nodes.Struct[sketch.FillNode]{Data: sketch.FillNode{
		Outlines: []nodes.LiftedPort[[]vector2.Float64]{
			outlines(squareAt(4, vector2.New(0., 0.)), squareAt(6, vector2.New(0., 0.))),
		},
		Holes: []nodes.LiftedPort[[]vector2.Float64]{outline(squareAt(2, vector2.New(0., 0.)))},
	}}, "Out").Value()

	require.Len(t, areas, 2)
	assert.InDelta(t, 12., areas[0].Size(), 1e-9)
	assert.InDelta(t, 32., areas[1].Size(), 1e-9)
}

func TestFillNodePairsAListOfHolesWithAListOfOutlines(t *testing.T) {
	areas := nodes.GetNodeOutputPort[[]sketch.Area](&nodes.Struct[sketch.FillNode]{Data: sketch.FillNode{
		Outlines: []nodes.LiftedPort[[]vector2.Float64]{
			outlines(squareAt(4, vector2.New(0., 0.)), squareAt(4, vector2.New(10., 0.))),
		},
		Holes: []nodes.LiftedPort[[]vector2.Float64]{
			outlines(squareAt(2, vector2.New(0., 0.)), squareAt(1, vector2.New(10., 0.))),
		},
	}}, "Out").Value()

	require.Len(t, areas, 2)
	assert.InDelta(t, 12., areas[0].Size(), 1e-9)
	assert.InDelta(t, 15., areas[1].Size(), 1e-9)
}
