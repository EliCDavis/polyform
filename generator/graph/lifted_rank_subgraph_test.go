package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testInstanceWithLiftedAndBoundaryTypes(t *testing.T) *graph.Instance {
	t.Helper()
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[floatArraySourceNode]](factory)
	refutil.RegisterType[nodes.Struct[liftedDoubleNode]](factory)
	refutil.RegisterType[nodes.Struct[floatArraySumNode]](factory)
	refutil.RegisterType[nodes.Struct[floatSourceNode]](factory)
	refutil.RegisterType[nodes.Struct[liftedTotalNode]](factory)
	refutil.RegisterType[nodes.Struct[floatConsumerNode]](factory)
	factory.RegisterBuilder(subgraph.InputNodeTypeKey, func() any {
		return subgraph.NewInputNode("", "")
	})
	factory.RegisterBuilder(subgraph.OutputNodeTypeKey, func() any {
		return subgraph.NewOutputNode("", "")
	})

	instance := graph.New(graph.Config{TypeFactory: factory})
	nodes.DiscoverPortTypes(factory)
	return instance
}

func onlyPort(t *testing.T, ports map[string]struct{}) string {
	t.Helper()
	require.Len(t, ports, 1)
	for name := range ports {
		return name
	}
	return ""
}

func instancePorts(t *testing.T, inst *graph.Instance, id string) (in, out string) {
	t.Helper()
	node := inst.Node(id)

	ins := map[string]struct{}{}
	for name := range node.Inputs() {
		ins[name] = struct{}{}
	}
	outs := map[string]struct{}{}
	for name := range node.Outputs() {
		outs[name] = struct{}{}
	}
	return onlyPort(t, ins), onlyPort(t, outs)
}

func liftedPassThroughSubgraph(t *testing.T, inst *graph.Instance, id string) (typePath, inPort, outPort string) {
	t.Helper()
	require.NoError(t, inst.CreateSubGraph(id, id, ""))

	child, err := inst.SubGraphInstance(id)
	require.NoError(t, err)

	_, boundaryIn, err := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	_, double, err := child.CreateNode(liftedDoubleType)
	require.NoError(t, err)
	_, boundaryOut, err := child.CreateBoundaryNode(subgraph.OutputNodeTypeKey, "float64")
	require.NoError(t, err)

	require.NoError(t, child.SetBoundaryNodeInfo(boundaryIn, "Value In"))
	require.NoError(t, child.SetBoundaryNodeInfo(boundaryOut, "Value Out"))
	child.ConnectNodes(boundaryIn, subgraph.ValuePortName, double, "In")
	child.ConnectNodes(double, "Out", boundaryOut, subgraph.ValuePortName)

	typePath, err = inst.RegisterSubGraphNodeType(id)
	require.NoError(t, err)
	return typePath, boundaryIn, boundaryOut
}

func TestAnArrayIntoASubgraphLiftsWhatIsInside(t *testing.T) {

	inst := testInstanceWithLiftedAndBoundaryTypes(t)
	typePath, _, _ := liftedPassThroughSubgraph(t, inst, "passthrough")

	_, placed, err := inst.CreateNode(typePath)
	require.NoError(t, err)
	_, array, err := inst.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, sum, err := inst.CreateNode(floatArraySumType)
	require.NoError(t, err)

	inName, outName := instancePorts(t, inst, placed)

	require.NotPanics(t, func() {
		inst.ConnectNodes(array, "Out", placed, inName)
	})
	require.NotPanics(t, func() {
		inst.ConnectNodes(placed, outName, sum, "In")
	})

	assert.InDelta(t, 12, nodes.GetNodeOutputPort[float64](inst.Node(sum), "Out").Value(), 1e-12,
		"1+2+3 doubled is 12, so the array lifted all the way through")
}

