package nodes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type subtractNode struct {
	A nodes.LiftedPort[float64]
	B nodes.LiftedPort[float64]
}

func (n subtractNode) Differences(out *nodes.Lifted[float64]) {
	nodes.Zip2(out, n.A, n.B, func(a, b float64) float64 { return a - b })
}

type sumNode struct {
	Values []nodes.LiftedPort[float64]
}

func (n sumNode) Out(out *nodes.Lifted[float64]) {
	nodes.ZipAll(out, n.Values, func(vals []float64) float64 {
		total := 0.
		for _, v := range vals {
			total += v
		}
		return total
	})
}

type negateNode struct {
	In nodes.LiftedPort[float64]
}

func (n negateNode) Out(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, n.In, func(v float64) float64 { return -v })
}

// countNode proves a lifted input can carry a slice element type: the
// scalar rank is []float64 and the array rank is [][]float64.
type sliceElementNode struct {
	In nodes.LiftedPort[[]float64]
}

func (n sliceElementNode) Out(out *nodes.Lifted[int]) {
	nodes.Zip1(out, n.In, func(v []float64) int { return len(v) })
}

func scalars(t *testing.T, node nodes.Node, port string, value float64) {
	t.Helper()
	connect(t, node, port, nodes.ConstOutput[float64]{Val: value})
}

func array(t *testing.T, node nodes.Node, port string, values ...float64) {
	t.Helper()
	connect(t, node, port, nodes.ConstOutput[[]float64]{Val: values})
}

func TestLiftedStaysScalarWhenEveryInputIs(t *testing.T) {
	node := &nodes.Struct[subtractNode]{}
	scalars(t, node, "A", 7)
	scalars(t, node, "B", 2)

	out := node.Outputs()["Differences"]
	assert.Equal(t, "float64", out.(nodes.Typed).Type())

	typed, ok := out.(nodes.Output[float64])
	require.True(t, ok, "a scalar-ranked output really is an Output[float64], got %T", out)
	assert.Equal(t, 5.0, typed.Value())
}

func TestLiftedBecomesAnArrayWhenAnyInputIs(t *testing.T) {
	node := &nodes.Struct[subtractNode]{}
	array(t, node, "A", 10, 20, 30)
	array(t, node, "B", 1, 2, 3)

	out := node.Outputs()["Differences"]
	assert.Equal(t, "[]float64", out.(nodes.Typed).Type())

	typed, ok := out.(nodes.Output[[]float64])
	require.True(t, ok, "got %T", out)
	assert.Equal(t, []float64{9, 18, 27}, typed.Value())
}

func TestLiftedBroadcastsAScalarAcrossAnArray(t *testing.T) {
	node := &nodes.Struct[subtractNode]{}
	array(t, node, "A", 10, 20, 30)
	scalars(t, node, "B", 1)

	assert.Equal(t, "[]float64", node.Outputs()["Differences"].(nodes.Typed).Type(),
		"one array input is enough to make the output an array")
	assert.Equal(t, []float64{9, 19, 29},
		nodes.GetNodeOutputPort[[]float64](node, "Differences").Value())
}

func TestLiftedBroadcastsFromEitherSide(t *testing.T) {
	node := &nodes.Struct[subtractNode]{}
	scalars(t, node, "A", 100)
	array(t, node, "B", 1, 2, 3)

	assert.Equal(t, []float64{99, 98, 97},
		nodes.GetNodeOutputPort[[]float64](node, "Differences").Value(),
		"the hand written twin only ever supported one of these orders")
}

func TestLiftedBroadcastsALengthOneArray(t *testing.T) {
	node := &nodes.Struct[subtractNode]{}
	array(t, node, "A", 10, 20, 30)
	array(t, node, "B", 1)

	assert.Equal(t, []float64{9, 19, 29},
		nodes.GetNodeOutputPort[[]float64](node, "Differences").Value())
}

func TestLiftedRefusesAGenuineLengthMismatch(t *testing.T) {
	node := &nodes.Struct[subtractNode]{}
	array(t, node, "A", 1, 2, 3, 4, 5)
	array(t, node, "B", 1, 2, 3)

	out := node.Outputs()["Differences"]
	assert.Empty(t, nodes.GetNodeOutputPort[[]float64](node, "Differences").Value(),
		"no truncated result that looks like it worked")

	errors := out.(nodes.ObservableExecution).ExecutionReport().Errors
	require.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "5 and 3")
}

func TestLiftedWithNothingWiredInIsScalar(t *testing.T) {
	node := &nodes.Struct[subtractNode]{}

	assert.Equal(t, "float64", node.Outputs()["Differences"].(nodes.Typed).Type())
	assert.Equal(t, 0.0, nodes.GetNodeOutputPort[float64](node, "Differences").Value())
}

func TestLiftedUnaryFollowsItsInput(t *testing.T) {
	scalar := &nodes.Struct[negateNode]{}
	scalars(t, scalar, "In", 3)
	assert.Equal(t, -3.0, nodes.GetNodeOutputPort[float64](scalar, "Out").Value())

	lifted := &nodes.Struct[negateNode]{}
	array(t, lifted, "In", 1, 2)
	assert.Equal(t, []float64{-1, -2}, nodes.GetNodeOutputPort[[]float64](lifted, "Out").Value())
}

