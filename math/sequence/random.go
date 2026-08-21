package sequence

import (
	"math/rand/v2"

	"github.com/EliCDavis/polyform/nodes"
)

type RandomFloatNode struct {
	Min     nodes.Output[float64] `description:"Lower bound of the random range, inclusive. Defaults to 0."`
	Max     nodes.Output[float64] `description:"Upper bound of the random range, exclusive. Defaults to 1."`
	Samples nodes.Output[int]     `description:"How many values to produce. Defaults to 0."`
}

func (snd RandomFloatNode) Description() string {
	return "Produces Samples random float64 values in [Min, Max). Uses a fixed seed. The same inputs always produce the same output."
}

func (snd RandomFloatNode) Out(out *nodes.StructOutput[[]float64]) {
	minV := nodes.TryGetOutputValue(out, snd.Min, 0.)
	maxV := nodes.TryGetOutputValue(out, snd.Max, 1.)
	rangeV := maxV - minV

	samples := max(nodes.TryGetOutputValue(out, snd.Samples, 0), 0)

	seed1 := uint64(12345)
	seed2 := uint64(67890)
	rnd := rand.New(rand.NewPCG(seed1, seed2))
	arr := make([]float64, samples)
	for i := range samples {
		v := minV + (rnd.Float64() * rangeV)
		arr[i] = v
	}

	out.Set(arr)
}

type RandomBoolNode struct {
	Samples nodes.Output[int] `description:"How many values to produce. Defaults to 0."`
}

func (snd RandomBoolNode) Description() string {
	return "Produces Samples random bool values (~50/50). Uses a fixed seed — the same inputs always produce the same output."
}

func (snd RandomBoolNode) Out(out *nodes.StructOutput[[]bool]) {
	samples := max(nodes.TryGetOutputValue(out, snd.Samples, 0), 0)

	seed1 := uint64(12345)
	seed2 := uint64(67890)
	rnd := rand.New(rand.NewPCG(seed1, seed2))
	arr := make([]bool, samples)
	for i := range samples {
		arr[i] = rnd.Float64() > 0.5
	}

	out.Set(arr)
}
