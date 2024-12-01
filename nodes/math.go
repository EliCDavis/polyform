package nodes

import (
	"math"

	"github.com/EliCDavis/vector"
)

// ============================================================================
type SumNode = StructNode[float64, SumData[float64]]

type SumData[T vector.Number] struct {
	Values []NodeOutput[T]
}

func (cn SumData[T]) Process() (T, error) {
	var total T
	for _, v := range cn.Values {
		total += v.Value()
	}
	return total, nil
}

// ============================================================================
type DifferenceNode = StructNode[float64, DifferenceData[float64]]

type DifferenceData[T vector.Number] struct {
	A NodeOutput[T]
	B NodeOutput[T]
}

func (cn DifferenceData[T]) Process() (T, error) {
	return cn.A.Value() - cn.B.Value(), nil
}

// ============================================================================
type DivideNode = StructNode[float64, DivideData[float64]]

type DivideData[T vector.Number] struct {
	Dividend NodeOutput[T]
	Divisor  NodeOutput[T]
}

func (cn DivideData[T]) Process() (T, error) {
	return cn.Dividend.Value() / cn.Divisor.Value(), nil
}

// ============================================================================
type Multiply = StructNode[float64, MultiplyData[float64]]

type MultiplyData[T vector.Number] struct {
	A NodeOutput[T]
	B NodeOutput[T]
}

func (cn MultiplyData[T]) Process() (T, error) {
	return cn.A.Value() * cn.B.Value(), nil
}

// ============================================================================
type Round = StructNode[int, RoundData[float64]]

type RoundData[T vector.Number] struct {
	A NodeOutput[T]
}

func (cn RoundData[T]) Process() (int, error) {
	if cn.A == nil {
		return 0, nil
	}
	return int(math.Round(float64(cn.A.Value()))), nil
}
