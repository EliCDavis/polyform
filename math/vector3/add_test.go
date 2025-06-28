package vector3_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/vector3"
	"github.com/EliCDavis/polyform/nodes"
	v3 "github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestSumNodeData(t *testing.T) {
	tests := map[string]struct {
		in  []nodes.Output[v3.Vector[float64]]
		out v3.Vector[float64]
	}{
		"nil => 0": {in: nil, out: v3.Zero[float64]()},
		"[(1,2,3)] => (1,2,3)": {
			in: []nodes.Output[v3.Vector[float64]]{
				nodes.ConstOutput[v3.Vector[float64]]{Val: v3.New(1., 2., 3.)},
			},
			out: v3.New(1., 2., 3.),
		},
		"[(1,2,3), (4,5,6)] => (5,7,9)": {
			in: []nodes.Output[v3.Vector[float64]]{
				nodes.ConstOutput[v3.Vector[float64]]{Val: v3.New(1., 2., 3.)},
				nodes.ConstOutput[v3.Vector[float64]]{Val: v3.New(4., 5., 6.)},
			},
			out: v3.New(5., 7., 9.),
		},
		"[(1,2,3), nil] => (1,2,3)": {
			in: []nodes.Output[v3.Vector[float64]]{
				nodes.ConstOutput[v3.Vector[float64]]{Val: v3.New(1., 2., 3.)},
				nil,
			},
			out: v3.New(1., 2., 3.),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[vector3.SumNodeData[float64]]{
				Data: vector3.SumNodeData[float64]{
					Values: tc.in,
				},
			}
			out := nodes.GetNodeOutputPort[v3.Vector[float64]](node, "Out").Value()
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestAddToArrayNodeData(t *testing.T) {
	tests := map[string]struct {
		amount nodes.Output[v3.Vector[float64]]
		array  nodes.Output[[]v3.Vector[float64]]
		out    []v3.Vector[float64]
	}{
		"(nil + nil) => nil": {amount: nil, array: nil, out: nil},
		"((1,2,3) + nil) => nil": {
			amount: nodes.ConstOutput[v3.Vector[float64]]{Val: v3.New(1., 2., 3.)},
			array:  nil,
			out:    nil,
		},
		"(nil + [(1,2,3)]) => [(1,2,3)]": {
			amount: nil,
			array: nodes.ConstOutput[[]v3.Vector[float64]]{
				Val: []v3.Float64{
					v3.New(1., 2., 3.),
				},
			},
			out: []v3.Float64{
				v3.New(1., 2., 3.),
			},
		},
		"((1,2,3) + [(1,1,1), (2,2,2)]) => [(2,3,4), (3,4,5)]": {
			amount: nodes.ConstOutput[v3.Vector[float64]]{Val: v3.New(1., 2., 3.)},
			array: nodes.ConstOutput[[]v3.Vector[float64]]{
				Val: []v3.Float64{
					v3.New(1., 1., 1.),
					v3.New(2., 2., 2.),
				},
			},
			out: []v3.Float64{
				v3.New(2., 3., 4.),
				v3.New(3., 4., 5.),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[vector3.AddToArrayNodeData[float64]]{
				Data: vector3.AddToArrayNodeData[float64]{
					Amount: tc.amount,
					Array:  tc.array,
				},
			}
			out := nodes.GetNodeOutputPort[[]v3.Vector[float64]](node, "Out").Value()
			assert.Equal(t, tc.out, out)
		})
	}
}
