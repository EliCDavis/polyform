package modeling

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector"
)

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

func (l Line) ClosestPoint(atr string, point vector.Vector3) vector.Vector3 {
	line3d := geometry.NewLine3D(
		l.mesh.v3Data[atr][l.startingIndex],
		l.mesh.v3Data[atr][l.startingIndex+1],
	)
	return line3d.ClosestPointOnLine(point)
}
