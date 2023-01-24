package curves

import (
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func CubicCurve(t, p0, p1, p2, p3 float64) float64 {
	return math.Pow(1-t, 3)*p0 +
		3*math.Pow(1-t, 2)*t*p1 +
		3*(1-t)*math.Pow(t, 2)*p2 +
		math.Pow(t, 3)*p3
}

func CubicBezierCurve2D(p0, p1, p2, p3 vector2.Float64) sample.FloatToVec2 {
	return func(t float64) vector2.Float64 {
		return vector2.New(
			CubicCurve(t, p0.X(), p1.X(), p2.X(), p3.X()),
			CubicCurve(t, p0.Y(), p1.Y(), p2.Y(), p3.Y()),
		)
	}
}

func CubicBezierCurve3D(p0, p1, p2, p3 vector3.Float64) sample.FloatToVec3 {
	return func(t float64) vector3.Float64 {
		return vector3.New(
			CubicCurve(t, p0.X(), p1.X(), p2.X(), p3.X()),
			CubicCurve(t, p0.Y(), p1.Y(), p2.Y(), p3.Y()),
			CubicCurve(t, p0.Z(), p1.Z(), p2.Z(), p3.Z()),
		)
	}
}
