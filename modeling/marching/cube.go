package marching

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

func BoundingBox(box modeling.AABB, strength float64) sample.Vec3ToFloat {
	return func(v vector.Vector3) float64 {

		if !box.Contains(v) {
			return 0
		}

		return v.Distance(box.Center()) * strength
	}
}
