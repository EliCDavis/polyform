package modeling

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
)

type scopedLine struct {
	data          []vector3.Float64
	startingIndex int
}

func (l scopedLine) BoundingBox() AABB {
	return NewAABBFromPoints(
		l.data[l.startingIndex],
		l.data[l.startingIndex+1],
	)
}

func (l scopedLine) ClosestPoint(point vector3.Float64) vector3.Float64 {
	line3d := geometry.NewLine3D(
		l.data[l.startingIndex],
		l.data[l.startingIndex+1],
	)
	return line3d.ClosestPointOnLine(point)
}

type Line struct {
	mesh          *Mesh
	startingIndex int
}

func (l Line) BoundingBox(atr string) AABB {
	return NewAABBFromPoints(
		l.mesh.v3Data[atr][l.startingIndex],
		l.mesh.v3Data[atr][l.startingIndex+1],
	)
}

func (l Line) ClosestPoint(atr string, point vector3.Float64) vector3.Float64 {
	line3d := geometry.NewLine3D(
		l.mesh.v3Data[atr][l.startingIndex],
		l.mesh.v3Data[atr][l.startingIndex+1],
	)
	return line3d.ClosestPointOnLine(point)
}

func (l Line) Scope(attribute string) ScopedPrimitive {
	return scopedLine{
		data:          l.mesh.v3Data[attribute],
		startingIndex: l.startingIndex,
	}
}
