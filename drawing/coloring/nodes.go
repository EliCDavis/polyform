package coloring

import (
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
	A    nodes.Output[WebColor]
	B    nodes.Output[WebColor]
	Time nodes.Output[float64]
}

func (n InterpolateNode) Out(out *nodes.StructOutput[WebColor]) {
	out.Set(WebColor{R: 0, G: 0, B: 0, A: 255})
	if n.A == nil && n.B == nil {
		return
	}

	if n.A == nil {
		out.Set(nodes.GetOutputValue(out, n.B))
		return
	}

	if n.B == nil {
		out.Set(nodes.GetOutputValue(out, n.A))
		return
	}

	i := Interpolate(
		nodes.GetOutputValue(out, n.A),
		nodes.GetOutputValue(out, n.B),
		nodes.TryGetOutputValue(out, n.Time, 0),
	)

	r, g, b, a := i.RGBA()
	out.Set(WebColor{
		R: byte(r >> 8),
		G: byte(g >> 8),
		B: byte(b >> 8),
		A: byte(a >> 8),
	})
}

type InterpolateToArrayNode struct {
	A    nodes.Output[WebColor]
	B    nodes.Output[WebColor]
	Time nodes.Output[[]float64]
}

func (n InterpolateToArrayNode) Out(out *nodes.StructOutput[[]WebColor]) {
	if n.Time == nil {
		return
	}

	times := nodes.GetOutputValue(out, n.Time)

	arr := make([]WebColor, len(times))
	out.Set(arr)

	if n.A == nil && n.B == nil {
		return
	}

	if n.A == nil {
		v := nodes.GetOutputValue(out, n.B)
		for i := range arr {
			arr[i] = v
		}
		return
	}

	if n.B == nil {
		v := nodes.GetOutputValue(out, n.A)
		for i := range arr {
			arr[i] = v
		}
		return
	}

	aV := nodes.GetOutputValue(out, n.A)
	bV := nodes.GetOutputValue(out, n.B)

	for i, t := range times {
		v := Interpolate(aV, bV, t)
		r, g, b, a := v.RGBA()
		arr[i] = WebColor{
			R: byte(r >> 8),
			G: byte(g >> 8),
			B: byte(b >> 8),
			A: byte(a >> 8),
		}
	}
}

type ToVectorNode struct {
	In nodes.Output[WebColor]
}

func (n ToVectorNode) vector4(c WebColor) vector4.Float64 {
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

func (n ToVectorNode) Vector3(out *nodes.StructOutput[vector3.Float64]) {
	out.Set(n.vector4(nodes.TryGetOutputValue(out, n.In, WebColor{})).XYZ())
}

func (n ToVectorNode) Vector4(out *nodes.StructOutput[vector4.Float64]) {
	out.Set(n.vector4(nodes.TryGetOutputValue(out, n.In, WebColor{})))
}

// ============================================================================

type ToVectorArrayNode struct {
	In nodes.Output[[]WebColor]
}

func (n ToVectorArrayNode) Vector3(out *nodes.StructOutput[[]vector3.Float64]) {
	in := nodes.TryGetOutputValue(out, n.In, nil)
	arr := make([]vector3.Float64, len(in))
	for i, c := range in {
		arr[i] = vector3.New(
			float64(c.R)/255.,
			float64(c.G)/255.,
			float64(c.B)/255.,
		)
	}
	out.Set(arr)
}

func (n ToVectorArrayNode) Vector4(out *nodes.StructOutput[[]vector4.Float64]) {
	in := nodes.TryGetOutputValue(out, n.In, nil)
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
}
