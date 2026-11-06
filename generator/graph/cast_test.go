package graph_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/generator/cast"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type castIntNode struct{}

func (castIntNode) Out(out *nodes.StructOutput[int]) { out.Set(3) }

type castFloatsNode struct{}

func (castFloatsNode) Out(out *nodes.StructOutput[[]float64]) { out.Set([]float64{1.4, 2.6}) }

type castSinkNode struct {
	Single nodes.Output[float64]
	Items  []nodes.Output[float64]
	Lifted nodes.LiftedPort[int]
}

func (n castSinkNode) Out(out *nodes.StructOutput[float64]) {
	total := nodes.GetOutputValue(out, n.Single)
	for _, v := range nodes.GetOutputValues(out, n.Items) {
		total += v
	}
	out.Set(total)
}

func (n castSinkNode) Doubled(out *nodes.Lifted[int]) {
	nodes.Zip1(out, n.Lifted, func(v int) int { return v * 2 })
}

const (
	castIntType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.castIntNode]"
	castFloatsType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.castFloatsNode]"
	castSinkType   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.castSinkNode]"
)

func castInstance(casts *cast.Table) *graph.Instance {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[castIntNode]](factory)
	refutil.RegisterType[nodes.Struct[castFloatsNode]](factory)
	refutil.RegisterType[nodes.Struct[castSinkNode]](factory)
	return graph.New(graph.Config{TypeFactory: factory, Casts: casts})
}

func numberCasts() *cast.Table {
	table := &cast.Table{}
	cast.Register(table, func(v int) float64 { return float64(v) })
	cast.Register(table, func(v float64) int { return int(math.Round(v)) })
	return table
}

func castGraph(t *testing.T, casts *cast.Table) (instance *graph.Instance, source, floats, sink string) {
	t.Helper()
	instance = castInstance(casts)
	_, source, err := instance.CreateNode(castIntType)
	require.NoError(t, err)
	_, floats, err = instance.CreateNode(castFloatsType)
	require.NoError(t, err)
	_, sink, err = instance.CreateNode(castSinkType)
	require.NoError(t, err)
	return instance, source, floats, sink
}

func TestCastFeedsSingleAndArrayInputs(t *testing.T) {
	instance, source, _, sink := castGraph(t, numberCasts())
	require.NoError(t, instance.ConnectNodes(source, "Out", sink, "Single"))
	require.NoError(t, instance.ConnectNodes(source, "Out", sink, "Items"))
	require.NoError(t, instance.ConnectNodes(source, "Out", sink, "Items"))

	assert.Equal(t, 9., nodes.GetNodeOutputPort[float64](instance.Node(sink), "Out").Value())
}

func TestCastMapsArraysIntoLiftedPorts(t *testing.T) {
	instance, _, floats, sink := castGraph(t, numberCasts())
	require.NoError(t, instance.ConnectNodes(floats, "Out", sink, "Lifted"))

	assert.Equal(t, []int{2, 6}, nodes.GetNodeOutputPort[[]int](instance.Node(sink), "Doubled").Value())
}

func TestCastSurvivesSaveAndLoad(t *testing.T) {
	instance, source, _, sink := castGraph(t, numberCasts())
	require.NoError(t, instance.ConnectNodes(source, "Out", sink, "Single"))
	saved, err := instance.EncodeToAppSchema()
	require.NoError(t, err)

	loaded := castInstance(numberCasts())
	require.NoError(t, loaded.ApplyAppSchema(saved))
	assert.Equal(t, 3., nodes.GetNodeOutputPort[float64](loaded.Node(sink), "Out").Value())
}

func TestUnregisteredCastIsRefused(t *testing.T) {
	instance, source, _, sink := castGraph(t, nil)
	err := instance.ConnectNodes(source, "Out", sink, "Single")
	assert.ErrorContains(t, err, `"Out" output is int, but node "`+sink+`"'s "Single" input takes float64`)
}
