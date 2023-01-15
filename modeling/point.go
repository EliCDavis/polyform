package modeling

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector"
)

type Point struct {
	mesh  *Mesh
	index int
}

func (p Point) BoundingBox(atr string) AABB {
	return NewAABB(p.mesh.v3Data[atr][p.index], vector.Vector3Zero())
}

func (p Point) ClosestPoint(atr string, point vector.Vector3) vector.Vector3 {
	return p.mesh.v3Data[atr][p.index]
}

func (p Point) Clips(plane geometry.Plane, atr string) bool {
	dist := plane.Normal().Dot(p.mesh.v3Data[atr][p.index].Sub(plane.Origin()))

	return dist < 0
}
