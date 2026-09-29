package vector3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type Scale[T vector.Number] struct {
	Vector nodes.LiftedPort[vector3.Vector[T]] `description:"The vector to scale"`
	Amount nodes.LiftedPort[float64]           `description:"The amount the scale by (defaults to 1.0)"`
}

func (cn Scale[T]) Description() string {
	return "Multiplies a vector by a number."
}

func (cn Scale[T]) Float64(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip2(
		out,
		cn.Vector,
		nodes.LiftedOr(cn.Amount, 1.),
		func(v vector3.Vector[T], amount float64) vector3.Float64 {
			return v.ToFloat64().Scale(amount)
		},
	)
}

func (cn Scale[T]) Int(out *nodes.Lifted[vector3.Int]) {
	nodes.Zip2(
		out,
		cn.Vector,
		nodes.LiftedOr(cn.Amount, 1.),
		func(v vector3.Vector[T], amount float64) vector3.Int {
			return v.ToFloat64().Scale(amount).RoundToInt()
		},
	)
}
