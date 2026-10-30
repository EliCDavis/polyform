package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type insertSourceNode struct{}

func (insertSourceNode) Out(out *nodes.StructOutput[float64]) {}

type insertSinkNode struct {
	Values []nodes.Output[float64]
}

func (insertSinkNode) Out(out *nodes.StructOutput[float64]) {}

const (
	insertSourceType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.insertSourceNode]"
	insertSinkType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.insertSinkNode]"
)

func insertGraph(t *testing.T) (*graph.Instance, nodes.Node, string, []string) {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[insertSourceNode]](factory)
	refutil.RegisterType[nodes.Struct[insertSinkNode]](factory)
	instance := graph.New(graph.Config{TypeFactory: factory})

	sink, sinkID, err := instance.CreateNode(insertSinkType)
	require.NoError(t, err)

	sources := make([]string, 3)
	for i := range sources {
		_, sources[i], err = instance.CreateNode(insertSourceType)
		require.NoError(t, err)
	}
	return instance, sink, sinkID, sources
}

func valuesFrom(instance *graph.Instance, sink nodes.Node) []nodes.Node {
	var from []nodes.Node
	for _, port := range sink.Inputs()["Values"].(nodes.ArrayValueInputPort).Value() {
		from = append(from, port.Node())
	}
	return from
}

func TestInsertingIntoAnArrayInputShiftsTheLaterElements(t *testing.T) {
	instance, sink, sinkID, sources := insertGraph(t)
	require.NoError(t, instance.ConnectNodes(sources[0], "Out", sinkID, "Values"))
	require.NoError(t, instance.ConnectNodes(sources[2], "Out", sinkID, "Values"))

	require.NoError(t, instance.InsertNodeConnection(sources[1], "Out", sinkID, "Values.1"))

	assert.Equal(t, []nodes.Node{
		instance.Node(sources[0]),
		instance.Node(sources[1]),
		instance.Node(sources[2]),
	}, valuesFrom(instance, sink))
}

func TestInsertingPastTheEndOfAnArrayInputIsRefused(t *testing.T) {
	instance, sink, sinkID, sources := insertGraph(t)
	require.NoError(t, instance.ConnectNodes(sources[0], "Out", sinkID, "Values"))

	assert.Error(t, instance.InsertNodeConnection(sources[1], "Out", sinkID, "Values.2"))
	assert.Len(t, valuesFrom(instance, sink), 1)
}
