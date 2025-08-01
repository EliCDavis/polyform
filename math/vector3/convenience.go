package vector3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

// ============================================================================

type Half[T vector.Number] struct {
	In nodes.Output[vector3.Vector[T]]
}

func (cn Half[T]) Float64(out *nodes.StructOutput[vector3.Float64]) {
	out.Set(nodes.TryGetOutputValue(out, cn.In, vector3.Zero[T]()).ToFloat64().Scale(0.5))
}

func (cn Half[T]) Int(out *nodes.StructOutput[vector3.Int]) {
	out.Set(nodes.TryGetOutputValue(out, cn.In, vector3.Zero[T]()).ToFloat64().Scale(0.5).ToInt())
}

// ============================================================================

type Double[T vector.Number] struct {
	In nodes.Output[vector3.Vector[T]]
}

func (cn Double[T]) Float64(out *nodes.StructOutput[vector3.Float64]) {
	out.Set(nodes.TryGetOutputValue(out, cn.In, vector3.Zero[T]()).ToFloat64().Scale(2))
}

func (cn Double[T]) Int(out *nodes.StructOutput[vector3.Int]) {
	out.Set(nodes.TryGetOutputValue(out, cn.In, vector3.Zero[T]()).ToFloat64().Scale(2).ToInt())
}

// ============================================================================

type Length[T vector.Number] struct {
	In nodes.Output[vector3.Vector[T]]
}

func (cn Length[T]) Float64(out *nodes.StructOutput[float64]) {
	out.Set(nodes.TryGetOutputValue(out, cn.In, vector3.Zero[T]()).ToFloat64().Length())
}

func (cn Length[T]) Int(out *nodes.StructOutput[int]) {
	out.Set(int(nodes.TryGetOutputValue(out, cn.In, vector3.Zero[T]()).ToFloat64().Length()))
}

// ============================================================================

type Scale[T vector.Number] struct {
	Vector nodes.Output[vector3.Vector[T]] `description:"The vector to scale"`
	Amount nodes.Output[float64]           `description:"The amount the scale by (defaults to 1.0)"`
}

func (cn Scale[T]) Float64(out *nodes.StructOutput[vector3.Float64]) {
	vec := nodes.TryGetOutputValue(out, cn.Vector, vector3.Zero[T]())
	out.Set(vec.ToFloat64().Scale(nodes.TryGetOutputValue(out, cn.Amount, 1)))
}

func (cn Scale[T]) Int(out *nodes.StructOutput[vector3.Int]) {
	vec := nodes.TryGetOutputValue(out, cn.Vector, vector3.Zero[T]())
	out.Set(vec.ToFloat64().Scale(nodes.TryGetOutputValue(out, cn.Amount, 1)).RoundToInt())
}

// ============================================================================

type Dot struct {
	A nodes.Output[vector3.Float64]
	B nodes.Output[vector3.Float64]
}

func (cn Dot) Dot(out *nodes.StructOutput[float64]) {
	if cn.A == nil || cn.B == nil {
		return
	}
	out.Set(nodes.GetOutputValue(out, cn.A).Dot(nodes.GetOutputValue(out, cn.B)))
}

func (cn Dot) DotDescription() string {
	return "The dot product of A and B. If either value is not set, then 0 is returned"
}

// ============================================================================

type Inverse[T vector.Number] struct {
	Vector nodes.Output[vector3.Vector[T]]
}

func (cn Inverse[T]) additive(in vector3.Float64) vector3.Float64 {
	return in.ToFloat64().Scale(-1)
}

func (cn Inverse[T]) multiplicative(in vector3.Float64) vector3.Float64 {
	out := vector3.Float64{}
	if in.X() != 0 {
		out = out.SetX(1. / in.X())
	}

	if in.Y() != 0 {
		out = out.SetY(1. / in.Y())
	}

	if in.Z() != 0 {
		out = out.SetZ(1. / in.Z())
	}

	return out
}

func (cn Inverse[T]) Additive(out *nodes.StructOutput[vector3.Float64]) {
	in := nodes.TryGetOutputValue(out, cn.Vector, vector3.Zero[T]())
	out.Set(cn.additive(in.ToFloat64()))
}

func (cn Inverse[T]) AdditiveInt(out *nodes.StructOutput[vector3.Int]) {
	in := nodes.TryGetOutputValue(out, cn.Vector, vector3.Zero[T]())
	out.Set(cn.additive(in.ToFloat64()).RoundToInt())
}

func (cn Inverse[T]) Multiplicative(out *nodes.StructOutput[vector3.Float64]) {
	in := nodes.TryGetOutputValue(out, cn.Vector, vector3.Zero[T]())
	out.Set(cn.multiplicative(in.ToFloat64()))
}

func (cn Inverse[T]) MultiplicativeInt(out *nodes.StructOutput[vector3.Int]) {
	in := nodes.TryGetOutputValue(out, cn.Vector, vector3.Zero[T]())
	out.Set(cn.multiplicative(in.ToFloat64()).RoundToInt())
}

// ============================================================================

type Normalize struct {
	In nodes.Output[vector3.Float64]
}

func (cn Normalize) Normalized(out *nodes.StructOutput[vector3.Float64]) {
	if cn.In != nil {
		out.Set(nodes.GetOutputValue(out, cn.In).Normalized())
	}
}

func (cn Normalize) NormalizeDescription() string {
	return "Returns the input vector scaled to have a length of 1. (0,0,0) is returned if no vector is provided"
}

// ============================================================================

type NormalizeArray struct {
	In nodes.Output[[]vector3.Float64]
}

func (cn NormalizeArray) Normalized(out *nodes.StructOutput[[]vector3.Float64]) {
	if cn.In == nil {
		return
	}

	in := nodes.GetOutputValue(out, cn.In)
	arr := make([]vector3.Float64, len(in))
	for i, v := range in {
		arr[i] = v.Normalized()
	}
	out.Set(arr)
}
