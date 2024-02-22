package vecn3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type New[T vector.Number] struct {
	nodes.StructData[vector3.Vector[T]]

	X nodes.NodeOutput[T]
	Y nodes.NodeOutput[T]
	Z nodes.NodeOutput[T]
}

func (cn *New[T]) Out() nodes.NodeOutput[vector3.Vector[T]] {
	return &nodes.StructNodeOutput[vector3.Vector[T]]{Definition: cn}
}

func (cn New[T]) Process() (vector3.Vector[T], error) {

	var x T
	if cn.X != nil {
		x = cn.X.Data()
	}

	var y T
	if cn.Y != nil {
		y = cn.Y.Data()
	}

	var z T
	if cn.Z != nil {
		z = cn.Z.Data()
	}

	return vector3.New[T](x, y, z), nil
}
