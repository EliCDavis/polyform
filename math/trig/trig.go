package trig

import (
	"math"

	"github.com/EliCDavis/polyform/nodes"
)

func wave(
	out *nodes.Lifted[float64],
	input nodes.LiftedPort[float64],
	amplitude nodes.LiftedPort[float64],
	shift nodes.LiftedPort[float64],
	f func(float64) float64,
) {
	nodes.Zip3(
		out,
		input,
		nodes.LiftedOr(amplitude, 1.),
		nodes.LiftedOr(shift, 0.),
		func(x, a, s float64) float64 {
			return f(x+s) * a
		},
	)
}

// ============================================================================

type SinNode struct {
	Input     nodes.LiftedPort[float64] `description:"Angle in radians."`
	Amplitude nodes.LiftedPort[float64] `description:"Multiplies every result: Amplitude * f(Input + Shift). Defaults to 1."`
	Shift     nodes.LiftedPort[float64] `description:"Added to the input before the function, a phase shift, not an output offset: Amplitude * f(Input + Shift). Defaults to 0."`
}

func (n SinNode) Description() string {
	return "Sine of an angle in radians"
}

func (n SinNode) Out(out *nodes.Lifted[float64]) {
	wave(out, n.Input, n.Amplitude, n.Shift, math.Sin)
}

// ============================================================================

type CosNode struct {
	Input     nodes.LiftedPort[float64] `description:"Angle in radians."`
	Amplitude nodes.LiftedPort[float64] `description:"Multiplies every result: Amplitude * f(Input + Shift). Defaults to 1."`
	Shift     nodes.LiftedPort[float64] `description:"Added to the input before the function, a phase shift, not an output offset: Amplitude * f(Input + Shift). Defaults to 0."`
}

func (n CosNode) Description() string {
	return "Cosine of an angle in radians"
}

func (n CosNode) Out(out *nodes.Lifted[float64]) {
	wave(out, n.Input, n.Amplitude, n.Shift, math.Cos)
}

// ============================================================================

type TanNode struct {
	Input     nodes.LiftedPort[float64] `description:"Angle in radians."`
	Amplitude nodes.LiftedPort[float64] `description:"Multiplies every result: Amplitude * f(Input + Shift). Defaults to 1."`
	Shift     nodes.LiftedPort[float64] `description:"Added to the input before the function, a phase shift, not an output offset: Amplitude * f(Input + Shift). Defaults to 0."`
}

func (n TanNode) Description() string {
	return "Tangent of an angle in radians"
}

func (n TanNode) Out(out *nodes.Lifted[float64]) {
	wave(out, n.Input, n.Amplitude, n.Shift, math.Tan)
}

// ============================================================================

type ArcSinNode struct {
	Input     nodes.LiftedPort[float64] `description:"The ratio."`
	Amplitude nodes.LiftedPort[float64] `description:"Multiplies every result: Amplitude * f(Input + Shift). Defaults to 1."`
	Shift     nodes.LiftedPort[float64] `description:"Added to the input before the function: Amplitude * f(Input + Shift). Defaults to 0."`
}

func (n ArcSinNode) Description() string {
	return "Arcsine (asin) of a ratio, returned as an angle in radians"
}

func (n ArcSinNode) Out(out *nodes.Lifted[float64]) {
	wave(out, n.Input, n.Amplitude, n.Shift, math.Asin)
}

// ============================================================================

type ArcCosNode struct {
	Input     nodes.LiftedPort[float64] `description:"The ratio."`
	Amplitude nodes.LiftedPort[float64] `description:"Multiplies every result: Amplitude * f(Input + Shift). Defaults to 1."`
	Shift     nodes.LiftedPort[float64] `description:"Added to the input before the function: Amplitude * f(Input + Shift). Defaults to 0."`
}

func (n ArcCosNode) Description() string {
	return "Arccosine (acos) of a ratio, returned as an angle in radians"
}

func (n ArcCosNode) Out(out *nodes.Lifted[float64]) {
	wave(out, n.Input, n.Amplitude, n.Shift, math.Acos)
}

// ============================================================================

type ArcTanNode struct {
	Input     nodes.LiftedPort[float64] `description:"The ratio."`
	Amplitude nodes.LiftedPort[float64] `description:"Multiplies every result: Amplitude * f(Input + Shift). Defaults to 1."`
	Shift     nodes.LiftedPort[float64] `description:"Added to the input before the function: Amplitude * f(Input + Shift). Defaults to 0."`
}

func (n ArcTanNode) Description() string {
	return "Arctangent (atan) of a ratio, returned as an angle in radians."
}

func (n ArcTanNode) Out(out *nodes.Lifted[float64]) {
	wave(out, n.Input, n.Amplitude, n.Shift, math.Atan)
}

// ============================================================================

type ArcTan2Node struct {
	Y nodes.LiftedPort[float64] `description:"Rise of the direction, or an array of rises."`
	X nodes.LiftedPort[float64] `description:"Run of the direction, or an array of runs."`
}

func (n ArcTan2Node) Description() string {
	return "Arctangent (atan2) of Y/X, returned as an angle in radians over the full circle."
}

func (n ArcTan2Node) Keywords() []string {
	return []string{"angle", "heading"}
}

func (n ArcTan2Node) Out(out *nodes.Lifted[float64]) {
	nodes.Zip2(out, n.Y, n.X, math.Atan2)
}

// ============================================================================

type DegreesToRadiansNode struct {
	Input nodes.LiftedPort[float64] `description:"Angle in degrees."`
}

func (n DegreesToRadiansNode) Description() string {
	return "Converts degrees to radians."
}

func (n DegreesToRadiansNode) Out(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, n.Input, func(v float64) float64 {
		return v * math.Pi / 180
	})
}

// ============================================================================

type RadiansToDegreesNode struct {
	Input nodes.LiftedPort[float64] `description:"Angle in radians."`
}

func (n RadiansToDegreesNode) Description() string {
	return "Converts radians to degrees."
}

func (n RadiansToDegreesNode) Out(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, n.Input, func(v float64) float64 {
		return v * 180 / math.Pi
	})
}
