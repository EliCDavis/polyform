package marching

import "math"

func weldPrecisionFor(cubesPerUnit float64) int {
	if cubesPerUnit <= 1 {
		return 3
	}
	return int(math.Ceil(math.Log10(cubesPerUnit))) + 3
}
