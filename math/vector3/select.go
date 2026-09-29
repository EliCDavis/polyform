package vector3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type Select[T vector.Number] struct {
	In nodes.LiftedPort[vector3.Vector[T]]
}

func (node Select[T]) Description() string {
	return "Splits a vector into its X, Y and Z components."
}

func (node Select[T]) X(out *nodes.Lifted[T]) {
	nodes.Zip1(out, node.In, vector3.Vector[T].X)
}

func (node Select[T]) Y(out *nodes.Lifted[T]) {
	nodes.Zip1(out, node.In, vector3.Vector[T].Y)
}

func (node Select[T]) Z(out *nodes.Lifted[T]) {
	nodes.Zip1(out, node.In, vector3.Vector[T].Z)
}
