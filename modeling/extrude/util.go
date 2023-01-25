package extrude

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// TODO
//
//			Pretty sure normal calculation is wrong. Need to determine what is and
//	     isn't a convex / concave point
func ProjectFace(center, normal, perpendicular vector3.Float64, shape []vector2.Float64) ([]vector3.Float64, []vector3.Float64) {
	cross := normal.Cross(perpendicular)
	transformation := modeling.Matrix{
		{cross.X(), perpendicular.X(), normal.X()},
		{cross.Y(), perpendicular.Y(), normal.Y()},
		{cross.Z(), perpendicular.Z(), normal.Z()},
	}

	outerPoints := make([]vector3.Float64, len(shape))
	outerNormals := make([]vector3.Float64, len(shape))

	for i := 0; i < len(shape); i++ {
		v := modeling.Multiply3x3by3x1(transformation, vector3.New(shape[i].X(), shape[i].Y(), 0))
		outerPoints[i] = v.Add(center)
	}

	for i := 0; i < len(shape); i++ {
		previous := i - 1
		if i == 0 {
			previous = len(shape) - 1
		}
		outerNormals[i] = outerPoints[i].Sub(outerPoints[previous]).Normalized()
	}

	return outerPoints, outerNormals
}

func GetPlaneOuterPoints(center, normal, perpendicular vector3.Float64, radius float64, sides int) ([]vector3.Float64, []vector3.Float64) {
	outerPoints := make([]vector3.Float64, sides)
	outerNormals := make([]vector3.Float64, sides)

	outerPoints[0] = perpendicular.Scale(radius).Add(center)
	outerNormals[0] = perpendicular

	angleIncrement := (math.Pi * 2) / float64(sides)

	for i := 1; i < sides; i++ {
		rot := modeling.UnitQuaternionFromTheta(angleIncrement*float64(i), normal)
		perp := rot.Rotate(perpendicular)
		outerPoints[i] = perp.Scale(radius).Add(center)
		outerNormals[i] = perp
	}

	return outerPoints, outerNormals
}
