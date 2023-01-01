package marching

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/trees"
	"github.com/EliCDavis/vector"
)

func Mesh(mesh modeling.Mesh, radius, strength float64) sample.Vec3ToFloat {
	octree := trees.FromMesh(mesh)
	bounds := mesh.BoundingBox(modeling.PositionAttribute)
	bounds.Expand(radius)
	return func(v vector.Vector3) float64 {

		// closestDist := math.MaxFloat64
		// for i := 0; i < mesh.TriCount(); i++ {
		// 	dist := mesh.Tri(i).ClosestPoint(v).Sub(v).Length()
		// 	if dist < closestDist {
		// 		closestDist = dist
		// 	}
		// }

		if !bounds.Contains(v) {
			return 0
		}

		closestDist := octree.ClosestPoint(v).Distance(v)

		if closestDist <= radius {
			return (radius - closestDist) * strength
		}
		return 0
	}
}
