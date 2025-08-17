package math

import (
	gomath "math"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

type MultiplyNode[T vector.Number] struct {
	Values []nodes.Output[T]
}

func (cn MultiplyNode[T]) val(out nodes.ExecutionRecorder) T {
	vals := nodes.GetOutputValues(out, cn.Values)
	if len(vals) == 0 {
		return 0
	}

	if len(vals) == 1 {
		return vals[0]
	}

	total := vals[0]
	for i := 1; i < len(vals); i++ {
		total *= vals[i]
	}
	return total
}

func (an MultiplyNode[T]) Float(out *nodes.StructOutput[float64]) {
	out.Set(float64(an.val(out)))
}

func (an MultiplyNode[T]) Int(out *nodes.StructOutput[int]) {
	out.Set(int(gomath.Round(float64(an.val(out)))))
}

// ============================================================================

type MultiplyToArrayNode[T vector.Number] struct {
	In    nodes.Output[T]
	Array nodes.Output[[]T]
}

func (cn MultiplyToArrayNode[T]) Products(out *nodes.StructOutput[[]T]) {
	out.Set(methodToArr(
		nodes.TryGetOutputValue(out, cn.In, 0),
		nodes.TryGetOutputValue(out, cn.Array, nil),
		func(a, b T) T {
			return a * b
		},
	))
}
