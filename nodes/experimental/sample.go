package experimental

import (
	"github.com/EliCDavis/polyform/nodes"
)

type SampleNode struct {
	Start   nodes.Output[float64]
	End     nodes.Output[float64]
	Samples nodes.Output[int]
}

func (snd SampleNode) Out(out *nodes.StructOutput[[]float64]) {
	start := nodes.TryGetOutputValue(out, snd.Start, 0.)
	end := nodes.TryGetOutputValue(out, snd.End, 1.)
	samples := max(nodes.TryGetOutputValue(out, snd.Samples, 0), 0)

	arr := make([]float64, samples)
	inc := (end - start) / float64(samples-1)
	for i := range samples {
		v := start + (float64(i) * inc)
		arr[i] = v
	}

	out.Set(arr)
}

type ShiftNode struct {
	In    nodes.Output[[]float64]
	Shift nodes.Output[float64]
}

func (snd ShiftNode) Out(out *nodes.StructOutput[[]float64]) {
	if snd.In == nil {
		return
	}

	in := nodes.GetOutputValue(out, snd.In)
	if snd.Shift == nil {
		out.Set(in)
		return
	}

	shift := nodes.GetOutputValue(out, snd.Shift)

	arr := make([]float64, len(in))
	for i, v := range in {
		arr[i] = v + shift
	}

	out.Set(arr)
}
