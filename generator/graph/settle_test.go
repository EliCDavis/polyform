package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pickChain struct {
	instance                        *graph.Instance
	source, first, second, consumer string
}

// source -> first.B -> second.B -> consumer, wired in the order given.
func wirePickChain(t *testing.T, order []int) pickChain {
	t.Helper()
	instance := testInstanceWithDynamicNodes(t)
	_, source, err := instance.CreateNode(vector3Param)
	require.NoError(t, err)
	_, first, err := instance.CreateNode(pickType)
	require.NoError(t, err)
	_, second, err := instance.CreateNode(pickType)
	require.NoError(t, err)
	_, consumer, err := instance.CreateNode(consumerType)
	require.NoError(t, err)

	edges := [][4]string{
		{source, "Value", first, "B"},
		{first, "Out", second, "B"},
		{second, "Out", consumer, "In"},
	}
	for _, i := range order {
		require.NoError(t, instance.ConnectNodes(edges[i][0], edges[i][1], edges[i][2], edges[i][3]))
	}

	param := instance.Node(source).(interface{ ApplyMessage([]byte) (bool, error) })
	_, err = param.ApplyMessage([]byte(`{"x":1,"y":2,"z":3}`))
	require.NoError(t, err)

	return pickChain{instance: instance, source: source, first: first, second: second, consumer: consumer}
}

func (c pickChain) read() vector3.Float64 {
	return nodes.GetNodeOutputPort[vector3.Float64](c.instance.Node(c.consumer), "Out").Value()
}

func TestTheSameEdgesInAnyOrderSettleTheSame(t *testing.T) {
	orders := [][]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}}
	for _, order := range orders {
		chain := wirePickChain(t, order)
		assert.Equal(t, vector3.New(1., 2., 3.), chain.read(), "order %v", order)
		assert.Equal(t, map[string]string{pickVariable: vector3Type}, dynamicTypesOf(t, chain.instance, chain.first), "order %v", order)
		assert.Equal(t, map[string]string{pickVariable: vector3Type}, dynamicTypesOf(t, chain.instance, chain.second), "order %v", order)
	}
}

func TestAChainWiredInAnyOrderLoadsTheSame(t *testing.T) {
	chain := wirePickChain(t, []int{2, 1, 0})
	saved, err := chain.instance.EncodeToAppSchema()
	require.NoError(t, err)

	reloaded := testInstanceWithDynamicNodes(t)
	require.NoError(t, reloaded.ApplyAppSchema(saved))
	resaved, err := reloaded.EncodeToAppSchema()
	require.NoError(t, err)
	assert.JSONEq(t, string(saved), string(resaved))
}

func TestTakingTheArrayOffALiftedNodeDropsAReaderThatOnlyTakesArrays(t *testing.T) {
	instance := testInstanceWithLiftedNodes(t)
	_, source, err := instance.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, double, err := instance.CreateNode(liftedDoubleType)
	require.NoError(t, err)
	_, sum, err := instance.CreateNode(floatArraySumType)
	require.NoError(t, err)
	require.NoError(t, instance.ConnectNodes(source, "Out", double, "In"))
	require.NoError(t, instance.ConnectNodes(double, "Out", sum, "In"))

	dropped, err := instance.DeleteNodeInputConnection(double, "In")
	require.NoError(t, err)

	held := instance.Node(sum).Inputs()["In"].(nodes.SingleValueInputPort).Value()
	assert.Nil(t, held, "a scalar cannot feed an input that takes []float64")
	require.Len(t, dropped, 1)
	assert.Equal(t, graph.DroppedEdge{
		From: double, FromPort: "Out", To: sum, ToPort: "In", Reason: dropped[0].Reason,
	}, dropped[0])
	assert.Contains(t, dropped[0].Reason, "cannot take")

	saved, err := instance.EncodeToAppSchema()
	require.NoError(t, err)
	require.NoError(t, testInstanceWithLiftedNodes(t).ApplyAppSchema(saved))
}

func TestAnEditInsideASubgraphReportsTheEdgeItDropsInTheGraphPlacingIt(t *testing.T) {
	inst := testInstanceWithLiftedAndBoundaryTypes(t)
	typePath, boundaryIn, _ := liftedPassThroughSubgraph(t, inst, "doubler")
	_, placed, err := inst.CreateNode(typePath)
	require.NoError(t, err)
	_, source, err := inst.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, sum, err := inst.CreateNode(floatArraySumType)
	require.NoError(t, err)

	in, out := instancePorts(t, inst, placed)
	require.NoError(t, inst.ConnectNodes(source, "Out", placed, in))
	require.NoError(t, inst.ConnectNodes(placed, out, sum, "In"))

	child, err := inst.SubGraphInstance("doubler")
	require.NoError(t, err)
	var double string
	for _, id := range child.NodeIds() {
		if id != boundaryIn && child.Node(id).Inputs()["In"] != nil {
			double = id
		}
	}

	dropped, err := child.DeleteNodeInputConnection(double, "In")
	require.NoError(t, err)

	require.Len(t, dropped, 1, "%+v", dropped)
	assert.Equal(t, "", dropped[0].Scope, "the dropped edge is in the root graph, not the subgraph")
	assert.Equal(t, placed, dropped[0].From)
	assert.Equal(t, sum, dropped[0].To)
}
