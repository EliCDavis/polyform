package geometry

import "github.com/EliCDavis/vector/vector3"

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
