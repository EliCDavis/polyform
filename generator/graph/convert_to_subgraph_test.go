package graph_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertSelectionToSubGraph_SimpleChain(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)

	param, paramID, err := inst.CreateNode("Float64")
	require.NoError(t, err)
	_, sumID, err := inst.CreateNode("Sum")
	require.NoError(t, err)
	_, outSumID, err := inst.CreateNode("Sum")
	require.NoError(t, err)

	require.NoError(t, inst.ConnectNodes(paramID, "Value", sumID, "Values"))
	require.NoError(t, inst.ConnectNodes(sumID, "Float", outSumID, "Values"))

	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{sumID}, "Adder", "adds things")
	require.NoError(t, err)
	assert.Equal(t, "Adder", result.SubGraphID)
	assert.Equal(t, "Adder", result.Name)
	assert.NotEmpty(t, result.RuntimeNodeID)
	assert.Equal(t, subgraph.RuntimeTypePath("Adder"), result.NodeType.Type)

	assert.False(t, inst.HasNodeWithId(sumID))
	assert.True(t, inst.HasNodeWithId(paramID))
	assert.True(t, inst.HasNodeWithId(outSumID))
	assert.True(t, inst.HasNodeWithId(result.RuntimeNodeID))

	child, err := inst.SubGraphInstance("Adder")
	require.NoError(t, err)
	assert.True(t, child.HasNodeWithId(sumID))

	ports, err := inst.CollectBoundaryPorts("Adder")
	require.NoError(t, err)
	require.Len(t, ports, 2)

	names := map[string]graph.BoundaryPortKind{}
	for _, p := range ports {
		names[p.Name] = p.Kind
	}
	assert.Equal(t, graph.BoundaryPortKindInput, names["Value"])
	assert.Equal(t, graph.BoundaryPortKindOutput, names["Float"])

	runtime := inst.Node(result.RuntimeNodeID)
	require.Contains(t, runtime.Inputs(), "Value")
	require.Contains(t, runtime.Outputs(), "Float")

	runtimeIn := runtime.Inputs()["Value"].(nodes.SingleValueInputPort)
	assert.Equal(t, param.Outputs()["Value"], runtimeIn.Value())

	outIn := inst.Node(outSumID).Inputs()["Values"].(nodes.ArrayValueInputPort)
	require.Len(t, outIn.Value(), 1)
	assert.Equal(t, runtime.Outputs()["Float"], outIn.Value()[0])

	got := nodes.GetNodeOutputPort[float64](inst.Node(outSumID), "Float").Value()
	assert.Equal(t, 2.0, got) // Float64 default CurrentValue is 2 in test factory
}

func TestConvertSelectionToSubGraph_NamesPortsAfterWhatTheyCarry(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, errOf(inst.NewVariable("Board/Width", &variable.TypeVariable[float64]{})))

	_, widthID, err := inst.CreateNode("Board/Width")
	require.NoError(t, err)
	_, firstID, err := inst.CreateNode("Float64")
	require.NoError(t, err)
	_, secondID, err := inst.CreateNode("Float64")
	require.NoError(t, err)
	_, sumID, err := inst.CreateNode("Sum")
	require.NoError(t, err)
	_, totalID, err := inst.CreateNode("Sum")
	require.NoError(t, err)
	for _, from := range []string{widthID, firstID, secondID} {
		require.NoError(t, inst.ConnectNodes(from, "Value", sumID, "Values"))
	}
	require.NoError(t, inst.ConnectNodes(sumID, "Float", totalID, "Values"))

	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{sumID}, "Adder", "")
	require.NoError(t, err)

	runtime := inst.Node(result.RuntimeNodeID)
	assert.ElementsMatch(t, []string{"Width", "Value", "Value 2"}, slices.Collect(maps.Keys(runtime.Inputs())))
	assert.ElementsMatch(t, []string{"Float"}, slices.Collect(maps.Keys(runtime.Outputs())))
}

func TestConvertSelectionToSubGraph_NameAndDescription(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)

	_, sumID, err := inst.CreateNode("Sum")
	require.NoError(t, err)

	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{sumID}, "My Graph", "a description")
	require.NoError(t, err)
	assert.Equal(t, "My_Graph", result.SubGraphID)
	assert.Equal(t, "My Graph", result.Name)

	encoded, err := inst.EncodeToAppSchema()
	require.NoError(t, err)
	app, err := jbtf.Unmarshal[persistence.App](encoded)
	require.NoError(t, err)
	def, ok := app.SubGraphs["My_Graph"]
	require.True(t, ok)
	assert.Equal(t, "My Graph", def.Name)
	assert.Equal(t, "a description", def.Description)
}

