package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testInstanceWithEveryPortKind(t *testing.T) *graph.Instance {
	t.Helper()
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[floatArraySourceNode]](factory)
	refutil.RegisterType[nodes.Struct[liftedDoubleNode]](factory)
	refutil.RegisterType[nodes.Struct[floatArraySumNode]](factory)
	refutil.RegisterType[nodes.Struct[floatSourceNode]](factory)
	refutil.RegisterType[nodes.Struct[liftedTotalNode]](factory)
	refutil.RegisterType[nodes.Struct[floatConsumerNode]](factory)
	refutil.RegisterType[nodes.Struct[pickNode]](factory)
	refutil.RegisterType[nodes.Struct[vector3ConsumerNode]](factory)
	refutil.RegisterType[nodes.Struct[lastOfNode]](factory)
	refutil.RegisterType[parameter.Value[vector3.Float64]](factory)
	refutil.RegisterType[parameter.Value[bool]](factory)
	factory.RegisterBuilder(subgraph.InputNodeTypeKey, func() any { return subgraph.NewInputNode("", "") })
	factory.RegisterBuilder(subgraph.OutputNodeTypeKey, func() any { return subgraph.NewOutputNode("", "") })

	instance := graph.New(graph.Config{TypeFactory: factory})
	nodes.DiscoverPortTypes(factory)
	return instance
}

func TestConvertSelection_ArrayFedFromOutsideKeepsTheLiftedRankInside(t *testing.T) {
	inst := testInstanceWithEveryPortKind(t)

	_, source, err := inst.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, double, err := inst.CreateNode(liftedDoubleType)
	require.NoError(t, err)
	_, sum, err := inst.CreateNode(floatArraySumType)
	require.NoError(t, err)
	_, consumer, err := inst.CreateNode(floatConsumerType)
	require.NoError(t, err)

	require.NoError(t, inst.ConnectNodes(source, "Out", double, "In"))
	require.NoError(t, inst.ConnectNodes(double, "Out", sum, "In"))
	require.NoError(t, inst.ConnectNodes(sum, "Out", consumer, "In"))
	require.InDelta(t, 12, nodes.GetNodeOutputPort[float64](inst.Node(consumer), "Out").Value(), 1e-12)

	// The array source stays outside: inside, double only carries an array
	// because of what the boundary feeds it.
	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{double, sum}, "Doubled Sum", "")
	require.NoError(t, err)

	assert.InDelta(t, 12, nodes.GetNodeOutputPort[float64](inst.Node(consumer), "Out").Value(), 1e-12)

	nested, err := inst.SubGraphInstance(result.SubGraphID)
	require.NoError(t, err)
	assert.Equal(t, "[]float64", nested.Node(double).Outputs()["Out"].(nodes.Typed).Type(),
		"the definition itself is wired for an array")
}

func TestConvertSelection_MixedLiftedListFedFromOutside(t *testing.T) {
	inst := testInstanceWithEveryPortKind(t)

	_, scalar, err := inst.CreateNode(floatSourceType)
	require.NoError(t, err)
	_, array, err := inst.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, total, err := inst.CreateNode(liftedTotalType)
	require.NoError(t, err)
	_, sum, err := inst.CreateNode(floatArraySumType)
	require.NoError(t, err)
	_, consumer, err := inst.CreateNode(floatConsumerType)
	require.NoError(t, err)

	require.NoError(t, inst.ConnectNodes(scalar, "Out", total, "In.0"))
	require.NoError(t, inst.ConnectNodes(array, "Out", total, "In.1"))
	require.NoError(t, inst.ConnectNodes(total, "Out", sum, "In"))
	require.NoError(t, inst.ConnectNodes(sum, "Out", consumer, "In"))
	want := nodes.GetNodeOutputPort[float64](inst.Node(consumer), "Out").Value()
	require.InDelta(t, 12, want, 1e-12)

	_, err = inst.ConvertSelectionToSubGraph(graph.RootScope, []string{total, sum}, "Totals", "")
	require.NoError(t, err)

	assert.InDelta(t, want, nodes.GetNodeOutputPort[float64](inst.Node(consumer), "Out").Value(), 1e-12)
}

func TestConvertSelection_DynamicNodeTypedFromOutsideOnBothSides(t *testing.T) {
	inst := testInstanceWithEveryPortKind(t)

	_, param, err := inst.CreateNode(vector3Param)
	require.NoError(t, err)
	_, pick, err := inst.CreateNode(pickType)
	require.NoError(t, err)
	_, consumer, err := inst.CreateNode(consumerType)
	require.NoError(t, err)

	require.NoError(t, inst.ConnectNodes(param, "Value", pick, "A"))
	require.NoError(t, inst.ConnectNodes(param, "Value", pick, "B"))
	require.NoError(t, inst.ConnectNodes(pick, "Out", consumer, "In"))

	// Nothing inside the selection has a concrete type of its own.
	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{pick}, "Picker", "")
	require.NoError(t, err)

	nested, err := inst.SubGraphInstance(result.SubGraphID)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{pickVariable: vector3Type}, dynamicTypesOf(t, nested, pick))
	assigned := inst.Node(consumer).Inputs()["In"].(nodes.SingleValueInputPort)
	require.NotNil(t, assigned.Value())
	assert.Equal(t, vector3Type, assigned.Value().(nodes.Typed).Type())
}

