package vector4_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector4"
	"github.com/EliCDavis/polyform/nodes"
	v4 "github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
)

func TestSelectNode(t *testing.T) {
	tests := map[string]struct {
		in nodes.Output[v4.Vector[float64]]
		x  float64
		y  float64
		z  float64
		w  float64
	}{
		"(nil) => 0,0,0,0": {},
		"(1,2,3,4) => 1,2,3,4": {
			in: nodes.ConstOutput[v4.Vector[float64]]{Val: v4.New(1., 2., 3., 4.)},
			x:  1.,
			y:  2.,
			z:  3.,
			w:  4.,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[vector4.Select[float64]]{
				Data: vector4.Select[float64]{
					In: tc.in,
				},
			}
			assert.Equal(t, tc.x, nodes.GetNodeOutputPort[float64](node, "X").Value())
			assert.Equal(t, tc.y, nodes.GetNodeOutputPort[float64](node, "Y").Value())
			assert.Equal(t, tc.z, nodes.GetNodeOutputPort[float64](node, "Z").Value())
			assert.Equal(t, tc.w, nodes.GetNodeOutputPort[float64](node, "W").Value())
		})
	}
}

func TestSelectArrayNode(t *testing.T) {
	tests := map[string]struct {
		in nodes.Output[[]v4.Vector[float64]]
		x  []float64
		y  []float64
		z  []float64
		w  []float64
	}{
		"(nil) => empty": {
			x: []float64{},
			y: []float64{},
			z: []float64{},
			w: []float64{},
		},
		"(1,2,3,4) => []1, []2, []3, []4": {
			in: nodes.ConstOutput[[]v4.Vector[float64]]{
				Val: []v4.Vector[float64]{
					v4.New(1., 2., 3., 4.),
				},
			},
			x: []float64{1.},
			y: []float64{2.},
			z: []float64{3.},
			w: []float64{4.},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[vector4.SelectArray[float64]]{
				Data: vector4.SelectArray[float64]{
					In: tc.in,
				},
			}
			assert.Equal(t, tc.x, nodes.GetNodeOutputPort[[]float64](node, "X").Value())
			assert.Equal(t, tc.y, nodes.GetNodeOutputPort[[]float64](node, "Y").Value())
			assert.Equal(t, tc.z, nodes.GetNodeOutputPort[[]float64](node, "Z").Value())
			assert.Equal(t, tc.w, nodes.GetNodeOutputPort[[]float64](node, "W").Value())
		})
	}
}
