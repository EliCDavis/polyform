package math

import (
	gomath "math"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

func methodToArr[T any, G any](in T, arr []T, f func(a, arrI T) G) []G {
	out := make([]G, len(arr))

	for i, v := range arr {
		out[i] = f(in, v)
	}

	return out
}

type SubtractNode[T vector.Number] struct {
	A nodes.Output[T]
	B nodes.Output[T]
}

func (cn SubtractNode[T]) val(out nodes.ExecutionRecorder) T {
	return nodes.TryGetOutputValue(out, cn.A, 0) - nodes.TryGetOutputValue(out, cn.B, 0)
}

func (an SubtractNode[T]) Float(out *nodes.StructOutput[float64]) {
	out.Set(float64(an.val(out)))
}

func (an SubtractNode[T]) Int(out *nodes.StructOutput[int]) {
	out.Set(int(gomath.Round(float64(an.val(out)))))
}

// ============================================================================

type SubtractToArrayNode[T vector.Number] struct {
	In    nodes.Output[T]
	Array nodes.Output[[]T]
}

func (cn SubtractToArrayNode[T]) Differences(out *nodes.StructOutput[[]T]) {
	out.Set(methodToArr(
		nodes.TryGetOutputValue(out, cn.In, 0),
		nodes.TryGetOutputValue(out, cn.Array, nil),
		func(a, b T) T {
			return b - a
		},
	))
}
