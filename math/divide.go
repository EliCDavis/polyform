package math

import (
	"errors"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

var cantDivideByZeroErr = errors.New("can't divide by 0")

type DivideNodeData[T vector.Number] struct {
	Dividend nodes.Output[T] `description:"the number being divided"`
	Divisor  nodes.Output[T] `description:"number doing the dividing"`
}

func (DivideNodeData[T]) Description() string {
	return "Dividend / Divisor"
}

func (cn DivideNodeData[T]) Out() nodes.StructOutput[T] {
	out := nodes.NewStructOutput[T](0)
	b := nodes.TryGetOutputValue(&out, cn.Divisor, 0)
	if b == 0 {
		out.CaptureError(cantDivideByZeroErr)
		return out
	}

	out.Set(nodes.TryGetOutputValue(&out, cn.Dividend, 0) / b)
	return out
}

// ============================================================================

type DivideToArrayNode[T vector.Number] struct {
	In    nodes.Output[T]
	Array nodes.Output[[]T]
}

func (cn DivideToArrayNode[T]) Out() nodes.StructOutput[[]T] {
	out := nodes.StructOutput[[]T]{}
	arr := nodes.TryGetOutputValue(&out, cn.Array, nil)
	if len(arr) == 0 {
		return out
	}

	b := nodes.TryGetOutputValue(&out, cn.In, 0)

	if b == 0 {
		out.Set(make([]T, len(arr)))
		out.CaptureError(cantDivideByZeroErr)
		return out
	}

	out.Set(methodToArr(
		b, arr,
		func(a, b T) T {
			return b / a
		},
	))
	return out
}
