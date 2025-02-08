package experimental

import (
	"github.com/EliCDavis/polyform/nodes"
)

type SampleNode = nodes.Struct[[]float64, SampleNodeData]

type SampleNodeData struct {
	Start   nodes.NodeOutput[float64]
	End     nodes.NodeOutput[float64]
	Samples nodes.NodeOutput[int]
}

func (snd SampleNodeData) Process() ([]float64, error) {
	start := nodes.TryGetOutputValue(snd.Start, 0.)
	end := nodes.TryGetOutputValue(snd.End, 1.)
	samples := max(nodes.TryGetOutputValue(snd.Samples, 0), 0)

	out := make([]float64, samples)
	inc := (end - start) / float64(samples-1)
	for i := 0; i < samples; i++ {
		v := start + (float64(i) * inc)
		out[i] = v
	}

	return out, nil
}

type ShiftNode = nodes.Struct[[]float64, ShiftNodeData]

type ShiftNodeData struct {
	In    nodes.NodeOutput[[]float64]
	Shift nodes.NodeOutput[float64]
}

func (snd ShiftNodeData) Process() ([]float64, error) {
	if snd.In == nil {
		return nil, nil
	}

	if snd.Shift == nil {
		return snd.In.Value(), nil
	}

	in := snd.In.Value()
	shift := snd.Shift.Value()

	out := make([]float64, len(in))
	for i, v := range in {
		out[i] = v + shift
	}

	return out, nil
}
