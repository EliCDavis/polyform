package math

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

// ============================================================================

type SumNode[T vector.Number] struct {
	Values []nodes.Output[T] `description:"The nodes to sum"`
}

func (sn SumNode[T]) Out(out *nodes.StructOutput[T]) {
	vals := nodes.GetOutputValues(out, sn.Values)
	var total T
	for _, v := range vals {
		total += v
	}
	out.Set(total)
}

// ============================================================================

type AddToArrayNode[T vector.Number] struct {
	In    nodes.Output[T]
	Array nodes.Output[[]T]
}

func (cn AddToArrayNode[T]) Out(out *nodes.StructOutput[[]T]) {
	out.Set(methodToArr(
		nodes.TryGetOutputValue(out, cn.In, 0),
		nodes.TryGetOutputValue(out, cn.Array, nil),
		func(a, b T) T {
			return a + b
		},
	))
}
