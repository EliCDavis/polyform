package vector2

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector2"
)

type NewNodeData[T vector.Number] struct {
	X nodes.Output[T]
	Y nodes.Output[T]
}

func (cn NewNodeData[T]) Out(out *nodes.StructOutput[vector2.Vector[T]]) {
	out.Set(vector2.New(
		nodes.TryGetOutputValue(out, cn.X, 0),
		nodes.TryGetOutputValue(out, cn.Y, 0),
	))
}

type ArrayFromComponentsNodeData[T vector.Number] struct {
	X nodes.Output[[]T]
	Y nodes.Output[[]T]
}

func (snd ArrayFromComponentsNodeData[T]) Out(out *nodes.StructOutput[[]vector2.Vector[T]]) {
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

type ArrayFromNodesNodeData[T vector.Number] struct {
	In []nodes.Output[vector2.Vector[T]]
}

func (node ArrayFromNodesNodeData[T]) Out(out *nodes.StructOutput[[]vector2.Vector[T]]) {
	out.Set(nodes.GetOutputValues(out, node.In))
}
