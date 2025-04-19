package vector3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type NewNodeData[T vector.Number] struct {
	X nodes.Output[T]
	Y nodes.Output[T]
	Z nodes.Output[T]
}

func (cn NewNodeData[T]) Out() nodes.StructOutput[vector3.Vector[T]] {
	return nodes.NewStructOutput(vector3.New(
		nodes.TryGetOutputValue(cn.X, 0),
		nodes.TryGetOutputValue(cn.Y, 0),
		nodes.TryGetOutputValue(cn.Z, 0),
	))
}

type ArrayFromComponentsNodeData[T vector.Number] struct {
	X nodes.Output[[]T]
	Y nodes.Output[[]T]
	Z nodes.Output[[]T]
}

func (snd ArrayFromComponentsNodeData[T]) Out() nodes.StructOutput[[]vector3.Vector[T]] {
	xArr := nodes.TryGetOutputValue(snd.X, nil)
	yArr := nodes.TryGetOutputValue(snd.Y, nil)
	zArr := nodes.TryGetOutputValue(snd.Z, nil)

	out := make([]vector3.Vector[T], max(len(xArr), len(yArr), len(zArr)))
	for i := range out {
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

		out[i] = vector3.New(x, y, z)
	}

	return nodes.NewStructOutput(out)
}

type ArrayFromNodesNodeData[T vector.Number] struct {
	In []nodes.Output[vector3.Vector[T]]
}

func (node ArrayFromNodesNodeData[T]) Out() nodes.StructOutput[[]vector3.Vector[T]] {
	out := make([]vector3.Vector[T], len(node.In))

	for i, n := range node.In {
		out[i] = nodes.TryGetOutputValue(n, vector3.Zero[T]())
	}

	return nodes.NewStructOutput(out)
}
