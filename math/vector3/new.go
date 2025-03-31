package vector3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type New = nodes.Struct[NewNodeData[float64]]

type NewNodeData[T vector.Number] struct {
	X nodes.Output[T]
	Y nodes.Output[T]
	Z nodes.Output[T]
}

func (cn NewNodeData[T]) Out() nodes.StructOutput[vector3.Vector[T]] {
	return nodes.NewStructOutput(vector3.New[T](
		nodes.TryGetOutputValue(cn.X, 0),
		nodes.TryGetOutputValue(cn.Y, 0),
		nodes.TryGetOutputValue(cn.Z, 0),
	))
}

type NewArray = nodes.Struct[NewArrayNodeData]

type NewArrayNodeData struct {
	X nodes.Output[[]float64]
	Y nodes.Output[[]float64]
	Z nodes.Output[[]float64]
}

func (snd NewArrayNodeData) Out() nodes.StructOutput[[]vector3.Float64] {
	xArr := nodes.TryGetOutputValue(snd.X, nil)
	yArr := nodes.TryGetOutputValue(snd.Y, nil)
	zArr := nodes.TryGetOutputValue(snd.Z, nil)

	out := make([]vector3.Float64, max(len(xArr), len(yArr), len(zArr)))
	for i := 0; i < len(out); i++ {
		x := 0.
		y := 0.
		z := 0.

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
