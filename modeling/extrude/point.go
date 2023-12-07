package extrude

import (
	"log"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func PathPoints(stencil []vector2.Float64, path []vector3.Float64) [][]vector3.Float64 {
	finalPaths := make([][]vector3.Float64, len(stencil))
	for i := range finalPaths {
		finalPaths[i] = make([]vector3.Vector[float64], len(path))
	}

	pointDirections := DirectionsOfPoints(path)

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

func PathPoints2(stencil []vector2.Float64, path []vector3.Float64) [][]vector3.Float64 {
	finalPaths := make([][]vector3.Float64, len(stencil))
	for i := range finalPaths {
		finalPaths[i] = make([]vector3.Vector[float64], len(path))
	}

	for pathIndex, pathPoint := range path {

		a := pathPoint
		c := pathPoint

		if pathIndex == 0 {
			c = path[1].Sub(pathPoint)
			a = c.Flip()
		} else if pathIndex == len(path)-1 {
			a = path[pathIndex-1].Sub(pathPoint)
			c = a.Flip()
		} else {
			a = path[pathIndex-1].Sub(pathPoint)
			c = path[pathIndex+1].Sub(pathPoint)
		}

		acCross := a.Cross(c).Normalized()

		log.Println(acCross)
		// r := a.Add(c)
		// ra := r.Cross(a.Normalized())

		// ral := r.Cross(a.Normalized()).Length()

		for stencilIndex, stencilPoint := range stencil {
			// point := vector3.New(stencilPoint.X(), 0, stencilPoint.Y())

			newVal := acCross.Scale(stencilPoint.X()).
				// Add(ra.Scale(stencilPoint.Y())).
				// DivByConstant(ral).
				Add(pathPoint)

			finalPaths[stencilIndex][pathIndex] = newVal
		}

	}

	return finalPaths
}
