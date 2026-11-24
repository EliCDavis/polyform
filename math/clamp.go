package math

import (
	"github.com/EliCDavis/polyform/math/scalar"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

type ClampNode[T vector.Number] struct {
	Value nodes.LiftedPort[T]
	Min   nodes.LiftedPort[T]
	Max   nodes.LiftedPort[T] `default:"1"`
}

func (ClampNode[T]) Description() string {
	return "Holds a number between Min and Max."
}

func (n ClampNode[T]) Out(out *nodes.Lifted[T]) {
	nodes.ZipAll(
		out,
		[]nodes.LiftedPort[T]{
			nodes.LiftedOr[T](n.Value, 0),
			nodes.LiftedOr[T](n.Min, 0),
			nodes.LiftedOr[T](n.Max, 1),
		},
		func(v []T) T { return scalar.Clamp(v[0], min(v[1], v[2]), max(v[1], v[2])) },
	)
}

type SmoothStepNode struct {
	Value nodes.LiftedPort[float64]
	From  nodes.LiftedPort[float64]
	To    nodes.LiftedPort[float64] `default:"1"`
}

func (SmoothStepNode) Description() string {
	return "Eases from 0 at From to 1 at To along a smooth S-curve."
}

func (n SmoothStepNode) Out(out *nodes.Lifted[float64]) {
	nodes.ZipAll(
		out,
		[]nodes.LiftedPort[float64]{
			nodes.LiftedOr[float64](n.Value, 0),
			nodes.LiftedOr[float64](n.From, 0),
			nodes.LiftedOr[float64](n.To, 1),
		},
		func(v []float64) float64 { return scalar.SmoothStep(v[0], v[1], v[2]) },
	)
}
