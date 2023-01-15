package geometry

import "github.com/EliCDavis/vector"

type Plane struct {
	normal   vector.Vector3
	distance float64
}

func NewPlaneFromPoints(a, b, c vector.Vector3) Plane {
	normal := b.Sub(a).Cross(c.Sub(a)).Normalized()
	return Plane{
		normal:   normal,
		distance: normal.Dot(a),
	}
}

func (p Plane) Origin() vector.Vector3 {
	return p.normal.MultByConstant(p.distance)
}

func (p Plane) Normal() vector.Vector3 {
	return p.normal
}

func (p Plane) ClosestPoint(point vector.Vector3) vector.Vector3 {
	distance := p.normal.Dot(point) - p.distance
	return point.Sub(p.normal.MultByConstant(distance))
}
