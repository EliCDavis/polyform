package vector3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type Select[T vector.Number] struct {
	In nodes.Output[vector3.Vector[T]]
}

func (node Select[T]) X() nodes.StructOutput[T] {
	out := nodes.StructOutput[T]{}
	out.Set(nodes.TryGetOutputValue(&out, node.In, vector3.Zero[T]()).X())
	return out
}

func (node Select[T]) Y() nodes.StructOutput[T] {
	out := nodes.StructOutput[T]{}
	out.Set(nodes.TryGetOutputValue(&out, node.In, vector3.Zero[T]()).Y())
	return out
}

func (node Select[T]) Z() nodes.StructOutput[T] {
	out := nodes.StructOutput[T]{}
	out.Set(nodes.TryGetOutputValue(&out, node.In, vector3.Zero[T]()).Z())
	return out
}

type SelectArray[T vector.Number] struct {
	In nodes.Output[[]vector3.Vector[T]]
}

func (node SelectArray[T]) arr(component int) nodes.StructOutput[[]T] {
	out := nodes.StructOutput[[]T]{}
	in := nodes.TryGetOutputValue(&out, node.In, nil)
	arr := make([]T, len(in))
	for i, v := range in {
		arr[i] = v.Component(component)
	}
	out.Set(arr)
	return out
}

func (node SelectArray[T]) X() nodes.StructOutput[[]T] {
	return node.arr(0)
}

func (node SelectArray[T]) Y() nodes.StructOutput[[]T] {
	return node.arr(1)
}

func (node SelectArray[T]) Z() nodes.StructOutput[[]T] {
	return node.arr(2)
}
