package opearations_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/opearations"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// One registered node now covers every element type, so the node has to keep
// working for a type its code never mentions.
func TestSeperateNodeCarriesAnyType(t *testing.T) {
	node := &nodes.Struct[opearations.SeperateNode]{}

	require.NoError(t, node.Inputs()["Array"].(nodes.SingleValueInputPort).
		Set(nodes.ConstOutput[[]trs.TRS]{Val: []trs.TRS{
			trs.Position(vector3.New(1., 0., 0.)),
			trs.Position(vector3.New(2., 0., 0.)),
			trs.Position(vector3.New(3., 0., 0.)),
		}}))
	require.NoError(t, node.Inputs()["Selection"].(nodes.SingleValueInputPort).
		Set(nodes.ConstOutput[[]bool]{Val: []bool{true, false, true}}))

	assert.Equal(t, "[]github.com/EliCDavis/polyform/math/trs.TRS",
		node.Outputs()["Selected"].(nodes.Typed).Type())

	selected := nodes.GetNodeOutputPort[[]trs.TRS](node, "Selected").Value()
	removed := nodes.GetNodeOutputPort[[]trs.TRS](node, "Removed").Value()

	require.Len(t, selected, 2)
	assert.Equal(t, 1., selected[0].Position().X())
	assert.Equal(t, 3., selected[1].Position().X())

	require.Len(t, removed, 1)
	assert.Equal(t, 2., removed[0].Position().X())
}

// The node's ordering has to match the function it replaced, reversed tail
// and all.
func TestSeperateNodeMatchesTheFunction(t *testing.T) {
	in := []int{1, 2, 3, 4}
	keep := []bool{true, false, true, false}

	node := &nodes.Struct[opearations.SeperateNode]{}
	require.NoError(t, node.Inputs()["Array"].(nodes.SingleValueInputPort).
		Set(nodes.ConstOutput[[]int]{Val: in}))
	require.NoError(t, node.Inputs()["Selection"].(nodes.SingleValueInputPort).
		Set(nodes.ConstOutput[[]bool]{Val: keep}))

	wantKept, wantRemoved := opearations.Seperate(in, keep)
	assert.Equal(t, wantKept, nodes.GetNodeOutputPort[[]int](node, "Selected").Value())
	assert.Equal(t, wantRemoved, nodes.GetNodeOutputPort[[]int](node, "Removed").Value())
}

func TestSeperate(t *testing.T) {

	tests := map[string]struct {
		in            []int
		keep          []bool
		keptResult    []int
		removedResult []int
	}{
		"basic": {
			in:            []int{1, 2, 3, 4},
			keep:          []bool{true, false, true, false},
			keptResult:    []int{1, 3},
			removedResult: []int{4, 2},
		},
		"lacking keep": {
			in:            []int{1, 2, 3, 4},
			keep:          []bool{true, false},
			keptResult:    []int{1},
			removedResult: []int{4, 3, 2},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			kept, removed := opearations.Seperate(tc.in, tc.keep)
			if assert.Len(t, kept, len(tc.keptResult)) {
				for i, v := range kept {
					assert.Equal(t, tc.keptResult[i], v)
				}
			}

			if assert.Len(t, removed, len(tc.removedResult)) {
				for i, v := range removed {
					assert.Equal(t, tc.removedResult[i], v)
				}
			}
		})
	}

}
