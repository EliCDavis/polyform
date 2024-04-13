package vecn3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type New = nodes.StructNode[vector3.Float64, NewData[float64]]

type NewData[T vector.Number] struct {
	X nodes.NodeOutput[T]
	Y nodes.NodeOutput[T]
	Z nodes.NodeOutput[T]
}

func (cn NewData[T]) Process() (vector3.Vector[T], error) {

	var x T
	if cn.X != nil {
		x = cn.X.Value()
	}

	var y T
	if cn.Y != nil {
		y = cn.Y.Value()
	}

	var z T
	if cn.Z != nil {
		z = cn.Z.Value()
	}

	return vector3.New[T](x, y, z), nil
}
