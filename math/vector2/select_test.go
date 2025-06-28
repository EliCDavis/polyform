package vector2_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector2"
	"github.com/EliCDavis/polyform/nodes"
	v2 "github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
)

func TestSelectNode(t *testing.T) {
	tests := map[string]struct {
		in nodes.Output[v2.Vector[float64]]
		x  float64
		y  float64
	}{
		"(nil) => 0,0,0": {},
		"(1,2,3) => 1,2,3": {
			in: nodes.ConstOutput[v2.Vector[float64]]{Val: v2.New(1., 2.)},
			x:  1.,
			y:  2.,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[vector2.Select[float64]]{
				Data: vector2.Select[float64]{
					In: tc.in,
				},
			}
			x := nodes.GetNodeOutputPort[float64](node, "X").Value()
			y := nodes.GetNodeOutputPort[float64](node, "Y").Value()
			assert.Equal(t, tc.x, x)
			assert.Equal(t, tc.y, y)
		})
	}
}

func TestSelectArrayNode(t *testing.T) {
	tests := map[string]struct {
		in nodes.Output[[]v2.Vector[float64]]
		x  []float64
		y  []float64
	}{
		"(nil) => nil": {
			x: []float64{},
			y: []float64{},
		},
		"1,2,3 => []1, []2, []3": {
			in: nodes.ConstOutput[[]v2.Vector[float64]]{
				Val: []v2.Vector[float64]{
					v2.New(1., 2.),
				},
			},
			x: []float64{1.},
			y: []float64{2.},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[vector2.SelectArray[float64]]{
				Data: vector2.SelectArray[float64]{
					In: tc.in,
				},
			}
			x := nodes.GetNodeOutputPort[[]float64](node, "X").Value()
			y := nodes.GetNodeOutputPort[[]float64](node, "Y").Value()
			assert.Equal(t, tc.x, x)
			assert.Equal(t, tc.y, y)
		})
	}
}
