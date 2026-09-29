package graph_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type floatArraySourceNode struct{}

func (floatArraySourceNode) Out(out *nodes.StructOutput[[]float64]) {
	out.Set([]float64{1, 2, 3})
}

type liftedDoubleNode struct {
	In nodes.LiftedPort[float64]
}

func (n liftedDoubleNode) Out(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, n.In, func(v float64) float64 { return v * 2 })
}

type floatArraySumNode struct {
	In nodes.Output[[]float64]
}

func (n floatArraySumNode) Out(out *nodes.StructOutput[float64]) {
	total := 0.
	for _, v := range nodes.TryGetOutputValue(out, n.In, nil) {
		total += v
	}
	out.Set(total)
}

type floatSourceNode struct{}

func (floatSourceNode) Out(out *nodes.StructOutput[float64]) {
	out.Set(2)
}

type liftedTotalNode struct {
	In []nodes.LiftedPort[float64]
}

func (n liftedTotalNode) Out(out *nodes.Lifted[float64]) {
	nodes.ZipAll(out, n.In, func(vals []float64) float64 {
		total := 0.
		for _, v := range vals {
			total += v
		}
		return total
	})
}

type floatConsumerNode struct {
	In nodes.Output[float64]
}

func (n floatConsumerNode) Out(out *nodes.StructOutput[float64]) {
	out.Set(nodes.TryGetOutputValue(out, n.In, 0))
}

const (
	floatArraySourceType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.floatArraySourceNode]"
	liftedDoubleType     = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.liftedDoubleNode]"
	floatArraySumType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.floatArraySumNode]"
	floatSourceType      = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.floatSourceNode]"
	liftedTotalType      = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.liftedTotalNode]"
	floatConsumerType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/graph_test.floatConsumerNode]"
)

func testInstanceWithLiftedNodes(t *testing.T) *graph.Instance {
	t.Helper()
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[floatArraySourceNode]](factory)
	refutil.RegisterType[nodes.Struct[liftedDoubleNode]](factory)
	refutil.RegisterType[nodes.Struct[floatArraySumNode]](factory)
	refutil.RegisterType[nodes.Struct[floatSourceNode]](factory)
	refutil.RegisterType[nodes.Struct[liftedTotalNode]](factory)
	refutil.RegisterType[nodes.Struct[floatConsumerNode]](factory)

	instance := graph.New(graph.Config{TypeFactory: factory})
	nodes.DiscoverPortTypes(factory)
	return instance
}

func TestConnectNodesRefusesAnArrayThatRetypesAConsumedLiftedOutput(t *testing.T) {
	instance := testInstanceWithLiftedNodes(t)

	_, scalar, err := instance.CreateNode(floatSourceType)
	require.NoError(t, err)
	_, array, err := instance.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, total, err := instance.CreateNode(liftedTotalType)
	require.NoError(t, err)
	_, consumer, err := instance.CreateNode(floatConsumerType)
	require.NoError(t, err)

	instance.ConnectNodes(scalar, "Out", total, "In.0")
	instance.ConnectNodes(total, "Out", consumer, "In")
	require.InDelta(t, 2, nodes.GetNodeOutputPort[float64](instance.Node(consumer), "Out").Value(), 1e-12)

	// A second element carrying an array makes Out an []float64, and the
	// consumer is still holding the float64 port it was handed.
	require.Panics(t, func() {
		instance.ConnectNodes(array, "Out", total, "In.1")
	})

	assert.Len(t, instance.Node(total).Inputs()["In"].(nodes.ArrayValueInputPort).Value(), 1,
		"the refused element came back off")
	assert.InDelta(t, 2, nodes.GetNodeOutputPort[float64](instance.Node(consumer), "Out").Value(), 1e-12,
		"and the consumer reads what it did before, rather than a zero")
}

func TestLiftedOutputRankSurvivesASaveAndLoad(t *testing.T) {
	instance := testInstanceWithLiftedNodes(t)

	_, source, err := instance.CreateNode(floatArraySourceType)
	require.NoError(t, err)
	_, double, err := instance.CreateNode(liftedDoubleType)
	require.NoError(t, err)
	_, sum, err := instance.CreateNode(floatArraySumType)
	require.NoError(t, err)

	instance.ConnectNodes(source, "Out", double, "In")
	instance.ConnectNodes(double, "Out", sum, "In")

	require.InDelta(t, 12, nodes.GetNodeOutputPort[float64](instance.Node(sum), "Out").Value(), 1e-12)

	saved, err := instance.EncodeToAppSchema()
	require.NoError(t, err)

	// Connections are restored in an order derived from a map, so a single
	// load could get away with wiring the consumer before the producer knows
	// it is carrying an array.
	for range 50 {
		reloaded := testInstanceWithLiftedNodes(t)
		require.NoError(t, reloaded.ApplyAppSchema(saved))
		assert.InDelta(t, 12, nodes.GetNodeOutputPort[float64](reloaded.Node(sum), "Out").Value(), 1e-12)
	}
}
