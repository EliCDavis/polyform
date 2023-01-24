package modeling

import "github.com/EliCDavis/vector/vector3"

type Primitive interface {
	BoundingBox(atr string) AABB
	ClosestPoint(atr string, p vector3.Float64) vector3.Float64
}
