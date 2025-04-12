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

	refutil.RegisterType[SumNode](factory)
	refutil.RegisterType[DifferenceNode](factory)
	refutil.RegisterType[DivideNode](factory)
	refutil.RegisterType[Multiply](factory)
	refutil.RegisterType[Round](factory)
	refutil.RegisterType[nodes.Struct[InverseNodeData[float64]]](factory)
	refutil.RegisterType[nodes.Struct[InverseNodeData[int]]](factory)

	generator.RegisterTypes(factory)
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
type SumNode = nodes.Struct[SumNodeData[float64]]

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
type DifferenceNode = nodes.Struct[DifferenceNodeData[float64]]

type DifferenceNodeData[T vector.Number] struct {
	A nodes.Output[T]
	B nodes.Output[T]
}

func (cn DifferenceNodeData[T]) Out() nodes.StructOutput[T] {
	var a T
	var b T

	if cn.A != nil {
		a = cn.A.Value()
	}

	if cn.B != nil {
		b = cn.B.Value()
	}
	return nodes.NewStructOutput(a - b)
}

// ============================================================================
type DivideNode = nodes.Struct[DivideNodeData[float64]]

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
type Multiply = nodes.Struct[MultiplyNodeData[float64]]

type MultiplyNodeData[T vector.Number] struct {
	A nodes.Output[T]
	B nodes.Output[T]
}

func (cn MultiplyNodeData[T]) Out() nodes.StructOutput[T] {
	var a T
	var b T

	if cn.A != nil {
		a = cn.A.Value()
	}

	if cn.B != nil {
		b = cn.B.Value()
	}

	return nodes.NewStructOutput(a * b)
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
