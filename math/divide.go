package math

import (
	"errors"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

var cantDivideByZeroError = errors.New("can't divide by 0")

type DivideNode[T vector.Number] struct {
	Dividend nodes.Output[T] `description:"the number being divided"`
	Divisor  nodes.Output[T] `description:"number doing the dividing"`
}

func (cn DivideNode[T]) Out() nodes.StructOutput[T] {
	b := nodes.TryGetOutputValue(cn.Divisor, 0)
	if b == 0 {
		out := nodes.NewStructOutput[T](0)
		out.CaptureError(cantDivideByZeroError)
		return out
	}

	return nodes.NewStructOutput(nodes.TryGetOutputValue(cn.Dividend, 0) / b)
}

// ============================================================================

type DivideToArrayNode[T vector.Number] struct {
	In    nodes.Output[T]
	Array nodes.Output[[]T]
}

func (cn DivideToArrayNode[T]) Out() nodes.StructOutput[[]T] {
	arr := nodes.TryGetOutputValue(cn.Array, nil)
	b := nodes.TryGetOutputValue(cn.In, 0)

	if b == 0 {
		out := nodes.NewStructOutput(make([]T, len(arr)))
		out.CaptureError(cantDivideByZeroError)
		return out
	}

	return nodes.NewStructOutput(methodToArr(
		b, arr,
		func(a, b T) T {
			return a / b
		},
	))
}
