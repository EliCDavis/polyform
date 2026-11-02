package vector2

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector2"
)

type NewNode[T vector.Number] struct {
	X nodes.LiftedPort[T]
	Y nodes.LiftedPort[T]
}

func (cn NewNode[T]) Description() string {
	return "Builds a vector2 from its X/Y scalar components."
}

func (cn NewNode[T]) Out(out *nodes.Lifted[vector2.Vector[T]]) {
	nodes.Zip2(
		out,
		nodes.LiftedOr[T](cn.X, 0),
		nodes.LiftedOr[T](cn.Y, 0),
		vector2.New[T],
	)
}
