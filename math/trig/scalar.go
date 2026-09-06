package trig

import (
	"math"

	"github.com/EliCDavis/polyform/nodes"
)

// ============================================================================

type SinNode struct {
	In nodes.Output[float64]
}

func (n SinNode) Description() string {
	return "Sine of an angle in radians"
}

func (n SinNode) Out(out *nodes.StructOutput[float64]) {
	out.Set(math.Sin(nodes.TryGetOutputValue(out, n.In, 0)))
}

// ============================================================================

type CosNode struct {
	In nodes.Output[float64]
}

func (n CosNode) Description() string {
	return "Cosine of an angle in radians"
}

func (n CosNode) Out(out *nodes.StructOutput[float64]) {
	out.Set(math.Cos(nodes.TryGetOutputValue(out, n.In, 0)))
}

// ============================================================================

type TanNode struct {
	In nodes.Output[float64]
}

func (n TanNode) Description() string {
	return "Tangent of an angle in radians"
}

func (n TanNode) Out(out *nodes.StructOutput[float64]) {
	out.Set(math.Tan(nodes.TryGetOutputValue(out, n.In, 0)))
}

// ============================================================================

type ArcSinNode struct {
	In nodes.Output[float64]
}

func (n ArcSinNode) Description() string {
	return "Arcsine (asin) of a ratio, returned as an angle in radians"
}

func (n ArcSinNode) Out(out *nodes.StructOutput[float64]) {
	out.Set(math.Asin(nodes.TryGetOutputValue(out, n.In, 0)))
}

// ============================================================================

type ArcCosNode struct {
	In nodes.Output[float64]
}

func (n ArcCosNode) Description() string {
	return "Arccosine (acos) of a ratio, returned as an angle in radians"
}

func (n ArcCosNode) Out(out *nodes.StructOutput[float64]) {
	out.Set(math.Acos(nodes.TryGetOutputValue(out, n.In, 0)))
}

// ============================================================================

type ArcTanNode struct {
	In nodes.Output[float64]
}

func (n ArcTanNode) Description() string {
	return "Arctangent (atan) of a ratio, returned as an angle in radians. For an angle built from separate rise and run values, prefer ArcTan2, which keeps the correct quadrant"
}

func (n ArcTanNode) Out(out *nodes.StructOutput[float64]) {
	out.Set(math.Atan(nodes.TryGetOutputValue(out, n.In, 0)))
}

// ============================================================================

type ArcTan2Node struct {
	Y nodes.Output[float64]
	X nodes.Output[float64]
}

func (n ArcTan2Node) Description() string {
	return "Arctangent (atan2) of Y/X, returned as an angle in radians over the full circle. Use for the angle of a slope from its rise (Y) and run (X)"
}

func (n ArcTan2Node) Out(out *nodes.StructOutput[float64]) {
	out.Set(math.Atan2(
		nodes.TryGetOutputValue(out, n.Y, 0),
		nodes.TryGetOutputValue(out, n.X, 0),
	))
}

// ============================================================================

type DegreesToRadiansNode struct {
	In nodes.Output[float64]
}

func (n DegreesToRadiansNode) Description() string {
	return "Converts degrees to radians, the unit every angle input in the graph expects"
}

func (n DegreesToRadiansNode) Out(out *nodes.StructOutput[float64]) {
	out.Set(nodes.TryGetOutputValue(out, n.In, 0) * math.Pi / 180)
}

// ============================================================================

type RadiansToDegreesNode struct {
	In nodes.Output[float64]
}

func (n RadiansToDegreesNode) Description() string {
	return "Converts radians to degrees, for reading an angle back in human terms"
}

func (n RadiansToDegreesNode) Out(out *nodes.StructOutput[float64]) {
	out.Set(nodes.TryGetOutputValue(out, n.In, 0) * 180 / math.Pi)
}
