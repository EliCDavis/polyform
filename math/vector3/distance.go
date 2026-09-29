package vector3

import (
	"math"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type Distance[T vector.Number] struct {
	A nodes.LiftedPort[vector3.Vector[T]]
	B nodes.LiftedPort[vector3.Vector[T]]
}

func (d Distance[T]) Description() string {
	return "Distance between two points. One point against an array measures to each of them; two arrays measure pair by pair."
}

func (d Distance[T]) between(a, b vector3.Vector[T]) float64 {
	return a.ToFloat64().Distance(b.ToFloat64())
}

func (d Distance[T]) Float64(out *nodes.Lifted[float64]) {
	nodes.Zip2(out, d.A, d.B, d.between)
}

func (d Distance[T]) Int(out *nodes.Lifted[int]) {
	nodes.Zip2(out, d.A, d.B, func(a, b vector3.Vector[T]) int {
		return int(math.Round(d.between(a, b)))
	})
}
