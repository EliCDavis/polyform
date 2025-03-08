package trig

import (
	"math"

	"github.com/EliCDavis/polyform/nodes"
)

type SinArray = nodes.Struct[SinArrayNodeData]

type SinArrayNodeData struct {
	Input     nodes.Output[[]float64]
	Amplitude nodes.Output[float64]
}

func (snd SinArrayNodeData) Out() nodes.StructOutput[[]float64] {
	if snd.Input == nil {
		return nodes.NewStructOutput[[]float64](nil)
	}

	scale := nodes.TryGetOutputValue(snd.Amplitude, 1)

	in := snd.Input.Value()
	out := make([]float64, len(in))
	for i, v := range in {
		out[i] = math.Sin(v) * scale
	}

	return nodes.NewStructOutput(out)
}

type CosArray = nodes.Struct[CosArrayNodeData]

type CosArrayNodeData struct {
	Input     nodes.Output[[]float64]
	Amplitude nodes.Output[float64]
}

func (snd CosArrayNodeData) Out() nodes.StructOutput[[]float64] {
	if snd.Input == nil {
		return nodes.NewStructOutput[[]float64](nil)
	}

	scale := nodes.TryGetOutputValue(snd.Amplitude, 1)

	in := snd.Input.Value()
	out := make([]float64, len(in))
	for i, v := range in {
		out[i] = math.Cos(v) * scale
	}

	return nodes.NewStructOutput(out)
}
