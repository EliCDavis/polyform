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

	generator.RegisterTypes(factory)
}

type AddNode = nodes.Struct[AddNodeData]

type AddNodeData struct {
	A nodes.Output[float64]
	B nodes.Output[float64]
}

func (and AddNodeData) Out() nodes.StructOutput[float64] {
	return nodes.NewStructOutput(and.A.Value() + and.B.Value())
}

// ============================================================================
type SumNode = nodes.Struct[SumData[float64]]

type SumData[T vector.Number] struct {
	Values []nodes.Output[T]
}

func (cn SumData[T]) Out() nodes.StructOutput[T] {
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
type DifferenceNode = nodes.Struct[DifferenceData[float64]]

type DifferenceData[T vector.Number] struct {
	A nodes.Output[T]
	B nodes.Output[T]
}

func (cn DifferenceData[T]) Out() nodes.StructOutput[T] {
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
type DivideNode = nodes.Struct[DivideData[float64]]

type DivideData[T vector.Number] struct {
	Dividend nodes.Output[T]
	Divisor  nodes.Output[T]
}

func (cn DivideData[T]) Out() nodes.StructOutput[T] {
	var a T
	var b T

	if cn.Dividend != nil {
		a = cn.Dividend.Value()
	}

	if cn.Divisor != nil {
		b = cn.Divisor.Value()
	}
	return nodes.NewStructOutput(a / b)
}

// ============================================================================
type Multiply = nodes.Struct[MultiplyData[float64]]

type MultiplyData[T vector.Number] struct {
	A nodes.Output[T]
	B nodes.Output[T]
}

func (cn MultiplyData[T]) Out() nodes.StructOutput[T] {
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
type Round = nodes.Struct[RoundData[float64]]

type RoundData[T vector.Number] struct {
	A nodes.Output[T]
}

func (cn RoundData[T]) Out() nodes.StructOutput[int] {
	if cn.A == nil {
		return nodes.NewStructOutput(0)
	}
	return nodes.NewStructOutput(int(math.Round(float64(cn.A.Value()))))
}
