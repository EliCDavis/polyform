package vector4_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector4"
	"github.com/EliCDavis/polyform/nodes"
	v4 "github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
)

func TestSumNode(t *testing.T) {
	tests := map[string]struct {
		in  []nodes.Output[v4.Vector[float64]]
		out v4.Vector[float64]
	}{
		"nil => 0": {in: nil, out: v4.Zero[float64]()},
		"[(1,2,3,4)] => (1,2,3,4)": {
			in: []nodes.Output[v4.Vector[float64]]{
				nodes.ConstOutput[v4.Vector[float64]]{Val: v4.New(1., 2., 3., 4.)},
			},
			out: v4.New(1., 2., 3., 4.),
		},
		"[(1,2,3,4), (5,6,7,8)] => (6,8,10,12)": {
			in: []nodes.Output[v4.Vector[float64]]{
				nodes.ConstOutput[v4.Vector[float64]]{Val: v4.New(1., 2., 3., 4.)},
				nodes.ConstOutput[v4.Vector[float64]]{Val: v4.New(5., 6., 7., 8.)},
			},
			out: v4.New(6., 8., 10., 12.),
		},
		"[(1,2,3,4), nil] => (1,2,3,4)": {
			in: []nodes.Output[v4.Vector[float64]]{
				nodes.ConstOutput[v4.Vector[float64]]{Val: v4.New(1., 2., 3., 4.)},
				nil,
			},
			out: v4.New(1., 2., 3., 4.),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[vector4.SumNode[float64]]{
				Data: vector4.SumNode[float64]{
					Values: tc.in,
				},
			}
			out := nodes.GetNodeOutputPort[v4.Vector[float64]](node, "Out").Value()
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestAddToArrayNode(t *testing.T) {
	tests := map[string]struct {
		amount nodes.Output[v4.Vector[float64]]
		array  nodes.Output[[]v4.Vector[float64]]
		out    []v4.Vector[float64]
	}{
		"(nil + nil) => nil": {amount: nil, array: nil, out: nil},
		"((1,2,3,4) + nil) => nil": {
			amount: nodes.ConstOutput[v4.Vector[float64]]{Val: v4.New(1., 2., 3., 4.)},
			array:  nil,
			out:    nil,
		},
		"(nil + [(1,2,3,4)]) => [(1,2,3,4)]": {
			amount: nil,
			array: nodes.ConstOutput[[]v4.Vector[float64]]{
				Val: []v4.Float64{
					v4.New(1., 2., 3., 4.),
				},
			},
			out: []v4.Float64{
				v4.New(1., 2., 3., 4.),
			},
		},
		"((1,2,3,4) + [(1,1,1,1), (2,2,2,2)]) => [(2,3,4,5), (3,4,5,6)]": {
			amount: nodes.ConstOutput[v4.Vector[float64]]{Val: v4.New(1., 2., 3., 4.)},
			array: nodes.ConstOutput[[]v4.Vector[float64]]{
				Val: []v4.Float64{
					v4.New(1., 1., 1., 1.),
					v4.New(2., 2., 2., 2.),
				},
			},
			out: []v4.Float64{
				v4.New(2., 3., 4., 5.),
				v4.New(3., 4., 5., 6.),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[vector4.AddToArrayNode[float64]]{
				Data: vector4.AddToArrayNode[float64]{
					Amount: tc.amount,
					Array:  tc.array,
				},
			}
			out := nodes.GetNodeOutputPort[[]v4.Vector[float64]](node, "Out").Value()
			assert.Equal(t, tc.out, out)
		})
	}
}
