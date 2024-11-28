package spz

import "math"

func halfToFloat(h uint16) float64 {
	exponent := ((h >> 10) & 0x1f)
	mantissa := (h & 0x3ff)

	signMul := 1.0
	if ((h >> 15) & 0x1) == 1 {
		signMul = -1.0
	}

	if exponent == 0 {
		// Subnormal numbers (no exponent, 0 in the mantissa decimal).
		return signMul * math.Pow(2.0, -14.0) * float64(mantissa) / 1024.0
	}

	if exponent == 31 {
		// Infinity or NaN.
		if mantissa != 0 {
			return math.NaN()
		} else {
			return math.Inf(int(signMul))
		}
	}

	// non-zero exponent implies 1 in the mantissa decimal.
	return signMul * math.Pow(2.0, float64(exponent)-15.0) * (1.0 + float64(mantissa)/1024.0)
}