func TestConvertSelectionToSubGraph_FanOut(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)

	_, paramID, err := inst.CreateNode("Float64")
	require.NoError(t, err)
	_, midID, err := inst.CreateNode("Sum")
	require.NoError(t, err)
	_, leftID, err := inst.CreateNode("Sum")
	require.NoError(t, err)
	_, rightID, err := inst.CreateNode("Sum")
	require.NoError(t, err)

	require.NoError(t, inst.ConnectNodes(paramID, "Value", midID, "Values"))
	require.NoError(t, inst.ConnectNodes(midID, "Float", leftID, "Values"))
	require.NoError(t, inst.ConnectNodes(midID, "Float", rightID, "Values"))

	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{midID}, "Fan", "")
	require.NoError(t, err)

	ports, err := inst.CollectBoundaryPorts(result.SubGraphID)
	require.NoError(t, err)
	require.Len(t, ports, 2) // one input, and one output shared by the fan-out

	runtime := inst.Node(result.RuntimeNodeID)
	leftIn := inst.Node(leftID).Inputs()["Values"].(nodes.ArrayValueInputPort)
	rightIn := inst.Node(rightID).Inputs()["Values"].(nodes.ArrayValueInputPort)
	require.Len(t, leftIn.Value(), 1)
	require.Len(t, rightIn.Value(), 1)
	assert.Equal(t, runtime.Outputs()["Float"], leftIn.Value()[0])
	assert.Equal(t, runtime.Outputs()["Float"], rightIn.Value()[0])
}

func TestConvertSelectionToSubGraph_FanInSharesOneInput(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)

	param, paramID, err := inst.CreateNode("Float64")
	require.NoError(t, err)
	_, leftID, err := inst.CreateNode("Sum")
	require.NoError(t, err)
	_, rightID, err := inst.CreateNode("Sum")
	require.NoError(t, err)

	require.NoError(t, inst.ConnectNodes(paramID, "Value", leftID, "Values"))
	require.NoError(t, inst.ConnectNodes(paramID, "Value", rightID, "Values"))

	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{leftID, rightID}, "Pair", "")
	require.NoError(t, err)

	ports, err := inst.CollectBoundaryPorts(result.SubGraphID)
	require.NoError(t, err)
	inputs := 0
	for _, p := range ports {
		if p.Kind == graph.BoundaryPortKindInput {
			inputs++
		}
	}
	assert.Equal(t, 1, inputs, "one outside source feeding two selected nodes is one port, not two")

	runtime := inst.Node(result.RuntimeNodeID)
	runtimeIn := runtime.Inputs()["Value"].(nodes.SingleValueInputPort)
	assert.Equal(t, param.Outputs()["Value"], runtimeIn.Value())

	child, err := inst.SubGraphInstance(result.SubGraphID)
	require.NoError(t, err)
	for _, id := range []string{leftID, rightID} {
		in := child.Node(id).Inputs()["Values"].(nodes.ArrayValueInputPort)
		require.Len(t, in.Value(), 1, id)
		_, isBoundary := in.Value()[0].Node().(*subgraph.InputNode)
		assert.True(t, isBoundary, "%s reads the shared boundary", id)
	}
}

func TestConvertSelectionToSubGraph_InternalEdgesPreserved(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)

	_, aID, err := inst.CreateNode("Float64")
	require.NoError(t, err)
	_, bID, err := inst.CreateNode("Sum")
	require.NoError(t, err)
	_, cID, err := inst.CreateNode("Sum")
	require.NoError(t, err)
	_, dID, err := inst.CreateNode("Sum")
	require.NoError(t, err)

	require.NoError(t, inst.ConnectNodes(aID, "Value", bID, "Values"))
	require.NoError(t, inst.ConnectNodes(bID, "Float", cID, "Values"))
	require.NoError(t, inst.ConnectNodes(cID, "Float", dID, "Values"))

	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{bID, cID}, "Inner", "")
	require.NoError(t, err)

	child, err := inst.SubGraphInstance(result.SubGraphID)
	require.NoError(t, err)
	assert.True(t, child.HasNodeWithId(bID))
	assert.True(t, child.HasNodeWithId(cID))

	cNode := child.Node(cID)
	cIn := cNode.Inputs()["Values"].(nodes.ArrayValueInputPort)
	require.Len(t, cIn.Value(), 1)
	assert.Equal(t, child.Node(bID).Outputs()["Float"], cIn.Value()[0])

	got := nodes.GetNodeOutputPort[float64](inst.Node(dID), "Float").Value()
	assert.Equal(t, 2.0, got)
}

func TestConvertSelectionToSubGraph_RejectsEmptyName(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)
	_, id, err := inst.CreateNode("Sum")
	require.NoError(t, err)

	_, err = inst.ConvertSelectionToSubGraph(graph.RootScope, []string{id}, "  ", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestConvertSelectionToSubGraph_RejectsMissingID(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)
	_, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{"Node-999"}, "X", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Node-999")
}

