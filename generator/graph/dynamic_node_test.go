package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pickVar nodes.DynamicType

type pickNode struct {
	Condition nodes.Output[bool]
	A         nodes.DynamicPort[pickVar]
	B         nodes.DynamicPort[pickVar]
}

func (p pickNode) Out(out *nodes.Dynamic[pickVar]) {
	if nodes.TryGetOutputValue(out, p.Condition, false) {
		out.Forward(p.A)
		return
	}
	out.Forward(p.B)
}

type vector3ConsumerNode struct {
	In nodes.Output[vector3.Float64]
}

func (n vector3ConsumerNode) Out(out *nodes.StructOutput[vector3.Float64]) {
	out.Set(nodes.TryGetOutputValue(out, n.In, vector3.Zero[float64]()))
}

const (
	pickType     = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.pickNode]"
	consumerType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.vector3ConsumerNode]"
	vector3Param = "github.com/EliCDavis/polyform/generator/parameter.Value[github.com/EliCDavis/vector/vector3.Vector[float64]]"

	pickVariable = "github.com/EliCDavis/polyform/generator/graph_test.pickVar"
	vector3Type  = "github.com/EliCDavis/vector/vector3.Vector[float64]"
)

func testInstanceWithDynamicNodes(t *testing.T) *graph.Instance {
	t.Helper()
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[pickNode]](factory)
	refutil.RegisterType[nodes.Struct[vector3ConsumerNode]](factory)
	refutil.RegisterType[parameter.Value[vector3.Float64]](factory)

	instance := graph.New(graph.Config{TypeFactory: factory})
	nodes.DiscoverPortTypes(factory)
	return instance
}

func dynamicTypesOf(t *testing.T, instance *graph.Instance, nodeID string) map[string]string {
	t.Helper()
	dynamic, ok := instance.Node(nodeID).(nodes.DynamicallyTyped)
	require.True(t, ok)
	return dynamic.DynamicTypes()
}

func TestConnectNodesBindsADynamicPortFromTheUpstreamType(t *testing.T) {
	instance := testInstanceWithDynamicNodes(t)

	_, param, err := instance.CreateNode(vector3Param)
	require.NoError(t, err)
	_, pick, err := instance.CreateNode(pickType)
	require.NoError(t, err)

	instance.ConnectNodes(param, "Value", pick, "A")

	assert.Equal(t, map[string]string{pickVariable: vector3Type}, dynamicTypesOf(t, instance, pick))
}

// Nothing upstream is wired yet, so the only type available is the one the
// consumer wants. Without binding from that end the connection is refused.
func TestConnectNodesBindsADynamicPortFromTheDownstreamType(t *testing.T) {
	instance := testInstanceWithDynamicNodes(t)

	_, pick, err := instance.CreateNode(pickType)
	require.NoError(t, err)
	_, consumer, err := instance.CreateNode(consumerType)
	require.NoError(t, err)

	instance.ConnectNodes(pick, "Out", consumer, "In")

	assert.Equal(t, map[string]string{pickVariable: vector3Type}, dynamicTypesOf(t, instance, pick))

	assigned := instance.Node(consumer).Inputs()["In"].(nodes.SingleValueInputPort)
	require.NotNil(t, assigned.Value(), "the connection really landed")
}

func TestConnectNodesRefusesAMismatchOnABoundDynamicPort(t *testing.T) {
	instance := testInstanceWithDynamicNodes(t)

	_, param, err := instance.CreateNode(vector3Param)
	require.NoError(t, err)
	_, pick, err := instance.CreateNode(pickType)
	require.NoError(t, err)
	_, other, err := instance.CreateNode(consumerType)
	require.NoError(t, err)

	instance.ConnectNodes(param, "Value", pick, "A")

	// Out is a vector3 now, so it cannot also be whatever B is offered.
	require.Panics(t, func() {
		instance.ConnectNodes(other, "Out", pick, "Condition")
	})
}

func TestARefusedConnectionLeavesNoDynamicTypeBound(t *testing.T) {
	instance := testInstanceWithDynamicNodes(t)

	_, pick, err := instance.CreateNode(pickType)
	require.NoError(t, err)
	_, consumer, err := instance.CreateNode(consumerType)
	require.NoError(t, err)

	// In takes a single value, so an element index is nonsense - but the
	// type has already been offered to Out by the time anything says so.
	require.Panics(t, func() {
		instance.ConnectNodes(pick, "Out", consumer, "In.0")
	})

	assert.Empty(t, dynamicTypesOf(t, instance, pick),
		"no connection was made, so nothing may be holding a type")

	// Which is the point of releasing it: there is no connection to remove
	// to get the node back, so a leftover binding would be permanent.
	_, param, err := instance.CreateNode(vector3Param)
	require.NoError(t, err)
	require.NotPanics(t, func() {
		instance.ConnectNodes(param, "Value", pick, "A")
	})
}

func TestDisconnectingTheLastPortReleasesTheDynamicType(t *testing.T) {
	instance := testInstanceWithDynamicNodes(t)

	_, param, err := instance.CreateNode(vector3Param)
	require.NoError(t, err)
	_, pick, err := instance.CreateNode(pickType)
	require.NoError(t, err)

	instance.ConnectNodes(param, "Value", pick, "A")
	require.NotEmpty(t, dynamicTypesOf(t, instance, pick))

	instance.DeleteNodeInputConnection(pick, "A")
	assert.Empty(t, dynamicTypesOf(t, instance, pick), "nothing holds the type any more")

	// The whole point of releasing it: the node can be rewired rather than
	// thrown away.
	_, other, err := instance.CreateNode(consumerType)
	require.NoError(t, err)
	require.NotPanics(t, func() {
		instance.ConnectNodes(other, "Out", pick, "A")
	})
}

