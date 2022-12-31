package curves

import (
	"math"

	"github.com/EliCDavis/polyform/math/sample"
)

func clamp(t, min, max float64) float64 {
	return math.Max(min, math.Min(t, max))
}

func clamp01(t float64) float64 {
	return clamp(t, 0, 1)
}

func PowerIn(power int) sample.FloatToFloat {
	return func(f float64) float64 {
		t := clamp01(f)
		return math.Pow(t, float64(power))
	}
}

func PowerOut(power int) sample.FloatToFloat {
	return func(f float64) float64 {
		t := 1.0 - clamp01(f)
		return 1.0 - math.Pow(t, float64(power))
	}
}

func PowerInOut(power int) sample.FloatToFloat {
	return func(f float64) float64 {

		t := clamp01(f)
		if t < 0.5 {
			return math.Pow(t*2, float64(power)) * 0.5
		} else {
			return 1.0 - math.Pow((1-t)*2, float64(power))*0.5
		}
	}
}
