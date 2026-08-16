package sdf

import (
	"errors"
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func Union(fields ...sample.Vec3ToFloat) sample.Vec3ToFloat {
	if len(fields) == 0 {
		panic("no fields to union")
	}

	if len(fields) == 1 {
		return fields[0]
	}

	if len(fields) == 2 {
		a := fields[0]
		b := fields[1]
		return func(v vector3.Float64) float64 {
			return math.Min(a(v), b(v))
		}
	}

	return func(v vector3.Float64) float64 {
		min := fields[0](v)
		for i := 1; i < len(fields); i++ {
			min = math.Min(min, fields[i](v))
		}
		return min
	}
}

func smoothUnionBlend(a, b, radius float64) float64 {
	h := math.Max(radius-math.Abs(a-b), 0) / radius
	return math.Min(a, b) - (h * h * radius * 0.25)
}

func SmoothUnion(radius float64, fields ...sample.Vec3ToFloat) sample.Vec3ToFloat {
	if len(fields) == 0 {
		panic("no fields to union")
	}

	if len(fields) == 1 {
		return fields[0]
	}

	if radius <= 0 {
		return Union(fields...)
	}

	if len(fields) == 2 {
		a := fields[0]
		b := fields[1]
		return func(v vector3.Float64) float64 {
			return smoothUnionBlend(a(v), b(v), radius)
		}
	}

	return func(v vector3.Float64) float64 {
		min1, min2 := math.Inf(1), math.Inf(1)
		for _, f := range fields {
			val := f(v)
			if val < min1 {
				min1, min2 = val, min1
			} else if val < min2 {
				min2 = val
			}
		}
		return smoothUnionBlend(min1, min2, radius)
	}
}

func Intersect(fields ...sample.Vec3ToFloat) sample.Vec3ToFloat {
	if len(fields) == 0 {
		panic("no fields to intersect")
	}

	if len(fields) == 1 {
		return fields[0]
	}

	return func(v vector3.Float64) float64 {
		max := fields[0](v)
		for i := 1; i < len(fields); i++ {
			max = math.Max(max, fields[i](v))
		}
		return max
	}
}

func Subtract(base, subtraction sample.Vec3ToFloat) sample.Vec3ToFloat {
	return func(f vector3.Float64) float64 {
		return math.Max(base(f), -subtraction(f))
	}
}

func SmoothSubtract(radius float64, base, subtraction sample.Vec3ToFloat) sample.Vec3ToFloat {
	if radius <= 0 {
		return Subtract(base, subtraction)
	}

	return func(f vector3.Float64) float64 {
		d1 := subtraction(f)
		d2 := base(f)
		h := math.Max(0, math.Min(1, 0.5-0.5*(d2+d1)/radius))
		return (d2*(1-h) + (-d1)*h) + radius*h*(1-h)
	}
}

type UnionNode struct {
	Fields []nodes.Output[sample.Vec3ToFloat] `description:"The fields to combine."`
}

func (n UnionNode) Description() string {
	return "Boolean union of two or more SDF fields."
}

func (n UnionNode) Union(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	fields := nodes.GetOutputValues(out, n.Fields)
	if len(fields) == 0 {
		out.CaptureError(errors.New("No fields provided to union"))
		return
	}
	out.Set(Union(fields...))
}

type SmoothUnionNode struct {
	Fields []nodes.Output[sample.Vec3ToFloat] `description:"The fields to combine."`
	Radius nodes.Output[float64]              `description:"Width of the blend region in world units. Zero or less is a regular union. Defaults to 0.1."`
}

func (n SmoothUnionNode) Description() string {
	return "Combines two or more SDF fields, blending them smoothly into each other where they meet."
}

func (n SmoothUnionNode) Union(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	fields := nodes.GetOutputValues(out, n.Fields)
	if len(fields) == 0 {
		out.CaptureError(errors.New("No fields provided to union"))
		return
	}
	out.Set(SmoothUnion(nodes.TryGetOutputValue(out, n.Radius, .1), fields...))
}

type IntersectionNode struct {
	Fields []nodes.Output[sample.Vec3ToFloat] `description:"The fields to combine."`
}

func (n IntersectionNode) Description() string {
	return "Boolean intersection of two or more SDF fields."
}

func (n IntersectionNode) Intersection(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	fields := nodes.GetOutputValues(out, n.Fields)
	if len(fields) == 0 {
		out.CaptureError(errors.New("No fields provided to intersect"))
		return
	}
	out.Set(Intersect(fields...))
}

type SubtractionNode struct {
	A nodes.Output[sample.Vec3ToFloat] `description:"The base shape. If unset, B passes through unmodified (not negated)."`
	B nodes.Output[sample.Vec3ToFloat] `description:"The shape carved out of A. If unset, A passes through unchanged."`
}

func (n SubtractionNode) Description() string {
	return "Carves B out of A."
}

func (n SubtractionNode) Subtract(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.A == nil && n.B == nil {
		return
	}

	if n.A == nil {
		b := nodes.GetOutputValue(out, n.B)
		out.Set(func(f vector3.Float64) float64 {
			return b(f)
		})
		return
	}

	if n.B == nil {
		out.Set(nodes.GetOutputValue(out, n.A))
		return
	}

	a := nodes.GetOutputValue(out, n.A)
	b := nodes.GetOutputValue(out, n.B)
	out.Set(Subtract(a, b))
}

type SmoothSubtractionNode struct {
	A      nodes.Output[sample.Vec3ToFloat] `description:"The base shape. If unset, B passes through unmodified (not negated)."`
	B      nodes.Output[sample.Vec3ToFloat] `description:"The shape carved out of A. If unset, A passes through unchanged."`
	Radius nodes.Output[float64]            `description:"Width of the blend region in world units. Zero or less is a hard subtraction. Defaults to 0.1."`
}

func (n SmoothSubtractionNode) Description() string {
	return "Carves B out of A but blends the cut smoothly into the surrounding surface."
}

func (n SmoothSubtractionNode) Subtract(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.A == nil && n.B == nil {
		return
	}

	if n.A == nil {
		b := nodes.GetOutputValue(out, n.B)
		out.Set(func(f vector3.Float64) float64 {
			return b(f)
		})
		return
	}

	if n.B == nil {
		out.Set(nodes.GetOutputValue(out, n.A))
		return
	}

	a := nodes.GetOutputValue(out, n.A)
	b := nodes.GetOutputValue(out, n.B)
	radius := nodes.TryGetOutputValue(out, n.Radius, .1)
	out.Set(SmoothSubtract(radius, a, b))
}
