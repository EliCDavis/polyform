package logic_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/logic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func maskErrors(node nodes.Node) []string {
	out := node.Outputs()["Out"]
	nodes.DynamicAnyValue(nil, out)
	return out.(nodes.ObservableExecution).ExecutionReport().Errors
}

func TestSelectWithAListOfConditionsPicksPerElement(t *testing.T) {
	node := &nodes.Struct[logic.SelectNode]{}
	wire(t, node, "Condition", nodes.ConstOutput[[]bool]{Val: []bool{true, false, true}})
	wire(t, node, "A", nodes.ConstOutput[[]int]{Val: []int{1, 2, 3}})
	wire(t, node, "B", nodes.ConstOutput[[]int]{Val: []int{-1, -1, -1}})

	out := node.Outputs()["Out"]
	assert.Equal(t, "[]int", out.(nodes.Typed).Type())
	assert.Equal(t, []int{1, -1, 3}, nodes.GetNodeOutputPort[[]int](node, "Out").Value())
	assert.Empty(t, maskErrors(node))
}

func TestSelectWithAListOfConditionsAndOneSideUnwiredUsesZeros(t *testing.T) {
	node := &nodes.Struct[logic.SelectNode]{}
	wire(t, node, "Condition", nodes.ConstOutput[[]bool]{Val: []bool{false, true}})
	wire(t, node, "A", nodes.ConstOutput[[]string]{Val: []string{"a", "b"}})

	assert.Equal(t, []string{"", "b"}, nodes.GetNodeOutputPort[[]string](node, "Out").Value())
}

func TestSelectWithAListOfConditionsBroadcastsAOneElementSide(t *testing.T) {
	node := &nodes.Struct[logic.SelectNode]{}
	wire(t, node, "Condition", nodes.ConstOutput[[]bool]{Val: []bool{true, false, false}})
	wire(t, node, "A", nodes.ConstOutput[[]int]{Val: []int{1, 2, 3}})
	wire(t, node, "B", nodes.ConstOutput[[]int]{Val: []int{-1}})

	assert.Equal(t, []int{1, -1, -1}, nodes.GetNodeOutputPort[[]int](node, "Out").Value())
	assert.Empty(t, maskErrors(node))
}

func TestSelectRefusesConditionsThatDoNotMatchTheLists(t *testing.T) {
	node := &nodes.Struct[logic.SelectNode]{}
	wire(t, node, "Condition", nodes.ConstOutput[[]bool]{Val: []bool{true, false}})
	wire(t, node, "A", nodes.ConstOutput[[]int]{Val: []int{1, 2, 3}})
	wire(t, node, "B", nodes.ConstOutput[[]int]{Val: []int{4, 5, 6}})

	require.NotEmpty(t, maskErrors(node))
}

func TestSelectWithASingleConditionStillForwardsWholeLists(t *testing.T) {
	node := &nodes.Struct[logic.SelectNode]{}
	wire(t, node, "Condition", nodes.ConstOutput[bool]{Val: false})
	wire(t, node, "A", nodes.ConstOutput[[]int]{Val: []int{1, 2}})
	wire(t, node, "B", nodes.ConstOutput[[]int]{Val: []int{7, 8, 9}})

	assert.Equal(t, []int{7, 8, 9}, nodes.GetNodeOutputPort[[]int](node, "Out").Value())
}
