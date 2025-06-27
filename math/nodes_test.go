package math_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

func TestSquareNode(t *testing.T) {
	tests := map[string]struct {
		in  float64
		out float64
	}{
		"1":  {in: 1, out: 1},
		"10": {in: 10, out: 100},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[math.SquareNode]{
				Data: math.SquareNode{
					In: nodes.ConstOutput[float64]{Val: tc.in},
				},
			}
			out := nodes.GetNodeOutputPort[float64](node, "Out").Value()
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestSquareRootNode(t *testing.T) {
	tests := map[string]struct {
		in  float64
		out float64
	}{
		"1":         {in: 1, out: 1},
		"100 => 10": {in: 100, out: 10},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[math.SquareRootNode]{
				Data: math.SquareRootNode{
					In: nodes.ConstOutput[float64]{Val: tc.in},
				},
			}
			out := nodes.GetNodeOutputPort[float64](node, "Out").Value()
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestRemapNode(t *testing.T) {
	tests := map[string]struct {
		value  float64
		inMin  float64
		inMax  float64
		outMin float64
		outMax float64

		result float64
	}{
		"-1,1 => 0, 10": {
			inMin:  -1,
			inMax:  1,
			outMin: 0,
			outMax: 10,
			value:  0,
			result: 5,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[math.RemapNode[float64]]{
				Data: math.RemapNode[float64]{
					InMin:  nodes.ConstOutput[float64]{Val: tc.inMin},
					InMax:  nodes.ConstOutput[float64]{Val: tc.inMax},
					OutMin: nodes.ConstOutput[float64]{Val: tc.outMin},
					OutMax: nodes.ConstOutput[float64]{Val: tc.outMax},
					Value:  nodes.ConstOutput[float64]{Val: tc.value},
				},
			}
			out := nodes.GetNodeOutputPort[float64](node, "Out").Value()
			assert.Equal(t, tc.result, out)
		})
	}
}

func TestRemapToArrayNode(t *testing.T) {
	tests := map[string]struct {
		value  []float64
		inMin  float64
		inMax  float64
		outMin float64
		outMax float64

		result []float64
	}{
		"-1,1 => 0, 10": {
			inMin:  -1,
			inMax:  1,
			outMin: 0,
			outMax: 10,
			value:  []float64{-1, 0, 1},
			result: []float64{0, 5, 10},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[math.RemapToArrayNode[float64]]{
				Data: math.RemapToArrayNode[float64]{
					InMin:  nodes.ConstOutput[float64]{Val: tc.inMin},
					InMax:  nodes.ConstOutput[float64]{Val: tc.inMax},
					OutMin: nodes.ConstOutput[float64]{Val: tc.outMin},
					OutMax: nodes.ConstOutput[float64]{Val: tc.outMax},
					Value:  nodes.ConstOutput[[]float64]{Val: tc.value},
				},
			}
			out := nodes.GetNodeOutputPort[[]float64](node, "Out").Value()
			assert.Equal(t, tc.result, out)
		})
	}
}

func TestDivideNode(t *testing.T) {
	tests := map[string]struct {
		a nodes.Output[float64]
		b nodes.Output[float64]

		result float64
	}{
		"-1 / 1": {
			a:      &nodes.ConstOutput[float64]{Val: -1},
			b:      &nodes.ConstOutput[float64]{Val: 1},
			result: -1,
		},
		"-1 / nil": {
			a:      &nodes.ConstOutput[float64]{Val: -1},
			b:      nil,
			result: 0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[math.DivideNodeData[float64]]{
				Data: math.DivideNodeData[float64]{
					Dividend: tc.a,
					Divisor:  tc.b,
				},
			}
			out := nodes.GetNodeOutputPort[float64](node, "Out").Value()
			assert.Equal(t, tc.result, out)
		})
	}
}
