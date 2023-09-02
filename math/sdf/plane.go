package sdf

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector/vector3"
)

// https://iquilezles.org/articles/distfunctions/
func Plane(pos, normal vector3.Float64, height float64) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		return v.Sub(pos).Dot(normal) + height
	}
}
