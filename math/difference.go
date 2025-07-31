package math

import (
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

type DifferenceNodeData[T vector.Number] struct {
	A nodes.Output[T]
	B nodes.Output[T]
}

func (cn DifferenceNodeData[T]) Out() nodes.StructOutput[T] {
	out := nodes.StructOutput[T]{}
	out.Set(nodes.TryGetOutputValue(&out, cn.A, 0) - nodes.TryGetOutputValue(&out, cn.B, 0))
	return out
}

// ============================================================================

type DifferencesToArrayNodeData[T vector.Number] struct {
	In    nodes.Output[T]
	Array nodes.Output[[]T]
}

func (cn DifferencesToArrayNodeData[T]) Out() nodes.StructOutput[[]T] {
	out := nodes.StructOutput[[]T]{}
	out.Set(methodToArr(
		nodes.TryGetOutputValue(&out, cn.In, 0),
		nodes.TryGetOutputValue(&out, cn.Array, nil),
		func(a, b T) T {
			return b - a
		},
	))
	return out
}
