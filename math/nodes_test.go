package math_test

import (
	"testing"

	gomath "math"

	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/nodes/nodetest"
	"github.com/stretchr/testify/assert"
)

func TestNodes(t *testing.T) {
	tests := map[string]struct {
		node       nodes.Node
		assertions []nodetest.Assertion
	}{
		"Square: 1 => 1": {
			node: nodetest.NewNode(math.SquareNode{
				In: nodetest.NewPortValue(1.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.NewAssertPortValue("Out", 1.),
			},
		},
		"Square: 10 => 100": {
			node: nodetest.NewNode(math.SquareNode{
				In: nodetest.NewPortValue(10.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.NewAssertPortValue("Out", 100.),
			},
		},
		"Round: nil => 0": {
			node: &nodes.Struct[math.RoundNodeData]{
				Data: math.RoundNodeData{},
			},
			assertions: []nodetest.Assertion{
				nodetest.NewAssertPortValue("Int", 0),
				nodetest.NewAssertPortValue("Float", 0.),
			},
		},
		"Round: 1.23 => 1": {
			node: &nodes.Struct[math.RoundNodeData]{
				Data: math.RoundNodeData{
					In: nodes.ConstOutput[float64]{Val: 1.23},
				},
			},
			assertions: []nodetest.Assertion{
				nodetest.NewAssertPortValue("Int", 1),
				nodetest.NewAssertPortValue("Float", 1.),
			},
		},
		"Circumference: nil => 0": {
			node: nodetest.NewNode(math.CircumferenceNode{}),
			assertions: []nodetest.Assertion{
				nodetest.NewAssertPortValue("Int", 0),
				nodetest.NewAssertPortValue("Float", 0.),
			},
		},
		"Circumference: 2 => 4pi": {
			node: nodetest.NewNode(math.CircumferenceNode{
				Radius: nodetest.NewPortValue(2.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.NewAssertPortValue("Int", 13),
				nodetest.NewAssertPortValue("Float", 4.*gomath.Pi),
				nodetest.AssertNodeDescription{Description: "Circumference of a circle"},
			},
		},
		"One": {
			node: nodetest.NewNode(math.OneNode{}),
			assertions: []nodetest.Assertion{
				nodetest.NewAssertPortValue("Int", 1),
				nodetest.NewAssertPortValue("Float 64", 1.),
				nodetest.AssertNodeDescription{Description: "Just the number 1"},
			},
		},
		"Zero": {
			node: nodetest.NewNode(math.ZeroNode{}),
			assertions: []nodetest.Assertion{
				nodetest.NewAssertPortValue("Int", 0),
				nodetest.NewAssertPortValue("Float 64", 0.),
				nodetest.AssertNodeDescription{Description: "Just the number 0"},
			},
		},
		"Double: nil => 0": {
			node: nodetest.NewNode(math.DoubleNode[float64]{}),
			assertions: []nodetest.Assertion{
				nodetest.NewAssertPortValue("Int", 0),
				nodetest.NewAssertPortValue("Float 64", 0.),
			},
		},
		"Double: 2 => 4": {
			node: nodetest.NewNode(math.DoubleNode[float64]{
				In: nodetest.NewPortValue(2.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.NewAssertPortValue("Int", 4),
				nodetest.NewAssertPortValue("Float 64", 4.),
				nodetest.AssertNodeDescription{Description: "Doubles the number provided"},
				nodetest.NewAssertInputPortDescription("In", "The number to double"),
			},
		},
		"Half: nil => 0": {
			node: nodetest.NewNode(math.HalfNode[float64]{}),
			assertions: []nodetest.Assertion{
				nodetest.NewAssertPortValue("Int", 0),
				nodetest.NewAssertPortValue("Float 64", 0.),
			},
		},
		"Half: 4 => 2": {
			node: nodetest.NewNode(math.HalfNode[float64]{
				In: nodetest.NewPortValue(4.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.NewAssertPortValue("Int", 2),
				nodetest.NewAssertPortValue("Float 64", 2.),
				nodetest.AssertNodeDescription{Description: "Divides the number in half"},
				nodetest.NewAssertInputPortDescription("In", "The number to halve"),
			},
		},
		"Negate: nil => 0": {
			node: nodetest.NewNode(math.NegateNode[float64]{}),
			assertions: []nodetest.Assertion{
				nodetest.NewAssertPortValue("Out", 0.),
				nodetest.NewAssertInputPortDescription("In", "The number to take the additive inverse of"),
				nodetest.AssertNodeDescription{Description: "The additive inverse of an element x, denoted −x, is the element that when added to x, yields the additive identity, 0"},
			},
		},
		"Negate: 4 => -4": {
			node: nodetest.NewNode(math.NegateNode[float64]{
				In: nodetest.NewPortValue(4.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.NewAssertPortValue("Out", -4.),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for _, assertion := range tc.assertions {
				assertion.Assert(t, tc.node)
			}
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
