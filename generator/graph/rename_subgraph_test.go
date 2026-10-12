package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Result = sum of A and B, with A fed from a Float64 literal at root.
func placedAdder(t *testing.T, inst *graph.Instance, id string) (placed, consumer string) {
	t.Helper()
	require.NoError(t, inst.CreateSubGraph(id, "Adder", ""))
	child, err := inst.SubGraphInstance(id)
	require.NoError(t, err)
	_, a, err := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, child.SetBoundaryNodeInfo(a, "A"))
	_, b, err := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, child.SetBoundaryNodeInfo(b, "B"))
	_, sum, err := child.CreateNode("Sum")
	require.NoError(t, err)
	require.NoError(t, child.ConnectNodes(a, "Value", sum, "Values"))
	require.NoError(t, child.ConnectNodes(b, "Value", sum, "Values"))
	_, result, err := child.CreateBoundaryNode(subgraph.OutputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, child.SetBoundaryNodeInfo(result, "Result"))
	require.NoError(t, child.ConnectNodes(sum, "Float", result, "Value"))

	_, placed, err = inst.CreateNode(subgraph.RuntimeTypePath(id))
	require.NoError(t, err)
	_, literal, err := inst.CreateNode("Float64")
	require.NoError(t, err)
	require.NoError(t, inst.ConnectNodes(literal, "Value", placed, "A"))
	_, consumer, err = inst.CreateNode("Sum")
	require.NoError(t, err)
	require.NoError(t, inst.ConnectNodes(placed, "Result", consumer, "Values"))
	return placed, consumer
}

func TestRenameSubGraphFollowsInstancesAndSurvivesAReload(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)
	placed, consumer := placedAdder(t, inst, "adder")
	require.InDelta(t, 2, nodes.GetNodeOutputPort[float64](inst.Node(consumer), "Float").Value(), 1e-9)

	require.NoError(t, inst.RenameSubGraph("adder", "plus"))

	runtime := inst.Node(placed).(*graph.SubgraphInstanceNode)
	assert.Equal(t, "plus", runtime.SubGraphID())
	sch := inst.Schema()
	assert.Contains(t, sch.SubGraphs, "plus")
	assert.NotContains(t, sch.SubGraphs, "adder")
	assert.Equal(t, subgraph.RuntimeTypePath("plus"), sch.Nodes[placed].Type)
	assert.InDelta(t, 2, nodes.GetNodeOutputPort[float64](inst.Node(consumer), "Float").Value(), 1e-9)

	_, _, err := inst.CreateNode(subgraph.RuntimeTypePath("plus"))
	require.NoError(t, err)
	_, _, err = inst.CreateNode(subgraph.RuntimeTypePath("adder"))
	assert.Error(t, err)

	saved, err := inst.EncodeToAppSchema()
	require.NoError(t, err)
	reloaded := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, reloaded.ApplyAppSchema(saved))
	assert.Contains(t, reloaded.Schema().SubGraphs, "plus")
	assert.InDelta(t, 2, nodes.GetNodeOutputPort[float64](reloaded.Node(consumer), "Float").Value(), 1e-9)
}

func TestRenameSubGraphRefusesMissingAndTakenIds(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, inst.CreateSubGraph("a", "A", ""))
	require.NoError(t, inst.CreateSubGraph("b", "B", ""))
	assert.Error(t, inst.RenameSubGraph("missing", "c"))
	assert.Error(t, inst.RenameSubGraph("a", "b"))
	assert.Error(t, inst.RenameSubGraph("a", " "))
	assert.NoError(t, inst.RenameSubGraph("a", "a"))
}

func TestReplaceSubgraphInstanceKeepsTheIdAndMatchingEdges(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)
	placed, consumer := placedAdder(t, inst, "adder")

	require.NoError(t, inst.CreateSubGraph("negator", "Negator", ""))
	child, err := inst.SubGraphInstance("negator")
	require.NoError(t, err)
	_, a, err := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, child.SetBoundaryNodeInfo(a, "A"))
	_, result, err := child.CreateBoundaryNode(subgraph.OutputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, child.SetBoundaryNodeInfo(result, "Result"))
	_, sum, err := child.CreateNode("Sum")
	require.NoError(t, err)
	require.NoError(t, child.ConnectNodes(a, "Value", sum, "Values"))
	require.NoError(t, child.ConnectNodes(a, "Value", sum, "Values"))
	require.NoError(t, child.ConnectNodes(sum, "Float", result, "Value"))

	replaced, err := inst.ReplaceSubgraphInstance(placed, "negator")
	require.NoError(t, err)
	assert.Equal(t, []string{"A"}, replaced.KeptInputs)
	assert.Equal(t, []string{"Result -> " + consumer + ".Values.0"}, replaced.KeptOutputs)
	assert.Empty(t, replaced.LostInputs, "B was never wired, so nothing was lost")
	assert.Empty(t, replaced.LostOutputs)

	runtime := inst.Node(placed).(*graph.SubgraphInstanceNode)
	assert.Equal(t, "negator", runtime.SubGraphID())
	assert.InDelta(t, 4, nodes.GetNodeOutputPort[float64](inst.Node(consumer), "Float").Value(), 1e-9, "A + A with A = 2")

	saved, err := inst.EncodeToAppSchema()
	require.NoError(t, err)
	reloaded := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, reloaded.ApplyAppSchema(saved))
	assert.InDelta(t, 4, nodes.GetNodeOutputPort[float64](reloaded.Node(consumer), "Float").Value(), 1e-9)
}

func TestReplaceSubgraphInstanceRefusesNonInstances(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)
	placed, _ := placedAdder(t, inst, "adder")
	_, literal, err := inst.CreateNode("Float64")
	require.NoError(t, err)
	assert.Error(t, errOf(inst.ReplaceSubgraphInstance(literal, "adder")))
	assert.Error(t, errOf(inst.ReplaceSubgraphInstance(placed, "adder")), "already an instance of it")
	assert.Error(t, errOf(inst.ReplaceSubgraphInstance(placed, "missing")))
}
