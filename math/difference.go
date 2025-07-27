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
	return nodes.NewStructOutput(nodes.TryGetOutputValue(cn.A, 0) - nodes.TryGetOutputValue(cn.B, 0))
}

// ============================================================================

type DifferencesToArrayNodeData[T vector.Number] struct {
	In    nodes.Output[T]
	Array nodes.Output[[]T]
}

func (cn DifferencesToArrayNodeData[T]) Out() nodes.StructOutput[[]T] {
	return nodes.NewStructOutput(methodToArr(
		nodes.TryGetOutputValue(cn.In, 0),
		nodes.TryGetOutputValue(cn.Array, nil),
		func(a, b T) T {
			return b - a
		},
	))
}
