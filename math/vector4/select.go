package vector4

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector4"
)

type Select[T vector.Number] struct {
	In nodes.Output[vector4.Vector[T]]
}

func (node Select[T]) X(out *nodes.StructOutput[T]) {
	out.Set(nodes.TryGetOutputValue(out, node.In, vector4.Zero[T]()).X())
}

func (node Select[T]) Y(out *nodes.StructOutput[T]) {
	out.Set(nodes.TryGetOutputValue(out, node.In, vector4.Zero[T]()).Y())
}

func (node Select[T]) Z(out *nodes.StructOutput[T]) {
	out.Set(nodes.TryGetOutputValue(out, node.In, vector4.Zero[T]()).Z())
}

func (node Select[T]) W(out *nodes.StructOutput[T]) {
	out.Set(nodes.TryGetOutputValue(out, node.In, vector4.Zero[T]()).W())
}

type SelectArray[T vector.Number] struct {
	In nodes.Output[[]vector4.Vector[T]]
}

func (node SelectArray[T]) X(out *nodes.StructOutput[[]T]) {
	in := nodes.TryGetOutputValue(out, node.In, nil)
	arr := make([]T, len(in))
	for i, v := range in {
		arr[i] = v.X()
	}
	out.Set(arr)
}

func (node SelectArray[T]) Y(out *nodes.StructOutput[[]T]) {
	in := nodes.TryGetOutputValue(out, node.In, nil)
	arr := make([]T, len(in))
	for i, v := range in {
		arr[i] = v.Y()
	}
	out.Set(arr)
}

func (node SelectArray[T]) Z(out *nodes.StructOutput[[]T]) {
	in := nodes.TryGetOutputValue(out, node.In, nil)
	arr := make([]T, len(in))
	for i, v := range in {
		arr[i] = v.Z()
	}
	out.Set(arr)
}

func (node SelectArray[T]) W(out *nodes.StructOutput[[]T]) {
	in := nodes.TryGetOutputValue(out, node.In, nil)
	arr := make([]T, len(in))
	for i, v := range in {
		arr[i] = v.W()
	}
	out.Set(arr)
}
