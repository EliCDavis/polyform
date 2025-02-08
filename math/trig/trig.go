package trig

import (
	"math"

	"github.com/EliCDavis/polyform/nodes"
)

type SinArray = nodes.Struct[[]float64, SinArrayNodeData]

type SinArrayNodeData struct {
	Input     nodes.NodeOutput[[]float64]
	Amplitude nodes.NodeOutput[float64]
}

func (snd SinArrayNodeData) Process() ([]float64, error) {
	if snd.Input == nil {
		return nil, nil
	}

	scale := nodes.TryGetOutputValue(snd.Amplitude, 1)

	in := snd.Input.Value()
	out := make([]float64, len(in))
	for i, v := range in {
		out[i] = math.Sin(v) * scale
	}

	return out, nil
}

type CosArray = nodes.Struct[[]float64, CosArrayNodeData]

type CosArrayNodeData struct {
	Input     nodes.NodeOutput[[]float64]
	Amplitude nodes.NodeOutput[float64]
}

func (snd CosArrayNodeData) Process() ([]float64, error) {
	if snd.Input == nil {
		return nil, nil
	}

	scale := nodes.TryGetOutputValue(snd.Amplitude, 1)

	in := snd.Input.Value()
	out := make([]float64, len(in))
	for i, v := range in {
		out[i] = math.Cos(v) * scale
	}

	return out, nil
}
