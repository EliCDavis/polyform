package sequence

import "github.com/EliCDavis/polyform/nodes"

type LinearNode struct {
	Start   nodes.Output[float64] `description:"Value of the first sample."`
	End     nodes.Output[float64] `description:"Value of the last sample." default:"1"`
	Samples nodes.Output[int]     `description:"How many evenly-spaced values to produce, including both Start and End."`
}

func (snd LinearNode) Description() string {
	return "Produces Samples evenly-spaced float64 values from Start to End, inclusive."
}

func (snd LinearNode) Out(out *nodes.StructOutput[[]float64]) {
	start := nodes.GetOutputValue(out, snd.Start)
	end := nodes.GetOutputValue(out, snd.End)
	samples := max(nodes.GetOutputValue(out, snd.Samples), 0)

	if samples == 0 {
		return
	}

	arr := make([]float64, samples)
	if samples == 1 {
		arr[0] = start
		out.Set(arr)
		return
	}

	inc := (end - start) / float64(samples-1)
	for i := range samples {
		v := start + (float64(i) * inc)
		arr[i] = v
	}

	out.Set(arr)
}
