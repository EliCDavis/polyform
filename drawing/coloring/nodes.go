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
	refutil.RegisterType[nodes.Struct[ToVectorNode]](factory)
	refutil.RegisterType[nodes.Struct[FromVectorNode]](factory)

	refutil.RegisterType[nodes.Struct[AdjustHSVNode]](factory)
	refutil.RegisterType[nodes.Struct[BrightnessNode]](factory)
	refutil.RegisterType[nodes.Struct[MultiplyNode]](factory)
	refutil.RegisterType[nodes.Struct[FromHSVNode]](factory)
	refutil.RegisterType[nodes.Struct[ToHSVNode]](factory)
	refutil.RegisterType[nodes.Struct[SRGBToLinearNode]](factory)
	refutil.RegisterType[nodes.Struct[LinearToSRGBNode]](factory)

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
	A    nodes.LiftedPort[Color]   `description:"Color at Time=0."`
	B    nodes.LiftedPort[Color]   `description:"Color at Time=1."`
	Time nodes.LiftedPort[float64] `description:"Blend factor between A and B, 0 to 1. Defaults to 0.5"`
}

func (n InterpolateNode) Description() string {
	return "Linearly interpolates between two colors."
}

func (n InterpolateNode) Out(out *nodes.Lifted[Color]) {
	// One end left unwired is the other end rather than black, so that a
	// half-built blend shows the color that is wired up.
	a, b := n.A, n.B
	switch {
	case a == nil && b == nil:
		opaqueBlack := nodes.ConstOutput[Color]{Val: Color{A: 1}}
		a, b = opaqueBlack, opaqueBlack
	case a == nil:
		a = b
	case b == nil:
		b = a
	}

	nodes.Zip3(out, a, b, nodes.LiftedOr(n.Time, 0.5), Color.Lerp)
}

// ============================================================================

type ToVectorNode struct {
	In nodes.LiftedPort[Color]
}

func (n ToVectorNode) Description() string {
	return "Splits a color into its components as a vector."
}

func (n ToVectorNode) Vector3(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip1(out, n.In, func(c Color) vector3.Float64 {
		return vector3.New(c.R, c.G, c.B)
	})
}

func (n ToVectorNode) Vector4(out *nodes.Lifted[vector4.Float64]) {
	nodes.Zip1(out, n.In, func(c Color) vector4.Float64 {
		return vector4.New(c.R, c.G, c.B, c.A)
	})
}

// ============================================================================

type Gradient1DNode struct {
	In []nodes.Output[GradientKey[float64]]
}

func (n Gradient1DNode) Description() string {
	return "Builds a gradient of numbers from a set of keys."
}

func (n Gradient1DNode) Gradient(out *nodes.StructOutput[Gradient[float64]]) {
	out.Set(NewGradient1D(nodes.GetOutputValues(out, n.In)...))
}

// ============================================================================

type Gradient2DNode struct {
	In []nodes.Output[GradientKey[vector2.Float64]]
}

func (n Gradient2DNode) Description() string {
	return "Builds a gradient of 2D vectors from a set of keys."
}

func (n Gradient2DNode) Gradient(out *nodes.StructOutput[Gradient[vector2.Float64]]) {
	out.Set(NewGradient2D(nodes.GetOutputValues(out, n.In)...))
}

// ============================================================================

type Gradient3DNode struct {
	In []nodes.Output[GradientKey[vector3.Float64]]
}

func (n Gradient3DNode) Description() string {
	return "Builds a gradient of 3D vectors from a set of keys."
}

func (n Gradient3DNode) Gradient(out *nodes.StructOutput[Gradient[vector3.Float64]]) {
	out.Set(NewGradient3D(nodes.GetOutputValues(out, n.In)...))
}

// ============================================================================

type Gradient4DNode struct {
	In []nodes.Output[GradientKey[vector4.Float64]]
}

func (n Gradient4DNode) Description() string {
	return "Builds a gradient of 4D vectors from a set of keys."
}

func (n Gradient4DNode) Gradient(out *nodes.StructOutput[Gradient[vector4.Float64]]) {
	out.Set(NewGradient4D(nodes.GetOutputValues(out, n.In)...))
}

// ============================================================================

type GradientColorNode struct {
	In []nodes.Output[GradientKey[Color]]
}

func (n GradientColorNode) Description() string {
	return "Builds a color ramp from a set of color keys."
}

func (n GradientColorNode) Gradient(out *nodes.StructOutput[Gradient[Color]]) {
	out.Set(NewGradientColor(nodes.GetOutputValues(out, n.In)...))
}

// ============================================================================

type GradientKeyNode[T any] struct {
	Value nodes.Output[T]
	Time  nodes.Output[float64]
}

func (n GradientKeyNode[T]) Description() string {
	return "One stop in a gradient: a value and the 0-1 time it lands at."
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

func (n RedNode) Description() string {
	return "Pure red, as a constant color"
}

func (n RedNode) Color(out *nodes.StructOutput[Color]) { out.Set(Color{1, 0, 0, 1}) }

type GreenNode struct{}

func (n GreenNode) Description() string {
	return "Pure green, as a constant color"
}

func (n GreenNode) Color(out *nodes.StructOutput[Color]) { out.Set(Color{0, 1, 0, 1}) }

type BlueNode struct{}

func (n BlueNode) Description() string {
	return "Pure blue, as a constant color"
}

func (n BlueNode) Color(out *nodes.StructOutput[Color]) { out.Set(Color{0, 0, 1, 1}) }

type YellowNode struct{}

func (n YellowNode) Description() string {
	return "Pure yellow, as a constant color"
}

func (n YellowNode) Color(out *nodes.StructOutput[Color]) {
	out.Set(Color{1, 1, 0, 1})
}

type MagentaNode struct{}

func (n MagentaNode) Description() string {
	return "Pure magenta, as a constant color"
}

func (n MagentaNode) Color(out *nodes.StructOutput[Color]) {
	out.Set(Color{1, 0, 1, 1})
}

type CyanNode struct{}

func (n CyanNode) Description() string {
	return "Pure cyan, as a constant color"
}

func (n CyanNode) Color(out *nodes.StructOutput[Color]) { out.Set(Color{0, 1, 1, 1}) }

type BlackNode struct{}

func (n BlackNode) Description() string {
	return "Pure black, as a constant color"
}

func (n BlackNode) Color(out *nodes.StructOutput[Color]) { out.Set(Color{0, 0, 0, 1}) }

type WhiteNode struct{}

func (n WhiteNode) Description() string {
	return "Pure white, as a constant color"
}

func (n WhiteNode) Color(out *nodes.StructOutput[Color]) { out.Set(Color{1, 1, 1, 1}) }
