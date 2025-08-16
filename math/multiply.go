package math

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

type MultiplyNode[T vector.Number] struct {
	Values []nodes.Output[T]
}

func (cn MultiplyNode[T]) Out(out *nodes.StructOutput[T]) {
	vals := nodes.GetOutputValues(out, cn.Values)
	if len(vals) == 0 {
		return
	}

	if len(vals) == 1 {
		out.Set(vals[0])
		return
	}

	total := vals[0]
	for i := 1; i < len(vals); i++ {
		total *= vals[i]
	}
	out.Set(total)
}

// ============================================================================

type MultiplyToArrayNode[T vector.Number] struct {
	In    nodes.Output[T]
	Array nodes.Output[[]T]
}

func (cn MultiplyToArrayNode[T]) Out(out *nodes.StructOutput[[]T]) {
	out.Set(methodToArr(
		nodes.TryGetOutputValue(out, cn.In, 0),
		nodes.TryGetOutputValue(out, cn.Array, nil),
		func(a, b T) T {
			return a * b
		},
	))
}
