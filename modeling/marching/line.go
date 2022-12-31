package marching

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

func Line(start, end vector.Vector3, radius float64) sample.Vec3ToFloat {
	line := modeling.NewLine3D(start, end)
	return func(v vector.Vector3) float64 {
		closestPoint := line.ClosestPointOnLine(v)
		dist := v.Distance(closestPoint)
		if dist <= radius {
			return radius - dist
		}
		return 0
	}
}
