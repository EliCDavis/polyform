package sdf

import "math"

func clampNormal(val float64) float64 {
	return math.Max(math.Min(val, 1), -1)
}
