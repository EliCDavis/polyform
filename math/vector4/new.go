package vector4

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector4"
)

type NewNode[T vector.Number] struct {
	X nodes.LiftedPort[T] `description:"Defaults to 0."`
	Y nodes.LiftedPort[T] `description:"Defaults to 0."`
	Z nodes.LiftedPort[T] `description:"Defaults to 0."`
	W nodes.LiftedPort[T] `description:"Defaults to 0."`
}

func (cn NewNode[T]) Description() string {
	return "Builds a vector4 from its X/Y/Z/W scalar components."
}

func (cn NewNode[T]) Out(out *nodes.Lifted[vector4.Vector[T]]) {
	nodes.ZipAll(
		out,
		[]nodes.LiftedPort[T]{
			nodes.LiftedOr[T](cn.X, 0),
			nodes.LiftedOr[T](cn.Y, 0),
			nodes.LiftedOr[T](cn.Z, 0),
			nodes.LiftedOr[T](cn.W, 0),
		},
		func(v []T) vector4.Vector[T] {
			return vector4.New(v[0], v[1], v[2], v[3])
		},
	)
}
