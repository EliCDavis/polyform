package vector2

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector2"
)

type Subtract[T vector.Number] struct {
	A nodes.LiftedPort[vector2.Vector[T]]
	B nodes.LiftedPort[vector2.Vector[T]]
}

func (d Subtract[T]) Description() string {
	return "Subtracts vector B from vector A."
}

func (d Subtract[T]) Out(out *nodes.Lifted[vector2.Vector[T]]) {
	nodes.Zip2(out, d.A, d.B, func(a, b vector2.Vector[T]) vector2.Vector[T] {
		return a.Sub(b)
	})
}
