package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type lastElement nodes.DynamicType

type lastOfNode struct {
	Array nodes.DynamicPort[[]lastElement]
}

func (n lastOfNode) Out(out *nodes.Dynamic[lastElement]) {
	values, ok := nodes.DynamicArrayValue(n.Array)
	if !ok || values.Len() == 0 {
		return
	}
	out.Set(values.At(values.Len() - 1))
}

const lastOfType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.lastOfNode]"

func testInstanceWithLiftedAndDynamicNodes(t *testing.T) *graph.Instance {
	t.Helper()
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[floatArraySourceNode]](factory)
	refutil.RegisterType[nodes.Struct[liftedDoubleNode]](factory)
	refutil.RegisterType[nodes.Struct[lastOfNode]](factory)
	refutil.RegisterType[nodes.Struct[floatConsumerNode]](factory)

	instance := graph.New(graph.Config{TypeFactory: factory})
	nodes.DiscoverPortTypes(factory)
	return instance
}

func liftedIntoArrayDynamic(t *testing.T) (*graph.Instance, string) {
	t.Helper()
	instance := testInstanceWithLiftedAndDynamicNodes(t)

	_, source, err := instance.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, double, err := instance.CreateNode(liftedDoubleType)
	require.NoError(t, err)
	_, last, err := instance.CreateNode(lastOfType)
	require.NoError(t, err)

	require.NoError(t, instance.ConnectNodes(source, "Out", double, "In"))
	require.NoError(t, instance.ConnectNodes(double, "Out", last, "Array"))
	require.InDelta(t, 6, nodes.GetNodeOutputPort[float64](instance.Node(last), "Out").Value(), 1e-12)
	return instance, last
}

func TestALiftedOutputIntoAnArrayDynamicPortSurvivesASaveAndLoad(t *testing.T) {
	instance, last := liftedIntoArrayDynamic(t)

	saved, err := instance.EncodeToAppSchema()
	require.NoError(t, err)

	for range 20 {
		reloaded := testInstanceWithLiftedAndDynamicNodes(t)
		require.NoError(t, reloaded.ApplyAppSchema(saved))
		assert.InDelta(t, 6, nodes.GetNodeOutputPort[float64](reloaded.Node(last), "Out").Value(), 1e-12)
	}
}

func TestEditingASubgraphWithALiftedOutputIntoAnArrayDynamicPortRebuildsItsInstances(t *testing.T) {
	inst := testInstanceWithLiftedAndDynamicNodes(t)
	require.NoError(t, inst.CreateSubGraph("pads", "Pads", ""))
	child, err := inst.SubGraphInstance("pads")
	require.NoError(t, err)

	_, source, err := child.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, double, err := child.CreateNode(liftedDoubleType)
	require.NoError(t, err)
	_, last, err := child.CreateNode(lastOfType)
	require.NoError(t, err)
	require.NoError(t, child.ConnectNodes(source, "Out", double, "In"))
	require.NoError(t, child.ConnectNodes(double, "Out", last, "Array"))

	typePath, err := inst.RegisterSubGraphNodeType("pads")
	require.NoError(t, err)
	_, _, err = inst.CreateNode(typePath)
	require.NoError(t, err)

	require.NoError(t, inst.History().Transact("edit inside", func() error {
		_, _, err := child.CreateNode(floatArraySourceType)
		return err
	}))
}

func TestAFailedStepAfterALiftedOutputIntoAnArrayDynamicPortRollsBack(t *testing.T) {
	instance, last := liftedIntoArrayDynamic(t)

	err := instance.History().Transact("fails", func() error {
		_, _, err := instance.CreateNode(floatArraySourceType)
		require.NoError(t, err)
		return assert.AnError
	})
	require.ErrorIs(t, err, assert.AnError)
	assert.NotContains(t, err.Error(), "could not be rolled back")
	assert.InDelta(t, 6, nodes.GetNodeOutputPort[float64](instance.Node(last), "Out").Value(), 1e-12)
}
