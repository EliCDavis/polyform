package trig

import (
	"math"

	"github.com/EliCDavis/polyform/nodes"
)

func runFunction(
	out *nodes.StructOutput[[]float64],
	Input nodes.Output[[]float64],
	Amplitude nodes.Output[float64],
	Shift nodes.Output[float64],
	f func(x float64) float64,
) {
	if Input == nil {
		return
	}

	scale := nodes.TryGetOutputValue(out, Amplitude, 1)
	shift := nodes.TryGetOutputValue(out, Shift, 0)

	in := nodes.GetOutputValue(out, Input)
	arr := make([]float64, len(in))
	for i, v := range in {
		arr[i] = f(v+shift) * scale
	}
	out.Set(arr)
}

// ============================================================================

type SinArray struct {
	Input     nodes.Output[[]float64]
	Amplitude nodes.Output[float64]
	Shift     nodes.Output[float64]
}

func (n SinArray) Out(out *nodes.StructOutput[[]float64]) {
	runFunction(out, n.Input, n.Amplitude, n.Shift, math.Sin)
}

// ============================================================================

type CosArray struct {
	Input     nodes.Output[[]float64]
	Amplitude nodes.Output[float64]
	Shift     nodes.Output[float64]
}

func (n CosArray) Out(out *nodes.StructOutput[[]float64]) {
	runFunction(out, n.Input, n.Amplitude, n.Shift, math.Cos)
}

// ============================================================================

type TanArray struct {
	Input     nodes.Output[[]float64]
	Amplitude nodes.Output[float64]
	Shift     nodes.Output[float64]
}

func (n TanArray) Out(out *nodes.StructOutput[[]float64]) {
	runFunction(out, n.Input, n.Amplitude, n.Shift, math.Tan)
}

// ============================================================================

type ArcSinArray struct {
	Input     nodes.Output[[]float64]
	Amplitude nodes.Output[float64]
	Shift     nodes.Output[float64]
}

func (n ArcSinArray) Out(out *nodes.StructOutput[[]float64]) {
	runFunction(out, n.Input, n.Amplitude, n.Shift, math.Asin)
}

// ============================================================================

type ArcCosArray struct {
	Input     nodes.Output[[]float64]
	Amplitude nodes.Output[float64]
	Shift     nodes.Output[float64]
}

func (n ArcCosArray) Out(out *nodes.StructOutput[[]float64]) {
	runFunction(out, n.Input, n.Amplitude, n.Shift, math.Acos)
}

// ============================================================================

type ArcTanArray struct {
	Input     nodes.Output[[]float64]
	Amplitude nodes.Output[float64]
	Shift     nodes.Output[float64]
}

func (n ArcTanArray) Out(out *nodes.StructOutput[[]float64]) {
	runFunction(out, n.Input, n.Amplitude, n.Shift, math.Atan)
}
