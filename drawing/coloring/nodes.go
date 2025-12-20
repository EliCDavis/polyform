package coloring

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[InterpolateNode]](factory)
	refutil.RegisterType[nodes.Struct[InterpolateToArrayNode]](factory)
	refutil.RegisterType[nodes.Struct[ToVectorNode]](factory)
	refutil.RegisterType[nodes.Struct[ToVectorArrayNode]](factory)

	refutil.RegisterType[nodes.Struct[Gradient1DNode]](factory)
	refutil.RegisterType[nodes.Struct[Gradient2DNode]](factory)
	refutil.RegisterType[nodes.Struct[Gradient3DNode]](factory)
	refutil.RegisterType[nodes.Struct[Gradient4DNode]](factory)
	refutil.RegisterType[nodes.Struct[GradientColorNode]](factory)

	refutil.RegisterType[nodes.Struct[GradientKeyNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[GradientKeyNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[GradientKeyNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[GradientKeyNode[vector4.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[GradientKeyNode[Color]]](factory)

	refutil.RegisterType[nodes.Struct[RedNode]](factory)
	refutil.RegisterType[nodes.Struct[GreenNode]](factory)
	refutil.RegisterType[nodes.Struct[BlueNode]](factory)
	refutil.RegisterType[nodes.Struct[MagentaNode]](factory)
	refutil.RegisterType[nodes.Struct[CyanNode]](factory)
	refutil.RegisterType[nodes.Struct[YellowNode]](factory)
	refutil.RegisterType[nodes.Struct[WhiteNode]](factory)
	refutil.RegisterType[nodes.Struct[BlackNode]](factory)

	generator.RegisterTypes(factory)
}

type InterpolateNode struct {
	A    nodes.Output[Color]
	B    nodes.Output[Color]
	Time nodes.Output[float64]
}

func (n InterpolateNode) Out(out *nodes.StructOutput[Color]) {
	out.Set(Color{R: 0, G: 0, B: 0, A: 1})
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

	out.Set(nodes.GetOutputValue(out, n.A).Lerp(
		nodes.GetOutputValue(out, n.B),
		nodes.TryGetOutputValue(out, n.Time, 0),
	))
}

type InterpolateToArrayNode struct {
	A    nodes.Output[Color]
	B    nodes.Output[Color]
	Time nodes.Output[[]float64]
}

func (n InterpolateToArrayNode) Out(out *nodes.StructOutput[[]Color]) {
	if n.Time == nil {
		return
	}

	times := nodes.GetOutputValue(out, n.Time)

	arr := make([]Color, len(times))
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
		arr[i] = aV.Lerp(bV, t)
	}
}

type ToVectorNode struct {
	In nodes.Output[Color]
}

func (n ToVectorNode) vector4(c Color) vector4.Float64 {
	if n.In == nil {
		return vector4.Zero[float64]()
	}

	return vector4.New(
		c.R,
		c.G,
		c.B,
		c.A,
	)
}

func (n ToVectorNode) Vector3(out *nodes.StructOutput[vector3.Float64]) {
	out.Set(n.vector4(nodes.TryGetOutputValue(out, n.In, Color{})).XYZ())
}

func (n ToVectorNode) Vector4(out *nodes.StructOutput[vector4.Float64]) {
	out.Set(n.vector4(nodes.TryGetOutputValue(out, n.In, Color{})))
}

// ============================================================================

type ToVectorArrayNode struct {
	In nodes.Output[[]Color]
}

func (n ToVectorArrayNode) Vector3(out *nodes.StructOutput[[]vector3.Float64]) {
	in := nodes.TryGetOutputValue(out, n.In, nil)
	arr := make([]vector3.Float64, len(in))
	for i, c := range in {
		arr[i] = vector3.New(
			c.R,
			c.G,
			c.B,
		)
	}
	out.Set(arr)
}

func (n ToVectorArrayNode) Vector4(out *nodes.StructOutput[[]vector4.Float64]) {
	in := nodes.TryGetOutputValue(out, n.In, nil)
	arr := make([]vector4.Float64, len(in))
	for i, c := range in {
		arr[i] = vector4.New(
			c.R,
			c.G,
			c.B,
			c.A,
		)
	}
	out.Set(arr)
}

// ============================================================================

type Gradient1DNode struct {
	In []nodes.Output[GradientKey[float64]]
}

func (n Gradient1DNode) Gradient(out *nodes.StructOutput[Gradient[float64]]) {
	out.Set(NewGradient1D(nodes.GetOutputValues(out, n.In)...))
}

// ============================================================================

type Gradient2DNode struct {
	In []nodes.Output[GradientKey[vector2.Float64]]
}

func (n Gradient2DNode) Gradient(out *nodes.StructOutput[Gradient[vector2.Float64]]) {
	out.Set(NewGradient2D(nodes.GetOutputValues(out, n.In)...))
}

// ============================================================================

type Gradient3DNode struct {
	In []nodes.Output[GradientKey[vector3.Float64]]
}

func (n Gradient3DNode) Gradient(out *nodes.StructOutput[Gradient[vector3.Float64]]) {
	out.Set(NewGradient3D(nodes.GetOutputValues(out, n.In)...))
}

// ============================================================================

type Gradient4DNode struct {
	In []nodes.Output[GradientKey[vector4.Float64]]
}

func (n Gradient4DNode) Gradient(out *nodes.StructOutput[Gradient[vector4.Float64]]) {
	out.Set(NewGradient4D(nodes.GetOutputValues(out, n.In)...))
}

// ============================================================================

type GradientColorNode struct {
	In []nodes.Output[GradientKey[Color]]
}

func (n GradientColorNode) Gradient(out *nodes.StructOutput[Gradient[Color]]) {
	out.Set(NewGradientColor(nodes.GetOutputValues(out, n.In)...))
}

// ============================================================================

type GradientKeyNode[T any] struct {
	Value nodes.Output[T]
	Time  nodes.Output[float64]
}

func (n GradientKeyNode[T]) Gradient(out *nodes.StructOutput[GradientKey[T]]) {
	var val T
	if n.Value != nil {
		val = n.Value.Value()
	}

	out.Set(GradientKey[T]{
		Time:  nodes.TryGetOutputValue(out, n.Time, 0),
		Value: val,
	})
}

// ============================================================================

type RedNode struct{}

func (n RedNode) Color(out *nodes.StructOutput[Color]) { out.Set(Color{1, 0, 0, 1}) }

type GreenNode struct{}

func (n GreenNode) Color(out *nodes.StructOutput[Color]) { out.Set(Color{0, 1, 0, 1}) }

type BlueNode struct{}

func (n BlueNode) Color(out *nodes.StructOutput[Color]) { out.Set(Color{0, 0, 1, 1}) }

type YellowNode struct{}

func (n YellowNode) Color(out *nodes.StructOutput[Color]) {
	out.Set(Color{1, 1, 0, 1})
}

type MagentaNode struct{}

func (n MagentaNode) Color(out *nodes.StructOutput[Color]) {
	out.Set(Color{1, 0, 1, 1})
}

type CyanNode struct{}

func (n CyanNode) Color(out *nodes.StructOutput[Color]) { out.Set(Color{0, 1, 1, 1}) }

type BlackNode struct{}

func (n BlackNode) Color(out *nodes.StructOutput[Color]) { out.Set(Color{0, 0, 0, 1}) }

type WhiteNode struct{}

func (n WhiteNode) Color(out *nodes.StructOutput[Color]) { out.Set(Color{1, 1, 1, 1}) }
