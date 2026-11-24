package scalar

import (
	"cmp"
	"math"
)

func Clamp[T cmp.Ordered](v, lo, hi T) T {
	return min(max(v, lo), hi)
}

func Fract(v float64) float64 {
	return v - math.Floor(v)
}

func Modulo[T int | float64](a, b T) T {
	if b == 0 {
		return 0
	}
	var remainder T
	switch x := any(a).(type) {
	case int:
		remainder = T(x % int(b))
	case float64:
		remainder = T(math.Mod(x, float64(b)))
	}
	if remainder != 0 && (remainder < 0) != (b < 0) {
		remainder += b
	}
	return remainder
}

func SmoothStep(v, from, to float64) float64 {
	if from == to {
		if v < from {
			return 0
		}
		return 1
	}
	t := Clamp((v-from)/(to-from), 0, 1)
	return t * t * (3 - 2*t)
}
