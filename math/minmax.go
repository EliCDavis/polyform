package math

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

// ============================================================================

type MinNode[T vector.Number] struct {
	In []nodes.Output[T]
}

func (n MinNode[T]) min(recorder nodes.ExecutionRecorder) T {
	var v T

	in := nodes.GetOutputValues(recorder, n.In)

	set := false
	for _, node := range in {
		if !set {
			set = true
			v = node
			continue
		}
		v = min(v, node)
	}
	return v
}

func (n MinNode[T]) Int(out *nodes.StructOutput[int]) {
	out.Set(int(n.min(out)))
}

func (n MinNode[T]) Float64(out *nodes.StructOutput[float64]) {
	out.Set(float64(n.min(out)))
}

// ============================================================================

type MinArrayNode[T vector.Number] struct {
	In nodes.Output[[]T]
}

func (n MinArrayNode[T]) min(recorder nodes.ExecutionRecorder) T {
	arr := nodes.TryGetOutputValue(recorder, n.In, nil)

	var v T
	if len(arr) == 0 {
		return v
	}

	v = arr[0]
	for i := 1; i < len(arr); i++ {
		v = min(v, arr[i])
	}

	return v
}

func (n MinArrayNode[T]) Int(out *nodes.StructOutput[int]) {
	out.Set(int(n.min(out)))
}

func (n MinArrayNode[T]) Float64(out *nodes.StructOutput[float64]) {
	out.Set(float64(n.min(out)))
}

// ============================================================================

type MaxNode[T vector.Number] struct {
	In []nodes.Output[T]
}

func (n MaxNode[T]) max(recorder nodes.ExecutionRecorder) T {
	var v T

	in := nodes.GetOutputValues(recorder, n.In)

	set := false
	for _, node := range in {
		if !set {
			set = true
			v = node
			continue
		}
		v = max(v, node)
	}
	return v
}

func (n MaxNode[T]) Int(out *nodes.StructOutput[int]) {
	out.Set(int(n.max(out)))
}

func (n MaxNode[T]) Float64(out *nodes.StructOutput[float64]) {
	out.Set(float64(n.max(out)))
}

// ============================================================================

type MaxArrayNode[T vector.Number] struct {
	In nodes.Output[[]T]
}

func (n MaxArrayNode[T]) max(recorder nodes.ExecutionRecorder) T {
	arr := nodes.TryGetOutputValue(recorder, n.In, nil)

	var v T
	if len(arr) == 0 {
		return v
	}

	v = arr[0]
	for i := 1; i < len(arr); i++ {
		v = max(v, arr[i])
	}

	return v
}

func (n MaxArrayNode[T]) Int(out *nodes.StructOutput[int]) {
	out.Set(int(n.max(out)))
}

func (n MaxArrayNode[T]) Float64(out *nodes.StructOutput[float64]) {
	out.Set(float64(n.max(out)))
}
