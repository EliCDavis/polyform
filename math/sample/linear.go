package sample

import "github.com/EliCDavis/vector"

func LinearFloatMapping(fromMin, fromMax, toMin, toMax float64) FloatToFloat {
	fromDif := fromMax - fromMin
	toDif := toMax - toMin
	return func(f float64) float64 {
		percentage := (f - fromMin) / fromDif
		return toMin + (percentage * toDif)
	}
}

func LinearVector2Mapping(fromMin, fromMax float64, toMin, toMax vector.Vector2) FloatToVec2 {
	fromDif := fromMax - fromMin
	toDif := toMax.Sub(toMin)
	return func(f float64) vector.Vector2 {
		percentage := (f - fromMin) / fromDif
		return toMin.Add(toDif.MultByConstant(percentage))
	}
}

func LinearVector3Mapping(fromMin, fromMax float64, toMin, toMax vector.Vector3) FloatToVec3 {
	fromDif := fromMax - fromMin
	toDif := toMax.Sub(toMin)
	return func(f float64) vector.Vector3 {
		percentage := (f - fromMin) / fromDif
		return toMin.Add(toDif.MultByConstant(percentage))
	}
}
