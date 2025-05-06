package math

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

type MultiplyNodeData[T vector.Number] struct {
	A nodes.Output[T]
	B nodes.Output[T]
}

func (cn MultiplyNodeData[T]) Out() nodes.StructOutput[T] {
	return nodes.NewStructOutput(nodes.TryGetOutputValue(cn.A, 0) * nodes.TryGetOutputValue(cn.B, 0))
}

// ============================================================================

type MultiplyToArrayNodeData[T vector.Number] struct {
	In    nodes.Output[T]
	Array nodes.Output[[]T]
}

func (cn MultiplyToArrayNodeData[T]) Out() nodes.StructOutput[[]T] {
	return nodes.NewStructOutput(methodToArr(
		nodes.TryGetOutputValue(cn.In, 0),
		nodes.TryGetOutputValue(cn.Array, nil),
		func(a, b T) T {
			return a * b
		},
	))
}
