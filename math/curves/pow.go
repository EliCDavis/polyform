package curves

import (
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/scalar"
)

func PowerIn(power int) sample.FloatToFloat {
	return func(f float64) float64 {
		t := scalar.Clamp(f, 0, 1)
		return math.Pow(t, float64(power))
	}
}

func PowerOut(power int) sample.FloatToFloat {
	return func(f float64) float64 {
		t := 1.0 - scalar.Clamp(f, 0, 1)
		return 1.0 - math.Pow(t, float64(power))
	}
}

func PowerInOut(power int) sample.FloatToFloat {
	return func(f float64) float64 {

		t := scalar.Clamp(f, 0, 1)
		if t < 0.5 {
			return math.Pow(t*2, float64(power)) * 0.5
		} else {
			return 1.0 - math.Pow((1-t)*2, float64(power))*0.5
		}
	}
}
