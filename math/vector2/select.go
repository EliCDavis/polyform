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
	out := nodes.StructOutput[T]{}
	out.Set(nodes.TryGetOutputValue(&out, node.In, vector2.Zero[T]()).X())
	return out
}

func (node Select[T]) Y() nodes.StructOutput[T] {
	out := nodes.StructOutput[T]{}
	out.Set(nodes.TryGetOutputValue(&out, node.In, vector2.Zero[T]()).Y())
	return out
}

type SelectArray[T vector.Number] struct {
	In nodes.Output[[]vector2.Vector[T]]
}

func (node SelectArray[T]) X() nodes.StructOutput[[]T] {
	out := nodes.StructOutput[[]T]{}
	in := nodes.TryGetOutputValue(&out, node.In, nil)
	arr := make([]T, len(in))
	for i, v := range in {
		arr[i] = v.X()
	}
	out.Set(arr)
	return out
}

func (node SelectArray[T]) Y() nodes.StructOutput[[]T] {
	out := nodes.StructOutput[[]T]{}
	in := nodes.TryGetOutputValue(&out, node.In, nil)
	arr := make([]T, len(in))
	for i, v := range in {
		arr[i] = v.Y()
	}
	out.Set(arr)
	return out
}
