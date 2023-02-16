package trees

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
)

type Tree interface {
	TraverseIntersectingRay(ray geometry.Ray, min, max float64, iterator func(i int, min, max *float64))
	ElementsContainingPoint(v vector3.Float64) []int
	ClosestPoint(v vector3.Float64) (int, vector3.Float64)
	ElementsIntersectingRay(ray geometry.Ray, min, max float64) []int
	BoundingBox() geometry.AABB
}
