package vector4

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector4"
)

type Subtract[T vector.Number] struct {
	A nodes.LiftedPort[vector4.Vector[T]]
	B nodes.LiftedPort[vector4.Vector[T]]
}

func (d Subtract[T]) Description() string {
	return "Subtracts vector B from vector A."
}

func (d Subtract[T]) Out(out *nodes.Lifted[vector4.Vector[T]]) {
	nodes.Zip2(out, d.A, d.B, func(a, b vector4.Vector[T]) vector4.Vector[T] {
		return a.Sub(b)
	})
}
