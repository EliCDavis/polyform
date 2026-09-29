package logic_test

import (
	"testing"

	"github.com/EliCDavis/polyform/logic"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func wire(t *testing.T, node nodes.Node, port string, source nodes.OutputPort) {
	t.Helper()
	input, ok := node.Inputs()[port].(nodes.SingleValueInputPort)
	require.True(t, ok, "node has no single-value input %q", port)
	require.NoError(t, input.Set(source))
}

func TestSelectRoutesWithoutCopyingTheValue(t *testing.T) {
	cube := nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[primitives.CubeNode]{}, "Out")
	sphere := nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[primitives.UvSphereNode]{}, "Out")
	condition := nodes.NewValue(true)

	node := &nodes.Struct[logic.SelectNode]{}
	wire(t, node, "A", cube)
	wire(t, node, "B", sphere)
	wire(t, node, "Condition", condition.Outputs()["Value"])

	out := node.Outputs()["Out"]
	assert.Equal(t, "github.com/EliCDavis/polyform/modeling.Mesh", out.(nodes.Typed).Type())
	typed, ok := out.(nodes.Output[modeling.Mesh])
	require.True(t, ok, "got %T", out)

	assert.Equal(t, cube.Value().PrimitiveCount(), typed.Value().PrimitiveCount())

	condition.Set(false)
	assert.Equal(t, sphere.Value().PrimitiveCount(), typed.Value().PrimitiveCount())
}

func TestSelectTakesItsTypeFromWhicheverEndIsWiredFirst(t *testing.T) {
	node := &nodes.Struct[logic.SelectNode]{}
	assert.Empty(t, node.DynamicTypes())

	wire(t, node, "B", nodes.ConstOutput[vector3.Float64]{Val: vector3.One[float64]()})
	assert.Equal(t, map[string]string{
		"github.com/EliCDavis/polyform/logic.SelectType": "github.com/EliCDavis/vector/vector3.Vector[float64]",
	}, node.DynamicTypes())
	assert.Equal(t, "github.com/EliCDavis/vector/vector3.Vector[float64]",
		node.Inputs()["A"].(nodes.Typed).Type(), "the other input took the same type")
}

func TestSelectRefusesAMixOfTypes(t *testing.T) {
	node := &nodes.Struct[logic.SelectNode]{}
	wire(t, node, "A", nodes.ConstOutput[float64]{Val: 1})

	err := node.Inputs()["B"].(nodes.SingleValueInputPort).Set(nodes.ConstOutput[int]{Val: 2})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already carrying float64")
}

func TestSwitchPicksACaseAndFallsBackOutsideTheList(t *testing.T) {
	index := nodes.NewValue(1)

	node := &nodes.Struct[logic.SwitchNode]{}
	cases := node.Inputs()["Cases"].(nodes.ArrayValueInputPort)
	require.NoError(t, cases.Add(nodes.ConstOutput[string]{Val: "zero"}))
	require.NoError(t, cases.Add(nodes.ConstOutput[string]{Val: "one"}))
	wire(t, node, "Fallback", nodes.ConstOutput[string]{Val: "none"})
	wire(t, node, "Index", index.Outputs()["Value"])

	typed, ok := node.Outputs()["Out"].(nodes.Output[string])
	require.True(t, ok, "got %T", node.Outputs()["Out"])

	assert.Equal(t, "one", typed.Value())

	index.Set(0)
	assert.Equal(t, "zero", typed.Value())

	index.Set(7)
	assert.Equal(t, "none", typed.Value())

	index.Set(-1)
	assert.Equal(t, "none", typed.Value())
}

// Downstream caching keys on the branch actually selected, so the unselected
// branch changing must not invalidate anything.
func TestSelectDoesNotInvalidateOnTheUnselectedBranch(t *testing.T) {
	chosen := nodes.NewValue(1.)
	other := nodes.NewValue(100.)
	condition := nodes.NewValue(true)

	selectNode := &nodes.Struct[logic.SelectNode]{}
	wire(t, selectNode, "A", chosen.Outputs()["Value"])
	wire(t, selectNode, "B", other.Outputs()["Value"])
	wire(t, selectNode, "Condition", condition.Outputs()["Value"])

	runs := 0
	counter := &nodes.Struct[countingNode]{Data: countingNode{Runs: &runs}}
	wire(t, counter, "In", selectNode.Outputs()["Out"])
	value := nodes.GetNodeOutputPort[float64](counter, "Out")

	assert.InDelta(t, 1, value.Value(), 1e-9)
	assert.Equal(t, 1, runs)

	value.Value()
	assert.Equal(t, 1, runs, "nothing changed, so nothing re-ran")

	other.Set(200)
	value.Value()
	assert.Equal(t, 1, runs, "the branch that changed is not the one being read")

	chosen.Set(2)
	assert.InDelta(t, 2, value.Value(), 1e-9)
	assert.Equal(t, 2, runs, "the selected branch changed, so it re-ran")
}

type countingNode struct {
	In   nodes.Output[float64]
	Runs *int
}

func (c countingNode) Out(out *nodes.StructOutput[float64]) {
	*c.Runs++
	out.Set(nodes.TryGetOutputValue(out, c.In, 0))
}
