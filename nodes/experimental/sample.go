package experimental

import (
	"github.com/EliCDavis/polyform/nodes"
)

type SampleNode = nodes.Struct[SampleNodeData]

type SampleNodeData struct {
	Start   nodes.Output[float64]
	End     nodes.Output[float64]
	Samples nodes.Output[int]
}

func (snd SampleNodeData) Out() nodes.StructOutput[[]float64] {
	start := nodes.TryGetOutputValue(snd.Start, 0.)
	end := nodes.TryGetOutputValue(snd.End, 1.)
	samples := max(nodes.TryGetOutputValue(snd.Samples, 0), 0)

	out := make([]float64, samples)
	inc := (end - start) / float64(samples-1)
	for i := 0; i < samples; i++ {
		v := start + (float64(i) * inc)
		out[i] = v
	}

	return nodes.NewStructOutput(out)
}

type ShiftNode = nodes.Struct[ShiftNodeData]

type ShiftNodeData struct {
	In    nodes.Output[[]float64]
	Shift nodes.Output[float64]
}

func (snd ShiftNodeData) Out() nodes.StructOutput[[]float64] {
	if snd.In == nil {
		return nodes.NewStructOutput[[]float64](nil)
	}

	if snd.Shift == nil {
		return nodes.NewStructOutput(snd.In.Value())
	}

	in := snd.In.Value()
	shift := snd.Shift.Value()

	out := make([]float64, len(in))
	for i, v := range in {
		out[i] = v + shift
	}

	return nodes.NewStructOutput(out)
}
