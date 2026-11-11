package chance

import (
	"errors"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator/registry"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[RandomNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[RandomNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[RandomNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[RandomNode[vector4.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[RandomBoolNode]](factory)
	refutil.RegisterType[nodes.Struct[RandomIndexNode]](factory)
	refutil.RegisterType[nodes.Struct[SampleNode]](factory)
	refutil.RegisterType[nodes.Struct[SeedNode]](factory)

	refutil.RegisterType[nodes.Struct[ConstantNode]](factory)
	refutil.RegisterType[nodes.Struct[NormalNode]](factory)
	refutil.RegisterType[nodes.Struct[LogNormalNode]](factory)
	refutil.RegisterType[nodes.Struct[ExponentialNode]](factory)
	refutil.RegisterType[nodes.Struct[TriangularNode]](factory)
	refutil.RegisterType[nodes.Struct[PowerNode]](factory)
	refutil.RegisterType[nodes.Struct[DiscreteNode]](factory)
	refutil.RegisterType[nodes.Struct[RampNode]](factory)
	refutil.RegisterType[nodes.Struct[Compose2DNode]](factory)
	refutil.RegisterType[nodes.Struct[Compose3DNode]](factory)
	refutil.RegisterType[nodes.Struct[Compose4DNode]](factory)
	refutil.RegisterType[nodes.Struct[UniformNode]](factory)
	refutil.RegisterType[nodes.Struct[Uniform2DNode]](factory)
	refutil.RegisterType[nodes.Struct[Uniform3DNode]](factory)
	refutil.RegisterType[nodes.Struct[Uniform4DNode]](factory)
	refutil.RegisterType[nodes.Struct[InsideCircleNode]](factory)
	refutil.RegisterType[nodes.Struct[InsideSphereNode]](factory)
	refutil.RegisterType[nodes.Struct[Direction2DNode]](factory)
	refutil.RegisterType[nodes.Struct[DirectionNode]](factory)
	refutil.RegisterType[nodes.Struct[RotationNode]](factory)
	refutil.RegisterType[nodes.Struct[RandomNode[quaternion.Quaternion]]](factory)

	registry.Register(factory)
}

// ============================================================================

type RandomNode[T any] struct {
	Key          nodes.LiftedPort[int]             `description:"The element each value is for: its index, or a stable id. A list gives one value per key."`
	Seed         nodes.LiftedPort[int]             `description:"Changes every value. The same seed and key always give the same value."`
	Distribution nodes.LiftedPort[Distribution[T]] `description:"The shape and range the values follow."`
}

func (RandomNode[T]) Description() string {
	return "A random value per key, shaped by Distribution. The same seed and key always give the same value, whatever changes around them."
}

func (RandomNode[T]) Keywords() []string {
	return []string{"jitter", "variation"}
}

func (r RandomNode[T]) Out(out *nodes.Lifted[T]) {
	if r.Distribution == nil {
		out.CaptureError(errors.New("Distribution is required"))
		return
	}
	nodes.Zip3(out, nodes.LiftedOr(r.Key, 0), nodes.LiftedOr(r.Seed, 0), r.Distribution,
		func(key, seed int, d Distribution[T]) T {
			var zero T
			if d == nil {
				return zero
			}
			return d(Roll(seed, key))
		})
}

// ============================================================================

type RandomBoolNode struct {
	Key         nodes.LiftedPort[int]     `description:"The element each value is for: its index, or a stable id. A list gives one value per key."`
	Seed        nodes.LiftedPort[int]     `description:"Changes every value. The same seed and key always give the same value."`
	Probability nodes.LiftedPort[float64] `description:"The share of keys that come out true, 0 to 1." default:"0.5"`
}

func (RandomBoolNode) Description() string {
	return "True or false for every key, true with the given probability."
}

func (RandomBoolNode) Keywords() []string {
	return []string{"coin", "mask"}
}

func (r RandomBoolNode) Out(out *nodes.Lifted[bool]) {
	nodes.Zip3(out, nodes.LiftedOr(r.Key, 0), nodes.LiftedOr(r.Seed, 0), r.Probability,
		func(key, seed int, p float64) bool { return Unit(seed, key) < p })
}

