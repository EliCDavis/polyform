package repeat_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformationNodeStacksEachSampleCountSeparately(t *testing.T) {
	floors := nodes.GetNodeOutputPort[[][]trs.TRS](&nodes.Struct[repeat.TransformationNode]{Data: repeat.TransformationNode{
		Transformation: nodes.ConstOutput[trs.TRS]{Val: trs.Position(vector3.New(0., 3., 0.))},
		Samples:        nodes.ConstOutput[[]int]{Val: []int{2, 4}},
	}}, "Out").Value()

	require.Len(t, floors, 2)
	assert.Len(t, floors[0], 2)
	require.Len(t, floors[1], 4)
	assert.Equal(t, vector3.New(0., 12., 0.), floors[1][3].Position())
}

func TestTRSNodePairsEachInputListWithItsOwnTransforms(t *testing.T) {
	rings := [][]trs.TRS{
		{trs.Position(vector3.New(1., 0., 0.)), trs.Position(vector3.New(2., 0., 0.))},
		{trs.Position(vector3.New(5., 0., 0.))},
	}
	floors := [][]trs.TRS{
		{trs.Position(vector3.New(0., 1., 0.))},
		{trs.Position(vector3.New(0., 1., 0.)), trs.Position(vector3.New(0., 2., 0.)), trs.Position(vector3.New(0., 3., 0.))},
	}

	windows := nodes.GetNodeOutputPort[[][]trs.TRS](&nodes.Struct[repeat.TRSNode]{Data: repeat.TRSNode{
		Input:      nodes.ConstOutput[[][]trs.TRS]{Val: rings},
		Transforms: nodes.ConstOutput[[][]trs.TRS]{Val: floors},
	}}, "Out").Value()

	require.Len(t, windows, 2)
	assert.Len(t, windows[0], 2)
	require.Len(t, windows[1], 3)
	assert.Equal(t, vector3.New(5., 3., 0.), windows[1][2].Position())
}

func TestTRSNodePassesInputThroughWithoutTransforms(t *testing.T) {
	rings := [][]trs.TRS{{trs.Position(vector3.New(1., 0., 0.))}, {}}

	assert.Equal(t, rings, nodes.GetNodeOutputPort[[][]trs.TRS](&nodes.Struct[repeat.TRSNode]{Data: repeat.TRSNode{
		Input: nodes.ConstOutput[[][]trs.TRS]{Val: rings},
	}}, "Out").Value())
	assert.Empty(t, nodes.GetNodeOutputPort[[]trs.TRS](&nodes.Struct[repeat.TRSNode]{}, "Out").Value())
}

func TestTRSKeepsEveryCopyWhenOneIsScaledToNothing(t *testing.T) {
	parents := []trs.TRS{trs.Position(vector3.New(10., 0., 0.))}
	slots := []trs.TRS{
		trs.Position(vector3.New(1., 0., 0.)),
		trs.New(vector3.New(2., 0., 0.), quaternion.Identity(), vector3.Zero[float64]()),
		trs.Position(vector3.New(3., 0., 0.)),
	}

	result, err := repeat.TRS(parents, slots)
	require.NoError(t, err)
	require.Len(t, result, 3)
	assert.Equal(t, vector3.New(11., 0., 0.), result[0].Position())
	assert.Equal(t, vector3.New(12., 0., 0.), result[1].Position())
	assert.Equal(t, vector3.Zero[float64](), result[1].Scale(), "the hidden slot stays hidden")
	assert.Equal(t, vector3.New(13., 0., 0.), result[2].Position())
}
