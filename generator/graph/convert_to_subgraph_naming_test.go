package graph_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type labelledNode struct {
	Name nodes.Output[string]
	In   nodes.Output[float64]
}

func (n labelledNode) Doubled(out *nodes.StructOutput[float64]) {
	out.Set(nodes.TryGetOutputValue(out, n.In, 0) * 2)
}

func (n labelledNode) Halved(out *nodes.StructOutput[float64]) {
	out.Set(nodes.TryGetOutputValue(out, n.In, 0) / 2)
}

const (
	labelledType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.labelledNode]"
	floatParamType  = "github.com/EliCDavis/polyform/generator/parameter.Value[float64]"
	stringParamType = "github.com/EliCDavis/polyform/generator/parameter.Value[string]"
)

func testInstanceForPortNaming(t *testing.T) *graph.Instance {
	t.Helper()
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[labelledNode]](factory)
	refutil.RegisterType[nodes.Struct[floatConsumerNode]](factory)
	refutil.RegisterType[parameter.Value[float64]](factory)
	refutil.RegisterType[parameter.Value[string]](factory)
	factory.RegisterBuilder(subgraph.InputNodeTypeKey, func() any { return subgraph.NewInputNode("", "") })
	factory.RegisterBuilder(subgraph.OutputNodeTypeKey, func() any { return subgraph.NewOutputNode("", "") })

	instance := graph.New(graph.Config{TypeFactory: factory})
	nodes.DiscoverPortTypes(factory)
	return instance
}

type nodeScope interface {
	CreateNode(nodeType string) (nodes.Node, string, error)
	Node(id string) nodes.Node
}

func createNode(t *testing.T, scope nodeScope, nodeType string) (nodes.Node, string) {
	t.Helper()
	node, id, err := scope.CreateNode(nodeType)
	require.NoError(t, err)
	return node, id
}

func convertedPortNames(t *testing.T, scope nodeScope, result graph.ConvertSelectionResult) (inputs, outputs []string) {
	t.Helper()
	runtime := scope.Node(result.RuntimeNodeID)
	return slices.Sorted(maps.Keys(runtime.Inputs())), slices.Sorted(maps.Keys(runtime.Outputs()))
}

func TestConvertSelection_InputNamedAfterFeedingParameterLabel(t *testing.T) {
	inst := testInstanceForPortNaming(t)
	param, radius := createNode(t, inst, floatParamType)
	param.(graph.Parameter).SetName("Radius")
	_, consumer := createNode(t, inst, floatConsumerType)
	require.NoError(t, inst.ConnectNodes(radius, "Value", consumer, "In"))

	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{consumer}, "Part", "")
	require.NoError(t, err)

	inputs, _ := convertedPortNames(t, inst, result)
	assert.Equal(t, []string{"Radius"}, inputs)
}

func TestConvertSelection_InputNamedAfterFeedingOutputWhenItSaysSomething(t *testing.T) {
	inst := testInstanceForPortNaming(t)
	_, source := createNode(t, inst, labelledType)
	_, consumer := createNode(t, inst, floatConsumerType)
	require.NoError(t, inst.ConnectNodes(source, "Doubled", consumer, "In"))

	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{consumer}, "Part", "")
	require.NoError(t, err)

	inputs, _ := convertedPortNames(t, inst, result)
	assert.Equal(t, []string{"Doubled"}, inputs)
}

func TestConvertSelection_InputFallsBackToTheConsumingInput(t *testing.T) {
	inst := testInstanceForPortNaming(t)
	_, source := createNode(t, inst, floatConsumerType)
	_, consumer := createNode(t, inst, floatConsumerType)
	require.NoError(t, inst.ConnectNodes(source, "Out", consumer, "In"))

	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{consumer}, "Part", "")
	require.NoError(t, err)

	inputs, _ := convertedPortNames(t, inst, result)
	assert.Equal(t, []string{"In"}, inputs)
}

func TestConvertSelection_InputNamedAfterTheOuterPortItCarries(t *testing.T) {
	inst := testInstanceForPortNaming(t)
	require.NoError(t, inst.CreateSubGraph("outer", "Outer", ""))
	outer, err := inst.SubGraphInstance("outer")
	require.NoError(t, err)

	_, radius, err := outer.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, outer.SetBoundaryNodeInfo(radius, "Radius"))
	_, consumer := createNode(t, outer, floatConsumerType)
	require.NoError(t, outer.ConnectNodes(radius, subgraph.ValuePortName, consumer, "In"))

	result, err := inst.ConvertSelectionToSubGraph(graph.SubGraphScope("outer"), []string{consumer}, "Inner", "")
	require.NoError(t, err)

	inputs, _ := convertedPortNames(t, outer, result)
	assert.Equal(t, []string{"Radius"}, inputs)
}

