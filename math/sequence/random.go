package sequence

import (
	"github.com/EliCDavis/polyform/math/chance"
	"github.com/EliCDavis/polyform/nodes"
)

type RandomFloatNode struct {
	Min     nodes.Output[float64] `description:"Lower bound of the random range, inclusive."`
	Max     nodes.Output[float64] `description:"Upper bound of the random range, exclusive." default:"1"`
	Samples nodes.Output[int]     `description:"How many values to produce."`
	Seed    nodes.Output[int]     `description:"The same seed always produces the same values."`
}

func (snd RandomFloatNode) Description() string {
	return "Produces Samples random float64 values in [Min, Max). The same inputs always produce the same output."
}

func (snd RandomFloatNode) Out(out *nodes.StructOutput[[]float64]) {
	minV := nodes.GetOutputValue(out, snd.Min)
	maxV := nodes.GetOutputValue(out, snd.Max)
	rangeV := maxV - minV

	samples := max(nodes.GetOutputValue(out, snd.Samples), 0)

	rnd := chance.FromSeed(nodes.GetOutputValue(out, snd.Seed))
	arr := make([]float64, samples)
	for i := range samples {
		v := minV + (rnd.Float64() * rangeV)
		arr[i] = v
	}

	out.Set(arr)
}

type RandomBoolNode struct {
	Samples nodes.Output[int] `description:"How many values to produce."`
	Seed    nodes.Output[int] `description:"The same seed always produces the same values."`
}

func (snd RandomBoolNode) Description() string {
	return "Produces Samples random bool values (~50/50). The same inputs always produce the same output."
}

func (snd RandomBoolNode) Out(out *nodes.StructOutput[[]bool]) {
	samples := max(nodes.GetOutputValue(out, snd.Samples), 0)

	rnd := chance.FromSeed(nodes.GetOutputValue(out, snd.Seed))
	arr := make([]bool, samples)
	for i := range samples {
		arr[i] = rnd.Float64() > 0.5
	}

	out.Set(arr)
}
