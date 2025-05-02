package math

import (
	"math"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[Round](factory)

	refutil.RegisterType[nodes.Struct[DifferenceNodeData[int]]](factory)
	refutil.RegisterType[nodes.Struct[DifferenceNodeData[float64]]](factory)

	refutil.RegisterType[nodes.Struct[SumNodeData[int]]](factory)
	refutil.RegisterType[nodes.Struct[SumNodeData[float64]]](factory)

	refutil.RegisterType[nodes.Struct[SumArraysNodeData[int]]](factory)
	refutil.RegisterType[nodes.Struct[SumArraysNodeData[float64]]](factory)

	refutil.RegisterType[nodes.Struct[DivideNodeData[int]]](factory)
	refutil.RegisterType[nodes.Struct[DivideNodeData[float64]]](factory)

	refutil.RegisterType[nodes.Struct[MultiplyNodeData[float64]]](factory)
	refutil.RegisterType[nodes.Struct[MultiplyNodeData[int]]](factory)

	refutil.RegisterType[nodes.Struct[InverseNodeData[float64]]](factory)
	refutil.RegisterType[nodes.Struct[InverseNodeData[int]]](factory)

	refutil.RegisterType[nodes.Struct[NegateNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[NegateNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[DoubleNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[DoubleNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[HalfNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[HalfNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[OneNode]](factory)
	refutil.RegisterType[nodes.Struct[ZeroNode]](factory)

	generator.RegisterTypes(factory)
}

// ============================================================================

type OneNode struct{}

func (cn OneNode) Int() nodes.StructOutput[int] {
	return nodes.NewStructOutput(1)
}

func (cn OneNode) Float64() nodes.StructOutput[float64] {
	return nodes.NewStructOutput(1.)
}

func (cn OneNode) Description() string {
	return "Just the number 1"
}

// ============================================================================

type ZeroNode struct{}

func (cn ZeroNode) Int() nodes.StructOutput[int] {
	return nodes.NewStructOutput(0)
}

func (cn ZeroNode) Float64() nodes.StructOutput[float64] {
	return nodes.NewStructOutput(0.)
}

func (cn ZeroNode) Description() string {
	return "Just the number 0"
}

// ============================================================================

type DoubleNode[T vector.Number] struct {
	In nodes.Output[T] `description:"The number to double"`
}

func (cn DoubleNode[T]) Int() nodes.StructOutput[int] {
	return nodes.NewStructOutput(int(nodes.TryGetOutputValue(cn.In, 0)) * 2)
}

func (cn DoubleNode[T]) Float64() nodes.StructOutput[float64] {
	return nodes.NewStructOutput(float64(nodes.TryGetOutputValue(cn.In, 0)) * 2)
}

func (cn DoubleNode[T]) Description() string {
	return "Doubles the number provided"
}

// ============================================================================

type HalfNode[T vector.Number] struct {
	In nodes.Output[T] `description:"The number to halve"`
}

func (cn HalfNode[T]) Int() nodes.StructOutput[int] {
	return nodes.NewStructOutput(int(float64(nodes.TryGetOutputValue(cn.In, 0)) * 0.5))
}

func (cn HalfNode[T]) Float64() nodes.StructOutput[float64] {
	return nodes.NewStructOutput(float64(nodes.TryGetOutputValue(cn.In, 0)) * 0.5)
}

func (cn HalfNode[T]) Description() string {
	return "Divides the number in half"
}

// ============================================================================

type NegateNode[T vector.Number] struct {
	In nodes.Output[T] `description:"The number to take the additive inverse of"`
}

func (cn NegateNode[T]) Out() nodes.StructOutput[T] {
	return nodes.NewStructOutput(nodes.TryGetOutputValue(cn.In, 0) * -1)
}

func (cn NegateNode[T]) Description() string {
	return "The additive inverse of an element x, denoted −x, is the element that when added to x, yields the additive identity, 0"
}

// ============================================================================
type InverseNodeData[T vector.Number] struct {
	In nodes.Output[T] `description:"The number to take the inverse of"`
}

func (cn InverseNodeData[T]) Additive() nodes.StructOutput[T] {
	return nodes.NewStructOutput(nodes.TryGetOutputValue(cn.In, 0) * -1)
}

func (cn InverseNodeData[T]) AdditiveDescription() string {
	return "The additive inverse of an element x, denoted −x, is the element that when added to x, yields the additive identity, 0"
}

func (cn InverseNodeData[T]) Multiplicative() nodes.StructOutput[T] {
	v := nodes.TryGetOutputValue(cn.In, 0)
	if v == 0 {
		var t T
		return nodes.NewStructOutput(t)
	}
	return nodes.NewStructOutput(1. / v)
}

func (cn InverseNodeData[T]) MultiplicativeDescription() string {
	return "The multiplicative inverse for a number x, denoted by 1/x or x^−1, is a number which when multiplied by x yields the multiplicative identity, 1"
}

// ============================================================================

type SumNodeData[T vector.Number] struct {
	Values []nodes.Output[T]
}

func (cn SumNodeData[T]) Out() nodes.StructOutput[T] {
	var total T
	for _, v := range cn.Values {
		if v == nil {
			continue
		}
		total += v.Value()
	}
	return nodes.NewStructOutput(total)
}

// ============================================================================

type SumArraysNodeData[T vector.Number] struct {
	Values []nodes.Output[[]T]
}

func (cn SumArraysNodeData[T]) Out() nodes.StructOutput[[]T] {
	size := 0
	values := make([][]T, 0)
	for _, v := range cn.Values {
		if v == nil {
			continue
		}

		val := v.Value()
		if len(val) == 0 {
			continue
		}

		values = append(values, val)
		size = max(size, len(val))
	}

	total := make([]T, size)
	for _, arrs := range values {
		for i, v := range arrs {
			total[i] = total[i] + v
		}
	}

	return nodes.NewStructOutput(total)
}

// ============================================================================
type DifferenceNodeData[T vector.Number] struct {
	A nodes.Output[T]
	B nodes.Output[T]
}

func (cn DifferenceNodeData[T]) Out() nodes.StructOutput[T] {
	return nodes.NewStructOutput(nodes.TryGetOutputValue(cn.A, 0) - nodes.TryGetOutputValue(cn.B, 0))
}

// ============================================================================

type DivideNodeData[T vector.Number] struct {
	Dividend nodes.Output[T]
	Divisor  nodes.Output[T]
}

func (DivideNodeData[T]) Description() string {
	return "Dividend / Divisor"
}

func (cn DivideNodeData[T]) Out() nodes.StructOutput[T] {
	var empty T
	a := nodes.TryGetOutputValue(cn.Dividend, empty)
	b := nodes.TryGetOutputValue(cn.Divisor, empty)

	// TODO: Eeeeehhhhhhhhhhhhhhhhhhhhh
	if b == 0 {
		return nodes.NewStructOutput(empty)
	}

	return nodes.NewStructOutput(a / b)
}

// ============================================================================

type MultiplyNodeData[T vector.Number] struct {
	A nodes.Output[T]
	B nodes.Output[T]
}

func (cn MultiplyNodeData[T]) Out() nodes.StructOutput[T] {
	return nodes.NewStructOutput(nodes.TryGetOutputValue(cn.A, 0) * nodes.TryGetOutputValue(cn.B, 0))
}

// ============================================================================

type Round = nodes.Struct[RoundNodeData]

type RoundNodeData struct {
	In nodes.Output[float64]
}

func (cn RoundNodeData) Int() nodes.StructOutput[int] {
	if cn.In == nil {
		return nodes.NewStructOutput(0)
	}
	return nodes.NewStructOutput(int(math.Round(cn.In.Value())))
}

func (cn RoundNodeData) Float() nodes.StructOutput[float64] {
	if cn.In == nil {
		return nodes.NewStructOutput(0.)
	}
	return nodes.NewStructOutput(math.Round(cn.In.Value()))
}
