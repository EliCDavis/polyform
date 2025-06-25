package geometry

import "github.com/EliCDavis/vector/vector3"

type Plane struct {
	normal   vector3.Float64
	distance float64
}

func NewPlane(position, normal vector3.Float64) Plane {
	return Plane{
		normal:   normal,
		distance: normal.Dot(position),
	}
}

func NewPlaneFromPoints(a, b, c vector3.Float64) Plane {
	normal := b.Sub(a).Cross(c.Sub(a)).Normalized()
	return Plane{
		normal:   normal,
		distance: normal.Dot(a),
	}
}

func (p Plane) Origin() vector3.Float64 {
	return p.normal.Scale(p.distance)
}

func (p Plane) Normal() vector3.Float64 {
	return p.normal
}

func (p Plane) ClosestPoint(point vector3.Float64) vector3.Float64 {
	distance := p.normal.Dot(point) - p.distance
	return point.Sub(p.normal.Scale(distance))
}
