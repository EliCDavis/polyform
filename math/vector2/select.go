package vector2

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector2"
)

type Select[T vector.Number] struct {
	In nodes.Output[vector2.Vector[T]]
}

func (node Select[T]) X() nodes.StructOutput[T] {
	if node.In == nil {
		var v T
		return nodes.NewStructOutput(v)
	}

	v := node.In.Value()
	return nodes.NewStructOutput(v.X())
}

func (node Select[T]) Y() nodes.StructOutput[T] {
	if node.In == nil {
		var v T
		return nodes.NewStructOutput(v)
	}

	v := node.In.Value()
	return nodes.NewStructOutput(v.Y())
}

type SelectArray[T vector.Number] struct {
	In nodes.Output[[]vector2.Vector[T]]
}

func (node SelectArray[T]) X() nodes.StructOutput[[]T] {
	in := nodes.TryGetOutputValue(node.In, nil)
	out := make([]T, len(in))
	for i, v := range in {
		out[i] = v.X()
	}
	return nodes.NewStructOutput(out)
}

func (node SelectArray[T]) Y() nodes.StructOutput[[]T] {
	in := nodes.TryGetOutputValue(node.In, nil)
	out := make([]T, len(in))
	for i, v := range in {
		out[i] = v.Y()
	}
	return nodes.NewStructOutput(out)
}
