package sdf

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector/vector3"
)

func Sphere(position vector3.Float64, radius float64) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		return v.Distance(position) - radius
	}
}
