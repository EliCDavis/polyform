package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestALiftedChainInsideASubgraphClonesWithItsRank(t *testing.T) {
	inst := testInstanceWithLiftedNodes(t)
	require.NoError(t, inst.CreateSubGraph("mottle", "Mottle", ""))

	child, err := inst.SubGraphInstance("mottle")
	require.NoError(t, err)

	_, source, err := child.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, double, err := child.CreateNode(liftedDoubleType)
	require.NoError(t, err)
	_, sum, err := child.CreateNode(floatArraySumType)
	require.NoError(t, err)

	child.ConnectNodes(source, "Out", double, "In")
	child.ConnectNodes(double, "Out", sum, "In")
	require.InDelta(t, 12, nodes.GetNodeOutputPort[float64](child.Node(sum), "Out").Value(), 1e-12)

	typePath, err := inst.RegisterSubGraphNodeType("mottle")
	require.NoError(t, err)

	var placed string
	require.NotPanics(t, func() {
		_, placed, err = inst.CreateNode(typePath)
	})
	require.NoError(t, err)
	assert.NotEmpty(t, placed)
}

func TestALiftedChainInsideASubgraphSurvivesASaveAndLoad(t *testing.T) {
	inst := testInstanceWithLiftedNodes(t)
	require.NoError(t, inst.CreateSubGraph("mottle", "Mottle", ""))

	child, err := inst.SubGraphInstance("mottle")
	require.NoError(t, err)

	_, source, err := child.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, double, err := child.CreateNode(liftedDoubleType)
	require.NoError(t, err)
	_, sum, err := child.CreateNode(floatArraySumType)
	require.NoError(t, err)

	child.ConnectNodes(source, "Out", double, "In")
	child.ConnectNodes(double, "Out", sum, "In")

	saved, err := inst.EncodeToAppSchema()
	require.NoError(t, err)

	for range 25 {
		reloaded := testInstanceWithLiftedNodes(t)
		require.NotPanics(t, func() {
			require.NoError(t, reloaded.ApplyAppSchema(saved))
		})

		back, err := reloaded.SubGraphInstance("mottle")
		require.NoError(t, err)
		assert.InDelta(t, 12, nodes.GetNodeOutputPort[float64](back.Node(sum), "Out").Value(), 1e-12)
	}
}

func TestConvertingALiftedChainIntoASubgraph(t *testing.T) {
	inst := testInstanceWithLiftedNodes(t)

	_, source, err := inst.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, double, err := inst.CreateNode(liftedDoubleType)
	require.NoError(t, err)
	_, sum, err := inst.CreateNode(floatArraySumType)
	require.NoError(t, err)

	inst.ConnectNodes(source, "Out", double, "In")
	inst.ConnectNodes(double, "Out", sum, "In")
	require.InDelta(t, 12, nodes.GetNodeOutputPort[float64](inst.Node(sum), "Out").Value(), 1e-12)

	var result graph.ConvertSelectionResult
	require.NotPanics(t, func() {
		result, err = inst.ConvertSelectionToSubGraph(graph.RootScope, []string{source, double, sum}, "Mottle", "")
	})
	require.NoError(t, err)
	assert.NotEmpty(t, result.SubGraphID)
}

func TestAMixedScalarAndArrayLiftedListSurvivesASaveAndLoad(t *testing.T) {
	inst := testInstanceWithLiftedNodes(t)

	_, scalar, err := inst.CreateNode(floatSourceType)
	require.NoError(t, err)
	_, array, err := inst.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, total, err := inst.CreateNode(liftedTotalType)
	require.NoError(t, err)
	_, sum, err := inst.CreateNode(floatArraySumType)
	require.NoError(t, err)

	inst.ConnectNodes(scalar, "Out", total, "In.0")
	inst.ConnectNodes(array, "Out", total, "In.1")
	inst.ConnectNodes(total, "Out", sum, "In")

	saved, err := inst.EncodeToAppSchema()
	require.NoError(t, err)

	for range 25 {
		reloaded := testInstanceWithLiftedNodes(t)
		require.NotPanics(t, func() {
			require.NoError(t, reloaded.ApplyAppSchema(saved))
		})
	}
}

func TestConvertingAMixedScalarAndArrayLiftedListIntoASubgraph(t *testing.T) {
	inst := testInstanceWithLiftedNodes(t)

	_, scalar, err := inst.CreateNode(floatSourceType)
	require.NoError(t, err)
	_, array, err := inst.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, total, err := inst.CreateNode(liftedTotalType)
	require.NoError(t, err)
	_, sum, err := inst.CreateNode(floatArraySumType)
	require.NoError(t, err)

	inst.ConnectNodes(scalar, "Out", total, "In.0")
	inst.ConnectNodes(array, "Out", total, "In.1")
	inst.ConnectNodes(total, "Out", sum, "In")

	require.NotPanics(t, func() {
		_, err = inst.ConvertSelectionToSubGraph(graph.RootScope,
			[]string{scalar, array, total, sum}, "Mottle", "")
	})
	require.NoError(t, err)
}

func TestConvertingALiftedChainOutOfAClonedSubgraph(t *testing.T) {
	inst := testInstanceWithLiftedNodes(t)
	require.NoError(t, inst.CreateSubGraph("outer", "Outer", ""))

	child, err := inst.SubGraphInstance("outer")
	require.NoError(t, err)

	_, source, err := child.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, double, err := child.CreateNode(liftedDoubleType)
	require.NoError(t, err)
	_, sum, err := child.CreateNode(floatArraySumType)
	require.NoError(t, err)
	child.ConnectNodes(source, "Out", double, "In")
	child.ConnectNodes(double, "Out", sum, "In")

	typePath, err := inst.RegisterSubGraphNodeType("outer")
	require.NoError(t, err)
	_, _, err = inst.CreateNode(typePath)
	require.NoError(t, err)

	var result graph.ConvertSelectionResult
	require.NotPanics(t, func() {
		result, err = inst.ConvertSelectionToSubGraph(
			graph.SubGraphScope("outer"), []string{source, double, sum}, "Mottle", "")
	})
	require.NoError(t, err)
	assert.NotEmpty(t, result.SubGraphID)

	nested, err := inst.SubGraphInstance(result.SubGraphID)
	require.NoError(t, err)
	assert.InDelta(t, 12, nodes.GetNodeOutputPort[float64](nested.Node(sum), "Out").Value(), 1e-12,
		"the moved chain still computes over the array")
}
