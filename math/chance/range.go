package chance

import (
	"math/rand"

	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Range1D struct {
	min  float64
	size float64
	rand *rand.Rand
}

func (r Range1D) Value() float64 {
	return r.min + (r.size * r.rand.Float64())
}

func NewRange1D(min, max float64, rand *rand.Rand) Range1D {
	return Range1D{
		min:  min,
		size: max - min,
		rand: rand,
	}
}

type Range2D struct {
	min  vector2.Float64
	size vector2.Float64
	rand *rand.Rand
}

func (r Range2D) Value() vector2.Float64 {
	return r.min.Add(vector2.New(r.size.X()*r.rand.Float64(), r.size.Y()*r.rand.Float64()))
}

func NewRange2D(min, max vector2.Float64, rand *rand.Rand) Range2D {
	return Range2D{
		min:  min,
		size: max.Sub(min),
		rand: rand,
	}
}

type Range3D struct {
	min  vector3.Float64
	size vector3.Float64
	rand *rand.Rand
}

func (r Range3D) Value() vector3.Float64 {
	return r.min.Add(vector3.New(
		r.size.X()*r.rand.Float64(),
		r.size.Y()*r.rand.Float64(),
		r.size.Z()*r.rand.Float64(),
	))
}

func NewRange3D(min, max vector3.Float64, rand *rand.Rand) Range3D {
	return Range3D{
		min:  min,
		size: max.Sub(min),
		rand: rand,
	}
}
