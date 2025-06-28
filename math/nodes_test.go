package math_test

import (
	"testing"

	gomath "math"

	"github.com/EliCDavis/polyform/math"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

type NodeAssert interface {
	Assert(t *testing.T, node nodes.Node)
}

type AssertPortValue[T any] struct {
	Port  string
	Value T
}

func (apv AssertPortValue[T]) Assert(t *testing.T, node nodes.Node) {
	out := nodes.GetNodeOutputPort[T](node, apv.Port).Value()
	assert.Equal(t, apv.Value, out)
}

type AssertNodeDescription struct {
	Description string
}

func (apv AssertNodeDescription) Assert(t *testing.T, node nodes.Node) {
	if describable, ok := node.(nodes.Describable); ok {
		assert.Equal(t, apv.Description, describable.Description())
		return
	}
	t.Error("node does not contain a description")
}

type AssertNodeInputPortDescription struct {
	Port        string
	Description string
}

func (apv AssertNodeInputPortDescription) Assert(t *testing.T, node nodes.Node) {
	outputs := node.Inputs()

	port, ok := outputs[apv.Port]
	if !ok {
		t.Error("node does not contain input port", apv.Port)
		return
	}

	describable, ok := port.(nodes.Describable)
	if !ok {
		t.Error("node input port does not contain a description", apv.Port)
		return
	}

	assert.Equal(t, apv.Description, describable.Description())
}

func NewAssertInputPortDescription(port, description string) AssertNodeInputPortDescription {
	return AssertNodeInputPortDescription{
		Port:        port,
		Description: description,
	}
}

func NewAssertPortValue[T any](port string, value T) AssertPortValue[T] {
	return AssertPortValue[T]{
		Port:  port,
		Value: value,
	}
}

func NewNode[T any](data T) nodes.Node {
	return &nodes.Struct[T]{
		Data: data,
	}
}

func NewPortValue[T any](data T) nodes.Output[T] {
	return nodes.ConstOutput[T]{Val: data}
}

func TestNodes(t *testing.T) {
	tests := map[string]struct {
		node       nodes.Node
		assertions []NodeAssert
	}{
		"Square: 1 => 1": {
			node: NewNode(math.SquareNode{
				In: NewPortValue(1.),
			}),
			assertions: []NodeAssert{
				NewAssertPortValue("Out", 1.),
			},
		},
		"Square: 10 => 100": {
			node: NewNode(math.SquareNode{
				In: NewPortValue(10.),
			}),
			assertions: []NodeAssert{
				NewAssertPortValue("Out", 100.),
			},
		},
		"Round: nil => 0": {
			node: &nodes.Struct[math.RoundNodeData]{
				Data: math.RoundNodeData{},
			},
			assertions: []NodeAssert{
				NewAssertPortValue("Int", 0),
				NewAssertPortValue("Float", 0.),
			},
		},
		"Round: 1.23 => 1": {
			node: &nodes.Struct[math.RoundNodeData]{
				Data: math.RoundNodeData{
					In: nodes.ConstOutput[float64]{Val: 1.23},
				},
			},
			assertions: []NodeAssert{
				NewAssertPortValue("Int", 1),
				NewAssertPortValue("Float", 1.),
			},
		},
		"Circumference: nil => 0": {
			node: NewNode(math.CircumferenceNode{}),
			assertions: []NodeAssert{
				NewAssertPortValue("Int", 0),
				NewAssertPortValue("Float", 0.),
			},
		},
		"Circumference: 2 => 4pi": {
			node: NewNode(math.CircumferenceNode{
				Radius: NewPortValue(2.),
			}),
			assertions: []NodeAssert{
				NewAssertPortValue("Int", 13),
				NewAssertPortValue("Float", 4.*gomath.Pi),
				AssertNodeDescription{Description: "Circumference of a circle"},
			},
		},
		"One": {
			node: NewNode(math.OneNode{}),
			assertions: []NodeAssert{
				NewAssertPortValue("Int", 1),
				NewAssertPortValue("Float 64", 1.),
				AssertNodeDescription{Description: "Just the number 1"},
			},
		},
		"Zero": {
			node: NewNode(math.ZeroNode{}),
			assertions: []NodeAssert{
				NewAssertPortValue("Int", 0),
				NewAssertPortValue("Float 64", 0.),
				AssertNodeDescription{Description: "Just the number 0"},
			},
		},
		"Double: nil => 0": {
			node: NewNode(math.DoubleNode[float64]{}),
			assertions: []NodeAssert{
				NewAssertPortValue("Int", 0),
				NewAssertPortValue("Float 64", 0.),
			},
		},
		"Double: 2 => 4": {
			node: NewNode(math.DoubleNode[float64]{
				In: NewPortValue(2.),
			}),
			assertions: []NodeAssert{
				NewAssertPortValue("Int", 4),
				NewAssertPortValue("Float 64", 4.),
				AssertNodeDescription{Description: "Doubles the number provided"},
				NewAssertInputPortDescription("In", "The number to double"),
			},
		},
		"Half: nil => 0": {
			node: NewNode(math.HalfNode[float64]{}),
			assertions: []NodeAssert{
				NewAssertPortValue("Int", 0),
				NewAssertPortValue("Float 64", 0.),
			},
		},
		"Half: 4 => 2": {
			node: NewNode(math.HalfNode[float64]{
				In: NewPortValue(4.),
			}),
			assertions: []NodeAssert{
				NewAssertPortValue("Int", 2),
				NewAssertPortValue("Float 64", 2.),
				AssertNodeDescription{Description: "Divides the number in half"},
				NewAssertInputPortDescription("In", "The number to halve"),
			},
		},
		"Negate: nil => 0": {
			node: NewNode(math.NegateNode[float64]{}),
			assertions: []NodeAssert{
				NewAssertPortValue("Out", 0.),
				NewAssertInputPortDescription("In", "The number to take the additive inverse of"),
				AssertNodeDescription{Description: "The additive inverse of an element x, denoted −x, is the element that when added to x, yields the additive identity, 0"},
			},
		},
		"Negate: 4 => -4": {
			node: NewNode(math.NegateNode[float64]{
				In: NewPortValue(4.),
			}),
			assertions: []NodeAssert{
				NewAssertPortValue("Out", -4.),
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
