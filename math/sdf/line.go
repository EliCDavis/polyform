package sdf

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector/vector3"
)

func Line(start, end vector3.Float64, radius float64) sample.Vec3ToFloat {
	line := geometry.NewLine3D(start, end)
	return func(v vector3.Float64) float64 {
		closestPoint := line.ClosestPointOnLine(v)
		return v.Distance(closestPoint) - radius
	}
}
