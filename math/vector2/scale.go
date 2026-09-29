package vector2

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector2"
)

type Scale[T vector.Number] struct {
	Vector nodes.LiftedPort[vector2.Vector[T]] `description:"The vector to scale"`
	Amount nodes.LiftedPort[float64]           `description:"The amount the scale by (defaults to 1.0)"`
}

func (cn Scale[T]) Description() string {
	return "Multiplies a vector by a number."
}

func (cn Scale[T]) Float64(out *nodes.Lifted[vector2.Float64]) {
	nodes.Zip2(
		out,
		cn.Vector,
		nodes.LiftedOr(cn.Amount, 1.),
		func(v vector2.Vector[T], amount float64) vector2.Float64 {
			return v.ToFloat64().Scale(amount)
		},
	)
}

func (cn Scale[T]) Int(out *nodes.Lifted[vector2.Int]) {
	nodes.Zip2(
		out,
		cn.Vector,
		nodes.LiftedOr(cn.Amount, 1.),
		func(v vector2.Vector[T], amount float64) vector2.Int {
			return v.ToFloat64().Scale(amount).RoundToInt()
		},
	)
}
