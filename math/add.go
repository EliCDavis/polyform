package math

import (
	gomath "math"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

type AddNode[T vector.Number] struct {
	Values []nodes.LiftedPort[T] `description:"The values to sum. Arrays are summed element by element."`
}

func (an AddNode[T]) Description() string {
	return "Adds two or more values together."
}

func (an AddNode[T]) sum(vals []T) T {
	var total T
	for _, v := range vals {
		total += v
	}
	return total
}

func (an AddNode[T]) Float(out *nodes.Lifted[float64]) {
	nodes.ZipAll(out, an.Values, func(vals []T) float64 {
		return float64(an.sum(vals))
	})
}

func (an AddNode[T]) Int(out *nodes.Lifted[int]) {
	nodes.ZipAll(out, an.Values, func(vals []T) int {
		return int(gomath.Round(float64(an.sum(vals))))
	})
}

// ============================================================================

type SumArrayNode[T vector.Number] struct {
	In nodes.Output[[]T]
}

func (SumArrayNode[T]) Description() string {
	return "Sum of every number in an array."
}

func (n SumArrayNode[T]) total(recorder nodes.ExecutionRecorder) T {
	return AddNode[T]{}.sum(nodes.TryGetOutputValue(recorder, n.In, nil))
}

func (n SumArrayNode[T]) Float(out *nodes.StructOutput[float64]) {
	out.Set(float64(n.total(out)))
}

func (n SumArrayNode[T]) Int(out *nodes.StructOutput[int]) {
	out.Set(int(gomath.Round(float64(n.total(out)))))
}
