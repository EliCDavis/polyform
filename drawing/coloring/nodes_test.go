package coloring_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

func TestInterpolateNode(t *testing.T) {
	tests := map[string]struct {
		a nodes.Output[coloring.WebColor]
		b nodes.Output[coloring.WebColor]
		t nodes.Output[float64]

		val coloring.WebColor
	}{
		"nil + nil = 0,0,0,1": {
			val: coloring.WebColor{R: 0, G: 0, B: 0, A: 1},
		},
		"black + white = grey": {
			a:   nodes.ConstOutput[coloring.WebColor]{Val: coloring.Black()},
			b:   nodes.ConstOutput[coloring.WebColor]{Val: coloring.White()},
			t:   nodes.ConstOutput[float64]{Val: 0.5},
			val: coloring.WebColor{R: 128 / 255., G: 128 / 255., B: 128 / 255., A: 1},
		},
		"black = black": {
			a:   nodes.ConstOutput[coloring.WebColor]{Val: coloring.Black()},
			t:   nodes.ConstOutput[float64]{Val: 0.5},
			val: coloring.Black(),
		},
		"white = white": {
			a:   nodes.ConstOutput[coloring.WebColor]{Val: coloring.White()},
			t:   nodes.ConstOutput[float64]{Val: 0.5},
			val: coloring.White(),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[coloring.InterpolateNode]{
				Data: coloring.InterpolateNode{
					A:    tc.a,
					B:    tc.b,
					Time: tc.t,
				},
			}
			out := nodes.GetNodeOutputPort[coloring.WebColor](node, "Out").Value()
			assert.InDelta(t, tc.val.R, out.R, 0.01)
			assert.InDelta(t, tc.val.G, out.G, 0.01)
			assert.InDelta(t, tc.val.B, out.B, 0.01)
			assert.InDelta(t, tc.val.A, out.A, 0.01)
			// assert.Equal(t, tc.val, out)
		})
	}
}
