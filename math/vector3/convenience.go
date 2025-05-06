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

func (cn Half[T]) Float64() nodes.StructOutput[vector3.Vector[float64]] {
	return nodes.NewStructOutput(nodes.TryGetOutputValue(cn.In, vector3.Zero[T]()).ToFloat64().Scale(0.5))
}

func (cn Half[T]) Int() nodes.StructOutput[vector3.Vector[int]] {
	return nodes.NewStructOutput(nodes.TryGetOutputValue(cn.In, vector3.Zero[T]()).Scale(0.5).ToInt())
}

// ============================================================================

type Double[T vector.Number] struct {
	In nodes.Output[vector3.Vector[T]]
}

func (cn Double[T]) Float64() nodes.StructOutput[vector3.Vector[float64]] {
	return nodes.NewStructOutput(nodes.TryGetOutputValue(cn.In, vector3.Zero[T]()).ToFloat64().Scale(2))
}

func (cn Double[T]) Int() nodes.StructOutput[vector3.Vector[int]] {
	return nodes.NewStructOutput(nodes.TryGetOutputValue(cn.In, vector3.Zero[T]()).Scale(2).ToInt())
}

// ============================================================================

type Length[T vector.Number] struct {
	In nodes.Output[vector3.Vector[T]]
}

func (cn Length[T]) Float64() nodes.StructOutput[float64] {
	return nodes.NewStructOutput(nodes.TryGetOutputValue(cn.In, vector3.Zero[T]()).ToFloat64().Length())
}

func (cn Length[T]) Int() nodes.StructOutput[int] {
	return nodes.NewStructOutput(int(nodes.TryGetOutputValue(cn.In, vector3.Zero[T]()).ToFloat64().Length()))
}

// ============================================================================

type Scale[T vector.Number] struct {
	Vector nodes.Output[vector3.Vector[T]] `description:"The vector to scale"`
	Amount nodes.Output[float64]           `description:"The amount the scale by (defaults to 1.0)"`
}

func (cn Scale[T]) result() vector3.Vector[float64] {
	vec := nodes.TryGetOutputValue(cn.Vector, vector3.Zero[T]())

	// TODO: Eeehhhhh. Is a default of 1 good? Does it matter that much?
	return vec.ToFloat64().Scale(nodes.TryGetOutputValue(cn.Amount, 1))
}

func (cn Scale[T]) Float64() nodes.StructOutput[vector3.Vector[float64]] {
	return nodes.NewStructOutput(cn.result())
}

func (cn Scale[T]) Int() nodes.StructOutput[vector3.Vector[int]] {
	return nodes.NewStructOutput(cn.result().RoundToInt())
}

// ============================================================================

type Dot struct {
	A nodes.Output[vector3.Float64]
	B nodes.Output[vector3.Float64]
}

func (cn Dot) Dot() nodes.StructOutput[float64] {
	if cn.A == nil || cn.B == nil {
		return nodes.NewStructOutput(0.)
	}
	return nodes.NewStructOutput(cn.A.Value().Dot(cn.B.Value()))
}

func (cn Dot) DotDescription() string {
	return "The dot product of A and B. If either value is not set, then 0 is returned"
}
