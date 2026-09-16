package geometry

import (
	"math"

	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

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

// Newell's method; a ring enclosing no area gives a zero normal.
func NewPlaneFromPolygon(ring []vector3.Float64) Plane {
	if len(ring) == 0 {
		return Plane{}
	}

	var center, normal vector3.Float64
	for i, p := range ring {
		center = center.Add(p)
		q := ring[(i+1)%len(ring)]
		normal = normal.Add(vector3.New(
			(p.Y()-q.Y())*(p.Z()+q.Z()),
			(p.Z()-q.Z())*(p.X()+q.X()),
			(p.X()-q.X())*(p.Y()+q.Y()),
		))
	}
	if normal.Length() == 0 {
		return Plane{}
	}

	normal = normal.Normalized()
	return Plane{
		normal:   normal,
		distance: normal.Dot(center.DivByConstant(float64(len(ring)))),
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

// Basis returns unit u and v in the plane with u × v = normal.
// "Building an Orthonormal Basis, Revisited", JCGT 2017.
func (p Plane) Basis() (u, v vector3.Float64) {
	n := p.normal
	sign := math.Copysign(1, n.Z())
	a := -1 / (sign + n.Z())
	b := n.X() * n.Y() * a
	u = vector3.New(1+sign*n.X()*n.X()*a, sign*b, -sign*n.X())
	v = vector3.New(b, sign+n.Y()*n.Y()*a, -n.Y())
	return u, v
}

// Project returns each point's coordinates in Basis; height above the plane is lost.
func (p Plane) Project(points []vector3.Float64) []vector2.Float64 {
	u, v := p.Basis()
	origin := p.Origin()
	flat := make([]vector2.Float64, len(points))
	for i, point := range points {
		d := point.Sub(origin)
		flat[i] = vector2.New(d.Dot(u), d.Dot(v))
	}
	return flat
}
