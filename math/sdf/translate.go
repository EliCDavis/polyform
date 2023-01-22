package sdf

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector"
)

func Translate(field sample.Vec3ToFloat, translation vector.Vector3) sample.Vec3ToFloat {
	return func(v vector.Vector3) float64 {
		return field(v.Sub(translation))
	}
}
