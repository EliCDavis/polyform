package coloring_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

func TestInterpolateArraysNode(t *testing.T) {
	black := coloring.Color{R: 0, G: 0, B: 0, A: 1}
	white := coloring.Color{R: 1, G: 1, B: 1, A: 1}
	red := coloring.Color{R: 1, G: 0, B: 0, A: 1}

	node := &nodes.Struct[coloring.InterpolateArraysNode]{Data: coloring.InterpolateArraysNode{
		A:    nodes.ConstOutput[[]coloring.Color]{Val: []coloring.Color{black, black, red}},
		B:    nodes.ConstOutput[[]coloring.Color]{Val: []coloring.Color{white, white, white}},
		Time: nodes.ConstOutput[[]float64]{Val: []float64{0, 1, 0.5}},
	}}

	out := nodes.GetNodeOutputPort[[]coloring.Color](node, "Out").Value()
	assert.Equal(t, []coloring.Color{
		black,
		white,
		{R: 1, G: 0.5, B: 0.5, A: 1},
	}, out)
}

func TestInterpolateArraysNodeLengthMismatchTruncatesAndReports(t *testing.T) {
	black := coloring.Color{A: 1}
	white := coloring.Color{R: 1, G: 1, B: 1, A: 1}

	node := &nodes.Struct[coloring.InterpolateArraysNode]{Data: coloring.InterpolateArraysNode{
		A:    nodes.ConstOutput[[]coloring.Color]{Val: []coloring.Color{black, black}},
		B:    nodes.ConstOutput[[]coloring.Color]{Val: []coloring.Color{white}},
		Time: nodes.ConstOutput[[]float64]{Val: []float64{1, 1}},
	}}

	port := nodes.GetNodeOutputPort[[]coloring.Color](node, "Out")
	assert.Equal(t, []coloring.Color{white}, port.Value())
	assert.NotEmpty(t, port.(nodes.ObservableExecution).ExecutionReport().Errors)
}
