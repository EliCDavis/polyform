package math_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

func TestMaxArrayNode(t *testing.T) {
	tests := map[string]struct {
		in  []float64
		out float64
	}{
		"nil => 0":    {in: []float64{}, out: 0},
		"[-1,1] => 1": {in: []float64{-1, 1}, out: 1},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[math.MaxArrayNode[float64]]{
				Data: math.MaxArrayNode[float64]{
					In: nodes.ConstOutput[[]float64]{Val: tc.in},
				},
			}
			out := nodes.GetNodeOutputPort[float64](node, "Float 64").Value()
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestMinArrayNode(t *testing.T) {
	tests := map[string]struct {
		in  []float64
		out float64
	}{
		"nil => 0":    {in: []float64{}, out: 0},
		"[-1,1] => 1": {in: []float64{-1, 1}, out: -1},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[math.MinArrayNode[float64]]{
				Data: math.MinArrayNode[float64]{
					In: nodes.ConstOutput[[]float64]{Val: tc.in},
				},
			}
			out := nodes.GetNodeOutputPort[float64](node, "Float 64").Value()
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestMinNode(t *testing.T) {
	tests := map[string]struct {
		in  []nodes.Output[float64]
		out float64
	}{
		"nil => 0": {in: nil, out: 0},
		"1 => 1": {
			in: []nodes.Output[float64]{
				nodes.ConstOutput[float64]{Val: 1},
			},
			out: 1,
		},
		"1, 2 => 2": {
			in: []nodes.Output[float64]{
				nodes.ConstOutput[float64]{Val: 1},
				nodes.ConstOutput[float64]{Val: 2},
			},
			out: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[math.MinNode[float64]]{
				Data: math.MinNode[float64]{
					In: tc.in,
				},
			}
			out := nodes.GetNodeOutputPort[float64](node, "Float 64").Value()
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestMaxNode(t *testing.T) {
	tests := map[string]struct {
		in  []nodes.Output[float64]
		out float64
	}{
		"nil => 0": {in: nil, out: 0},
		"1 => 1": {
			in: []nodes.Output[float64]{
				nodes.ConstOutput[float64]{Val: 1},
			},
			out: 1,
		},
		"1, 2 => 2": {
			in: []nodes.Output[float64]{
				nodes.ConstOutput[float64]{Val: 1},
				nodes.ConstOutput[float64]{Val: 2},
			},
			out: 2,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[math.MaxNode[float64]]{
				Data: math.MaxNode[float64]{
					In: tc.in,
				},
			}
			out := nodes.GetNodeOutputPort[float64](node, "Float 64").Value()
			assert.Equal(t, tc.out, out)
		})
	}
}
