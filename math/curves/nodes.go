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
	Spline   nodes.Output[Spline]
	Distance nodes.Output[float64] `description:"distance along the spline where the point lies"`
}

func (tn PositionNode) Position(out *nodes.StructOutput[vector3.Float64]) {
	spline := nodes.TryGetOutputValue(out, tn.Spline, nil)
	if spline != nil {
		out.Set(spline.At(nodes.TryGetOutputValue(out, tn.Distance, 0)))
	}
}

type PositionsForArrayNode struct {
	Spline    nodes.Output[Spline]
	Distances nodes.Output[[]float64] `description:"distances along the spline where the points lie"`
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
	Spline nodes.Output[Spline]
}

func (ln LengthNode) Out(out *nodes.StructOutput[float64]) {
	spline := nodes.TryGetOutputValue(out, ln.Spline, nil)
	if spline != nil {
		out.Set(spline.Length())
	}
}

type TangentNode struct {
	Spline   nodes.Output[Spline]
	Distance nodes.Output[float64] `description:"distance the point is along the spline where we take the tangent"`
}

func (tn TangentNode) Tangent(out *nodes.StructOutput[vector3.Float64]) {
	spline := nodes.TryGetOutputValue(out, tn.Spline, nil)
	if spline != nil {
		out.Set(spline.Tangent(nodes.TryGetOutputValue(out, tn.Distance, 0)))
	}
}

type TangentsForArrayNode struct {
	Spline nodes.Output[Spline]
	Times  nodes.Output[[]float64] `description:"distances the points are along the spline where we take the tangents"`
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
