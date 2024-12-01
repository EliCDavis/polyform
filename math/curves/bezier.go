package curves

import (
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type CubicCurve struct {
	P0, P1, P2, P3 vector3.Float64
}

func (crc CubicCurve) At(t float64) vector3.Float64 {
	return vector3.New(
		Cubic(t, crc.P0.X(), crc.P1.X(), crc.P2.X(), crc.P3.X()),
		Cubic(t, crc.P0.Y(), crc.P1.Y(), crc.P2.Y(), crc.P3.Y()),
		Cubic(t, crc.P0.Z(), crc.P1.Z(), crc.P2.Z(), crc.P3.Z()),
	)
}

func Cubic(t, p0, p1, p2, p3 float64) float64 {
	return math.Pow(1-t, 3)*p0 +
		3*math.Pow(1-t, 2)*t*p1 +
		3*(1-t)*math.Pow(t, 2)*p2 +
		math.Pow(t, 3)*p3
}

func CubicBezierCurve2DSampler(p0, p1, p2, p3 vector2.Float64) sample.FloatToVec2 {
	return func(t float64) vector2.Float64 {
		return vector2.New(
			Cubic(t, p0.X(), p1.X(), p2.X(), p3.X()),
			Cubic(t, p0.Y(), p1.Y(), p2.Y(), p3.Y()),
		)
	}
}

func CubicBezierCurve3DSampler(p0, p1, p2, p3 vector3.Float64) sample.FloatToVec3 {
	return func(t float64) vector3.Float64 {
		return vector3.New(
			Cubic(t, p0.X(), p1.X(), p2.X(), p3.X()),
			Cubic(t, p0.Y(), p1.Y(), p2.Y(), p3.Y()),
			Cubic(t, p0.Z(), p1.Z(), p2.Z(), p3.Z()),
		)
	}
}
