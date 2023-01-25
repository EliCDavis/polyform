package modeling

import "github.com/EliCDavis/vector/vector3"

type ScopedPrimitive interface {
	BoundingBox() AABB
	ClosestPoint(p vector3.Float64) vector3.Float64
}

type Primitive interface {
	BoundingBox(attribute string) AABB
	ClosestPoint(attribute string, p vector3.Float64) vector3.Float64
	Scope(attribute string) ScopedPrimitive
}