func TestConvertSelectionToSubGraph_RejectsBoundaryNode(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, inst.CreateSubGraph("existing", "Existing", ""))
	child, err := inst.SubGraphInstance("existing")
	require.NoError(t, err)

	_, inID, err := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, child.SetBoundaryNodeInfo(inID, "A"))

	_, err = inst.ConvertSelectionToSubGraph(graph.SubGraphScope("existing"), []string{inID}, "Nope", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boundary")
}

func TestConvertSelectionToSubGraph_ScopedParent(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, inst.CreateSubGraph("outer", "Outer", ""))
	child, err := inst.SubGraphInstance("outer")
	require.NoError(t, err)

	_, paramID, err := child.CreateNode("Float64")
	require.NoError(t, err)
	_, sumID, err := child.CreateNode("Sum")
	require.NoError(t, err)
	_, outID, err := child.CreateNode("Sum")
	require.NoError(t, err)
	require.NoError(t, child.ConnectNodes(paramID, "Value", sumID, "Values"))
	require.NoError(t, child.ConnectNodes(sumID, "Float", outID, "Values"))

	result, err := inst.ConvertSelectionToSubGraph(graph.SubGraphScope("outer"), []string{sumID}, "Nested", "")
	require.NoError(t, err)

	assert.False(t, child.HasNodeWithId(sumID))
	assert.True(t, child.HasNodeWithId(result.RuntimeNodeID))

	nested, err := inst.SubGraphInstance(result.SubGraphID)
	require.NoError(t, err)
	assert.True(t, nested.HasNodeWithId(sumID))

	got := nodes.GetNodeOutputPort[float64](child.Node(outID), "Float").Value()
	assert.Equal(t, 2.0, got)
}

func TestConvertSelectionToSubGraph_UniqueIDWhenTaken(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, inst.CreateSubGraph("Taken", "Taken", ""))

	_, id, err := inst.CreateNode("Sum")
	require.NoError(t, err)

	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{id}, "Taken", "")
	require.NoError(t, err)
	assert.Equal(t, "Taken_2", result.SubGraphID)
}

func TestConvertSelectionToSubGraph_RoundTrip(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)

	_, paramID, err := inst.CreateNode("Float64")
	require.NoError(t, err)
	_, sumID, err := inst.CreateNode("Sum")
	require.NoError(t, err)
	_, outID, err := inst.CreateNode("Sum")
	require.NoError(t, err)
	require.NoError(t, inst.ConnectNodes(paramID, "Value", sumID, "Values"))
	require.NoError(t, inst.ConnectNodes(sumID, "Float", outID, "Values"))

	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{sumID}, "Round", "trip")
	require.NoError(t, err)

	payload, err := inst.EncodeToAppSchema()
	require.NoError(t, err)

	restored := testInstanceWithSubGraphTypesExtended(t)
	require.NoError(t, restored.ApplyAppSchema(payload))

	assert.True(t, restored.HasNodeWithId(result.RuntimeNodeID))
	child, err := restored.SubGraphInstance(result.SubGraphID)
	require.NoError(t, err)
	assert.True(t, child.HasNodeWithId(sumID))

	got := nodes.GetNodeOutputPort[float64](restored.Node(outID), "Float").Value()
	assert.Equal(t, 2.0, got)
}

func TestConvertSelectionToSubGraph_RefusesACycleBeforeChangingAnything(t *testing.T) {
	inst := testInstanceWithSubGraphTypesExtended(t)

	// a -> mid -> b -> a's input: selecting {a, b} without mid would need
	// mid to both read from and feed the new instance.
	_, aID, err := inst.CreateNode("Sum")
	require.NoError(t, err)
	_, midID, err := inst.CreateNode("Sum")
	require.NoError(t, err)
	_, bID, err := inst.CreateNode("Sum")
	require.NoError(t, err)
	require.NoError(t, inst.ConnectNodes(aID, "Float", midID, "Values"))
	require.NoError(t, inst.ConnectNodes(midID, "Float", bID, "Values"))

	before := len(inst.NodeIds())
	subgraphsBefore := len(inst.Schema().SubGraphs)

	_, err = inst.ConvertSelectionToSubGraph(graph.RootScope, []string{aID, bID}, "Loop", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cycle")
	assert.Contains(t, err.Error(), midID)

	assert.Len(t, inst.NodeIds(), before, "a refused conversion creates no instance")
	assert.Len(t, inst.Schema().SubGraphs, subgraphsBefore, "and no definition")
	assert.True(t, inst.HasNodeWithId(aID))
	assert.True(t, inst.HasNodeWithId(bID))
	bIn := inst.Node(bID).Inputs()["Values"].(nodes.ArrayValueInputPort)
	require.Len(t, bIn.Value(), 1)
	assert.Equal(t, inst.Node(midID).Outputs()["Float"], bIn.Value()[0], "b still reads mid, not a dead instance")
}
