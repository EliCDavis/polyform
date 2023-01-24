package sample

import (
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func LinearFloatMapping(fromMin, fromMax, toMin, toMax float64) FloatToFloat {
	fromDif := fromMax - fromMin
	toDif := toMax - toMin
	return func(f float64) float64 {
		percentage := (f - fromMin) / fromDif
		return toMin + (percentage * toDif)
	}
}

func LinearVector2Mapping(fromMin, fromMax float64, toMin, toMax vector2.Float64) FloatToVec2 {
	fromDif := fromMax - fromMin
	toDif := toMax.Sub(toMin)
	return func(f float64) vector2.Float64 {
		percentage := (f - fromMin) / fromDif
		return toMin.Add(toDif.MultByConstant(percentage))
	}
}

func LinearVector3Mapping(fromMin, fromMax float64, toMin, toMax vector3.Float64) FloatToVec3 {
	fromDif := fromMax - fromMin
	toDif := toMax.Sub(toMin)
	return func(f float64) vector3.Float64 {
		percentage := (f - fromMin) / fromDif
		return toMin.Add(toDif.MultByConstant(percentage))
	}
}
