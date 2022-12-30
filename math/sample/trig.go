package sample

import "math"

func Sin(amplitude, frequency float64) FloatToFloat {
	return func(f float64) float64 {
		return math.Sin(f*frequency*math.Pi*2) * amplitude
	}
}

func Cos(amplitude, frequency float64) FloatToFloat {
	return func(f float64) float64 {
		return math.Cos(f*frequency*math.Pi*2) * amplitude
	}
}