// ============================================================================

type RandomIndexNode struct {
	Key     nodes.LiftedPort[int]   `description:"The element each value is for: its index, or a stable id. A list gives one value per key."`
	Seed    nodes.LiftedPort[int]   `description:"Changes every value. The same seed and key always give the same value."`
	Weights nodes.Output[[]float64] `description:"How likely each index is, relative to the others. Index i comes up Weights[i] / sum of Weights of the time; 0 never comes up."`
}

func (RandomIndexNode) Description() string {
	return "An index into Weights for every key, each index coming up in proportion to its weight."
}

func (RandomIndexNode) Keywords() []string {
	return []string{"weighted", "pick"}
}

func (r RandomIndexNode) Out(out *nodes.Lifted[int]) {
	weights := nodes.TryGetOutputValue(out, r.Weights, nil)
	indices := make([]float64, len(weights))
	for i := range indices {
		indices[i] = float64(i)
	}
	pick, err := discreteQuantile(indices, weights)
	if err != nil {
		out.CaptureError(err)
		return
	}
	nodes.Zip2(out, nodes.LiftedOr(r.Key, 0), nodes.LiftedOr(r.Seed, 0),
		func(key, seed int) int { return int(pick(Unit(seed, key))) })
}

// ============================================================================

type SampleNode struct {
	Distribution nodes.LiftedPort[Distribution[float64]] `description:"The shape to read from."`
	Fraction     nodes.LiftedPort[float64]               `description:"Where to read, 0 to 1: the share of draws that fall below the result. 0.5 is the middle value, 0.9 a value only 1 draw in 10 exceeds."`
}

func (SampleNode) Description() string {
	return "The value a distribution gives at Fraction. Shapes any 0 to 1 value, random or not, the way the distribution would."
}

func (SampleNode) Keywords() []string {
	return []string{"quantile", "inverse cdf", "remap"}
}

func (s SampleNode) Out(out *nodes.Lifted[float64]) {
	nodes.Zip2(out, nodes.LiftedOr(s.Distribution, Uniform(0, 1)), nodes.LiftedOr(s.Fraction, 0),
		func(d Distribution[float64], u float64) float64 { return d(Fixed(u)) })
}

// ============================================================================

type SeedNode struct {
	Seed  nodes.LiftedPort[int]    `description:"The master seed to derive from."`
	Label nodes.LiftedPort[string] `description:"Names what the derived seed is for."`
}

func (SeedNode) Description() string {
	return "A seed of its own for each label, derived from one master seed, so separate parts never draw the same values."
}

func (SeedNode) Keywords() []string {
	return []string{"derive", "salt"}
}

func (s SeedNode) Out(out *nodes.Lifted[int]) {
	nodes.Zip2(out, nodes.LiftedOr(s.Seed, 0), nodes.LiftedOr(s.Label, ""), Derive)
}

// ============================================================================

type ConstantNode struct {
	Value nodes.LiftedPort[float64]
}

func (ConstantNode) Description() string {
	return "A distribution that always gives Value."
}

func (c ConstantNode) Out(out *nodes.Lifted[Distribution[float64]]) {
	nodes.Zip1(out, nodes.LiftedOr(c.Value, 0), Constant)
}

type UniformNode struct {
	Min nodes.LiftedPort[float64]
	Max nodes.LiftedPort[float64] `default:"1"`
}

func (UniformNode) Description() string {
	return "Every value between Min and Max equally likely."
}

func (u UniformNode) Out(out *nodes.Lifted[Distribution[float64]]) {
	nodes.Zip2(out, nodes.LiftedOr(u.Min, 0), u.Max, Uniform)
}

type NormalNode struct {
	Mean      nodes.LiftedPort[float64]
	Deviation nodes.LiftedPort[float64] `description:"The standard deviation: about 68% of values fall within one of Mean, 95% within two." default:"1"`
	Clamp     nodes.LiftedPort[float64] `description:"No value falls more than this many deviations from Mean. 0 leaves the tails open." default:"3"`
}

