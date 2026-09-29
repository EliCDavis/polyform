package vector3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type Multiply[T vector.Number] struct {
	A nodes.LiftedPort[vector3.Vector[T]]
	B nodes.LiftedPort[vector3.Vector[T]] `description:"Defaults to (1,1,1)"`
}

func (Multiply[T]) Description() string {
	return "Multiplies two vectors component by component."
}

func (Multiply[T]) Keywords() []string {
	return []string{"component-wise", "hadamard"}
}

func (n Multiply[T]) Float64(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip2(
		out,
		n.A,
		nodes.LiftedOr(n.B, vector3.One[T]()),
		func(a, b vector3.Vector[T]) vector3.Float64 {
			return a.MultByVector(b).ToFloat64()
		},
	)
}

func (n Multiply[T]) Int(out *nodes.Lifted[vector3.Int]) {
	nodes.Zip2(
		out,
		n.A,
		nodes.LiftedOr(n.B, vector3.One[T]()),
		func(a, b vector3.Vector[T]) vector3.Int {
			return a.MultByVector(b).ToFloat64().RoundToInt()
		},
	)
}
