package math

import (
	"errors"
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

var errNothingToCompare = errors.New("no values to compare")

func pick[T vector.Number](recorder nodes.ExecutionRecorder, vals []T, choose func([]T) T) T {
	if len(vals) == 0 {
		recorder.CaptureError(errNothingToCompare)
		var zero T
		return zero
	}
	return choose(vals)
}

func (n MinNode[T]) Int(out *nodes.Lifted[int]) {
	nodes.ZipAll(out, n.In, func(vals []T) int { return int(pick(out, vals, slices.Min[[]T])) })
}

func (n MinNode[T]) Float64(out *nodes.Lifted[float64]) {
	nodes.ZipAll(out, n.In, func(vals []T) float64 { return float64(pick(out, vals, slices.Min[[]T])) })
}

// ============================================================================

type MinArrayNode[T vector.Number] struct {
	In nodes.Output[[]T]
}

func (n MinArrayNode[T]) Description() string {
	return "Smallest number in an array, reduced to a single value."
}

func (n MinArrayNode[T]) min(recorder nodes.ExecutionRecorder) T {
	return pick(recorder, nodes.TryGetOutputValue(recorder, n.In, nil), slices.Min[[]T])
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

func (n MaxNode[T]) Int(out *nodes.Lifted[int]) {
	nodes.ZipAll(out, n.In, func(vals []T) int { return int(pick(out, vals, slices.Max[[]T])) })
}

func (n MaxNode[T]) Float64(out *nodes.Lifted[float64]) {
	nodes.ZipAll(out, n.In, func(vals []T) float64 { return float64(pick(out, vals, slices.Max[[]T])) })
}

// ============================================================================

type MaxArrayNode[T vector.Number] struct {
	In nodes.Output[[]T]
}

func (n MaxArrayNode[T]) Description() string {
	return "Largest number in an array, reduced to a single value."
}

func (n MaxArrayNode[T]) max(recorder nodes.ExecutionRecorder) T {
	return pick(recorder, nodes.TryGetOutputValue(recorder, n.In, nil), slices.Max[[]T])
}

func (n MaxArrayNode[T]) Int(out *nodes.StructOutput[int]) {
	out.Set(int(n.max(out)))
}

func (n MaxArrayNode[T]) Float64(out *nodes.StructOutput[float64]) {
	out.Set(float64(n.max(out)))
}
