package extrude

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// One normal per shape vertex means creases get averaged away. Callers that
// need a hard edge have to duplicate the vertex themselves.
func ProjectFace(center, normal, perpendicular vector3.Float64, shape []vector2.Float64) ([]vector3.Float64, []vector3.Float64) {
	cross := normal.Cross(perpendicular)

	outerPoints := make([]vector3.Float64, len(shape))
	for i := range shape {
		outerPoints[i] = center.
			Add(cross.Scale(shape[i].X())).
			Add(perpendicular.Scale(shape[i].Y()))
	}

	// Which side is outward flips with the outline's winding, so it is read
	// off the signed area instead of assumed.
	facing := 1.
	if geometry.Shape(shape).SignedArea() < 0 {
		facing = -1
	}

	edges := make([]vector2.Float64, len(shape))
	for i := range shape {
		along := shape[(i+1)%len(shape)].Sub(shape[i])
		if along.Length() == 0 {
			continue
		}
		edges[i] = vector2.New(along.Y(), -along.X()).Normalized().Scale(facing)
	}

	outerNormals := make([]vector3.Float64, len(shape))
	for i := range shape {
		averaged := edges[(i+len(shape)-1)%len(shape)].Add(edges[i])
		if averaged.Length() == 0 {
			continue
		}
		averaged = averaged.Normalized()
		outerNormals[i] = cross.Scale(averaged.X()).
			Add(perpendicular.Scale(averaged.Y()))
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
		rot := quaternion.FromTheta(angleIncrement*float64(i), normal)
		perp := rot.Rotate(perpendicular)
		outerPoints[i] = perp.Scale(radius).Add(center)
		outerNormals[i] = perp
	}

	return outerPoints, outerNormals
}
