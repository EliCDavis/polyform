package trees

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
)

type Element interface {
	BoundingBox() geometry.AABB
	ClosestPoint(p vector3.Float64) vector3.Float64
}

type elementReference struct {
	primitive     Element
	bounds        geometry.AABB
	originalIndex int
}
