package marching

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/trees"
	"github.com/EliCDavis/vector/vector3"
)

func Mesh(mesh modeling.Mesh, radius, strength float64) Field {
	octree := trees.FromMesh(mesh)
	bounds := mesh.BoundingBox(modeling.PositionAttribute)
	bounds.Expand(radius)
	return Field{
		Domain: bounds,
		Float1Functions: map[string]sample.Vec3ToFloat{
			modeling.PositionAttribute: func(v vector3.Float64) float64 {
				if !bounds.Contains(v) {
					return 0
				}

				_, closestPoint := octree.ClosestPoint(v)
				closestDist := closestPoint.Distance(v)

				if closestDist < radius {
					return (radius - closestDist) * strength
				}
				return 0
			},
		},
	}
}
