package vector2

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector2"
)

type Select[T vector.Number] struct {
	In nodes.LiftedPort[vector2.Vector[T]]
}

func (node Select[T]) Description() string {
	return "Splits a vector into its X and Y components."
}

func (node Select[T]) X(out *nodes.Lifted[T]) {
	nodes.Zip1(out, node.In, vector2.Vector[T].X)
}

func (node Select[T]) Y(out *nodes.Lifted[T]) {
	nodes.Zip1(out, node.In, vector2.Vector[T].Y)
}
