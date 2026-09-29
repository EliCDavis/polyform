package vector3

import (
	"fmt"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

// ============================================================================

type Half[T vector.Number] struct {
	In nodes.LiftedPort[vector3.Vector[T]]
}

func (cn Half[T]) Description() string {
	return "Halves a vector."
}

func (cn Half[T]) Float64(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip1(out, cn.In, func(v vector3.Vector[T]) vector3.Float64 {
		return v.ToFloat64().Scale(0.5)
	})
}

func (cn Half[T]) Int(out *nodes.Lifted[vector3.Int]) {
	nodes.Zip1(out, cn.In, func(v vector3.Vector[T]) vector3.Int {
		return v.ToFloat64().Scale(0.5).ToInt()
	})
}

// ============================================================================

type Double[T vector.Number] struct {
	In nodes.LiftedPort[vector3.Vector[T]]
}

func (cn Double[T]) Description() string {
	return "Doubles a vector."
}

func (cn Double[T]) Float64(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip1(out, cn.In, func(v vector3.Vector[T]) vector3.Float64 {
		return v.ToFloat64().Scale(2)
	})
}

func (cn Double[T]) Int(out *nodes.Lifted[vector3.Int]) {
	nodes.Zip1(out, cn.In, func(v vector3.Vector[T]) vector3.Int {
		return v.ToFloat64().Scale(2).ToInt()
	})
}

// ============================================================================

type Length[T vector.Number] struct {
	In nodes.LiftedPort[vector3.Vector[T]]
}

func (cn Length[T]) Description() string {
	return "Length of a vector."
}

func (cn Length[T]) Float64(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, cn.In, func(v vector3.Vector[T]) float64 {
		return v.ToFloat64().Length()
	})
}

func (cn Length[T]) Int(out *nodes.Lifted[int]) {
	nodes.Zip1(out, cn.In, func(v vector3.Vector[T]) int {
		return int(v.ToFloat64().Length())
	})
}

// ============================================================================

type Dot struct {
	A nodes.LiftedPort[vector3.Float64]
	B nodes.LiftedPort[vector3.Float64]
}

func (cn Dot) Description() string {
	return "Dot product of two vectors."
}

func (cn Dot) Dot(out *nodes.Lifted[float64]) {
	nodes.Zip2(out, cn.A, cn.B, vector3.Float64.Dot)
}

func (cn Dot) DotDescription() string {
	return "The dot product of A and B. If either value is not set, then 0 is returned"
}

// ============================================================================

type Inverse[T vector.Number] struct {
	Vector nodes.LiftedPort[vector3.Vector[T]]
}

func (cn Inverse[T]) Description() string {
	return "Inverts a vector, either by negating it (additive) or by dividing one by each component (multiplicative)."
}

func (cn Inverse[T]) additive(in vector3.Vector[T]) vector3.Float64 {
	return in.ToFloat64().Scale(-1)
}

func (cn Inverse[T]) multiplicative(v vector3.Vector[T]) vector3.Float64 {
	in := v.ToFloat64()
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

func (cn Inverse[T]) Additive(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip1(out, cn.Vector, cn.additive)
}

func (cn Inverse[T]) AdditiveInt(out *nodes.Lifted[vector3.Int]) {
	nodes.Zip1(out, cn.Vector, func(v vector3.Vector[T]) vector3.Int {
		return cn.additive(v).RoundToInt()
	})
}

func (cn Inverse[T]) Multiplicative(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip1(out, cn.Vector, cn.multiplicative)
}

func (cn Inverse[T]) MultiplicativeInt(out *nodes.Lifted[vector3.Int]) {
	nodes.Zip1(out, cn.Vector, func(v vector3.Vector[T]) vector3.Int {
		return cn.multiplicative(v).RoundToInt()
	})
}

// ============================================================================

type Normalize struct {
	In nodes.LiftedPort[vector3.Float64]
}

func (cn Normalize) Description() string {
	return "Scales a vector to unit length."
}

func (cn Normalize) Normalized(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip1(out, cn.In, func(v vector3.Float64) vector3.Float64 {
		// Dividing by its own length would make a zero vector NaN.
		if v.Length() == 0 {
			return vector3.Zero[float64]()
		}
		return v.Normalized()
	})
}

func (cn Normalize) NormalizeDescription() string {
	return "Returns the input vector scaled to have a length of 1. (0,0,0) is returned if no vector is provided"
}

// ============================================================================

type NormalizeArray struct {
	In nodes.Output[[]vector3.Float64]
}

func (cn NormalizeArray) Description() string {
	return "Scales a whole array of vectors by the length of the longest one, so their relative lengths survive."
}

func (cn NormalizeArray) Global(out *nodes.StructOutput[[]vector3.Float64]) {
	if cn.In == nil {
		return
	}

	in := nodes.GetOutputValue(out, cn.In)
	if len(in) == 0 {
		return
	}

	maxMagnitude := 0.
	for _, v := range in {
		maxMagnitude = max(maxMagnitude, v.Length())
	}

	if maxMagnitude == 0 {
		out.CaptureError(fmt.Errorf("all vector data has a magnitude of 0"))
		out.Set(in)
		return
	}

	arr := make([]vector3.Float64, len(in))
	for i, v := range in {
		arr[i] = v.DivByConstant(maxMagnitude)
	}
	out.Set(arr)
}

func (cn NormalizeArray) GlobalDescription() string {
	return "Scales each vector by the inverse of the magnitude of the longest vector"
}
