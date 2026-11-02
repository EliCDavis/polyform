package vector3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type NewNode[T vector.Number] struct {
	X nodes.LiftedPort[T]
	Y nodes.LiftedPort[T]
	Z nodes.LiftedPort[T]
}

func (cn NewNode[T]) Description() string {
	return "Builds a vector3 from its X/Y/Z scalar components."
}

func (cn NewNode[T]) Out(out *nodes.Lifted[vector3.Vector[T]]) {
	nodes.Zip3(
		out,
		nodes.LiftedOr[T](cn.X, 0),
		nodes.LiftedOr[T](cn.Y, 0),
		nodes.LiftedOr[T](cn.Z, 0),
		vector3.New[T],
	)
}
