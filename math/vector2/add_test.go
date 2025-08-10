package vector2_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector2"
	"github.com/EliCDavis/polyform/nodes"
	v2 "github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
)

func TestSumNode(t *testing.T) {
	tests := map[string]struct {
		in  []nodes.Output[v2.Vector[float64]]
		out v2.Vector[float64]
	}{
		"nil => 0": {in: nil, out: v2.Zero[float64]()},
		"[(1,2,3)] => (1,2,3)": {
			in: []nodes.Output[v2.Vector[float64]]{
				nodes.ConstOutput[v2.Vector[float64]]{Val: v2.New(1., 2.)},
			},
			out: v2.New(1., 2.),
		},
		"[(1,2,3), (4,5,6)] => (5,7,9)": {
			in: []nodes.Output[v2.Vector[float64]]{
				nodes.ConstOutput[v2.Vector[float64]]{Val: v2.New(1., 2.)},
				nodes.ConstOutput[v2.Vector[float64]]{Val: v2.New(4., 5.)},
			},
			out: v2.New(5., 7.),
		},
		"[(1,2,3), nil] => (1,2,3)": {
			in: []nodes.Output[v2.Vector[float64]]{
				nodes.ConstOutput[v2.Vector[float64]]{Val: v2.New(1., 2.)},
				nil,
			},
			out: v2.New(1., 2.),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[vector2.SumNode[float64]]{
				Data: vector2.SumNode[float64]{
					Values: tc.in,
				},
			}
			out := nodes.GetNodeOutputPort[v2.Vector[float64]](node, "Out").Value()
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestAddToArrayNode(t *testing.T) {
	tests := map[string]struct {
		amount nodes.Output[v2.Vector[float64]]
		array  nodes.Output[[]v2.Vector[float64]]
		out    []v2.Vector[float64]
	}{
		"(nil + nil) => nil": {amount: nil, array: nil, out: nil},
		"((1,2,3) + nil) => nil": {
			amount: nodes.ConstOutput[v2.Vector[float64]]{Val: v2.New(1., 2.)},
			array:  nil,
			out:    nil,
		},
		"(nil + [(1,2,3)]) => [(1,2,3)]": {
			amount: nil,
			array: nodes.ConstOutput[[]v2.Vector[float64]]{
				Val: []v2.Float64{
					v2.New(1., 2.),
				},
			},
			out: []v2.Float64{
				v2.New(1., 2.),
			},
		},
		"((1,2,3) + [(1,1,1), (2,2,2)]) => [(2,3,4), (3,4,5)]": {
			amount: nodes.ConstOutput[v2.Vector[float64]]{Val: v2.New(1., 2.)},
			array: nodes.ConstOutput[[]v2.Vector[float64]]{
				Val: []v2.Float64{
					v2.New(1., 1.),
					v2.New(2., 2.),
				},
			},
			out: []v2.Float64{
				v2.New(2., 3.),
				v2.New(3., 4.),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[vector2.AddToArrayNode[float64]]{
				Data: vector2.AddToArrayNode[float64]{
					Amount: tc.amount,
					Array:  tc.array,
				},
			}
			out := nodes.GetNodeOutputPort[[]v2.Vector[float64]](node, "Out").Value()
			assert.Equal(t, tc.out, out)
		})
	}
}