func (NormalNode) Description() string {
	return "Values cluster around Mean and thin out with distance, the bell curve."
}

func (NormalNode) Keywords() []string {
	return []string{"gaussian", "bell"}
}

func (n NormalNode) Out(out *nodes.Lifted[Distribution[float64]]) {
	nodes.Zip3(out, nodes.LiftedOr(n.Mean, 0), n.Deviation, n.Clamp, Normal)
}

type LogNormalNode struct {
	Median nodes.LiftedPort[float64] `description:"Half the values fall below this, half above." default:"1"`
	Spread nodes.LiftedPort[float64] `description:"How far values range, as a ratio: about 68% fall between Median / e^Spread and Median * e^Spread." default:"0.5"`
}

func (LogNormalNode) Description() string {
	return "Always-positive values clustered around Median, with a long tail upward, suited to sizes and scales."
}

func (l LogNormalNode) Out(out *nodes.Lifted[Distribution[float64]]) {
	nodes.Zip2(out, l.Median, l.Spread, LogNormal)
}

type ExponentialNode struct {
	Median nodes.LiftedPort[float64] `description:"Half the values fall below this, half above." default:"1"`
}

func (ExponentialNode) Description() string {
	return "Values from 0 upward, most of them small, with a thinning tail of large ones."
}

func (e ExponentialNode) Out(out *nodes.Lifted[Distribution[float64]]) {
	nodes.Zip1(out, e.Median, Exponential)
}

type TriangularNode struct {
	Min  nodes.LiftedPort[float64]
	Mode nodes.LiftedPort[float64] `description:"The most likely value, between Min and Max." default:"0.5"`
	Max  nodes.LiftedPort[float64] `default:"1"`
}

func (TriangularNode) Description() string {
	return "Values between Min and Max, most likely at Mode and falling off straight to either end."
}

func (t TriangularNode) Out(out *nodes.Lifted[Distribution[float64]]) {
	nodes.Zip3(out, nodes.LiftedOr(t.Min, 0), t.Mode, t.Max, Triangular)
}

type PowerNode struct {
	Min      nodes.LiftedPort[float64]
	Max      nodes.LiftedPort[float64] `default:"1"`
	Exponent nodes.LiftedPort[float64] `description:"Above 1 crowds values toward Min, below 1 toward Max, 1 is even." default:"2"`
}

func (PowerNode) Description() string {
	return "Values between Min and Max, crowded toward one end by Exponent."
}

func (PowerNode) Keywords() []string {
	return []string{"skew"}
}

func (p PowerNode) Out(out *nodes.Lifted[Distribution[float64]]) {
	nodes.Zip3(out, nodes.LiftedOr(p.Min, 0), p.Max, p.Exponent, Power)
}

type DiscreteNode struct {
	Values  nodes.Output[[]float64] `description:"The values that can come up."`
	Weights nodes.Output[[]float64] `description:"How likely each value is, relative to the others; 0 never comes up. Unwired, all are equally likely."`
}

func (DiscreteNode) Description() string {
	return "One of a fixed set of values, each as likely as its weight."
}

func (DiscreteNode) Keywords() []string {
	return []string{"weighted", "pick"}
}

func (d DiscreteNode) Out(out *nodes.StructOutput[Distribution[float64]]) {
	dist, err := Discrete(nodes.TryGetOutputValue(out, d.Values, nil), nodes.TryGetOutputValue(out, d.Weights, nil))
	if err != nil {
		out.CaptureError(err)
		return
	}
	out.Set(dist)
}

type RampNode struct {
	Gradient nodes.LiftedPort[coloring.Gradient[float64]] `description:"The distribution as a curve: the value at time t is what a share t of draws fall below, so a steep stretch is rare and a flat one common."`
}

func (RampNode) Description() string {
	return "A distribution shaped by a gradient of numbers over 0 to 1."
}

func (RampNode) Keywords() []string {
	return []string{"curve", "custom"}
}

