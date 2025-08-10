package vector3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type NewNode[T vector.Number] struct {
	X nodes.Output[T]
	Y nodes.Output[T]
	Z nodes.Output[T]
}

func (cn NewNode[T]) Out(out *nodes.StructOutput[vector3.Vector[T]]) {
	out.Set(vector3.New(
		nodes.TryGetOutputValue(out, cn.X, 0),
		nodes.TryGetOutputValue(out, cn.Y, 0),
		nodes.TryGetOutputValue(out, cn.Z, 0),
	))
}

type ArrayFromComponentsNode[T vector.Number] struct {
	X nodes.Output[[]T]
	Y nodes.Output[[]T]
	Z nodes.Output[[]T]
}

func (snd ArrayFromComponentsNode[T]) Out(out *nodes.StructOutput[[]vector3.Vector[T]]) {
	xArr := nodes.TryGetOutputValue(out, snd.X, nil)
	yArr := nodes.TryGetOutputValue(out, snd.Y, nil)
	zArr := nodes.TryGetOutputValue(out, snd.Z, nil)

	arr := make([]vector3.Vector[T], max(len(xArr), len(yArr), len(zArr)))
	for i := range arr {
		var x T
		var y T
		var z T

		if i < len(xArr) {
			x = xArr[i]
		}

		if i < len(yArr) {
			y = yArr[i]
		}

		if i < len(zArr) {
			z = zArr[i]
		}

		arr[i] = vector3.New(x, y, z)
	}

	out.Set(arr)
}

type ArrayFromNodesNode[T vector.Number] struct {
	In []nodes.Output[vector3.Vector[T]]
}

func (node ArrayFromNodesNode[T]) Out(out *nodes.StructOutput[[]vector3.Vector[T]]) {
	out.Set(nodes.GetOutputValues(out, node.In))
}
