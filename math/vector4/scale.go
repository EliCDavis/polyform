package vector4

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector4"
)

type Scale[T vector.Number] struct {
	Vector nodes.Output[vector4.Vector[T]] `description:"The vector to scale"`
	Amount nodes.Output[float64]           `description:"The amount the scale by (defaults to 1.0)"`
}

func (cn Scale[T]) Float64(out *nodes.StructOutput[vector4.Float64]) {
	vec := nodes.TryGetOutputValue(out, cn.Vector, vector4.Zero[T]())
	out.Set(vec.ToFloat64().Scale(nodes.TryGetOutputValue(out, cn.Amount, 1)))
}

func (cn Scale[T]) Int(out *nodes.StructOutput[vector4.Int]) {
	vec := nodes.TryGetOutputValue(out, cn.Vector, vector4.Zero[T]())
	out.Set(vec.ToFloat64().Scale(nodes.TryGetOutputValue(out, cn.Amount, 1)).RoundToInt())
}
