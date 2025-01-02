package basics

import "github.com/EliCDavis/polyform/nodes"

type SampleNode = nodes.Struct[[]float64, SampleNodeData]

type SampleNodeData struct {
	Start   nodes.NodeOutput[float64]
	End     nodes.NodeOutput[float64]
	Samples nodes.NodeOutput[int]
}

func (snd SampleNodeData) Process() ([]float64, error) {
	if snd.Samples == nil {
		return nil, nil
	}

	start := 0.
	end := 1.
	samples := max(snd.Samples.Value(), 0)

	if snd.Start != nil {
		start = snd.Start.Value()
	}

	if snd.End != nil {
		end = snd.End.Value()
	}

	out := make([]float64, samples)
	inc := (end - start) / float64(samples-1)
	for i := 0; i < samples; i++ {
		v := start + (float64(i) * inc)
		out[i] = v
	}

	return out, nil
}
