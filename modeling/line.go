package modeling

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector3"
)

type scopedLine struct {
	data []vector3.Float64
	p1   int
	p2   int
}

func (l scopedLine) BoundingBox() geometry.AABB {
	return geometry.NewAABBFromPoints(
		l.data[l.p1],
		l.data[l.p2],
	)
}

func (l scopedLine) ClosestPoint(point vector3.Float64) vector3.Float64 {
	line3d := geometry.NewLine3D(
		l.data[l.p1],
		l.data[l.p2],
	)
	return line3d.ClosestPointOnLine(point)
}

type Line struct {
	mesh          *Mesh
	startingIndex int
}

// P1 is the first point on our triangle, which is an index to the vertices array of a mesh
func (l Line) P1() int {
	return l.mesh.indices[l.startingIndex]
}

// P2 is the second point on our triangle, which is an index to the vertices array of a mesh
func (l Line) P2() int {
	return l.mesh.indices[l.startingIndex+1]
}

func (l Line) BoundingBox(atr string) geometry.AABB {
	return geometry.NewAABBFromPoints(
		l.mesh.v3Data[atr][l.P1()],
		l.mesh.v3Data[atr][l.P2()],
	)
}

func (l Line) ClosestPoint(atr string, point vector3.Float64) vector3.Float64 {
	line3d := geometry.NewLine3D(
		l.mesh.v3Data[atr][l.P1()],
		l.mesh.v3Data[atr][l.P2()],
	)
	return line3d.ClosestPointOnLine(point)
}

func (l Line) Scope(attribute string) trees.Element {
	return scopedLine{
		data: l.mesh.v3Data[attribute],
		p1:   l.P1(),
		p2:   l.P2(),
	}
}
