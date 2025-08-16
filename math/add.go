package math

import (
	gomath "math"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

// ============================================================================

type AddNode[T vector.Number] struct {
	Values []nodes.Output[T] `description:"The nodes to sum"`
}

func (an AddNode[T]) val(out nodes.ExecutionRecorder) T {
	vals := nodes.GetOutputValues(out, an.Values)
	var total T
	for _, v := range vals {
		total += v
	}
	return total
}

func (an AddNode[T]) Float(out *nodes.StructOutput[float64]) {
	out.Set(float64(an.val(out)))
}

func (an AddNode[T]) Int(out *nodes.StructOutput[int]) {
	out.Set(int(gomath.Round(float64(an.val(out)))))
}

// ============================================================================

type AddToArrayNode[T vector.Number] struct {
	In    nodes.Output[T]
	Array nodes.Output[[]T]
}

func (cn AddToArrayNode[T]) Sums(out *nodes.StructOutput[[]T]) {
	out.Set(methodToArr(
		nodes.TryGetOutputValue(out, cn.In, 0),
		nodes.TryGetOutputValue(out, cn.Array, nil),
		func(a, b T) T {
			return a + b
		},
	))
}
