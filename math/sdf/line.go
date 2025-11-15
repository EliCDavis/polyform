package sdf

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func lineSDF(l geometry.Line3D, radius float64, v vector3.Float64) float64 {
	closestPoint := l.ClosestPointOnLine(v)
	return v.Distance(closestPoint) - radius
}

func Line(start, end vector3.Float64, radius float64) sample.Vec3ToFloat {
	line := geometry.NewLine3D(start, end)
	return func(v vector3.Float64) float64 {
		return lineSDF(line, radius, v)
	}
}

func MultipointLine(points []vector3.Float64, radius float64) sample.Vec3ToFloat {

	switch len(points) {
	case 0:
		return func(f vector3.Float64) float64 {
			return math.Inf(1)
		}

	case 1:
		return Sphere(points[0], radius)

	case 2:
		return Line(points[0], points[1], radius)

	}

	lines := make([]geometry.Line3D, len(points)-1)
	for i := 0; i < len(points)-1; i++ {
		lines[i] = geometry.NewLine3D(points[i], points[i+1])
	}

	return func(v vector3.Float64) float64 {
		closestPoint := lineSDF(lines[0], radius, v)
		for i := 1; i < len(lines); i++ {
			closestPoint = min(closestPoint, lineSDF(lines[i], radius, v))
		}
		return closestPoint
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

type LineNode struct {
	Start  nodes.Output[vector3.Float64]
	End    nodes.Output[vector3.Float64]
	Radius nodes.Output[float64]
}

func (cn LineNode) Field(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	out.Set(Line(
		nodes.TryGetOutputValue(out, cn.Start, vector3.Zero[float64]()),
		nodes.TryGetOutputValue(out, cn.End, vector3.One[float64]()),
		nodes.TryGetOutputValue(out, cn.Radius, .25),
	))
}

type LinesNode struct {
	Points nodes.Output[[]vector3.Float64]
	Radius nodes.Output[float64]
}

func (cn LinesNode) Field(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	out.Set(MultipointLine(
		nodes.TryGetOutputValue(out, cn.Points, nil),
		nodes.TryGetOutputValue(out, cn.Radius, .25),
	))
}
