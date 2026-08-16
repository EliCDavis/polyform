package sequence

import "github.com/EliCDavis/polyform/nodes"

type LinearNode struct {
	Start   nodes.Output[float64] `description:"Value of the first sample. Defaults to 0."`
	End     nodes.Output[float64] `description:"Value of the last sample. Defaults to 1."`
	Samples nodes.Output[int]     `description:"How many evenly-spaced values to produce, including both Start and End. Defaults to 0."`
}

func (snd LinearNode) Description() string {
	return "Produces Samples evenly-spaced float64 values from Start to End, inclusive."
}

func (snd LinearNode) Out(out *nodes.StructOutput[[]float64]) {
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
