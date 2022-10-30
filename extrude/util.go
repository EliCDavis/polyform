package extrude

import (
	"math"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/vector"
)

func ProjectFace(center, normal, perpendicular vector.Vector3, shape []vector.Vector2) ([]vector.Vector3, []vector.Vector3) {
	cross := normal.Cross(perpendicular)
	transformation := mesh.Matrix{
		{cross.X(), perpendicular.X(), normal.X()},
		{cross.Y(), perpendicular.Y(), normal.Y()},
		{cross.Z(), perpendicular.Z(), normal.Z()},
	}

	// transformation = mesh.Matrix{
	// 	{cross.X(), cross.Y(), cross.Z()},
	// 	{perpendicular.X(), perpendicular.Y(), perpendicular.Z()},
	// 	{normal.X(), normal.Y(), normal.Z()},
	// }

	outerPoints := make([]vector.Vector3, len(shape))
	outerNormals := make([]vector.Vector3, len(shape))

	for i := 0; i < len(shape); i++ {
		v := mesh.Multiply3x3by3x1(transformation, vector.NewVector3(shape[i].X(), shape[i].Y(), 0))

		outerPoints[i] = v.Add(center)
		outerNormals[i] = v.Normalized()
	}

	return outerPoints, outerNormals
}

func GetPlaneOuterPoints(center, normal, perpendicular vector.Vector3, radius float64, sides int) ([]vector.Vector3, []vector.Vector3) {
	outerPoints := make([]vector.Vector3, sides)
	outerNormals := make([]vector.Vector3, sides)

	outerPoints[0] = perpendicular.MultByConstant(radius).Add(center)
	outerNormals[0] = perpendicular

	angleIncrement := (math.Pi * 2) / float64(sides)

	for i := 1; i < sides; i++ {
		rot := mesh.UnitQuaternionFromTheta(angleIncrement*float64(i), normal)
		perp := rot.Rotate(perpendicular)
		outerPoints[i] = perp.MultByConstant(radius).Add(center)
		outerNormals[i] = perp
	}

	return outerPoints, outerNormals
}
