package math

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

// ============================================================================

type SumNodeData[T vector.Number] struct {
	Values []nodes.Output[T]
}

func (cn SumNodeData[T]) Out() nodes.StructOutput[T] {
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

type AddToArrayNodeData[T vector.Number] struct {
	In    nodes.Output[T]
	Array nodes.Output[[]T]
}

func (cn AddToArrayNodeData[T]) Out() nodes.StructOutput[[]T] {
	return nodes.NewStructOutput(methodToArr(
		nodes.TryGetOutputValue(cn.In, 0),
		nodes.TryGetOutputValue(cn.Array, nil),
		func(a, b T) T {
			return a + b
		},
	))
}
