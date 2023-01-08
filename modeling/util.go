package modeling

import (
	"math"
)

func Clamp(v, min, max float64) float64 {
	return math.Max(math.Min(v, max), min)
}
