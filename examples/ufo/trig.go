package main

import (
	"math"

	"github.com/EliCDavis/polyform/nodes"
)

type SinNode = nodes.Struct[[]float64, SinNodeData]

type SinNodeData struct {
	Input nodes.NodeOutput[[]float64]
	Scale nodes.NodeOutput[float64]
}

func (snd SinNodeData) Process() ([]float64, error) {
	if snd.Input == nil {
		return nil, nil
	}

	scale := 1.
	if snd.Scale != nil {
		scale = snd.Scale.Value()
	}

	in := snd.Input.Value()
	out := make([]float64, len(in))
	for i, v := range in {
		out[i] = math.Sin(v) * scale
	}

	return out, nil
}

type CosNode = nodes.Struct[[]float64, CosNodeData]

type CosNodeData struct {
	Input nodes.NodeOutput[[]float64]
	Scale nodes.NodeOutput[float64]
}

func (snd CosNodeData) Process() ([]float64, error) {
	if snd.Input == nil {
		return nil, nil
	}

	scale := 1.
	if snd.Scale != nil {
		scale = snd.Scale.Value()
	}

	in := snd.Input.Value()
	out := make([]float64, len(in))
	for i, v := range in {
		out[i] = math.Cos(v) * scale
	}

	return out, nil
}