func (r RampNode) Out(out *nodes.Lifted[Distribution[float64]]) {
	if r.Gradient == nil {
		out.CaptureError(errors.New("Gradient is required"))
		return
	}
	nodes.Zip1(out, r.Gradient, Ramp)
}

type Compose2DNode struct {
	X nodes.LiftedPort[Distribution[float64]] `description:"Shapes the X component. Unwired, it spreads evenly over [0, 1)."`
	Y nodes.LiftedPort[Distribution[float64]] `description:"Shapes the Y component. Unwired, it spreads evenly over [0, 1)."`
}

func (Compose2DNode) Description() string {
	return "A 2D distribution from one distribution per axis, each drawn independently."
}

func (c Compose2DNode) Out(out *nodes.Lifted[Distribution[vector2.Float64]]) {
	nodes.Zip2(out, nodes.LiftedOr(c.X, Uniform(0, 1)), nodes.LiftedOr(c.Y, Uniform(0, 1)), Compose2D)
}

type Compose3DNode struct {
	X nodes.LiftedPort[Distribution[float64]] `description:"Shapes the X component. Unwired, it spreads evenly over [0, 1)."`
	Y nodes.LiftedPort[Distribution[float64]] `description:"Shapes the Y component. Unwired, it spreads evenly over [0, 1)."`
	Z nodes.LiftedPort[Distribution[float64]] `description:"Shapes the Z component. Unwired, it spreads evenly over [0, 1)."`
}

func (Compose3DNode) Description() string {
	return "A 3D distribution from one distribution per axis, each drawn independently."
}

func (c Compose3DNode) Out(out *nodes.Lifted[Distribution[vector3.Float64]]) {
	nodes.Zip3(out, nodes.LiftedOr(c.X, Uniform(0, 1)), nodes.LiftedOr(c.Y, Uniform(0, 1)), nodes.LiftedOr(c.Z, Uniform(0, 1)), Compose3D)
}

type Compose4DNode struct {
	X nodes.LiftedPort[Distribution[float64]] `description:"Shapes the X component. Unwired, it spreads evenly over [0, 1)."`
	Y nodes.LiftedPort[Distribution[float64]] `description:"Shapes the Y component. Unwired, it spreads evenly over [0, 1)."`
	Z nodes.LiftedPort[Distribution[float64]] `description:"Shapes the Z component. Unwired, it spreads evenly over [0, 1)."`
	W nodes.LiftedPort[Distribution[float64]] `description:"Shapes the W component. Unwired, it spreads evenly over [0, 1)."`
}

func (Compose4DNode) Description() string {
	return "A 4D distribution from one distribution per axis, each drawn independently."
}

func (c Compose4DNode) Out(out *nodes.Lifted[Distribution[vector4.Float64]]) {
	nodes.Zip4(out, nodes.LiftedOr(c.X, Uniform(0, 1)), nodes.LiftedOr(c.Y, Uniform(0, 1)), nodes.LiftedOr(c.Z, Uniform(0, 1)), nodes.LiftedOr(c.W, Uniform(0, 1)), Compose4D)
}

type Uniform2DNode struct {
	Min nodes.LiftedPort[vector2.Float64]
	Max nodes.LiftedPort[vector2.Float64] `default:"1,1"`
}

func (Uniform2DNode) Description() string {
	return "Every vector between Min and Max equally likely, each component on its own."
}

func (Uniform2DNode) Keywords() []string {
	return []string{"jitter", "offset", "rectangle"}
}

func (u Uniform2DNode) Out(out *nodes.Lifted[Distribution[vector2.Float64]]) {
	nodes.Zip2(out, nodes.LiftedOr(u.Min, vector2.Zero[float64]()), u.Max, Uniform2D)
}

type Uniform3DNode struct {
	Min nodes.LiftedPort[vector3.Float64]
	Max nodes.LiftedPort[vector3.Float64] `default:"1,1,1"`
}

func (Uniform3DNode) Description() string {
	return "Every vector between Min and Max equally likely, each component on its own."
}

