package vector2

import (
	"math"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector2"
)

type Length[T vector.Number] struct {
	In nodes.LiftedPort[vector2.Vector[T]]
}

func (cn Length[T]) Description() string {
	return "Length of a vector."
}

func (cn Length[T]) Float64(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, cn.In, func(v vector2.Vector[T]) float64 {
		return v.ToFloat64().Length()
	})
}

func (cn Length[T]) Int(out *nodes.Lifted[int]) {
	nodes.Zip1(out, cn.In, func(v vector2.Vector[T]) int {
		return int(v.ToFloat64().Length())
	})
}

type Dot struct {
	A nodes.LiftedPort[vector2.Float64]
	B nodes.LiftedPort[vector2.Float64]
}

func (cn Dot) Description() string {
	return "Dot product of two vectors."
}

func (cn Dot) Dot(out *nodes.Lifted[float64]) {
	nodes.Zip2(out, cn.A, cn.B, vector2.Float64.Dot)
}

type Distance[T vector.Number] struct {
	A nodes.LiftedPort[vector2.Vector[T]]
	B nodes.LiftedPort[vector2.Vector[T]]
}

func (d Distance[T]) Description() string {
	return "Distance between two points."
}

func (d Distance[T]) between(a, b vector2.Vector[T]) float64 {
	return a.ToFloat64().Distance(b.ToFloat64())
}

func (d Distance[T]) Float64(out *nodes.Lifted[float64]) {
	nodes.Zip2(out, d.A, d.B, d.between)
}

func (d Distance[T]) Int(out *nodes.Lifted[int]) {
	nodes.Zip2(out, d.A, d.B, func(a, b vector2.Vector[T]) int {
		return int(math.Round(d.between(a, b)))
	})
}
