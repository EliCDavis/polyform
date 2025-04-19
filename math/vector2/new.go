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

func (cn NewNodeData[T]) Out() nodes.StructOutput[vector2.Vector[T]] {
	return nodes.NewStructOutput(vector2.New(
		nodes.TryGetOutputValue(cn.X, 0),
		nodes.TryGetOutputValue(cn.Y, 0),
	))
}

type ArrayFromComponentsNodeData[T vector.Number] struct {
	X nodes.Output[[]T]
	Y nodes.Output[[]T]
}

func (snd ArrayFromComponentsNodeData[T]) Out() nodes.StructOutput[[]vector2.Vector[T]] {
	xArr := nodes.TryGetOutputValue(snd.X, nil)
	yArr := nodes.TryGetOutputValue(snd.Y, nil)

	out := make([]vector2.Vector[T], max(len(xArr), len(yArr)))
	for i := range out {
		var x T
		var y T

		if i < len(xArr) {
			x = xArr[i]
		}

		if i < len(yArr) {
			y = yArr[i]
		}

		out[i] = vector2.New(x, y)
	}

	return nodes.NewStructOutput(out)
}

type ArrayFromNodesNodeData[T vector.Number] struct {
	In []nodes.Output[vector2.Vector[T]]
}

func (node ArrayFromNodesNodeData[T]) Out() nodes.StructOutput[[]vector2.Vector[T]] {
	out := make([]vector2.Vector[T], len(node.In))

	for i, n := range node.In {
		out[i] = nodes.TryGetOutputValue(n, vector2.Zero[T]())
	}

	return nodes.NewStructOutput(out)
}
