package curves

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[CatmullRomSplineNode]](factory)
	refutil.RegisterType[nodes.Struct[LengthNode]](factory)

	refutil.RegisterType[nodes.Struct[PositionNode]](factory)
	refutil.RegisterType[nodes.Struct[TangentNode]](factory)

	generator.RegisterTypes(factory)
}

// An unwired spline arrives as a nil interface, which a direct call panics on.
func onSpline(spline Spline, at func(Spline) vector3.Float64) vector3.Float64 {
	if spline == nil {
		return vector3.Zero[float64]()
	}
	return at(spline)
}

type PositionNode struct {
	Spline   nodes.LiftedPort[Spline]  `description:"The spline to sample."`
	Distance nodes.LiftedPort[float64] `description:"Distance along the spline, from 0 at the start. Defaults to 0."`
}

func (tn PositionNode) Description() string {
	return "The position at a given distance along a spline."
}

func (tn PositionNode) Position(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip2(out, tn.Spline, tn.Distance, func(spline Spline, distance float64) vector3.Float64 {
		return onSpline(spline, func(s Spline) vector3.Float64 { return s.At(distance) })
	})
}

type LengthNode struct {
	Spline nodes.Output[Spline] `description:"The spline to measure."`
}

func (ln LengthNode) Description() string {
	return "Total arc length of a spline."
}

func (ln LengthNode) Out(out *nodes.StructOutput[float64]) {
	spline := nodes.TryGetOutputValue(out, ln.Spline, nil)
	if spline != nil {
		out.Set(spline.Length())
	}
}

type TangentNode struct {
	Spline   nodes.LiftedPort[Spline]  `description:"The spline to sample."`
	Distance nodes.LiftedPort[float64] `description:"Distance along the spline, from 0 at the start. Defaults to 0."`
}

func (tn TangentNode) Description() string {
	return "The normalized direction a spline is heading at a given distance."
}

func (tn TangentNode) Tangent(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip2(out, tn.Spline, tn.Distance, func(spline Spline, distance float64) vector3.Float64 {
		return onSpline(spline, func(s Spline) vector3.Float64 { return s.Tangent(distance) })
	})
}

