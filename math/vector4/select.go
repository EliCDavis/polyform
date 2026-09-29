package vector4

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector4"
)

type Select[T vector.Number] struct {
	In nodes.LiftedPort[vector4.Vector[T]]
}

func (node Select[T]) Description() string {
	return "Splits a vector into its X, Y, Z and W components."
}

func (node Select[T]) X(out *nodes.Lifted[T]) {
	nodes.Zip1(out, node.In, vector4.Vector[T].X)
}

func (node Select[T]) Y(out *nodes.Lifted[T]) {
	nodes.Zip1(out, node.In, vector4.Vector[T].Y)
}

func (node Select[T]) Z(out *nodes.Lifted[T]) {
	nodes.Zip1(out, node.In, vector4.Vector[T].Z)
}

func (node Select[T]) W(out *nodes.Lifted[T]) {
	nodes.Zip1(out, node.In, vector4.Vector[T].W)
}
