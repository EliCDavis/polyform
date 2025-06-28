package vector3_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector3"
	"github.com/EliCDavis/polyform/nodes"
	v3 "github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestSelectNode(t *testing.T) {
	tests := map[string]struct {
		in nodes.Output[v3.Vector[float64]]
		x  float64
		y  float64
		z  float64
	}{
		"(nil) => 0,0,0": {},
		"(1,2,3) => 1,2,3": {
			in: nodes.ConstOutput[v3.Vector[float64]]{Val: v3.New(1., 2., 3.)},
			x:  1.,
			y:  2.,
			z:  3.,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[vector3.Select[float64]]{
				Data: vector3.Select[float64]{
					In: tc.in,
				},
			}
			x := nodes.GetNodeOutputPort[float64](node, "X").Value()
			y := nodes.GetNodeOutputPort[float64](node, "Y").Value()
			z := nodes.GetNodeOutputPort[float64](node, "Z").Value()
			assert.Equal(t, tc.x, x)
			assert.Equal(t, tc.y, y)
			assert.Equal(t, tc.z, z)
		})
	}
}

func TestSelectArrayNode(t *testing.T) {
	tests := map[string]struct {
		in nodes.Output[[]v3.Vector[float64]]
		x  []float64
		y  []float64
		z  []float64
	}{
		"(nil) => nil": {
			x: []float64{},
			y: []float64{},
			z: []float64{},
		},
		"1,2,3 => []1, []2, []3": {
			in: nodes.ConstOutput[[]v3.Vector[float64]]{
				Val: []v3.Vector[float64]{
					v3.New(1., 2., 3.),
				},
			},
			x: []float64{1.},
			y: []float64{2.},
			z: []float64{3.},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[vector3.SelectArray[float64]]{
				Data: vector3.SelectArray[float64]{
					In: tc.in,
				},
			}
			x := nodes.GetNodeOutputPort[[]float64](node, "X").Value()
			y := nodes.GetNodeOutputPort[[]float64](node, "Y").Value()
			z := nodes.GetNodeOutputPort[[]float64](node, "Z").Value()
			assert.Equal(t, tc.x, x)
			assert.Equal(t, tc.y, y)
			assert.Equal(t, tc.z, z)
		})
	}
}
