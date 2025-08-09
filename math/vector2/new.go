package vector2

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector2"
)

type NewNode[T vector.Number] struct {
	X nodes.Output[T]
	Y nodes.Output[T]
}

func (cn NewNode[T]) Out(out *nodes.StructOutput[vector2.Vector[T]]) {
	out.Set(vector2.New(
		nodes.TryGetOutputValue(out, cn.X, 0),
		nodes.TryGetOutputValue(out, cn.Y, 0),
	))
}

type ArrayFromComponentsNode[T vector.Number] struct {
	X nodes.Output[[]T]
	Y nodes.Output[[]T]
}

func (snd ArrayFromComponentsNode[T]) Out(out *nodes.StructOutput[[]vector2.Vector[T]]) {
	xArr := nodes.TryGetOutputValue(out, snd.X, nil)
	yArr := nodes.TryGetOutputValue(out, snd.Y, nil)

	arr := make([]vector2.Vector[T], max(len(xArr), len(yArr)))
	for i := range arr {
		var x T
		var y T

		if i < len(xArr) {
			x = xArr[i]
		}

		if i < len(yArr) {
			y = yArr[i]
		}

		arr[i] = vector2.New(x, y)
	}

	out.Set(arr)
}

type ArrayFromNodesNode[T vector.Number] struct {
	In []nodes.Output[vector2.Vector[T]]
}

func (node ArrayFromNodesNode[T]) Out(out *nodes.StructOutput[[]vector2.Vector[T]]) {
	out.Set(nodes.GetOutputValues(out, node.In))
}
