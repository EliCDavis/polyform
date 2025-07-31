package math

import (
	"math"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[Round](factory)

	refutil.RegisterType[nodes.Struct[CircumferenceNode]](factory)

	refutil.RegisterType[nodes.Struct[DifferenceNodeData[int]]](factory)
	refutil.RegisterType[nodes.Struct[DifferenceNodeData[float64]]](factory)
	refutil.RegisterType[nodes.Struct[DifferencesToArrayNodeData[int]]](factory)
	refutil.RegisterType[nodes.Struct[DifferencesToArrayNodeData[float64]]](factory)

	refutil.RegisterType[nodes.Struct[SumNodeData[int]]](factory)
	refutil.RegisterType[nodes.Struct[SumNodeData[float64]]](factory)

	refutil.RegisterType[nodes.Struct[AddToArrayNodeData[int]]](factory)
	refutil.RegisterType[nodes.Struct[AddToArrayNodeData[float64]]](factory)

	refutil.RegisterType[nodes.Struct[DivideNodeData[int]]](factory)
	refutil.RegisterType[nodes.Struct[DivideNodeData[float64]]](factory)
	refutil.RegisterType[nodes.Struct[DivideToArrayNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[DivideToArrayNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[MultiplyNodeData[float64]]](factory)
	refutil.RegisterType[nodes.Struct[MultiplyNodeData[int]]](factory)
	refutil.RegisterType[nodes.Struct[MultiplyToArrayNodeData[float64]]](factory)
	refutil.RegisterType[nodes.Struct[MultiplyToArrayNodeData[int]]](factory)

	refutil.RegisterType[nodes.Struct[InverseNodeData[float64]]](factory)
	refutil.RegisterType[nodes.Struct[InverseNodeData[int]]](factory)

	refutil.RegisterType[nodes.Struct[NegateNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[NegateNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[DoubleNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[DoubleNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[HalfNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[HalfNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[OneNode]](factory)
	refutil.RegisterType[nodes.Struct[ZeroNode]](factory)

	refutil.RegisterType[nodes.Struct[MinNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[MinNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[MinArrayNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[MinArrayNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[MaxNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[MaxNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[MaxArrayNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[MaxArrayNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[IntToFloatNode]](factory)

	refutil.RegisterType[nodes.Struct[PlaneFromNormalNode]](factory)
	refutil.RegisterType[nodes.Struct[SquareNode]](factory)
	refutil.RegisterType[nodes.Struct[SquareRootNode]](factory)
	refutil.RegisterType[nodes.Struct[HypotenuseNode]](factory)

	refutil.RegisterType[nodes.Struct[RemapNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[RemapToArrayNode[float64]]](factory)

	generator.RegisterTypes(factory)
}

// ============================================================================

type PlaneFromNormalNode struct {
	Normal   nodes.Output[vector3.Float64]
	Position nodes.Output[vector3.Float64]
}

func (n PlaneFromNormalNode) Out() nodes.StructOutput[geometry.Plane] {
	out := nodes.StructOutput[geometry.Plane]{}
	out.Set(geometry.NewPlane(
		nodes.TryGetOutputValue(&out, n.Position, vector3.Zero[float64]()),
		nodes.TryGetOutputValue(&out, n.Normal, vector3.Up[float64]()),
	))
	return out
}

// ============================================================================

type OneNode struct{}

func (cn OneNode) Int() nodes.StructOutput[int] {
	return nodes.NewStructOutput(1)
}

func (cn OneNode) Float64() nodes.StructOutput[float64] {
	return nodes.NewStructOutput(1.)
}

func (cn OneNode) Description() string {
	return "Just the number 1"
}

// ============================================================================

type ZeroNode struct{}

func (cn ZeroNode) Int() nodes.StructOutput[int] {
	return nodes.NewStructOutput(0)
}

func (cn ZeroNode) Float64() nodes.StructOutput[float64] {
	return nodes.NewStructOutput(0.)
}

func (cn ZeroNode) Description() string {
	return "Just the number 0"
}

// ============================================================================

type DoubleNode[T vector.Number] struct {
	In nodes.Output[T] `description:"The number to double"`
}

func (cn DoubleNode[T]) Int() nodes.StructOutput[int] {
	out := nodes.StructOutput[int]{}
	out.Set(int(float64(nodes.TryGetOutputValue(&out, cn.In, 0)) * 2))
	return out
}

func (cn DoubleNode[T]) Float64() nodes.StructOutput[float64] {
	out := nodes.StructOutput[float64]{}
	out.Set(float64(nodes.TryGetOutputValue(&out, cn.In, 0)) * 2)
	return out
}

func (cn DoubleNode[T]) Description() string {
	return "Doubles the number provided"
}

// ============================================================================

type HalfNode[T vector.Number] struct {
	In nodes.Output[T] `description:"The number to halve"`
}

func (cn HalfNode[T]) Int() nodes.StructOutput[int] {
	out := nodes.StructOutput[int]{}
	out.Set(int(float64(nodes.TryGetOutputValue(&out, cn.In, 0)) * 0.5))
	return out
}

func (cn HalfNode[T]) Float64() nodes.StructOutput[float64] {
	out := nodes.StructOutput[float64]{}
	out.Set(float64(nodes.TryGetOutputValue(&out, cn.In, 0)) * 0.5)
	return out
}

func (cn HalfNode[T]) Description() string {
	return "Divides the number in half"
}

// ============================================================================

type IntToFloatNode struct {
	In nodes.Output[int]
}

func (cn IntToFloatNode) Out() nodes.StructOutput[float64] {
	out := nodes.StructOutput[float64]{}
	out.Set(float64(nodes.TryGetOutputValue(&out, cn.In, 0)))
	return out
}

// ============================================================================

type NegateNode[T vector.Number] struct {
	In nodes.Output[T] `description:"The number to take the additive inverse of"`
}

func (cn NegateNode[T]) Out() nodes.StructOutput[T] {
	out := nodes.StructOutput[T]{}
	out.Set(nodes.TryGetOutputValue(&out, cn.In, 0) * -1)
	return out
}

func (cn NegateNode[T]) Description() string {
	return "The additive inverse of an element x, denoted −x, is the element that when added to x, yields the additive identity, 0"
}

// ============================================================================
type InverseNodeData[T vector.Number] struct {
	In nodes.Output[T] `description:"The number to take the inverse of"`
}

func (cn InverseNodeData[T]) Additive() nodes.StructOutput[T] {
	out := nodes.StructOutput[T]{}
	out.Set(nodes.TryGetOutputValue(&out, cn.In, 0) * -1)
	return out
}

func (cn InverseNodeData[T]) AdditiveDescription() string {
	return "The additive inverse of an element x, denoted −x, is the element that when added to x, yields the additive identity, 0"
}

func (cn InverseNodeData[T]) Multiplicative() nodes.StructOutput[T] {
	out := nodes.StructOutput[T]{}
	v := nodes.TryGetOutputValue(&out, cn.In, 0)
	if v == 0 {
		out.CaptureError(cantDivideByZeroErr)
		return out
	}
	out.Set(1. / v)
	return out
}

func (cn InverseNodeData[T]) MultiplicativeDescription() string {
	return "The multiplicative inverse for a number x, denoted by 1/x or x^−1, is a number which when multiplied by x yields the multiplicative identity, 1"
}

// ============================================================================

type Round = nodes.Struct[RoundNodeData]

type RoundNodeData struct {
	In nodes.Output[float64]
}

func (cn RoundNodeData) Int() nodes.StructOutput[int] {
	out := nodes.StructOutput[int]{}
	out.Set(int(math.Round(nodes.TryGetOutputValue(&out, cn.In, 0.))))
	return out
}

func (cn RoundNodeData) Float() nodes.StructOutput[float64] {
	out := nodes.StructOutput[float64]{}
	out.Set(math.Round(nodes.TryGetOutputValue(&out, cn.In, 0.)))
	return out
}

// ============================================================================

type CircumferenceNode struct {
	Radius nodes.Output[float64]
}

func (cn CircumferenceNode) Description() string {
	return "Circumference of a circle"
}

func (cn CircumferenceNode) Int() nodes.StructOutput[int] {
	out := nodes.StructOutput[int]{}
	out.Set(int(math.Round(nodes.TryGetOutputValue(&out, cn.Radius, 0.) * 2 * math.Pi)))
	return out
}

func (cn CircumferenceNode) Float() nodes.StructOutput[float64] {
	out := nodes.StructOutput[float64]{}
	out.Set(nodes.TryGetOutputValue(&out, cn.Radius, 0.) * 2 * math.Pi)
	return out
}

// ============================================================================
type SquareNode struct {
	In nodes.Output[float64]
}

func (cn SquareNode) Out() nodes.StructOutput[float64] {
	out := nodes.StructOutput[float64]{}
	v := nodes.TryGetOutputValue(&out, cn.In, 0)
	out.Set(v * v)
	return out
}

type SquareRootNode struct {
	In nodes.Output[float64]
}

func (cn SquareRootNode) Out() nodes.StructOutput[float64] {
	out := nodes.StructOutput[float64]{}
	out.Set(math.Sqrt(nodes.TryGetOutputValue(&out, cn.In, 0)))
	return out
}

// ============================================================================

type HypotenuseNode struct {
	P nodes.Output[float64]
	Q nodes.Output[float64]
}

func (cn HypotenuseNode) Out() nodes.StructOutput[float64] {
	out := nodes.StructOutput[float64]{}
	out.Set(math.Hypot(
		nodes.TryGetOutputValue(&out, cn.P, 0),
		nodes.TryGetOutputValue(&out, cn.Q, 0),
	))
	return out
}

// ============================================================================

type RemapNode[T vector.Number] struct {
	Value nodes.Output[T]

	InMin nodes.Output[T]
	InMax nodes.Output[T]

	OutMin nodes.Output[T]
	OutMax nodes.Output[T]
}

func (n RemapNode[T]) Out() nodes.StructOutput[T] {
	out := nodes.StructOutput[T]{}

	inMin := nodes.TryGetOutputValue(&out, n.InMin, 0)
	inMax := nodes.TryGetOutputValue(&out, n.InMax, 1)
	inRange := inMax - inMin

	outMin := nodes.TryGetOutputValue(&out, n.OutMin, 0)
	outMax := nodes.TryGetOutputValue(&out, n.OutMax, 1)
	outRange := outMax - outMin

	v := nodes.TryGetOutputValue(&out, n.Value, 0)

	in := (v - inMin) / inRange
	out.Set((in * outRange) + outMin)

	return out
}

type RemapToArrayNode[T vector.Number] struct {
	Value nodes.Output[[]T]

	InMin nodes.Output[T]
	InMax nodes.Output[T]

	OutMin nodes.Output[T]
	OutMax nodes.Output[T]

	Clamp nodes.Output[bool]
}

func (n RemapToArrayNode[T]) Out() nodes.StructOutput[[]T] {
	out := nodes.StructOutput[[]T]{}

	inMin := nodes.TryGetOutputValue(&out, n.InMin, 0)
	inMax := nodes.TryGetOutputValue(&out, n.InMax, 1)
	inRange := inMax - inMin

	outMin := nodes.TryGetOutputValue(&out, n.OutMin, 0)
	outMax := nodes.TryGetOutputValue(&out, n.OutMax, 1)
	outRange := outMax - outMin

	values := nodes.TryGetOutputValue(&out, n.Value, nil)
	arr := make([]T, len(values))

	clamped := nodes.TryGetOutputValue(&out, n.Clamp, true)

	for i, v := range values {
		in := (v - inMin) / inRange
		arr[i] = (in * outRange) + outMin
		if clamped {
			arr[i] = clamp(arr[i], outMin, outMax)
		}
	}

	out.Set(arr)
	return out
}

func clamp[T vector.Number](t, minV, maxV T) T {
	return min(max(t, minV), maxV)
}
