package vector3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type Select[T vector.Number] struct {
	In nodes.Output[vector3.Vector[T]]
}

func (node Select[T]) X(out *nodes.StructOutput[T]) {
	out.Set(nodes.TryGetOutputValue(out, node.In, vector3.Zero[T]()).X())
}

func (node Select[T]) Y(out *nodes.StructOutput[T]) {
	out.Set(nodes.TryGetOutputValue(out, node.In, vector3.Zero[T]()).Y())
}

func (node Select[T]) Z(out *nodes.StructOutput[T]) {
	out.Set(nodes.TryGetOutputValue(out, node.In, vector3.Zero[T]()).Z())
}

type SelectArray[T vector.Number] struct {
	In nodes.Output[[]vector3.Vector[T]]
}

func (node SelectArray[T]) arr(out *nodes.StructOutput[[]T], component int) []T {
	in := nodes.TryGetOutputValue(out, node.In, nil)
	arr := make([]T, len(in))
	for i, v := range in {
		arr[i] = v.Component(component)
	}
	return arr
}

func (node SelectArray[T]) X(out *nodes.StructOutput[[]T]) {
	out.Set(node.arr(out, 0))
}

func (node SelectArray[T]) Y(out *nodes.StructOutput[[]T]) {
	out.Set(node.arr(out, 1))
}

func (node SelectArray[T]) Z(out *nodes.StructOutput[[]T]) {
	out.Set(node.arr(out, 2))
}
