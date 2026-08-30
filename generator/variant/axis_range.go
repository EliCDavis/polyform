package variant

import (
	"math"
	"math/rand"
)

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

// lerp interpolates between min and max by t, where t is typically in [0,1].
func lerp(min, max, t float64) float64 {
	return min + (max-min)*t
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

// intAxisValue is axisValue rounded to the nearest whole number, so it can
// back an int-typed Dimension without producing fractional JSON values.
func intAxisValue(min, max, samples, index int) int {
	return int(math.Round(axisValue(float64(min), float64(max), samples, index)))
}

// intLerp is lerp rounded to the nearest whole number.
func intLerp(min, max int, t float64) int {
	return int(math.Round(lerp(float64(min), float64(max), t)))
}
