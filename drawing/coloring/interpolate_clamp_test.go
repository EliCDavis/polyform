package coloring_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

func interpolate(time float64, clamp nodes.Output[bool]) coloring.Color {
	node := &nodes.Struct[coloring.InterpolateNode]{
		Data: coloring.InterpolateNode{
			A:     nodes.ConstOutput[coloring.Color]{Val: coloring.Color{R: 0, G: 0, B: 0, A: 1}},
			B:     nodes.ConstOutput[coloring.Color]{Val: coloring.Color{R: 0.5, G: 0.5, B: 0.5, A: 1}},
			Time:  nodes.ConstOutput[float64]{Val: time},
			Clamp: clamp,
		},
	}
	return nodes.GetNodeOutputPort[coloring.Color](node, "Out").Value()
}

func TestInterpolateNodeExtrapolatesByDefault(t *testing.T) {
	assert.InDelta(t, 1, interpolate(2, nil).R, 1e-9)
	assert.InDelta(t, -0.5, interpolate(-1, nil).R, 1e-9)
}

func TestInterpolateNodeClampHoldsTimeInsideZeroToOne(t *testing.T) {
	clamp := nodes.ConstOutput[bool]{Val: true}

	assert.InDelta(t, 0.5, interpolate(2, clamp).R, 1e-9)
	assert.InDelta(t, 0, interpolate(-1, clamp).R, 1e-9)
	assert.InDelta(t, 0.25, interpolate(0.5, clamp).R, 1e-9)
}
