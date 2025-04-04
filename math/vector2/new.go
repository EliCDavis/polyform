package vector2

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector2"
)

type New = nodes.Struct[NewNodeData[float64]]

type NewNodeData[T vector.Number] struct {
	X nodes.Output[T]
	Y nodes.Output[T]
}

func (cn NewNodeData[T]) Out() nodes.StructOutput[vector2.Vector[T]] {
	return nodes.NewStructOutput(vector2.New[T](
		nodes.TryGetOutputValue(cn.X, 0),
		nodes.TryGetOutputValue(cn.Y, 0),
	))
}

type NewArray = nodes.Struct[NewArrayNodeData]

type NewArrayNodeData struct {
	X nodes.Output[[]float64]
	Y nodes.Output[[]float64]
}

func (snd NewArrayNodeData) Out() nodes.StructOutput[[]vector2.Float64] {
	xArr := nodes.TryGetOutputValue(snd.X, nil)
	yArr := nodes.TryGetOutputValue(snd.Y, nil)

	out := make([]vector2.Float64, max(len(xArr), len(yArr)))
	for i := 0; i < len(out); i++ {
		x := 0.
		y := 0.

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
