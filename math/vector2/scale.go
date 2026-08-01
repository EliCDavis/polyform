package vector2

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector2"
)

type Scale[T vector.Number] struct {
	Vector nodes.Output[vector2.Vector[T]] `description:"The vector to scale"`
	Amount nodes.Output[float64]           `description:"The amount the scale by (defaults to 1.0)"`
}

func (cn Scale[T]) Float64(out *nodes.StructOutput[vector2.Float64]) {
	vec := nodes.TryGetOutputValue(out, cn.Vector, vector2.Zero[T]())
	out.Set(vec.ToFloat64().Scale(nodes.TryGetOutputValue(out, cn.Amount, 1)))
}

func (cn Scale[T]) Int(out *nodes.StructOutput[vector2.Int]) {
	vec := nodes.TryGetOutputValue(out, cn.Vector, vector2.Zero[T]())
	out.Set(vec.ToFloat64().Scale(nodes.TryGetOutputValue(out, cn.Amount, 1)).RoundToInt())
}

type ScaleArray[T vector.Number] struct {
	Vector nodes.Output[[]vector2.Vector[T]] `description:"The vector array to scale"`
	Amount nodes.Output[float64]             `description:"The amount the scale by (defaults to 1.0)"`
}

func (cn ScaleArray[T]) Float64(out *nodes.StructOutput[[]vector2.Float64]) {
	if cn.Vector == nil {
		return
	}

	inV := nodes.GetOutputValue(out, cn.Vector)
	outV := make([]vector2.Float64, len(inV))

	if cn.Amount == nil {
		for i, v := range inV {
			outV[i] = v.ToFloat64()
		}
		out.Set(outV)
		return
	}

	amount := nodes.GetOutputValue(out, cn.Amount)
	for i, v := range inV {
		outV[i] = v.ToFloat64().Scale(amount)
	}
	out.Set(outV)
}

func (cn ScaleArray[T]) Int(out *nodes.StructOutput[[]vector2.Int]) {
	if cn.Vector == nil {
		return
	}

	inV := nodes.GetOutputValue(out, cn.Vector)
	outV := make([]vector2.Int, len(inV))

	if cn.Amount == nil {
		for i, v := range inV {
			outV[i] = v.ToInt()
		}
		out.Set(outV)
		return
	}

	amount := nodes.GetOutputValue(out, cn.Amount)
	for i, v := range inV {
		outV[i] = v.Scale(amount).RoundToInt()
	}
	out.Set(outV)
}
