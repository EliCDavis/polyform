package math_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

func arrays(a, b []float64) (nodes.Output[[]float64], nodes.Output[[]float64]) {
	return nodes.ConstOutput[[]float64]{Val: a}, nodes.ConstOutput[[]float64]{Val: b}
}

func TestMultiplyArraysNode(t *testing.T) {
	a, b := arrays([]float64{1, 2, 3}, []float64{10, 0.5, -1})
	node := &nodes.Struct[math.MultiplyArraysNode[float64]]{Data: math.MultiplyArraysNode[float64]{A: a, B: b}}
	assert.Equal(t, []float64{10, 1, -3}, nodes.GetNodeOutputPort[[]float64](node, "Products").Value())
}

func TestAddArraysNode(t *testing.T) {
	a, b := arrays([]float64{1, 2}, []float64{10, 20})
	node := &nodes.Struct[math.AddArraysNode[float64]]{Data: math.AddArraysNode[float64]{A: a, B: b}}
	assert.Equal(t, []float64{11, 22}, nodes.GetNodeOutputPort[[]float64](node, "Sums").Value())
}

func TestSubtractArraysNode(t *testing.T) {
	a, b := arrays([]float64{10, 20}, []float64{1, 2})
	node := &nodes.Struct[math.SubtractArraysNode[float64]]{Data: math.SubtractArraysNode[float64]{A: a, B: b}}
	assert.Equal(t, []float64{9, 18}, nodes.GetNodeOutputPort[[]float64](node, "Differences").Value())
}

func TestDivideArraysNode(t *testing.T) {
	a, b := arrays([]float64{10, 20, 5}, []float64{2, 0, 5})
	node := &nodes.Struct[math.DivideArraysNode[float64]]{Data: math.DivideArraysNode[float64]{A: a, B: b}}
	port := nodes.GetNodeOutputPort[[]float64](node, "Quotients")
	assert.Equal(t, []float64{5, 0, 1}, port.Value())
	assert.NotEmpty(t, port.(nodes.ObservableExecution).ExecutionReport().Errors, "divide by zero should be reported")
}

func TestMinMaxArraysNode(t *testing.T) {
	a, b := arrays([]float64{1, 5, 3}, []float64{4, 2, 3})
	minNode := &nodes.Struct[math.MinArraysNode[float64]]{Data: math.MinArraysNode[float64]{A: a, B: b}}
	maxNode := &nodes.Struct[math.MaxArraysNode[float64]]{Data: math.MaxArraysNode[float64]{A: a, B: b}}
	assert.Equal(t, []float64{1, 2, 3}, nodes.GetNodeOutputPort[[]float64](minNode, "Minimums").Value())
	assert.Equal(t, []float64{4, 5, 3}, nodes.GetNodeOutputPort[[]float64](maxNode, "Maximums").Value())
}

func TestElementwiseLengthMismatchTruncatesAndReports(t *testing.T) {
	a, b := arrays([]float64{1, 2, 3}, []float64{10, 20})
	node := &nodes.Struct[math.AddArraysNode[float64]]{Data: math.AddArraysNode[float64]{A: a, B: b}}
	port := nodes.GetNodeOutputPort[[]float64](node, "Sums")
	assert.Equal(t, []float64{11, 22}, port.Value())
	assert.NotEmpty(t, port.(nodes.ObservableExecution).ExecutionReport().Errors)
}

func TestElementwiseUnwiredInputsYieldEmpty(t *testing.T) {
	node := &nodes.Struct[math.AddArraysNode[float64]]{}
	assert.Empty(t, nodes.GetNodeOutputPort[[]float64](node, "Sums").Value())
}
