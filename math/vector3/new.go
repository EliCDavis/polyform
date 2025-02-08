package vector3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type New = nodes.Struct[vector3.Float64, NewNodeData[float64]]

type NewNodeData[T vector.Number] struct {
	X nodes.NodeOutput[T]
	Y nodes.NodeOutput[T]
	Z nodes.NodeOutput[T]
}

func (cn NewNodeData[T]) Process() (vector3.Vector[T], error) {
	return vector3.New[T](
		nodes.TryGetOutputValue(cn.X, 0),
		nodes.TryGetOutputValue(cn.Y, 0),
		nodes.TryGetOutputValue(cn.Z, 0),
	), nil
}

type NewArray = nodes.Struct[[]vector3.Float64, NewArrayNodeData]

type NewArrayNodeData struct {
	X nodes.NodeOutput[[]float64]
	Y nodes.NodeOutput[[]float64]
	Z nodes.NodeOutput[[]float64]
}

func (snd NewArrayNodeData) Process() ([]vector3.Float64, error) {
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

	return out, nil
}
