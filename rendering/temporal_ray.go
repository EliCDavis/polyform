package rendering

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
)

type TemporalRay struct {
	origin    vector3.Float64
	direction vector3.Float64
	time      float64
}

func NewTemporalRay(origin vector3.Float64, direction vector3.Float64, time float64) TemporalRay {
	return TemporalRay{
		origin:    origin,
		direction: direction.Normalized(),
	}
}

func (r TemporalRay) Origin() vector3.Float64 {
	return r.origin
}

func (r TemporalRay) Direction() vector3.Float64 {
	return r.direction
}

func (r TemporalRay) At(t float64) vector3.Float64 {
	return r.origin.Add(r.direction.Scale(t))
}

func (r TemporalRay) Ray() geometry.Ray {
	return geometry.NewRay(r.origin, r.direction)
}