func TestLiftedInputAdvertisesBothRanks(t *testing.T) {
	node := &nodes.Struct[subtractNode]{}

	input := node.Inputs()["A"]
	assert.Equal(t, "", input.(nodes.Typed).Type(), "unconnected, either rank is still allowed")
	assert.Equal(t, []string{"float64", "[]float64"}, input.(nodes.TypeOptions).AcceptedTypes())

	scalars(t, node, "A", 1)
	assert.Equal(t, "float64", node.Inputs()["A"].(nodes.Typed).Type(),
		"connected, it reports what it actually took")
}

func TestLiftedInputRefusesAnUnrelatedType(t *testing.T) {
	node := &nodes.Struct[subtractNode]{}

	input := node.Inputs()["A"].(nodes.SingleValueInputPort)
	err := input.Set(nodes.ConstOutput[string]{Val: "no"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "float64")
}

func TestLiftedResolvesASliceElementType(t *testing.T) {
	scalar := &nodes.Struct[sliceElementNode]{}
	connect(t, scalar, "In", nodes.ConstOutput[[]float64]{Val: []float64{1, 2, 3}})
	assert.Equal(t, "int", scalar.Outputs()["Out"].(nodes.Typed).Type(),
		"[]float64 against a LiftedPort[[]float64] is the scalar rank")
	assert.Equal(t, 3, nodes.GetNodeOutputPort[int](scalar, "Out").Value())

	lifted := &nodes.Struct[sliceElementNode]{}
	connect(t, lifted, "In", nodes.ConstOutput[[][]float64]{Val: [][]float64{{1, 2}, {3}}})
	assert.Equal(t, "[]int", lifted.Outputs()["Out"].(nodes.Typed).Type(),
		"[][]float64 is the array rank of the same port")
	assert.Equal(t, []int{2, 1}, nodes.GetNodeOutputPort[[]int](lifted, "Out").Value())
}

func addTo(t *testing.T, node nodes.Node, port string, source nodes.OutputPort) {
	t.Helper()
	input, ok := node.Inputs()[port].(nodes.ArrayValueInputPort)
	require.True(t, ok, "node has no array input %q", port)
	require.NoError(t, input.Add(source))
}

func TestZipAllStaysScalarWhenEveryConnectionIs(t *testing.T) {
	node := &nodes.Struct[sumNode]{}
	addTo(t, node, "Values", nodes.ConstOutput[float64]{Val: 1})
	addTo(t, node, "Values", nodes.ConstOutput[float64]{Val: 2})
	addTo(t, node, "Values", nodes.ConstOutput[float64]{Val: 4})

	assert.Equal(t, "float64", node.Outputs()["Out"].(nodes.Typed).Type())
	assert.Equal(t, 7.0, nodes.GetNodeOutputPort[float64](node, "Out").Value())
}

func TestZipAllBroadcastsScalarsAcrossOneArray(t *testing.T) {
	node := &nodes.Struct[sumNode]{}
	addTo(t, node, "Values", nodes.ConstOutput[[]float64]{Val: []float64{1, 2, 3}})
	addTo(t, node, "Values", nodes.ConstOutput[float64]{Val: 10})

	assert.Equal(t, "[]float64", node.Outputs()["Out"].(nodes.Typed).Type())
	assert.Equal(t, []float64{11, 12, 13}, nodes.GetNodeOutputPort[[]float64](node, "Out").Value())
}

func TestZipAllCombinesTwoArraysElementwise(t *testing.T) {
	node := &nodes.Struct[sumNode]{}
	addTo(t, node, "Values", nodes.ConstOutput[[]float64]{Val: []float64{1, 2, 3}})
	addTo(t, node, "Values", nodes.ConstOutput[[]float64]{Val: []float64{10, 20, 30}})

	assert.Equal(t, []float64{11, 22, 33}, nodes.GetNodeOutputPort[[]float64](node, "Out").Value())
}

func TestZipAllRefusesAGenuineLengthMismatch(t *testing.T) {
	node := &nodes.Struct[sumNode]{}
	addTo(t, node, "Values", nodes.ConstOutput[[]float64]{Val: []float64{1, 2, 3}})
	addTo(t, node, "Values", nodes.ConstOutput[[]float64]{Val: []float64{10, 20}})

	out := node.Outputs()["Out"]
	assert.Empty(t, nodes.GetNodeOutputPort[[]float64](node, "Out").Value())

	errors := out.(nodes.ObservableExecution).ExecutionReport().Errors
	require.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "3 and 2")
}

func TestZipAllWithNothingWiredInIsScalar(t *testing.T) {
	node := &nodes.Struct[sumNode]{}

	assert.Equal(t, "float64", node.Outputs()["Out"].(nodes.Typed).Type())
	assert.Equal(t, 0.0, nodes.GetNodeOutputPort[float64](node, "Out").Value())
}

func TestZipAllInputReportsBothRanks(t *testing.T) {
	node := &nodes.Struct[sumNode]{}
	accepted := node.Inputs()["Values"].(nodes.TypeOptions).AcceptedTypes()
	assert.Equal(t, []string{"float64", "[]float64"}, accepted)
}

func TestZipAllInputRefusesAnUnrelatedType(t *testing.T) {
	node := &nodes.Struct[sumNode]{}
	input := node.Inputs()["Values"].(nodes.ArrayValueInputPort)
	assert.Error(t, input.Add(nodes.ConstOutput[string]{Val: "nope"}))
}
