package sdf

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector/vector3"
)

func Translate(field sample.Vec3ToFloat, translation vector3.Float64) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		return field(v.Sub(translation))
	}
}
