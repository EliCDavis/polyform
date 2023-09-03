package sdf

import (
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// https://iquilezles.org/articles/distfunctions/
func RoundedCylinder(pos vector3.Float64, radius, topHeight, bodyHeight float64) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		p := v.Sub(pos)
		d := vector2.New(p.XZ().Length()-2.0*radius+topHeight, math.Abs(p.Y())-bodyHeight)
		return math.Min(math.Max(d.X(), d.Y()), 0.0) + vector2.New(math.Max(d.X(), 0.0), math.Max(d.Y(), 0.0)).Length() - topHeight
	}
}