func TestASubgraphFedAnArraySurvivesASaveAndLoad(t *testing.T) {

	inst := testInstanceWithLiftedAndBoundaryTypes(t)
	typePath, _, _ := liftedPassThroughSubgraph(t, inst, "passthrough")

	_, placed, err := inst.CreateNode(typePath)
	require.NoError(t, err)
	_, array, err := inst.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, sum, err := inst.CreateNode(floatArraySumType)
	require.NoError(t, err)

	inName, outName := instancePorts(t, inst, placed)
	inst.ConnectNodes(array, "Out", placed, inName)
	inst.ConnectNodes(placed, outName, sum, "In")

	saved, err := inst.EncodeToAppSchema()
	require.NoError(t, err)

	for range 20 {
		reloaded := testInstanceWithLiftedAndBoundaryTypes(t)
		require.NotPanics(t, func() {
			require.NoError(t, reloaded.ApplyAppSchema(saved))
		})
		assert.InDelta(t, 12, nodes.GetNodeOutputPort[float64](reloaded.Node(sum), "Out").Value(), 1e-12)
	}
}

func TestASubgraphRefusesAnArrayConsumerUntilItCarriesOne(t *testing.T) {

	inst := testInstanceWithLiftedAndBoundaryTypes(t)
	typePath, _, _ := liftedPassThroughSubgraph(t, inst, "passthrough")

	_, placed, err := inst.CreateNode(typePath)
	require.NoError(t, err)
	_, array, err := inst.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, sum, err := inst.CreateNode(floatArraySumType)
	require.NoError(t, err)

	inName, outName := instancePorts(t, inst, placed)

	require.Panics(t, func() {
		inst.ConnectNodes(placed, outName, sum, "In")
	}, "the instance is still scalar, so a []float64 consumer does not fit yet")

	require.NotPanics(t, func() {
		inst.ConnectNodes(array, "Out", placed, inName)
	})
	require.NotPanics(t, func() {
		inst.ConnectNodes(placed, outName, sum, "In")
	})

	assert.InDelta(t, 12, nodes.GetNodeOutputPort[float64](inst.Node(sum), "Out").Value(), 1e-12)
}

func TestAnArrayLiftsThroughTwoNestedSubgraphs(t *testing.T) {

	inst := testInstanceWithLiftedAndBoundaryTypes(t)
	innerType, _, _ := liftedPassThroughSubgraph(t, inst, "inner")

	require.NoError(t, inst.CreateSubGraph("outer", "Outer", ""))
	outer, err := inst.SubGraphInstance("outer")
	require.NoError(t, err)

	_, outerIn, err := outer.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, outer.SetBoundaryNodeInfo(outerIn, "Value In"))
	require.NoError(t, err)
	_, nested, err := outer.CreateNode(innerType)
	require.NoError(t, err)
	_, outerOut, err := outer.CreateBoundaryNode(subgraph.OutputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, outer.SetBoundaryNodeInfo(outerOut, "Value Out"))
	require.NoError(t, err)

	nestedNode := outer.Node(nested)
	var nIn, nOut string
	for name := range nestedNode.Inputs() {
		nIn = name
	}
	for name := range nestedNode.Outputs() {
		nOut = name
	}
	outer.ConnectNodes(outerIn, subgraph.ValuePortName, nested, nIn)
	outer.ConnectNodes(nested, nOut, outerOut, subgraph.ValuePortName)

	outerType, err := inst.RegisterSubGraphNodeType("outer")
	require.NoError(t, err)

	_, placed, err := inst.CreateNode(outerType)
	require.NoError(t, err)
	_, array, err := inst.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, sum, err := inst.CreateNode(floatArraySumType)
	require.NoError(t, err)

	pIn, pOut := instancePorts(t, inst, placed)

	require.NotPanics(t, func() {
		inst.ConnectNodes(array, "Out", placed, pIn)
		inst.ConnectNodes(placed, pOut, sum, "In")
	})
	assert.InDelta(t, 12, nodes.GetNodeOutputPort[float64](inst.Node(sum), "Out").Value(), 1e-12)
}
