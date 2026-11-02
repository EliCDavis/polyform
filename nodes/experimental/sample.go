package experimental

import (
	"github.com/EliCDavis/polyform/nodes"
)

type SampleNode struct {
	Start   nodes.Output[float64]
	End     nodes.Output[float64] `default:"1"`
	Samples nodes.Output[int]
}

func (snd SampleNode) Description() string {
	return "Evenly spaced numbers from Start to End."
}

func (snd SampleNode) Out(out *nodes.StructOutput[[]float64]) {
	start := nodes.GetOutputValue(out, snd.Start)
	end := nodes.GetOutputValue(out, snd.End)
	samples := max(nodes.GetOutputValue(out, snd.Samples), 0)

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

func (snd ShiftNode) Description() string {
	return "Adds a number to every entry of an array."
}

func (snd ShiftNode) Out(out *nodes.StructOutput[[]float64]) {
	if snd.In == nil {
		return
	}

	in := nodes.GetOutputValue(out, snd.In)
	shift := nodes.GetOutputValue(out, snd.Shift)

	arr := make([]float64, len(in))
	for i, v := range in {
		arr[i] = v + shift
	}

	out.Set(arr)
}