func TestConvertSelection_OutputNamedAfterTheProducersNameLiteral(t *testing.T) {
	inst := testInstanceForPortNaming(t)
	_, label := createNode(t, inst, stringParamType)
	_, err := inst.UpdateParameter(label, []byte(`"Body"`))
	require.NoError(t, err)
	_, body := createNode(t, inst, labelledType)
	_, reader := createNode(t, inst, floatConsumerType)
	require.NoError(t, inst.ConnectNodes(label, "Value", body, "Name"))
	require.NoError(t, inst.ConnectNodes(body, "Doubled", reader, "In"))

	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{label, body}, "Part", "")
	require.NoError(t, err)

	inputs, outputs := convertedPortNames(t, inst, result)
	assert.Empty(t, inputs)
	assert.Equal(t, []string{"Body"}, outputs)
}

func TestConvertSelection_SeveralOutputsOfOneNamedNodeAreToldApart(t *testing.T) {
	inst := testInstanceForPortNaming(t)
	_, label := createNode(t, inst, stringParamType)
	_, err := inst.UpdateParameter(label, []byte(`"Body"`))
	require.NoError(t, err)
	_, body := createNode(t, inst, labelledType)
	_, first := createNode(t, inst, floatConsumerType)
	_, second := createNode(t, inst, floatConsumerType)
	require.NoError(t, inst.ConnectNodes(label, "Value", body, "Name"))
	require.NoError(t, inst.ConnectNodes(body, "Doubled", first, "In"))
	require.NoError(t, inst.ConnectNodes(body, "Halved", second, "In"))

	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{label, body}, "Part", "")
	require.NoError(t, err)

	_, outputs := convertedPortNames(t, inst, result)
	assert.Equal(t, []string{"Body Doubled", "Body Halved"}, outputs)
}

func TestConvertSelection_ClashingNamesAreSuffixed(t *testing.T) {
	inst := testInstanceForPortNaming(t)
	_, left := createNode(t, inst, labelledType)
	_, right := createNode(t, inst, labelledType)
	_, first := createNode(t, inst, floatConsumerType)
	_, second := createNode(t, inst, floatConsumerType)
	require.NoError(t, inst.ConnectNodes(left, "Doubled", first, "In"))
	require.NoError(t, inst.ConnectNodes(right, "Doubled", second, "In"))

	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{left, right}, "Pair", "")
	require.NoError(t, err)

	_, outputs := convertedPortNames(t, inst, result)
	assert.Equal(t, []string{"Doubled", "Doubled 2"}, outputs)
}

func TestRenameBoundaryPorts_NamesTradePlacesAndInstancesFollow(t *testing.T) {
	inst := testInstanceForPortNaming(t)
	require.NoError(t, inst.CreateSubGraph("part", "Part", ""))
	child, err := inst.SubGraphInstance("part")
	require.NoError(t, err)

	_, a, err := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, child.SetBoundaryNodeInfo(a, "A"))
	_, b, err := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, child.SetBoundaryNodeInfo(b, "B"))

	param, paramID := createNode(t, inst, floatParamType)
	_, placed := createNode(t, inst, subgraph.RuntimeTypePath("part"))
	require.NoError(t, inst.ConnectNodes(paramID, "Value", placed, "A"))

	require.NoError(t, child.RenameBoundaryPorts(map[string]string{a: "B", b: "A"}))

	assert.Equal(t, "B", child.Node(a).(*subgraph.InputNode).PortName)
	assert.Equal(t, "A", child.Node(b).(*subgraph.InputNode).PortName)
	fed := inst.Node(placed).Inputs()["B"].(nodes.SingleValueInputPort)
	assert.Equal(t, param.Outputs()["Value"], fed.Value(), "the wire follows the port it was on")
}

func TestRenameBoundaryPorts_RefusesAClashInTheFinalNames(t *testing.T) {
	inst := testInstanceForPortNaming(t)
	require.NoError(t, inst.CreateSubGraph("part", "Part", ""))
	child, err := inst.SubGraphInstance("part")
	require.NoError(t, err)

	_, a, err := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, child.SetBoundaryNodeInfo(a, "A"))
	_, b, err := child.CreateBoundaryNode(subgraph.InputNodeTypeKey, "float64")
	require.NoError(t, err)
	require.NoError(t, child.SetBoundaryNodeInfo(b, "B"))

	err = child.RenameBoundaryPorts(map[string]string{a: "B"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "would be used by both")
	assert.Equal(t, "A", child.Node(a).(*subgraph.InputNode).PortName)

	err = child.RenameBoundaryPorts(map[string]string{"Node-999": "C"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no node exists")
}
