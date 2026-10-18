package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/arrays"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const spreadType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/arrays.SpreadNode]"

func TestOnlyListInputsTakeASpread(t *testing.T) {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[floatArraySourceNode]](factory)
	refutil.RegisterType[nodes.Struct[floatConsumerNode]](factory)
	refutil.RegisterType[nodes.Struct[liftedDoubleNode]](factory)
	refutil.RegisterType[nodes.Struct[liftedTotalNode]](factory)
	refutil.RegisterType[nodes.Struct[arrays.SpreadNode]](factory)
	instance := graph.New(graph.Config{TypeFactory: factory})

	_, values, err := instance.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, spread, err := instance.CreateNode(spreadType)
	require.NoError(t, err)
	require.NoError(t, instance.ConnectNodes(values, "Out", spread, "Array"))

	_, single, err := instance.CreateNode(floatConsumerType)
	require.NoError(t, err)
	_, lifted, err := instance.CreateNode(liftedDoubleType)
	require.NoError(t, err)
	total, list, err := instance.CreateNode(liftedTotalType)
	require.NoError(t, err)

	assert.ErrorContains(t, instance.ConnectNodes(spread, "Out", single, "In"), "not a spread")
	assert.ErrorContains(t, instance.ConnectNodes(spread, "Out", lifted, "In"), "not a spread")
	require.NoError(t, instance.ConnectNodes(spread, "Out", list, "In"))
	assert.Equal(t, 6., nodes.GetNodeOutputPort[float64](total, "Out").Value())
}