func (Uniform3DNode) Keywords() []string {
	return []string{"jitter", "offset", "box"}
}

func (u Uniform3DNode) Out(out *nodes.Lifted[Distribution[vector3.Float64]]) {
	nodes.Zip2(out, nodes.LiftedOr(u.Min, vector3.Zero[float64]()), u.Max, Uniform3D)
}

type Uniform4DNode struct {
	Min nodes.LiftedPort[vector4.Float64]
	Max nodes.LiftedPort[vector4.Float64] `default:"1,1,1,1"`
}

func (Uniform4DNode) Description() string {
	return "Every vector between Min and Max equally likely, each component on its own."
}

func (Uniform4DNode) Keywords() []string {
	return []string{"jitter", "offset"}
}

func (u Uniform4DNode) Out(out *nodes.Lifted[Distribution[vector4.Float64]]) {
	nodes.Zip2(out, nodes.LiftedOr(u.Min, vector4.Zero[float64]()), u.Max, Uniform4D)
}

type InsideCircleNode struct {
	Radius nodes.LiftedPort[float64] `default:"1"`
}

func (InsideCircleNode) Description() string {
	return "Points spread evenly through a disc around the origin."
}

func (InsideCircleNode) Keywords() []string {
	return []string{"disc"}
}

func (i InsideCircleNode) Out(out *nodes.Lifted[Distribution[vector2.Float64]]) {
	nodes.Zip1(out, i.Radius, InsideCircle)
}

type Direction2DNode struct {
	Axis      nodes.LiftedPort[vector2.Float64] `description:"The centre of the arc." default:"0,1"`
	ConeAngle nodes.LiftedPort[float64]         `description:"Degrees either side of Axis a direction may lean, 0 to 180. 0 always gives Axis, 180 any direction."`
}

func (Direction2DNode) Description() string {
	return "Unit vectors spread evenly within an arc around Axis."
}

func (Direction2DNode) Keywords() []string {
	return []string{"orientation", "tilt"}
}

func (d Direction2DNode) Out(out *nodes.Lifted[Distribution[vector2.Float64]]) {
	nodes.Zip2(out, d.Axis, nodes.LiftedOr(d.ConeAngle, 0), Direction2D)
}

type RotationNode struct {
	MaxAngle nodes.LiftedPort[float64] `description:"Degrees a rotation may turn from no rotation, 0 to 180. 180 gives any rotation." default:"180"`
}

func (RotationNode) Description() string {
	return "Rotations spread evenly within Max Angle of no rotation, about any axis."
}

func (RotationNode) Keywords() []string {
	return []string{"quaternion"}
}

func (r RotationNode) Out(out *nodes.Lifted[Distribution[quaternion.Quaternion]]) {
	nodes.Zip1(out, r.MaxAngle, Rotation)
}

type InsideSphereNode struct {
	Radius nodes.LiftedPort[float64] `default:"1"`
}

func (InsideSphereNode) Description() string {
	return "Points spread evenly through a solid sphere around the origin."
}

func (InsideSphereNode) Keywords() []string {
	return []string{"jitter", "ball"}
}

func (i InsideSphereNode) Out(out *nodes.Lifted[Distribution[vector3.Float64]]) {
	nodes.Zip1(out, i.Radius, InsideSphere)
}

type DirectionNode struct {
	Axis      nodes.LiftedPort[vector3.Float64] `description:"The centre of the cone." default:"0,1,0"`
	ConeAngle nodes.LiftedPort[float64]         `description:"Degrees from Axis a direction may lean, 0 to 180. 0 always gives Axis, 180 any direction."`
}

func (DirectionNode) Description() string {
	return "Unit vectors spread evenly within a cone around Axis."
}

func (DirectionNode) Keywords() []string {
	return []string{"orientation", "tilt"}
}

func (d DirectionNode) Out(out *nodes.Lifted[Distribution[vector3.Float64]]) {
	nodes.Zip2(out, d.Axis, nodes.LiftedOr(d.ConeAngle, 0), Direction)
}
