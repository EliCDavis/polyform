package modeling

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
)

type Point struct {
	mesh  *Mesh
	index int
}

func (p Point) BoundingBox(atr string) AABB {
	return NewAABB(p.mesh.v3Data[atr][p.index], vector3.Zero[float64]())
}

func (p Point) ClosestPoint(atr string, point vector3.Float64) vector3.Float64 {
	return p.mesh.v3Data[atr][p.index]
}

func (p Point) Clips(plane geometry.Plane, atr string) bool {
	dist := plane.Normal().Dot(p.mesh.v3Data[atr][p.index].Sub(plane.Origin()))

	return dist < 0
}
