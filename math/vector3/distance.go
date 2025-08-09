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

func (d Distance[T]) distance(recorder nodes.ExecutionRecorder) float64 {
	a := nodes.TryGetOutputValue(recorder, d.A, vector3.Zero[T]()).ToFloat64()
	b := nodes.TryGetOutputValue(recorder, d.B, vector3.Zero[T]()).ToFloat64()
	return a.Distance(b)
}

func (d Distance[T]) Float64(out *nodes.StructOutput[float64]) {
	out.Set(d.distance(out))
}

func (d Distance[T]) Int(out *nodes.StructOutput[int]) {
	out.Set(int(math.Round(d.distance(out))))
}

// ============================================================================

// Returns an array of floats, representing the distance between A to every element in B
type DistancesToArray[T vector.Number] struct {
	In    nodes.Output[vector3.Vector[T]]
	Array nodes.Output[[]vector3.Vector[T]]
}

func (d DistancesToArray[T]) Distances(out *nodes.StructOutput[[]float64]) {
	a := nodes.TryGetOutputValue(out, d.In, vector3.Zero[T]()).ToFloat64()
	arr := nodes.TryGetOutputValue(out, d.Array, nil)
	result := make([]float64, len(arr))

	for i, v := range arr {
		result[i] = a.Distance(v.ToFloat64())
	}

	out.Set(result)
}

// ============================================================================

// Returns an array of floats, representing the distance between A to every node connected to B
type DistancesToNodes[T vector.Number] struct {
	In    nodes.Output[vector3.Vector[T]]
	Nodes []nodes.Output[vector3.Vector[T]]
}

func (d DistancesToNodes[T]) Distances(out *nodes.StructOutput[[]float64]) {
	a := nodes.TryGetOutputValue(out, d.In, vector3.Zero[T]()).ToFloat64()

	resolvedNodes := nodes.GetOutputValues(out, d.Nodes)
	arr := make([]float64, len(resolvedNodes))
	for i, v := range resolvedNodes {
		arr[i] = a.Distance(v.ToFloat64())
	}

	out.Set(arr)
}

// ============================================================================

// Returns an array of floats, representing distance(a[i], b[i])
type Distances[T vector.Number] struct {
	A nodes.Output[[]vector3.Vector[T]]
	B nodes.Output[[]vector3.Vector[T]]
}

func (d Distances[T]) Distances(out *nodes.StructOutput[[]float64]) {
	a := nodes.TryGetOutputValue(out, d.A, nil)
	b := nodes.TryGetOutputValue(out, d.B, nil)
	result := make([]float64, max(len(a), len(b)))

	for i := range min(len(a), len(b)) {
		result[i] = a[i].ToFloat64().Distance(b[i].ToFloat64())
	}
	out.Set(result)
}
