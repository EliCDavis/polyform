package modeling

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector3"
)

type Primitive interface {
	BoundingBox(attribute string) geometry.AABB
	ClosestPoint(attribute string, p vector3.Float64) vector3.Float64
	Scope(attribute string) trees.Element
}
