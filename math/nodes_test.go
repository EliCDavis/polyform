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
				nodetest.AssertOutput("Out", 1.),
			},
		},
		"Square: 10 => 100": {
			node: nodetest.NewNode(math.SquareNode{
				In: nodetest.NewPortValue(10.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Out", 100.),
			},
		},
		"Difference: nil - nil = 0": {
			node: nodetest.NewNode(math.DifferenceNode[float64]{}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Out", 0.),
			},
		},
		"Difference: 1 - nil = 1": {
			node: nodetest.NewNode(math.DifferenceNode[float64]{
				A: nodetest.NewPortValue(1.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Out", 1.),
			},
		},
		"Difference: nil - 1 = -1": {
			node: nodetest.NewNode(math.DifferenceNode[float64]{
				B: nodetest.NewPortValue(1.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Out", -1.),
			},
		},
		"Difference: 1 - 1 = 0": {
			node: nodetest.NewNode(math.DifferenceNode[float64]{
				A: nodetest.NewPortValue(1.),
				B: nodetest.NewPortValue(1.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Out", 0.),
			},
		},
		"Difference: []{1,2,3} - nil = []{1,2,3}": {
			node: nodetest.NewNode(math.DifferencesToArrayNode[float64]{
				Array: nodetest.NewPortValue([]float64{1., 2., 3.}),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Out", []float64{1., 2., 3.}),
			},
		},
		"Difference: []{1,2,3} - 1 = []{0,1,2}": {
			node: nodetest.NewNode(math.DifferencesToArrayNode[float64]{
				In:    nodetest.NewPortValue(1.),
				Array: nodetest.NewPortValue([]float64{1., 2., 3.}),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Out", []float64{0., 1., 2.}),
			},
		},
		"Divide: nil / nil = 0": {
			node: nodetest.NewNode(math.DivideNode[float64]{}),
			assertions: []nodetest.Assertion{
				nodetest.NewAssertInputPortDescription("Dividend", "the number being divided"),
				nodetest.NewAssertInputPortDescription("Divisor", "number doing the dividing"),
				nodetest.AssertNodeDescription{Description: "Dividend / Divisor"},
				nodetest.AssertOutputPortValue[float64]{
					Port:  "Out",
					Value: 0.,
					ExecutionReport: &nodes.ExecutionReport{
						Errors: []string{"can't divide by 0"},
					},
				},
			},
		},
		"Divide: 1 / nil = 0": {
			node: nodetest.NewNode(math.DivideNode[float64]{
				Dividend: nodetest.NewPortValue(1.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutputPortValue[float64]{
					Port:  "Out",
					Value: 0.,
					ExecutionReport: &nodes.ExecutionReport{
						Errors: []string{"can't divide by 0"},
					},
				},
			},
		},
		"Divide: 1 / 2 = 0.5": {
			node: nodetest.NewNode(math.DivideNode[float64]{
				Dividend: nodetest.NewPortValue(1.),
				Divisor:  nodetest.NewPortValue(2.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutputPortValue[float64]{
					Port:            "Out",
					Value:           0.5,
					ExecutionReport: &nodes.ExecutionReport{},
				},
			},
		},
		"Divide: []nil / nil = nil": {
			node: nodetest.NewNode(math.DivideToArrayNode[float64]{}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutputPortValue[[]float64]{
					Port:            "Out",
					Value:           nil,
					ExecutionReport: &nodes.ExecutionReport{},
				},
			},
		},
		"Divide: []{1, 2, 3} / nil = []{0, 0, 0}": {
			node: nodetest.NewNode(math.DivideToArrayNode[float64]{
				Array: nodetest.NewPortValue([]float64{1, 2, 3}),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutputPortValue[[]float64]{
					Port:  "Out",
					Value: []float64{0, 0, 0},
					ExecutionReport: &nodes.ExecutionReport{
						Errors: []string{"can't divide by 0"},
					},
				},
			},
		},
		"Divide: []{1, 2, 4} / 2 = []{0.5, 1, 2}": {
			node: nodetest.NewNode(math.DivideToArrayNode[float64]{
				Array: nodetest.NewPortValue([]float64{1, 2, 4}),
				In:    nodetest.NewPortValue(2.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutputPortValue[[]float64]{
					Port:            "Out",
					Value:           []float64{0.5, 1, 2},
					ExecutionReport: &nodes.ExecutionReport{},
				},
			},
		},
		"Divide: []nil / 2 = []nil": {
			node: nodetest.NewNode(math.DivideToArrayNode[float64]{
				In: nodetest.NewPortValue(2.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutputPortValue[[]float64]{
					Port:            "Out",
					Value:           nil,
					ExecutionReport: &nodes.ExecutionReport{},
				},
			},
		},
		"Inverse Multiplicative(nil) = 0, Additive(nil) = 0": {
			node: nodetest.NewNode(math.InverseNode[float64]{}),
			assertions: []nodetest.Assertion{
				nodetest.NewAssertInputPortDescription("In", "The number to take the inverse of"),
				nodetest.AssertOutputPortValue[float64]{
					Port:            "Additive",
					Value:           0.,
					ExecutionReport: &nodes.ExecutionReport{},
				},
				nodetest.AssertOutputPortValue[float64]{
					Port:  "Multiplicative",
					Value: 0.,
					ExecutionReport: &nodes.ExecutionReport{
						Errors: []string{"can't divide by 0"},
					},
				},
				nodetest.AssertNodeOutputPortDescription{
					Port:        "Additive",
					Description: "The additive inverse of an element x, denoted −x, is the element that when added to x, yields the additive identity, 0",
				},
				nodetest.AssertNodeOutputPortDescription{
					Port:        "Multiplicative",
					Description: "The multiplicative inverse for a number x, denoted by 1/x or x^−1, is a number which when multiplied by x yields the multiplicative identity, 1",
				},
			},
		},
		"Inverse Multiplicative(2) = 1/2, Additive(2) = -2": {
			node: nodetest.NewNode(math.InverseNode[float64]{
				In: nodetest.NewPortValue(2.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutputPortValue[float64]{
					Port:            "Additive",
					Value:           -2.,
					ExecutionReport: &nodes.ExecutionReport{},
				},
				nodetest.AssertOutputPortValue[float64]{
					Port:            "Multiplicative",
					Value:           1 / 2.,
					ExecutionReport: &nodes.ExecutionReport{},
				},
			},
		},
		"Round: nil => 0": {
			node: &nodes.Struct[math.RoundNode]{
				Data: math.RoundNode{},
			},
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Int", 0),
				nodetest.AssertOutput("Float", 0.),
			},
		},
		"Round: 1.23 => 1": {
			node: &nodes.Struct[math.RoundNode]{
				Data: math.RoundNode{
					In: nodes.ConstOutput[float64]{Val: 1.23},
				},
			},
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Int", 1),
				nodetest.AssertOutput("Float", 1.),
			},
		},
		"Circumference: nil => 0": {
			node: nodetest.NewNode(math.CircumferenceNode{}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Int", 0),
				nodetest.AssertOutput("Float", 0.),
			},
		},
		"Circumference: 2 => 4pi": {
			node: nodetest.NewNode(math.CircumferenceNode{
				Radius: nodetest.NewPortValue(2.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Int", 13),
				nodetest.AssertOutput("Float", 4.*gomath.Pi),
				nodetest.AssertNodeDescription{Description: "Circumference of a circle"},
			},
		},
		"One": {
			node: nodetest.NewNode(math.OneNode{}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Int", 1),
				nodetest.AssertOutput("Float 64", 1.),
				nodetest.AssertNodeDescription{Description: "Just the number 1"},
			},
		},
		"Zero": {
			node: nodetest.NewNode(math.ZeroNode{}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Int", 0),
				nodetest.AssertOutput("Float 64", 0.),
				nodetest.AssertNodeDescription{Description: "Just the number 0"},
			},
		},
		"Double: nil => 0": {
			node: nodetest.NewNode(math.DoubleNode[float64]{}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Int", 0),
				nodetest.AssertOutput("Float 64", 0.),
			},
		},
		"Double: 2 => 4": {
			node: nodetest.NewNode(math.DoubleNode[float64]{
				In: nodetest.NewPortValue(2.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Int", 4),
				nodetest.AssertOutput("Float 64", 4.),
				nodetest.AssertNodeDescription{Description: "Doubles the number provided"},
				nodetest.NewAssertInputPortDescription("In", "The number to double"),
			},
		},
		"Half: nil => 0": {
			node: nodetest.NewNode(math.HalfNode[float64]{}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Int", 0),
				nodetest.AssertOutput("Float 64", 0.),
			},
		},
		"Half: 4 => 2": {
			node: nodetest.NewNode(math.HalfNode[float64]{
				In: nodetest.NewPortValue(4.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Int", 2),
				nodetest.AssertOutput("Float 64", 2.),
				nodetest.AssertNodeDescription{Description: "Divides the number in half"},
				nodetest.NewAssertInputPortDescription("In", "The number to halve"),
			},
		},
		"Negate: nil => 0": {
			node: nodetest.NewNode(math.NegateNode[float64]{}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Out", 0.),
				nodetest.NewAssertInputPortDescription("In", "The number to take the additive inverse of"),
				nodetest.AssertNodeDescription{Description: "The additive inverse of an element x, denoted −x, is the element that when added to x, yields the additive identity, 0"},
			},
		},
		"Negate: 4 => -4": {
			node: nodetest.NewNode(math.NegateNode[float64]{
				In: nodetest.NewPortValue(4.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Out", -4.),
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
			node := &nodes.Struct[math.DivideNode[float64]]{
				Data: math.DivideNode[float64]{
					Dividend: tc.a,
					Divisor:  tc.b,
				},
			}
			out := nodes.GetNodeOutputPort[float64](node, "Out").Value()
			assert.Equal(t, tc.result, out)
		})
	}
}