func TestDisconnectingOnePortKeepsATypeAnotherStillHolds(t *testing.T) {
	instance := testInstanceWithDynamicNodes(t)

	_, param, err := instance.CreateNode(vector3Param)
	require.NoError(t, err)
	_, pick, err := instance.CreateNode(pickType)
	require.NoError(t, err)

	instance.ConnectNodes(param, "Value", pick, "A")
	instance.ConnectNodes(param, "Value", pick, "B")

	instance.DeleteNodeInputConnection(pick, "A")
	assert.Equal(t, map[string]string{pickVariable: vector3Type}, dynamicTypesOf(t, instance, pick),
		"B is still carrying it")
}

func TestADownstreamConsumerHoldsTheDynamicTypeOnItsOwn(t *testing.T) {
	instance := testInstanceWithDynamicNodes(t)

	_, pick, err := instance.CreateNode(pickType)
	require.NoError(t, err)
	_, consumer, err := instance.CreateNode(consumerType)
	require.NoError(t, err)

	instance.ConnectNodes(pick, "Out", consumer, "In")
	require.NotEmpty(t, dynamicTypesOf(t, instance, pick))

	instance.DeleteNodeInputConnection(consumer, "In")
	assert.Empty(t, dynamicTypesOf(t, instance, pick),
		"the output was the only thing holding the type, and it is now unwired")
}

func TestDeletingTheUpstreamNodeReleasesTheDynamicType(t *testing.T) {
	instance := testInstanceWithDynamicNodes(t)

	_, param, err := instance.CreateNode(vector3Param)
	require.NoError(t, err)
	_, pick, err := instance.CreateNode(pickType)
	require.NoError(t, err)

	instance.ConnectNodes(param, "Value", pick, "A")
	require.NotEmpty(t, dynamicTypesOf(t, instance, pick))

	instance.DeleteNodeById(param)
	assert.Empty(t, dynamicTypesOf(t, instance, pick))
}

func TestDynamicTypeSurvivesASaveAndLoad(t *testing.T) {
	instance := testInstanceWithDynamicNodes(t)

	_, pick, err := instance.CreateNode(pickType)
	require.NoError(t, err)
	_, consumer, err := instance.CreateNode(consumerType)
	require.NoError(t, err)
	instance.ConnectNodes(pick, "Out", consumer, "In")

	saved, err := instance.EncodeToAppSchema()
	require.NoError(t, err)

	reloaded := testInstanceWithDynamicNodes(t)
	require.NoError(t, reloaded.ApplyAppSchema(saved))

	assert.Equal(t, map[string]string{pickVariable: vector3Type}, dynamicTypesOf(t, reloaded, pick),
		"a node only its output was wired from has nothing to infer the type back from")

	restored := reloaded.Node(consumer).Inputs()["In"].(nodes.SingleValueInputPort)
	assert.NotNil(t, restored.Value(), "so the connection has to come back too")
}

func TestDynamicTypeSurvivesASaveAndLoadThroughAChain(t *testing.T) {
	instance := testInstanceWithDynamicNodes(t)

	_, first, err := instance.CreateNode(pickType)
	require.NoError(t, err)
	_, second, err := instance.CreateNode(pickType)
	require.NoError(t, err)
	_, consumer, err := instance.CreateNode(consumerType)
	require.NoError(t, err)

	instance.ConnectNodes(second, "Out", consumer, "In")
	instance.ConnectNodes(first, "Out", second, "A")
	require.Equal(t, map[string]string{pickVariable: vector3Type}, dynamicTypesOf(t, instance, first))

	saved, err := instance.EncodeToAppSchema()
	require.NoError(t, err)

	reloaded := testInstanceWithDynamicNodes(t)
	require.NoError(t, reloaded.ApplyAppSchema(saved))

	assert.Equal(t, map[string]string{pickVariable: vector3Type}, dynamicTypesOf(t, reloaded, second))
	assert.Equal(t, map[string]string{pickVariable: vector3Type}, dynamicTypesOf(t, reloaded, first),
		"the only concrete type in the graph is two hops downstream")
}

func TestDynamicPortsAppearInTheNodeTypeSchema(t *testing.T) {
	instance := testInstanceWithDynamicNodes(t)

	var pickSchema = func() (found bool) {
		for _, nodeType := range instance.BuildSchemaForAllNodeTypes() {
			if nodeType.Type != pickType {
				continue
			}
			found = true

			out := nodeType.Outputs["Out"]
			assert.True(t, out.Dynamic, "the editor needs to know this port has no fixed type")
			assert.Equal(t, pickVariable, out.Type, "unbound, it shows the type variable")

			a := nodeType.Inputs["A"]
			assert.True(t, a.Dynamic)
			assert.Equal(t, pickVariable, a.Type)

			condition := nodeType.Inputs["Condition"]
			assert.False(t, condition.Dynamic, "a fixed port beside dynamic ones is unaffected")
			assert.Equal(t, "bool", condition.Type)
		}
		return
	}
	require.True(t, pickSchema(), "the node type was not in the catalog")
}
