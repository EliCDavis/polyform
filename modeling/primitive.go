package modeling

import "github.com/EliCDavis/vector"

type Primitive interface {
	BoundingBox(atr string) AABB
	ClosestPoint(atr string, p vector.Vector3) vector.Vector3
}
