package trig

import (
	"math"

	"github.com/EliCDavis/polyform/nodes"
)

func runFunction(
	Input nodes.Output[[]float64],
	Amplitude nodes.Output[float64],
	Shift nodes.Output[float64],
	f func(x float64) float64,
) []float64 {
	if Input == nil {
		return nil
	}

	scale := nodes.TryGetOutputValue(Amplitude, 1)
	shift := nodes.TryGetOutputValue(Shift, 0)

	in := Input.Value()
	out := make([]float64, len(in))
	for i, v := range in {
		out[i] = f(v+shift) * scale
	}

	return out
}

// ============================================================================

type SinArray = nodes.Struct[SinArrayNodeData]

type SinArrayNodeData struct {
	Input     nodes.Output[[]float64]
	Amplitude nodes.Output[float64]
	Shift     nodes.Output[float64]
}

func (n SinArrayNodeData) Out() nodes.StructOutput[[]float64] {
	return nodes.NewStructOutput(runFunction(n.Input, n.Amplitude, n.Shift, math.Sin))
}

// ============================================================================

type CosArray = nodes.Struct[CosArrayNodeData]

type CosArrayNodeData struct {
	Input     nodes.Output[[]float64]
	Amplitude nodes.Output[float64]
	Shift     nodes.Output[float64]
}

func (n CosArrayNodeData) Out() nodes.StructOutput[[]float64] {
	return nodes.NewStructOutput(runFunction(n.Input, n.Amplitude, n.Shift, math.Cos))
}

// ============================================================================

type TanArray struct {
	Input     nodes.Output[[]float64]
	Amplitude nodes.Output[float64]
	Shift     nodes.Output[float64]
}

func (n TanArray) Out() nodes.StructOutput[[]float64] {
	return nodes.NewStructOutput(runFunction(n.Input, n.Amplitude, n.Shift, math.Tan))
}

// ============================================================================

type ArcSinArray struct {
	Input     nodes.Output[[]float64]
	Amplitude nodes.Output[float64]
	Shift     nodes.Output[float64]
}

func (n ArcSinArray) Out() nodes.StructOutput[[]float64] {
	return nodes.NewStructOutput(runFunction(n.Input, n.Amplitude, n.Shift, math.Asin))
}

// ============================================================================

type ArcCosArray struct {
	Input     nodes.Output[[]float64]
	Amplitude nodes.Output[float64]
	Shift     nodes.Output[float64]
}

func (n ArcCosArray) Out() nodes.StructOutput[[]float64] {
	return nodes.NewStructOutput(runFunction(n.Input, n.Amplitude, n.Shift, math.Acos))
}

// ============================================================================

type ArcTanArray struct {
	Input     nodes.Output[[]float64]
	Amplitude nodes.Output[float64]
	Shift     nodes.Output[float64]
}

func (n ArcTanArray) Out() nodes.StructOutput[[]float64] {
	return nodes.NewStructOutput(runFunction(n.Input, n.Amplitude, n.Shift, math.Atan))
}
