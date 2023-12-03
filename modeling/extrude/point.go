package extrude

import (
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func PathPoints(stencil []vector2.Float64, path []vector3.Float64) [][]vector3.Float64 {
	finalPaths := make([][]vector3.Float64, len(stencil))
	for i := range finalPaths {
		finalPaths[i] = make([]vector3.Vector[float64], len(path))
	}

	pointDirections := directionsOfPoints(path)

	lastDir := vector3.Up[float64]()
	lastRot := quaternion.New(vector3.Zero[float64](), 1)

	for pathIndex, pathPoint := range path {
		dir := pointDirections[pathIndex]

		rot := quaternion.RotationTo(lastDir, dir)

		for stencilIndex, stencilPoint := range stencil {
			point := vector3.New(stencilPoint.X(), 0, stencilPoint.Y())
			point = lastRot.Rotate(point)
			point = rot.Rotate(point)

			finalPaths[stencilIndex][pathIndex] = point.Add(pathPoint)
		}

		lastRot = rot.Multiply(lastRot)
		lastDir = dir
	}

	return finalPaths
}
