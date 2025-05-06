package math

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

// ============================================================================

type MinNode[T vector.Number] struct {
	In []nodes.Output[T]
}

func (n MinNode[T]) min() T {
	var v T
	set := false
	for _, node := range n.In {
		if node == nil {
			continue
		}
		if !set {
			set = true
			v = node.Value()
			continue
		}
		v = min(v, node.Value())
	}
	return v
}

func (n MinNode[T]) Int() nodes.StructOutput[int] {
	return nodes.NewStructOutput(int(n.min()))
}

func (n MinNode[T]) Float64() nodes.StructOutput[float64] {
	return nodes.NewStructOutput(float64(n.min()))
}

// ============================================================================

type MinArrayNode[T vector.Number] struct {
	In nodes.Output[[]T]
}

func (n MinArrayNode[T]) min() T {
	arr := nodes.TryGetOutputValue(n.In, nil)

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

func (n MinArrayNode[T]) Int() nodes.StructOutput[int] {
	return nodes.NewStructOutput(int(n.min()))
}

func (n MinArrayNode[T]) Float64() nodes.StructOutput[float64] {
	return nodes.NewStructOutput(float64(n.min()))
}

// ============================================================================

type MaxNode[T vector.Number] struct {
	In []nodes.Output[T]
}

func (n MaxNode[T]) max() T {
	var v T
	set := false
	for _, node := range n.In {
		if node == nil {
			continue
		}
		if !set {
			set = true
			v = node.Value()
			continue
		}
		v = max(v, node.Value())
	}
	return v
}

func (n MaxNode[T]) Int() nodes.StructOutput[int] {
	return nodes.NewStructOutput(int(n.max()))
}

func (n MaxNode[T]) Float64() nodes.StructOutput[float64] {
	return nodes.NewStructOutput(float64(n.max()))
}

// ============================================================================

type MaxArrayNode[T vector.Number] struct {
	In nodes.Output[[]T]
}

func (n MaxArrayNode[T]) max() T {
	arr := nodes.TryGetOutputValue(n.In, nil)

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

func (n MaxArrayNode[T]) Int() nodes.StructOutput[int] {
	return nodes.NewStructOutput(int(n.max()))
}

func (n MaxArrayNode[T]) Float64() nodes.StructOutput[float64] {
	return nodes.NewStructOutput(float64(n.max()))
}
