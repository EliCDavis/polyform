package math

import (
	"slices"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

// ============================================================================

type MinNode[T vector.Number] struct {
	In []nodes.LiftedPort[T] `description:"The values to compare. Arrays are compared element by element."`
}

func (n MinNode[T]) Description() string {
	return "Smallest of the given numbers."
}

func (n MinNode[T]) smallest(vals []T) T {
	if len(vals) == 0 {
		var zero T
		return zero
	}
	return slices.Min(vals)
}

func (n MinNode[T]) Int(out *nodes.Lifted[int]) {
	nodes.ZipAll(out, n.In, func(vals []T) int { return int(n.smallest(vals)) })
}

func (n MinNode[T]) Float64(out *nodes.Lifted[float64]) {
	nodes.ZipAll(out, n.In, func(vals []T) float64 { return float64(n.smallest(vals)) })
}

// ============================================================================

type MinArrayNode[T vector.Number] struct {
	In nodes.Output[[]T]
}

func (n MinArrayNode[T]) Description() string {
	return "Smallest number in an array, reduced to a single value."
}

func (n MinArrayNode[T]) min(recorder nodes.ExecutionRecorder) T {
	arr := nodes.TryGetOutputValue(recorder, n.In, nil)
	if len(arr) == 0 {
		var zero T
		return zero
	}
	return slices.Min(arr)
}

func (n MinArrayNode[T]) Int(out *nodes.StructOutput[int]) {
	out.Set(int(n.min(out)))
}

func (n MinArrayNode[T]) Float64(out *nodes.StructOutput[float64]) {
	out.Set(float64(n.min(out)))
}

// ============================================================================

type MaxNode[T vector.Number] struct {
	In []nodes.LiftedPort[T] `description:"The values to compare. Arrays are compared element by element."`
}

func (n MaxNode[T]) Description() string {
	return "Largest of the given numbers."
}

func (n MaxNode[T]) largest(vals []T) T {
	if len(vals) == 0 {
		var zero T
		return zero
	}
	return slices.Max(vals)
}

func (n MaxNode[T]) Int(out *nodes.Lifted[int]) {
	nodes.ZipAll(out, n.In, func(vals []T) int { return int(n.largest(vals)) })
}

func (n MaxNode[T]) Float64(out *nodes.Lifted[float64]) {
	nodes.ZipAll(out, n.In, func(vals []T) float64 { return float64(n.largest(vals)) })
}

// ============================================================================

type MaxArrayNode[T vector.Number] struct {
	In nodes.Output[[]T]
}

func (n MaxArrayNode[T]) Description() string {
	return "Largest number in an array, reduced to a single value."
}

func (n MaxArrayNode[T]) max(recorder nodes.ExecutionRecorder) T {
	arr := nodes.TryGetOutputValue(recorder, n.In, nil)
	if len(arr) == 0 {
		var zero T
		return zero
	}
	return slices.Max(arr)
}

func (n MaxArrayNode[T]) Int(out *nodes.StructOutput[int]) {
	out.Set(int(n.max(out)))
}

func (n MaxArrayNode[T]) Float64(out *nodes.StructOutput[float64]) {
	out.Set(float64(n.max(out)))
}
