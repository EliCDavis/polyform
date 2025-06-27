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
	if n.A == nil && n.B == nil {
		return nodes.NewStructOutput(coloring.WebColor{R: 0, G: 0, B: 0, A: 255})
	}

	if n.A == nil {
		return nodes.NewStructOutput(n.B.Value())
	}

	if n.B == nil {
		return nodes.NewStructOutput(n.A.Value())
	}

	i := Interpolate(
		n.A.Value(),
		n.B.Value(),
		nodes.TryGetOutputValue(n.Time, 0),
	)

	r, g, b, a := i.RGBA()

	return nodes.NewStructOutput(coloring.WebColor{
		R: byte(r >> 8),
		G: byte(g >> 8),
		B: byte(b >> 8),
		A: byte(a >> 8),
	})
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

	times := n.Time.Value()
	out := make([]coloring.WebColor, len(times))

	if n.A == nil && n.B == nil {
		return nodes.NewStructOutput(out)
	}

	if n.A == nil {
		v := n.B.Value()
		for i := range out {
			out[i] = v
		}

		return nodes.NewStructOutput(out)
	}

	if n.B == nil {
		v := n.A.Value()
		for i := range out {
			out[i] = v
		}

		return nodes.NewStructOutput(out)
	}

	aV := n.A.Value()
	bV := n.B.Value()

	for i, t := range times {
		v := Interpolate(aV, bV, t)
		r, g, b, a := v.RGBA()
		out[i] = coloring.WebColor{
			R: byte(r >> 8),
			G: byte(g >> 8),
			B: byte(b >> 8),
			A: byte(a >> 8),
		}
	}

	return nodes.NewStructOutput(out)
}

type ToVectorNode struct {
	In nodes.Output[coloring.WebColor]
}

func (n ToVectorNode) vector4() vector4.Float64 {
	if n.In == nil {
		return vector4.Zero[float64]()
	}

	c := n.In.Value()
	return vector4.New(
		float64(c.R)/255.,
		float64(c.G)/255.,
		float64(c.B)/255.,
		float64(c.A)/255.,
	)
}

func (n ToVectorNode) Vector3() nodes.StructOutput[vector3.Float64] {
	return nodes.NewStructOutput(n.vector4().XYZ())
}

func (n ToVectorNode) Vector4() nodes.StructOutput[vector4.Float64] {
	return nodes.NewStructOutput(n.vector4())
}

// ============================================================================

type ToVectorArrayNode struct {
	In nodes.Output[[]coloring.WebColor]
}

func (n ToVectorArrayNode) Vector3() nodes.StructOutput[[]vector3.Float64] {
	if n.In == nil {
		return nodes.NewStructOutput([]vector3.Float64{})
	}

	in := n.In.Value()
	out := make([]vector3.Float64, len(in))
	for i, c := range in {
		out[i] = vector3.New(
			float64(c.R)/255.,
			float64(c.G)/255.,
			float64(c.B)/255.,
		)
	}
	return nodes.NewStructOutput(out)
}

func (n ToVectorArrayNode) Vector4() nodes.StructOutput[[]vector4.Float64] {
	if n.In == nil {
		return nodes.NewStructOutput([]vector4.Float64{})
	}

	in := n.In.Value()
	out := make([]vector4.Float64, len(in))
	for i, c := range in {
		out[i] = vector4.New(
			float64(c.R)/255.,
			float64(c.G)/255.,
			float64(c.B)/255.,
			float64(c.A)/255.,
		)
	}
	return nodes.NewStructOutput(out)
}
