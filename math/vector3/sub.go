package vector3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type Subtract[T vector.Number] struct {
	A nodes.LiftedPort[vector3.Vector[T]]
	B nodes.LiftedPort[vector3.Vector[T]]
}

func (d Subtract[T]) Description() string {
	return "Subtracts vector B from vector A."
}

func (d Subtract[T]) Out(out *nodes.Lifted[vector3.Vector[T]]) {
	nodes.Zip2(out, d.A, d.B, func(a, b vector3.Vector[T]) vector3.Vector[T] {
		return a.Sub(b)
	})
}
