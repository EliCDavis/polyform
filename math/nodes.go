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

	refutil.RegisterType[nodes.Struct[RoundNode]](factory)
	refutil.RegisterType[nodes.Struct[AbsNode]](factory)
	refutil.RegisterType[nodes.Struct[ModuloNode]](factory)

	refutil.RegisterType[nodes.Struct[CircumferenceNode]](factory)

	refutil.RegisterType[nodes.Struct[SubtractNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[SubtractNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[AddNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[AddNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[DivideNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[DivideNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[MultiplyNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[MultiplyNode[int]]](factory)

	refutil.RegisterType[nodes.Struct[InverseNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[InverseNode[int]]](factory)

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
	refutil.RegisterType[nodes.Struct[CompareNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[CompareNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[BoolToNumberNode]](factory)
	refutil.RegisterType[nodes.Struct[NotNode]](factory)

	refutil.RegisterType[nodes.Struct[PlaneFromNormalNode]](factory)
	refutil.RegisterType[nodes.Struct[SquareNode]](factory)
	refutil.RegisterType[nodes.Struct[SquareRootNode]](factory)
	refutil.RegisterType[nodes.Struct[HypotenuseNode]](factory)

	refutil.RegisterType[nodes.Struct[RemapNode[float64]]](factory)

	generator.RegisterTypes(factory)
}

// ============================================================================

type PlaneFromNormalNode struct {
	Normal   nodes.LiftedPort[vector3.Float64]
	Position nodes.LiftedPort[vector3.Float64]
}

func (n PlaneFromNormalNode) Description() string {
	return "Builds an infinite plane from a point on it and its normal."
}

func (n PlaneFromNormalNode) Out(out *nodes.Lifted[geometry.Plane]) {
	nodes.Zip2(
		out,
		nodes.LiftedOr(n.Position, vector3.Zero[float64]()),
		nodes.LiftedOr(n.Normal, vector3.Up[float64]()),
		geometry.NewPlane,
	)
}

// ============================================================================

type OneNode struct{}

func (cn OneNode) Int(out *nodes.StructOutput[int]) {
	out.Set(1)
}

func (cn OneNode) Float64(out *nodes.StructOutput[float64]) {
	out.Set(1)
}

func (cn OneNode) Description() string {
	return "Just the number 1"
}

// ============================================================================

type ZeroNode struct{}

func (cn ZeroNode) Int(out *nodes.StructOutput[int]) {
}

func (cn ZeroNode) Float64(out *nodes.StructOutput[float64]) {
}

func (cn ZeroNode) Description() string {
	return "Just the number 0"
}

// ============================================================================

type DoubleNode[T vector.Number] struct {
	In nodes.LiftedPort[T] `description:"The number to double"`
}

func (cn DoubleNode[T]) Int(out *nodes.Lifted[int]) {
	nodes.Zip1(out, cn.In, func(v T) int { return int(float64(v) * 2) })
}

func (cn DoubleNode[T]) Float64(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, cn.In, func(v T) float64 { return float64(v) * 2 })
}

func (cn DoubleNode[T]) Description() string {
	return "Doubles the number provided"
}

// ============================================================================

type HalfNode[T vector.Number] struct {
	In nodes.LiftedPort[T] `description:"The number to halve"`
}

func (cn HalfNode[T]) Int(out *nodes.Lifted[int]) {
	nodes.Zip1(out, cn.In, func(v T) int { return int(float64(v) * 0.5) })
}

func (cn HalfNode[T]) Float64(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, cn.In, func(v T) float64 { return float64(v) * 0.5 })
}

func (cn HalfNode[T]) Description() string {
	return "Divides the number in half"
}

// ============================================================================

type IntToFloatNode struct {
	In nodes.LiftedPort[int]
}

func (cn IntToFloatNode) Description() string {
	return "Converts a whole number to a decimal one."
}

func (cn IntToFloatNode) Out(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, cn.In, func(v int) float64 { return float64(v) })
}

// ============================================================================

type NegateNode[T vector.Number] struct {
	In nodes.LiftedPort[T] `description:"The number to take the additive inverse of"`
}

func (cn NegateNode[T]) Out(out *nodes.Lifted[T]) {
	nodes.Zip1(out, cn.In, func(v T) T { return v * -1 })
}

func (cn NegateNode[T]) Description() string {
	return "The additive inverse of an element x, denoted −x, is the element that when added to x, yields the additive identity, 0"
}

// ============================================================================

type InverseNode[T vector.Number] struct {
	In nodes.LiftedPort[T] `description:"The number to take the inverse of"`
}

func (cn InverseNode[T]) Description() string {
	return "Inverts a number, either by negating it (additive) or by dividing one by it (multiplicative)."
}

func (cn InverseNode[T]) Additive(out *nodes.Lifted[T]) {
	nodes.Zip1(out, cn.In, func(v T) T { return v * -1 })
}

func (cn InverseNode[T]) AdditiveDescription() string {
	return "The additive inverse of an element x, denoted −x, is the element that when added to x, yields the additive identity, 0"
}

func (cn InverseNode[T]) Multiplicative(out *nodes.Lifted[T]) {
	dividedByZero := false
	nodes.Zip1(out, cn.In, func(v T) T {
		if v == 0 {
			dividedByZero = true
			return 0
		}
		return 1. / v
	})
	if dividedByZero {
		out.CaptureError(cantDivideByZeroErr)
	}
}

func (cn InverseNode[T]) MultiplicativeDescription() string {
	return "The multiplicative inverse for a number x, denoted by 1/x or x^−1, is a number which when multiplied by x yields the multiplicative identity, 1"
}

// ============================================================================

type RoundNode struct {
	In nodes.LiftedPort[float64]
}

func (cn RoundNode) Description() string {
	return "Rounds a number to the nearest whole number."
}

func (cn RoundNode) Int(out *nodes.Lifted[int]) {
	nodes.Zip1(out, cn.In, func(v float64) int { return int(math.Round(v)) })
}

func (cn RoundNode) Float(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, cn.In, math.Round)
}

func (cn RoundNode) Down(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, cn.In, math.Floor)
}

func (cn RoundNode) Up(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, cn.In, math.Ceil)
}

// ============================================================================

type AbsNode struct {
	In nodes.LiftedPort[float64]
}

func (cn AbsNode) Description() string {
	return "Magnitude of a number, dropping its sign."
}

func (cn AbsNode) Out(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, cn.In, math.Abs)
}

// ============================================================================

type ModuloNode struct {
	A nodes.LiftedPort[float64]
	B nodes.LiftedPort[float64] `description:"The divisor. Zero gives zero rather than NaN."`
}

func (cn ModuloNode) Description() string {
	return "Remainder of A divided by B, always with B's sign."
}

func (cn ModuloNode) Keywords() []string {
	return []string{"mod", "remainder"}
}

func (cn ModuloNode) Out(out *nodes.Lifted[float64]) {
	nodes.Zip2(out, cn.A, cn.B, func(a, b float64) float64 {
		if b == 0 {
			return 0
		}
		remainder := math.Mod(a, b)
		if remainder != 0 && (remainder < 0) != (b < 0) {
			remainder += b
		}
		return remainder
	})
}

// ============================================================================

type CircumferenceNode struct {
	Radius nodes.LiftedPort[float64]
}

func (cn CircumferenceNode) Description() string {
	return "Circumference of a circle"
}

func (cn CircumferenceNode) Int(out *nodes.Lifted[int]) {
	nodes.Zip1(out, cn.Radius, func(r float64) int {
		return int(math.Round(r * 2 * math.Pi))
	})
}

func (cn CircumferenceNode) Float(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, cn.Radius, func(r float64) float64 { return r * 2 * math.Pi })
}

// ============================================================================

type SquareNode struct {
	In nodes.LiftedPort[float64]
}

func (cn SquareNode) Description() string {
	return "Multiplies a number by itself."
}

func (cn SquareNode) Out(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, cn.In, func(v float64) float64 { return v * v })
}

