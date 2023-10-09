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

type LinePoint struct {
	Point  vector3.Float64
	Radius float64
}

func VarryingThicknessLine(linePoints []LinePoint) sample.Vec3ToFloat {
	if len(linePoints) < 2 {
		panic("can not create a line segment field with less than 2 points")
	}

	sdfs := make([]sample.Vec3ToFloat, 0, len(linePoints)-1)
	for i := 1; i < len(linePoints); i++ {
		start := linePoints[i-1]
		end := linePoints[i]

		sdfs = append(sdfs, RoundedCone(start.Point, end.Point, start.Radius, end.Radius))
	}

	return Union(sdfs...)
}
