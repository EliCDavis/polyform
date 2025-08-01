package math

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

type MultiplyNodeData[T vector.Number] struct {
	A nodes.Output[T]
	B nodes.Output[T]
}

func (cn MultiplyNodeData[T]) Out(out *nodes.StructOutput[T]) {
	out.Set(nodes.TryGetOutputValue(out, cn.A, 0) * nodes.TryGetOutputValue(out, cn.B, 0))
}

// ============================================================================

type MultiplyToArrayNodeData[T vector.Number] struct {
	In    nodes.Output[T]
	Array nodes.Output[[]T]
}

func (cn MultiplyToArrayNodeData[T]) Out(out *nodes.StructOutput[[]T]) {
	out.Set(methodToArr(
		nodes.TryGetOutputValue(out, cn.In, 0),
		nodes.TryGetOutputValue(out, cn.Array, nil),
		func(a, b T) T {
			return a * b
		},
	))
}
