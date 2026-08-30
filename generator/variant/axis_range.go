package variant

import "math/rand"

// axisRange is a min/max/samples spec shared by range Dimensions - not
// itself a Dimension.
type axisRange struct {
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Samples int     `json:"samples"`
}

// count is never less than 1.
func (r axisRange) count() int {
	return axisCount(r.Samples)
}

func (r axisRange) value(index int) float64 {
	return axisValue(r.Min, r.Max, r.Samples, index)
}

func (r axisRange) random(rng *rand.Rand) float64 {
	return randomBetween(r.Min, r.Max, rng)
}

func randomBetween(min, max float64, rng *rand.Rand) float64 {
	return min + (max-min)*rng.Float64()
}

// axisCount is never less than 1.
func axisCount(samples int) int {
	if samples < 1 {
		return 1
	}
	return samples
}

func axisValue(min, max float64, samples, index int) float64 {
	n := axisCount(samples)
	if n <= 1 {
		return min
	}
	step := (max - min) / float64(n-1)
	return min + step*float64(index)
}
