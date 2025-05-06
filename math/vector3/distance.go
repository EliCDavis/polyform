package vector3

import (
	"math"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

// Returns a single float, representing the distance between A and B
type Distance[T vector.Number] struct {
	A nodes.Output[vector3.Vector[T]]
	B nodes.Output[vector3.Vector[T]]
}

func (d Distance[T]) distance() float64 {
	a := nodes.TryGetOutputValue(d.A, vector3.Zero[T]()).ToFloat64()
	b := nodes.TryGetOutputValue(d.B, vector3.Zero[T]()).ToFloat64()
	return a.Distance(b)
}

func (d Distance[T]) Float64() nodes.StructOutput[float64] {
	return nodes.NewStructOutput(d.distance())
}

func (d Distance[T]) Int() nodes.StructOutput[int] {
	return nodes.NewStructOutput(int(math.Round(d.distance())))
}

// ============================================================================

// Returns an array of floats, representing the distance between A to every element in B
type DistancesToArray[T vector.Number] struct {
	In    nodes.Output[vector3.Vector[T]]
	Array nodes.Output[[]vector3.Vector[T]]
}

func (d DistancesToArray[T]) Distances() nodes.StructOutput[[]float64] {
	a := nodes.TryGetOutputValue(d.In, vector3.Zero[T]()).ToFloat64()
	b := nodes.TryGetOutputValue(d.Array, nil)
	out := make([]float64, len(b))

	for i, v := range b {
		out[i] = a.Distance(v.ToFloat64())
	}

	return nodes.NewStructOutput(out)
}

// ============================================================================

// Returns an array of floats, representing the distance between A to every node connected to B
type DistancesToNodes[T vector.Number] struct {
	In    nodes.Output[vector3.Vector[T]]
	Nodes []nodes.Output[vector3.Vector[T]]
}

func (d DistancesToNodes[T]) Distances() nodes.StructOutput[[]float64] {
	a := nodes.TryGetOutputValue(d.In, vector3.Zero[T]()).ToFloat64()
	out := make([]float64, 0, len(d.Nodes))

	for _, v := range d.Nodes {
		if v == nil {
			continue
		}
		out = append(out, a.Distance(v.Value().ToFloat64()))
	}

	return nodes.NewStructOutput(out)
}

// ============================================================================

// Returns an array of floats, representing distance(a[i], b[i])
type Distances[T vector.Number] struct {
	A nodes.Output[[]vector3.Vector[T]]
	B nodes.Output[[]vector3.Vector[T]]
}

func (d Distances[T]) Distances() nodes.StructOutput[[]float64] {
	a := nodes.TryGetOutputValue(d.A, nil)
	b := nodes.TryGetOutputValue(d.B, nil)
	out := make([]float64, max(len(a), len(b)))

	for i := range min(len(a), len(b)) {
		out[i] = a[i].ToFloat64().Distance(b[i].ToFloat64())
	}

	return nodes.NewStructOutput(out)
}
