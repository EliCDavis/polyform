package extrude

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type LinePoint struct {
	Point   vector3.Float64
	Up      vector3.Float64
	Width   float64
	Height  float64
	Uv      vector2.Float64
	UvWidth float64
}

func directionsOfLinePoints(points []LinePoint) []vector3.Float64 {
	pointVec := make([]vector3.Float64, len(points))
	for i, point := range points {
		pointVec[i] = point.Point
	}
	return directionOfPoints(pointVec)
}

func Line(linePoints []LinePoint) modeling.Mesh {
	if len(linePoints) < 2 {
		panic("extruding a line requires 2 or more points")
	}

	vertices := make([]vector3.Float64, 0)
	normals := make([]vector3.Float64, 0)
	directions := directionsOfLinePoints(linePoints)
	uvs := make([]vector2.Float64, 0)
	for i, p := range linePoints {

		low := p.Point.Add(p.Up.Scale(p.Height))
		outDir := directions[i].Cross(p.Up).Scale(p.Width)

		rightPoint := low.Add(outDir)
		leftPoint := low.Sub(outDir)

		rightNormal := p.Up
		leftNormal := p.Up

		if p.Width != 0 {
			rightNormal = rightPoint.Sub(p.Point).Normalized().Cross(directions[i]).Scale(-1)
			leftNormal = leftPoint.Sub(p.Point).Normalized().Cross(directions[i]).Scale(-1)
		}

		vertices = append(
			vertices,
			p.Point,
			rightPoint,
			leftPoint,
		)

		normals = append(
			normals,
			p.Up,
			rightNormal,
			leftNormal,
		)

		var uvAPoint vector2.Float64
		var uvBPoint vector2.Float64
		if i == 0 {
			uvAPoint = linePoints[0].Uv
			uvBPoint = linePoints[1].Uv
		} else {
			uvAPoint = linePoints[i-1].Uv
			uvBPoint = linePoints[i].Uv
		}
		uvDir := uvBPoint.Sub(uvAPoint)
		uvs = append(
			uvs,
			linePoints[i].Uv,
			linePoints[i].Uv.Add(uvDir.Perpendicular().Normalized().Scale(linePoints[i].UvWidth/2)),
			linePoints[i].Uv.Add(uvDir.Perpendicular().Normalized().Scale(-linePoints[i].UvWidth/2)),
		)
	}

	tris := make([]int, 0)
	for i := 1; i < len(linePoints); i++ {
		front := i * 3
		back := (i - 1) * 3

		frontMiddle := front
		frontRight := front + 1
		frontLeft := front + 2

		backMiddle := back
		backRight := back + 1
		backLeft := back + 2

		tris = append(
			tris,

			// Right Side
			frontMiddle, backMiddle, backRight,
			frontMiddle, backRight, frontRight,

			// Left Side
			frontMiddle, frontLeft, backMiddle,
			frontLeft, backLeft, backMiddle,
		)
	}

	return modeling.NewMesh(tris).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: vertices,
			modeling.NormalAttribute:   normals,
		}).
		SetFloat2Data(map[string][]vector2.Float64{
			modeling.TexCoordAttribute: uvs,
		})
}
