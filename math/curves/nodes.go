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
	refutil.RegisterType[nodes.Struct[PositionsForArrayNode]](factory)

	refutil.RegisterType[nodes.Struct[TangentNode]](factory)
	refutil.RegisterType[nodes.Struct[TangentsForArrayNode]](factory)

	generator.RegisterTypes(factory)
}

type PositionNode struct {
	Spline   nodes.Output[Spline]  `description:"The spline to sample."`
	Distance nodes.Output[float64] `description:"Distance along the spline, from 0 at the start. Defaults to 0."`
}

func (tn PositionNode) Description() string {
	return "The position at a given distance along a spline."
}

func (tn PositionNode) Position(out *nodes.StructOutput[vector3.Float64]) {
	spline := nodes.TryGetOutputValue(out, tn.Spline, nil)
	if spline != nil {
		out.Set(spline.At(nodes.TryGetOutputValue(out, tn.Distance, 0)))
	}
}

type PositionsForArrayNode struct {
	Spline    nodes.Output[Spline]    `description:"The spline to sample."`
	Distances nodes.Output[[]float64] `description:"Distances along the spline to sample, one per output point."`
}

func (tn PositionsForArrayNode) Description() string {
	return "Samples a spline's position at each given distance."
}

func (tn PositionsForArrayNode) Position(out *nodes.StructOutput[[]vector3.Float64]) {
	if tn.Spline == nil || tn.Distances == nil {
		return
	}

	spline := nodes.GetOutputValue(out, tn.Spline)
	if spline == nil {
		return
	}

	times := nodes.GetOutputValue(out, tn.Distances)
	if len(times) == 0 {
		return
	}

	result := make([]vector3.Float64, len(times))
	for i, t := range times {
		result[i] = spline.At(t)
	}
	out.Set(result)
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
	Spline   nodes.Output[Spline]  `description:"The spline to sample."`
	Distance nodes.Output[float64] `description:"Distance along the spline, from 0 at the start. Defaults to 0."`
}

func (tn TangentNode) Description() string {
	return "The normalized direction a spline is heading at a given distance."
}

func (tn TangentNode) Tangent(out *nodes.StructOutput[vector3.Float64]) {
	spline := nodes.TryGetOutputValue(out, tn.Spline, nil)
	if spline != nil {
		out.Set(spline.Tangent(nodes.TryGetOutputValue(out, tn.Distance, 0)))
	}
}

type TangentsForArrayNode struct {
	Spline nodes.Output[Spline]    `description:"The spline to sample."`
	Times  nodes.Output[[]float64] `description:"Distances along the spline to sample, one per output tangent."`
}

func (tn TangentsForArrayNode) Description() string {
	return "Samples a spline's tangent direction at each given distance."
}

func (tn TangentsForArrayNode) Tangents(out *nodes.StructOutput[[]vector3.Float64]) {
	if tn.Spline == nil || tn.Times == nil {
		return
	}

	spline := nodes.GetOutputValue(out, tn.Spline)
	if spline == nil {
		return
	}

	times := nodes.GetOutputValue(out, tn.Times)
	if len(times) == 0 {
		return
	}

	result := make([]vector3.Float64, len(times))
	for i, t := range times {
		result[i] = spline.Tangent(t)
	}
	out.Set(result)
}
