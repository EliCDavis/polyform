package sdf

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector"
)

func Sphere(pos vector.Vector3, radius float64) sample.Vec3ToFloat {
	return func(v vector.Vector3) float64 {
		return v.Distance(pos) - radius
	}
}
