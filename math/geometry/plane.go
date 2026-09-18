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

// Distance is signed, positive on the side the normal points to.
func (p Plane) Distance(point vector3.Float64) float64 {
	return p.normal.Dot(point) - p.distance
}

func (p Plane) ClosestPoint(point vector3.Float64) vector3.Float64 {
	return point.Sub(p.normal.Scale(p.Distance(point)))
}

// Holds reports whether every corner of t lies within tolerance of the plane.
func (p Plane) Holds(t Triangle, tolerance float64) bool {
	for _, corner := range t {
		if math.Abs(p.Distance(corner)) > tolerance {
			return false
		}
	}
	return true
}

// DominantAxis is the axis the normal points along most: 0, 1 or 2.
func (p Plane) DominantAxis() int {
	return DominantAxis(p.normal)
}

// Lift puts back the coordinate DropAxis removed, placing the point on the
// plane. The plane must not be parallel to axis; DominantAxis is always safe.
func (p Plane) Lift(flat vector2.Float64, axis int) vector3.Float64 {
	n := p.normal
	switch axis {
	case 0:
		y, z := flat.X(), flat.Y()
		return vector3.New((p.distance-n.Y()*y-n.Z()*z)/n.X(), y, z)
	case 1:
		z, x := flat.X(), flat.Y()
		return vector3.New(x, (p.distance-n.Z()*z-n.X()*x)/n.Y(), z)
	}
	x, y := flat.X(), flat.Y()
	return vector3.New(x, y, (p.distance-n.X()*x-n.Y()*y)/n.Z())
}

func DominantAxis(normal vector3.Float64) int {
	x, y, z := normal.Abs().Values()
	switch {
	case x >= y && x >= z:
		return 0
	case y >= z:
		return 1
	}
	return 2
}

// DropAxis flattens a point by removing one coordinate, keeping the other two
// exact. Dropping a plane's DominantAxis keeps at least 1/√3 of any area in it.
func DropAxis(point vector3.Float64, axis int) vector2.Float64 {
	switch axis {
	case 0:
		return vector2.New(point.Y(), point.Z())
	case 1:
		return vector2.New(point.Z(), point.X())
	}
	return vector2.New(point.X(), point.Y())
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

// Project returns each point's coordinates relative to the plane's frame.
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
