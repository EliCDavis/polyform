package trees

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
)

type elementReference struct {
	primitive     Element
	bounds        geometry.AABB
	originalIndex int
}

type Element interface {
	BoundingBox() geometry.AABB
	ClosestPoint(p vector3.Float64) vector3.Float64
}

type BoundingBoxElement geometry.AABB

func (be BoundingBoxElement) BoundingBox() geometry.AABB {
	return geometry.AABB(be)
}

func (be BoundingBoxElement) ClosestPoint(p vector3.Float64) vector3.Float64 {
	return geometry.AABB(be).ClosestPoint(p)
}