func TestConvertSelection_DynamicChainTypedOnlyFromOutside(t *testing.T) {
	inst := testInstanceWithEveryPortKind(t)

	_, param, err := inst.CreateNode(vector3Param)
	require.NoError(t, err)
	_, first, err := inst.CreateNode(pickType)
	require.NoError(t, err)
	_, second, err := inst.CreateNode(pickType)
	require.NoError(t, err)
	_, consumer, err := inst.CreateNode(consumerType)
	require.NoError(t, err)

	require.NoError(t, inst.ConnectNodes(param, "Value", first, "A"))
	require.NoError(t, inst.ConnectNodes(first, "Out", second, "A"))
	require.NoError(t, inst.ConnectNodes(second, "Out", consumer, "In"))

	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{first, second}, "Chain", "")
	require.NoError(t, err)

	nested, err := inst.SubGraphInstance(result.SubGraphID)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{pickVariable: vector3Type}, dynamicTypesOf(t, nested, first))
	assert.Equal(t, map[string]string{pickVariable: vector3Type}, dynamicTypesOf(t, nested, second))
}

func TestConvertSelection_LiftedIntoArrayDynamicFedFromOutside(t *testing.T) {
	inst := testInstanceWithEveryPortKind(t)

	_, source, err := inst.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, double, err := inst.CreateNode(liftedDoubleType)
	require.NoError(t, err)
	_, last, err := inst.CreateNode(lastOfType)
	require.NoError(t, err)
	_, consumer, err := inst.CreateNode(floatConsumerType)
	require.NoError(t, err)

	require.NoError(t, inst.ConnectNodes(source, "Out", double, "In"))
	require.NoError(t, inst.ConnectNodes(double, "Out", last, "Array"))
	require.NoError(t, inst.ConnectNodes(last, "Out", consumer, "In"))
	require.InDelta(t, 6, nodes.GetNodeOutputPort[float64](inst.Node(consumer), "Out").Value(), 1e-12)

	_, err = inst.ConvertSelectionToSubGraph(graph.RootScope, []string{double, last}, "Last Doubled", "")
	require.NoError(t, err)

	assert.InDelta(t, 6, nodes.GetNodeOutputPort[float64](inst.Node(consumer), "Out").Value(), 1e-12)
}

func TestConvertSelection_SelectedVariableReferenceMovesIn(t *testing.T) {
	inst := testInstanceWithEveryPortKind(t)
	require.NoError(t, errOf(inst.NewVariable("Scale", &variable.TypeVariable[float64]{})))

	_, ref, err := inst.CreateNode("Scale")
	require.NoError(t, err)
	_, consumer, err := inst.CreateNode(floatConsumerType)
	require.NoError(t, err)
	_, outside, err := inst.CreateNode(floatConsumerType)
	require.NoError(t, err)
	require.NoError(t, inst.ConnectNodes(ref, "Value", consumer, "In"))
	require.NoError(t, inst.ConnectNodes(consumer, "Out", outside, "In"))

	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{ref, consumer}, "Scaled", "")
	require.NoError(t, err)

	assert.False(t, inst.HasNodeWithId(ref))
	nested, err := inst.SubGraphInstance(result.SubGraphID)
	require.NoError(t, err)
	_, isRef := nested.Node(ref).(variable.Reference)
	assert.True(t, isRef, "the reference lives inside now")

	runtime := inst.Node(result.RuntimeNodeID)
	assert.Empty(t, runtime.Inputs(), "a variable read inside is not a port")
}

func TestConvertSelection_SelectedVariableReferenceStillReadOutsideIsCopiedNotCut(t *testing.T) {
	inst := testInstanceWithEveryPortKind(t)
	require.NoError(t, errOf(inst.NewVariable("Scale", &variable.TypeVariable[float64]{})))

	_, ref, err := inst.CreateNode("Scale")
	require.NoError(t, err)
	_, inside, err := inst.CreateNode(floatConsumerType)
	require.NoError(t, err)
	_, outside, err := inst.CreateNode(floatConsumerType)
	require.NoError(t, err)
	require.NoError(t, inst.ConnectNodes(ref, "Value", inside, "In"))
	require.NoError(t, inst.ConnectNodes(ref, "Value", outside, "In"))

	result, err := inst.ConvertSelectionToSubGraph(graph.RootScope, []string{ref, inside}, "Scaled", "")
	require.NoError(t, err)

	assert.True(t, inst.HasNodeWithId(ref), "the outside reader keeps its reference")
	outsideIn := inst.Node(outside).Inputs()["In"].(nodes.SingleValueInputPort)
	require.NotNil(t, outsideIn.Value())
	assert.Equal(t, inst.Node(ref).Outputs()["Value"], outsideIn.Value())

	nested, err := inst.SubGraphInstance(result.SubGraphID)
	require.NoError(t, err)
	_, isRef := nested.Node(ref).(variable.Reference)
	assert.True(t, isRef)

	runtime := inst.Node(result.RuntimeNodeID)
	assert.Empty(t, runtime.Inputs())
	assert.Empty(t, runtime.Outputs(), "the outside reader was never cut, so nothing is exposed for it")
}
