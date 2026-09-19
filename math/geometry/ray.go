package geometry

import (
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Ray struct {
	origin    vector3.Float64
	direction vector3.Float64
}

func NewRay(origin vector3.Float64, direction vector3.Float64) Ray {
	return Ray{
		origin:    origin,
		direction: direction.Normalized(),
	}
}

func (r Ray) Origin() vector3.Float64 {
	return r.origin
}

func (r Ray) Direction() vector3.Float64 {
	return r.direction
}

func (r Ray) At(t float64) vector3.Float64 {
	return r.origin.Add(r.direction.Scale(t))
}

func (r Ray) TimeOnRay(v vector3.Float64) float64 {
	adjusted := v.Sub(r.origin)
	return adjusted.Dot(r.direction)
}

// TriangleHit is where a ray meets a triangle's plane, which may be outside
// the triangle.
type TriangleHit struct {
	// How far along the ray the hit is.
	Distance float64

	// Barycentric weights of the second and third corners.
	UV vector2.Float64
}

// Inside allows the weights to fall short of the rim by slack.
func (h TriangleHit) Inside(slack float64) bool {
	return h.UV.X() >= -slack && h.UV.Y() >= -slack && h.UV.X()+h.UV.Y() <= 1+slack
}

// Margin is how far inside the rim the hit lies in barycentric terms,
// negative when outside.
func (h TriangleHit) Margin() float64 {
	return min(h.UV.X(), h.UV.Y(), 1-h.UV.X()-h.UV.Y())
}

// RayHit is Möller-Trumbore. Not ok when the ray is parallel to the triangle.
// Pointer receiver for speed: passing the triangle by value halves throughput.
func (t *Triangle) RayHit(r Ray) (TriangleHit, bool) {
	edge1 := t[1].Sub(t[0])
	edge2 := t[2].Sub(t[0])

	perpendicular := r.direction.Cross(edge2)
	determinant := edge1.Dot(perpendicular)
	if determinant == 0 {
		return TriangleHit{}, false
	}
	inverse := 1 / determinant

	fromCorner := r.origin.Sub(t[0])
	crossed := fromCorner.Cross(edge1)

	return TriangleHit{
		Distance: edge2.Dot(crossed) * inverse,
		UV: vector2.New(
			fromCorner.Dot(perpendicular)*inverse,
			r.direction.Dot(crossed)*inverse,
		),
	}, true
}
