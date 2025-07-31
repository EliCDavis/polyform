package colors

import (
	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[InterpolateNode]](factory)
	refutil.RegisterType[nodes.Struct[InterpolateToArrayNode]](factory)
	refutil.RegisterType[nodes.Struct[ToVectorNode]](factory)
	refutil.RegisterType[nodes.Struct[ToVectorArrayNode]](factory)

	generator.RegisterTypes(factory)
}

type InterpolateNode struct {
	A    nodes.Output[coloring.WebColor]
	B    nodes.Output[coloring.WebColor]
	Time nodes.Output[float64]
}

func (n InterpolateNode) Out() nodes.StructOutput[coloring.WebColor] {
	out := nodes.NewStructOutput(coloring.WebColor{R: 0, G: 0, B: 0, A: 255})
	if n.A == nil && n.B == nil {
		return out
	}

	if n.A == nil {
		out.Set(nodes.GetOutputValue(&out, n.B))
		return out
	}

	if n.B == nil {
		out.Set(nodes.GetOutputValue(&out, n.A))
		return out
	}

	i := Interpolate(
		nodes.GetOutputValue(&out, n.A),
		nodes.GetOutputValue(&out, n.B),
		nodes.TryGetOutputValue(&out, n.Time, 0),
	)

	r, g, b, a := i.RGBA()
	out.Set(coloring.WebColor{
		R: byte(r >> 8),
		G: byte(g >> 8),
		B: byte(b >> 8),
		A: byte(a >> 8),
	})

	return out
}

type InterpolateToArrayNode struct {
	A    nodes.Output[coloring.WebColor]
	B    nodes.Output[coloring.WebColor]
	Time nodes.Output[[]float64]
}

func (n InterpolateToArrayNode) Out() nodes.StructOutput[[]coloring.WebColor] {
	if n.Time == nil {
		return nodes.NewStructOutput([]coloring.WebColor{})
	}

	out := nodes.StructOutput[[]coloring.WebColor]{}
	times := nodes.GetOutputValue(&out, n.Time)

	arr := make([]coloring.WebColor, len(times))
	out.Set(arr)

	if n.A == nil && n.B == nil {
		return out
	}

	if n.A == nil {
		v := nodes.GetOutputValue(&out, n.B)
		for i := range arr {
			arr[i] = v
		}

		return out
	}

	if n.B == nil {
		v := nodes.GetOutputValue(&out, n.A)
		for i := range arr {
			arr[i] = v
		}

		return out
	}

	aV := nodes.GetOutputValue(&out, n.A)
	bV := nodes.GetOutputValue(&out, n.B)

	for i, t := range times {
		v := Interpolate(aV, bV, t)
		r, g, b, a := v.RGBA()
		arr[i] = coloring.WebColor{
			R: byte(r >> 8),
			G: byte(g >> 8),
			B: byte(b >> 8),
			A: byte(a >> 8),
		}
	}

	return out
}

type ToVectorNode struct {
	In nodes.Output[coloring.WebColor]
}

func (n ToVectorNode) vector4(c coloring.WebColor) vector4.Float64 {
	if n.In == nil {
		return vector4.Zero[float64]()
	}

	return vector4.New(
		float64(c.R)/255.,
		float64(c.G)/255.,
		float64(c.B)/255.,
		float64(c.A)/255.,
	)
}

func (n ToVectorNode) Vector3() nodes.StructOutput[vector3.Float64] {
	out := nodes.StructOutput[vector3.Float64]{}
	out.Set(n.vector4(nodes.TryGetOutputValue(&out, n.In, coloring.WebColor{})).XYZ())
	return out
}

func (n ToVectorNode) Vector4() nodes.StructOutput[vector4.Float64] {
	out := nodes.StructOutput[vector4.Float64]{}
	out.Set(n.vector4(nodes.TryGetOutputValue(&out, n.In, coloring.WebColor{})))
	return out
}

// ============================================================================

type ToVectorArrayNode struct {
	In nodes.Output[[]coloring.WebColor]
}

func (n ToVectorArrayNode) Vector3() nodes.StructOutput[[]vector3.Float64] {
	out := nodes.StructOutput[[]vector3.Float64]{}
	in := nodes.TryGetOutputValue(&out, n.In, nil)
	arr := make([]vector3.Float64, len(in))
	for i, c := range in {
		arr[i] = vector3.New(
			float64(c.R)/255.,
			float64(c.G)/255.,
			float64(c.B)/255.,
		)
	}
	out.Set(arr)
	return out
}

func (n ToVectorArrayNode) Vector4() nodes.StructOutput[[]vector4.Float64] {
	out := nodes.StructOutput[[]vector4.Float64]{}
	in := nodes.TryGetOutputValue(&out, n.In, nil)
	arr := make([]vector4.Float64, len(in))
	for i, c := range in {
		arr[i] = vector4.New(
			float64(c.R)/255.,
			float64(c.G)/255.,
			float64(c.B)/255.,
			float64(c.A)/255.,
		)
	}
	out.Set(arr)
	return out
}
