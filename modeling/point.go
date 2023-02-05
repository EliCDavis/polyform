package modeling

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector3"
)

type scopedPoint vector3.Float64

func (p scopedPoint) BoundingBox() geometry.AABB {
	return geometry.NewAABB(vector3.Float64(p), vector3.Zero[float64]())
}

func (p scopedPoint) ClosestPoint(point vector3.Float64) vector3.Float64 {
	return vector3.Float64(p)
}

type Point struct {
	mesh  *Mesh
	index int
}

func (p Point) BoundingBox(atr string) geometry.AABB {
	return geometry.NewAABB(p.mesh.v3Data[atr][p.index], vector3.Zero[float64]())
}

func (p Point) ClosestPoint(atr string, point vector3.Float64) vector3.Float64 {
	return p.mesh.v3Data[atr][p.index]
}

func (p Point) Clips(plane geometry.Plane, atr string) bool {
	dist := plane.Normal().Dot(p.mesh.v3Data[atr][p.index].Sub(plane.Origin()))

	return dist < 0
}

func (p Point) Scope(attribute string) trees.Element {
	return scopedPoint(p.mesh.v3Data[attribute][p.index])
}
