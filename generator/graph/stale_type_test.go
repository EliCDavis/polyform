package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testInstanceWithPickAndLiftedNodes(t *testing.T) *graph.Instance {
	t.Helper()
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[floatArraySourceNode]](factory)
	refutil.RegisterType[nodes.Struct[liftedDoubleNode]](factory)
	refutil.RegisterType[nodes.Struct[floatArraySumNode]](factory)
	refutil.RegisterType[nodes.Struct[pickNode]](factory)

	instance := graph.New(graph.Config{TypeFactory: factory})
	nodes.DiscoverPortTypes(factory)
	return instance
}

// A lifted input holds a typed wrapper around a dynamic output; once that output
// loses its type, the wrapper must not keep typing the chain through it.
func TestUnbindingADynamicProducerUnliftsWhatReadsItAndTheGraphReloads(t *testing.T) {
	inst := testInstanceWithPickAndLiftedNodes(t)
	_, source, err := inst.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, pick, err := inst.CreateNode(pickType)
	require.NoError(t, err)
	_, double, err := inst.CreateNode(liftedDoubleType)
	require.NoError(t, err)
	_, sum, err := inst.CreateNode(floatArraySumType)
	require.NoError(t, err)
	require.NoError(t, inst.ConnectNodes(source, "Out", pick, "A"))
	require.NoError(t, inst.ConnectNodes(pick, "Out", double, "In"))
	require.NoError(t, inst.ConnectNodes(double, "Out", sum, "In"))
	require.Equal(t, "[]float64", inst.Node(double).Outputs()["Out"].(nodes.Typed).Type())

	dropped, err := inst.DeleteNodeInputConnection(pick, "A")
	require.NoError(t, err)

	require.Len(t, dropped, 1, "%+v", dropped)
	assert.Equal(t, sum, dropped[0].To)
	assert.Empty(t, dynamicTypesOf(t, inst, pick), "nothing types the pick any more")
	assert.Equal(t, "float64", inst.Node(double).Outputs()["Out"].(nodes.Typed).Type())

	saved, err := inst.EncodeToAppSchema()
	require.NoError(t, err)
	require.NoError(t, testInstanceWithPickAndLiftedNodes(t).ApplyAppSchema(saved), "what the live graph holds is what loads")
}
