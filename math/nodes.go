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

// ============================================================================
type SumNode = nodes.Struct[float64, SumData[float64]]

type SumData[T vector.Number] struct {
	Values []nodes.NodeOutput[T]
}

func (cn SumData[T]) Process() (T, error) {
	var total T
	for _, v := range cn.Values {
		if v == nil {
			continue
		}
		total += v.Value()
	}
	return total, nil
}

// ============================================================================
type DifferenceNode = nodes.Struct[float64, DifferenceData[float64]]

type DifferenceData[T vector.Number] struct {
	A nodes.NodeOutput[T]
	B nodes.NodeOutput[T]
}

func (cn DifferenceData[T]) Process() (T, error) {
	var a T
	var b T

	if cn.A != nil {
		a = cn.A.Value()
	}

	if cn.B != nil {
		b = cn.B.Value()
	}
	return a - b, nil
}

// ============================================================================
type DivideNode = nodes.Struct[float64, DivideData[float64]]

type DivideData[T vector.Number] struct {
	Dividend nodes.NodeOutput[T]
	Divisor  nodes.NodeOutput[T]
}

func (cn DivideData[T]) Process() (T, error) {
	var a T
	var b T

	if cn.Dividend != nil {
		a = cn.Dividend.Value()
	}

	if cn.Divisor != nil {
		b = cn.Divisor.Value()
	}
	return a / b, nil
}

// ============================================================================
type Multiply = nodes.Struct[float64, MultiplyData[float64]]

type MultiplyData[T vector.Number] struct {
	A nodes.NodeOutput[T]
	B nodes.NodeOutput[T]
}

func (cn MultiplyData[T]) Process() (T, error) {
	var a T
	var b T

	if cn.A != nil {
		a = cn.A.Value()
	}

	if cn.B != nil {
		b = cn.B.Value()
	}

	return a * b, nil
}

// ============================================================================
type Round = nodes.Struct[int, RoundData[float64]]

type RoundData[T vector.Number] struct {
	A nodes.NodeOutput[T]
}

func (cn RoundData[T]) Process() (int, error) {
	if cn.A == nil {
		return 0, nil
	}
	return int(math.Round(float64(cn.A.Value()))), nil
}
