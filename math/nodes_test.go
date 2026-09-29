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
			node: nodetest.NewNode(math.SubtractNode[float64]{}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Float", 0.),
			},
		},
		"Difference: 1 - nil = 1": {
			node: nodetest.NewNode(math.SubtractNode[float64]{
				A: nodetest.NewPortValue(1.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Float", 1.),
			},
		},
		"Difference: nil - 1 = -1": {
			node: nodetest.NewNode(math.SubtractNode[float64]{
				B: nodetest.NewPortValue(1.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Float", -1.),
			},
		},
		"Difference: 1 - 1 = 0": {
			node: nodetest.NewNode(math.SubtractNode[float64]{
				A: nodetest.NewPortValue(1.),
				B: nodetest.NewPortValue(1.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Float", 0.),
			},
		},
		"Difference: []{1,2,3} - nil = []{1,2,3}": {
			node: nodetest.NewNode(math.SubtractNode[float64]{
				A: nodetest.NewPortValue([]float64{1., 2., 3.}),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Float", []float64{1., 2., 3.}),
			},
		},
		"Difference: []{1,2,3} - 1 = []{0,1,2}": {
			node: nodetest.NewNode(math.SubtractNode[float64]{
				A: nodetest.NewPortValue([]float64{1., 2., 3.}),
				B: nodetest.NewPortValue(1.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Float", []float64{0., 1., 2.}),
			},
		},
		"Difference: 1 - []{1,2,3} = []{0,-1,-2}": {
			node: nodetest.NewNode(math.SubtractNode[float64]{
				A: nodetest.NewPortValue(1.),
				B: nodetest.NewPortValue([]float64{1., 2., 3.}),
			}),
			assertions: []nodetest.Assertion{
				// The hand written twin could only broadcast one way round.
				nodetest.AssertOutput("Float", []float64{0., -1., -2.}),
			},
		},
		"Divide: nil / nil = 0": {
			node: nodetest.NewNode(math.DivideNode[float64]{}),
			assertions: []nodetest.Assertion{
				nodetest.NewAssertInputPortDescription("Dividend", "the number being divided"),
				nodetest.NewAssertInputPortDescription("Divisor", "number doing the dividing"),
				nodetest.AssertNodeDescription{Description: "Dividend / Divisor"},
				nodetest.AssertOutputPortValue[float64]{
					Port:  "Float",
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
					Port:  "Float",
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
					Port:            "Float",
					Value:           0.5,
					ExecutionReport: &nodes.ExecutionReport{},
				},
			},
		},
		"Divide: []{1, 2, 3} / nil = []{0, 0, 0}": {
			node: nodetest.NewNode(math.DivideNode[float64]{
				Dividend: nodetest.NewPortValue([]float64{1, 2, 3}),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutputPortValue[[]float64]{
					Port:  "Float",
					Value: []float64{0, 0, 0},
					ExecutionReport: &nodes.ExecutionReport{
						Errors: []string{"can't divide by 0"},
					},
				},
			},
		},
		"Divide: []{1, 2, 4} / 2 = []{0.5, 1, 2}": {
			node: nodetest.NewNode(math.DivideNode[float64]{
				Dividend: nodetest.NewPortValue([]float64{1, 2, 4}),
				Divisor:  nodetest.NewPortValue(2.),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutputPortValue[[]float64]{
					Port:            "Float",
					Value:           []float64{0.5, 1, 2},
					ExecutionReport: &nodes.ExecutionReport{},
				},
				nodetest.AssertOutputType("Float", "[]float64"),
			},
		},
		"Divide: 8 / []{2, 4} = []{4, 2}": {
			node: nodetest.NewNode(math.DivideNode[float64]{
				Dividend: nodetest.NewPortValue(8.),
				Divisor:  nodetest.NewPortValue([]float64{2, 4}),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Float", []float64{4, 2}),
			},
		},
		"Divide: []{1, 2, 3} / []{1, 2} is refused": {
			node: nodetest.NewNode(math.DivideNode[float64]{
				Dividend: nodetest.NewPortValue([]float64{1, 2, 3}),
				Divisor:  nodetest.NewPortValue([]float64{1, 2}),
			}),
			assertions: []nodetest.Assertion{
				nodetest.AssertOutput("Float", []float64{}),
				nodetest.AssertOutputError("Float", "3 and 2"),
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

func remap(value nodes.LiftedPort[float64], inMin, inMax, outMin, outMax float64) nodes.Node {
	return nodetest.NewNode(math.RemapNode[float64]{
		Value:  value,
		InMin:  nodetest.NewPortValue(inMin),
		InMax:  nodetest.NewPortValue(inMax),
		OutMin: nodetest.NewPortValue(outMin),
		OutMax: nodetest.NewPortValue(outMax),
	})
}

func TestRemapNodeOverArrays(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"-1,1 => 0,10",
			remap(nodetest.NewPortValue([]float64{-1, 0, 1}), -1, 1, 0, 10),
			nodetest.AssertOutput("Out", []float64{0, 5, 10}),
			nodetest.AssertOutputType("Out", "[]float64"),
			nodetest.AssertLiftedInput("Value", "float64"),
		),
		nodetest.NewTestCase(
			"-1,1 => 10,0 (descending out range)",
			remap(nodetest.NewPortValue([]float64{-1, 0, 1}), -1, 1, 10, 0),
			nodetest.AssertOutput("Out", []float64{10, 5, 0}),
		),
		nodetest.NewTestCase(
			"clamps to the output range by default",
			remap(nodetest.NewPortValue([]float64{-2, 2}), -1, 1, 0, 10),
			nodetest.AssertOutput("Out", []float64{0, 10}),
		),
		nodetest.NewTestCase(
			"a range end can itself be an array",
			nodetest.NewNode(math.RemapNode[float64]{
				Value:  nodetest.NewPortValue(1.),
				InMin:  nodetest.NewPortValue(0.),
				InMax:  nodetest.NewPortValue(1.),
				OutMin: nodetest.NewPortValue(0.),
				OutMax: nodetest.NewPortValue([]float64{10, 100}),
			}),
			nodetest.AssertOutput("Out", []float64{10, 100}),
		),
		nodetest.NewTestCase(
			"unclamped keeps values outside the range",
			nodetest.NewNode(math.RemapNode[float64]{
				Value:  nodetest.NewPortValue([]float64{-2, 2}),
				InMin:  nodetest.NewPortValue(-1.),
				InMax:  nodetest.NewPortValue(1.),
				OutMin: nodetest.NewPortValue(0.),
				OutMax: nodetest.NewPortValue(10.),
				Clamp:  nodetest.NewPortValue(false),
			}),
			nodetest.AssertOutput("Out", []float64{-5, 15}),
		),
	).Run(t)
}

func TestSquareRootNodeOverArrays(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"scalar stays scalar",
			nodetest.NewNode(math.SquareRootNode{In: nodetest.NewPortValue(100.)}),
			nodetest.AssertOutput("Out", 10.),
			nodetest.AssertOutputType("Out", "float64"),
		),
		nodetest.NewTestCase(
			"array of roots",
			nodetest.NewNode(math.SquareRootNode{In: nodetest.NewPortValue([]float64{1, 4, 9})}),
			nodetest.AssertOutput("Out", []float64{1, 2, 3}),
			nodetest.AssertOutputType("Out", "[]float64"),
		),
	).Run(t)
}

func TestAddAndMultiplyOverArrays(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"Add: 1 + 2 + 4 => 7",
			nodetest.NewNode(math.AddNode[float64]{Values: nodetest.NewPortValues(1., 2., 4.)}),
			nodetest.AssertOutput("Float", 7.),
			nodetest.AssertOutputType("Float", "float64"),
			nodetest.AssertLiftedInput("Values", "float64"),
		),
		nodetest.NewTestCase(
			"Add: []{1,2,3} + 10 => []{11,12,13}",
			nodetest.NewNode(math.AddNode[float64]{Values: []nodes.LiftedPort[float64]{
				nodetest.NewPortValue([]float64{1, 2, 3}),
				nodetest.NewPortValue(10.),
			}}),
			nodetest.AssertOutput("Float", []float64{11, 12, 13}),
			nodetest.AssertOutputType("Float", "[]float64"),
		),
		nodetest.NewTestCase(
			"Add: []{1,2,3} + []{10,20,30} => []{11,22,33}",
			nodetest.NewNode(math.AddNode[float64]{Values: []nodes.LiftedPort[float64]{
				nodetest.NewPortValue([]float64{1, 2, 3}),
				nodetest.NewPortValue([]float64{10, 20, 30}),
			}}),
			nodetest.AssertOutput("Float", []float64{11, 22, 33}),
		),
		nodetest.NewTestCase(
			"Multiply: 2 * 3 * 4 => 24",
			nodetest.NewNode(math.MultiplyNode[float64]{Values: nodetest.NewPortValues(2., 3., 4.)}),
			nodetest.AssertOutput("Float", 24.),
		),
		nodetest.NewTestCase(
			"Multiply: []{1,2,3} * 2 => []{2,4,6}",
			nodetest.NewNode(math.MultiplyNode[float64]{Values: []nodes.LiftedPort[float64]{
				nodetest.NewPortValue([]float64{1, 2, 3}),
				nodetest.NewPortValue(2.),
			}}),
			nodetest.AssertOutput("Float", []float64{2, 4, 6}),
		),
		nodetest.NewTestCase(
			"Multiply: mismatched lengths are refused",
			nodetest.NewNode(math.MultiplyNode[float64]{Values: []nodes.LiftedPort[float64]{
				nodetest.NewPortValue([]float64{1, 2, 3}),
				nodetest.NewPortValue([]float64{1, 2}),
			}}),
			nodetest.AssertOutput("Float", []float64{}),
			nodetest.AssertOutputError("Float", "3 and 2"),
		),
	).Run(t)
}

func TestUnaryNodesFollowTheirInput(t *testing.T) {
	nodetest.NewSuite(
		nodetest.NewTestCase(
			"Abs over an array",
			nodetest.NewNode(math.AbsNode{In: nodetest.NewPortValue([]float64{-1, 2, -3})}),
			nodetest.AssertOutput("Out", []float64{1, 2, 3}),
		),
		nodetest.NewTestCase(
			"Round over an array",
			nodetest.NewNode(math.RoundNode{In: nodetest.NewPortValue([]float64{1.4, 1.5, -1.5})}),
			nodetest.AssertOutput("Float", []float64{1, 2, -2}),
			nodetest.AssertOutput("Down", []float64{1, 1, -2}),
			nodetest.AssertOutput("Up", []float64{2, 2, -1}),
			nodetest.AssertOutput("Int", []int{1, 2, -2}),
		),
		nodetest.NewTestCase(
			"Square over an array",
			nodetest.NewNode(math.SquareNode{In: nodetest.NewPortValue([]float64{2, 3})}),
			nodetest.AssertOutput("Out", []float64{4, 9}),
		),
		nodetest.NewTestCase(
			"Negate over an array",
			nodetest.NewNode(math.NegateNode[float64]{In: nodetest.NewPortValue([]float64{1, -2})}),
			nodetest.AssertOutput("Out", []float64{-1, 2}),
		),
		nodetest.NewTestCase(
			"Not over an array",
			nodetest.NewNode(math.NotNode{In: nodetest.NewPortValue([]bool{true, false})}),
			nodetest.AssertOutput("Out", []bool{false, true}),
			nodetest.AssertLiftedInput("In", "bool"),
		),
		nodetest.NewTestCase(
			"Bool to number over an array",
			nodetest.NewNode(math.BoolToNumberNode{In: nodetest.NewPortValue([]bool{true, false})}),
			nodetest.AssertOutput("Float 64", []float64{1, 0}),
			nodetest.AssertOutput("Int", []int{1, 0}),
		),
		nodetest.NewTestCase(
			"Int to float over an array",
			nodetest.NewNode(math.IntToFloatNode{In: nodetest.NewPortValue([]int{1, 2})}),
			nodetest.AssertOutput("Out", []float64{1, 2}),
		),
		nodetest.NewTestCase(
			"Inverse over an array reports the zero it could not divide by",
			nodetest.NewNode(math.InverseNode[float64]{In: nodetest.NewPortValue([]float64{2, 0})}),
			nodetest.AssertOutput("Multiplicative", []float64{0.5, 0}),
			nodetest.AssertOutputError("Multiplicative", "can't divide by 0"),
			nodetest.AssertOutput("Additive", []float64{-2, 0}),
		),
		nodetest.NewTestCase(
			"Compare over an array",
			nodetest.NewNode(math.CompareNode[float64]{
				A: nodetest.NewPortValue([]float64{1, 2, 3}),
				B: nodetest.NewPortValue(2.),
			}),
			nodetest.AssertOutput("Less", []bool{true, false, false}),
			nodetest.AssertOutput("Equal", []bool{false, true, false}),
			nodetest.AssertOutput("Greater", []bool{false, false, true}),
			nodetest.AssertOutputType("Greater", "[]bool"),
		),
		nodetest.NewTestCase(
			"Modulo over an array",
			nodetest.NewNode(math.ModuloNode{
				A: nodetest.NewPortValue([]float64{-1, 3, 5}),
				B: nodetest.NewPortValue(4.),
			}),
			nodetest.AssertOutput("Out", []float64{3, 3, 1}),
		),
		nodetest.NewTestCase(
			"Hypotenuse over an array",
			nodetest.NewNode(math.HypotenuseNode{
				P: nodetest.NewPortValue([]float64{3, 5}),
				Q: nodetest.NewPortValue([]float64{4, 12}),
			}),
			nodetest.AssertOutput("Out", []float64{5, 13}),
		),
	).Run(t)
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
			out := nodes.GetNodeOutputPort[float64](node, "Float").Value()
			assert.Equal(t, tc.result, out)
		})
	}
}
