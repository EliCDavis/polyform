package curves

import (
	"github.com/EliCDavis/vector/vector3"
)

type Curve interface {
	At(t float64) vector3.Float64
}

type Spline interface {
	Length() float64
	At(distance float64) vector3.Float64
	Tangent(distance float64) vector3.Float64
}