func (cn SquareNode) Int(out *nodes.Lifted[int]) {
	nodes.Zip1(out, cn.In, func(v float64) int { return int(math.Round(v * v)) })
}

// ============================================================================

type SquareRootNode struct {
	In nodes.LiftedPort[float64]
}

func (cn SquareRootNode) Description() string {
	return "Square root of a number."
}

func (cn SquareRootNode) Out(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, cn.In, math.Sqrt)
}

// ============================================================================

type HypotenuseNode struct {
	P nodes.LiftedPort[float64]
	Q nodes.LiftedPort[float64]
}

func (cn HypotenuseNode) Description() string {
	return "Length of a right triangle's hypotenuse from its two legs."
}

func (cn HypotenuseNode) Out(out *nodes.Lifted[float64]) {
	nodes.Zip2(out, cn.P, cn.Q, math.Hypot)
}

// ============================================================================

type RemapNode[T vector.Number] struct {
	Value nodes.LiftedPort[T]

	InMin nodes.LiftedPort[T]
	InMax nodes.LiftedPort[T]

	OutMin nodes.LiftedPort[T]
	OutMax nodes.LiftedPort[T]

	Clamp nodes.Output[bool] `description:"Holds results inside the output range. Defaults to true."`
}

func (n RemapNode[T]) Description() string {
	return "Rescales a number from one range to another."
}

func (n RemapNode[T]) Out(out *nodes.Lifted[T]) {
	clamped := nodes.TryGetOutputValue(out, n.Clamp, true)

	nodes.ZipAll(
		out,
		[]nodes.LiftedPort[T]{
			nodes.LiftedOr[T](n.Value, 0),
			nodes.LiftedOr[T](n.InMin, 0),
			nodes.LiftedOr[T](n.InMax, 1),
			nodes.LiftedOr[T](n.OutMin, 0),
			nodes.LiftedOr[T](n.OutMax, 1),
		},
		func(v []T) T {
			value, inMin, inMax, outMin, outMax := v[0], v[1], v[2], v[3], v[4]
			scaled := ((value - inMin) / (inMax - inMin) * (outMax - outMin)) + outMin
			if clamped {
				return clamp(scaled, outMin, outMax)
			}
			return scaled
		},
	)
}

func clamp[T vector.Number](t, minV, maxV T) T {
	if minV > maxV {
		minV, maxV = maxV, minV
	}
	return min(max(t, minV), maxV)
}
