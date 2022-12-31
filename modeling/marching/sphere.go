package marching

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector"
)

func Sphere(pos vector.Vector3, radius, strength float64) sample.Vec3ToFloat {
	return func(v vector.Vector3) float64 {
		dist := v.Distance(pos)
		if dist <= radius {
			return (radius - dist) * strength
		}
		return 0
	}
}
