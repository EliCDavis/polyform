package vector4

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector4"
)

type NewNode[T vector.Number] struct {
	X nodes.Output[T]
	Y nodes.Output[T]
	Z nodes.Output[T]
	W nodes.Output[T]
}

func (cn NewNode[T]) Out(out *nodes.StructOutput[vector4.Vector[T]]) {
	out.Set(vector4.New(
		nodes.TryGetOutputValue(out, cn.X, 0),
		nodes.TryGetOutputValue(out, cn.Y, 0),
		nodes.TryGetOutputValue(out, cn.Z, 0),
		nodes.TryGetOutputValue(out, cn.W, 0),
	))
}

type ArrayFromComponentsNode[T vector.Number] struct {
	X nodes.Output[[]T]
	Y nodes.Output[[]T]
	Z nodes.Output[[]T]
	W nodes.Output[[]T]
}

func (snd ArrayFromComponentsNode[T]) Out(out *nodes.StructOutput[[]vector4.Vector[T]]) {
	xArr := nodes.TryGetOutputValue(out, snd.X, nil)
	yArr := nodes.TryGetOutputValue(out, snd.Y, nil)
	zArr := nodes.TryGetOutputValue(out, snd.Z, nil)
	wArr := nodes.TryGetOutputValue(out, snd.W, nil)

	arr := make([]vector4.Vector[T], max(len(xArr), len(yArr), len(zArr), len(wArr)))
	for i := range arr {
		var x T
		var y T
		var z T
		var w T

		if i < len(xArr) {
			x = xArr[i]
		}

		if i < len(yArr) {
			y = yArr[i]
		}

		if i < len(zArr) {
			z = zArr[i]
		}

		if i < len(wArr) {
			w = wArr[i]
		}

		arr[i] = vector4.New(x, y, z, w)
	}

	out.Set(arr)
}

type ArrayFromNodesNode[T vector.Number] struct {
	In []nodes.Output[vector4.Vector[T]]
}

func (node ArrayFromNodesNode[T]) Out(out *nodes.StructOutput[[]vector4.Vector[T]]) {
	out.Set(nodes.GetOutputValues(out, node.In))
}
