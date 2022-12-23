package extrude

import (
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

func directionOfPoints(points []vector.Vector3) []vector.Vector3 {
	directions := make([]vector.Vector3, len(points))

	for i, point := range points {
		if i == 0 {
			directions[i] = points[1].Sub(point).Normalized()
			continue
		}

		if i == len(points)-1 {
			directions[i] = point.Sub(points[i-1]).Normalized()
			continue
		}

		dirA := point.Sub(points[i-1]).Normalized()
		dirB := points[i+1].Sub(point).Normalized()
		directions[i] = dirA.Add(dirB).Normalized()
	}

	return directions
}

func directionsOfExtrusionPoints(points []ExtrusionPoint) []vector.Vector3 {
	pointVec := make([]vector.Vector3, len(points))
	for i, point := range points {
		pointVec[i] = point.Point
	}
	return directionOfPoints(pointVec)
}

// TODO: Pretty sure this breaks for paths that have multiple points in the
// same direction.
func polygon(sides int, points []ExtrusionPoint, closed bool) modeling.Mesh {
	if len(points) < 2 {
		panic(fmt.Errorf("can not extrude polygon with %d points", len(points)))
	}

	vertCount := sides + 1
	vertices := make([]vector.Vector3, 0, len(points)*vertCount)
	uvs := make([]vector.Vector2, 0, len(points)*vertCount)
	normals := make([]vector.Vector3, 0, len(points)*vertCount)

	circlePoints := make([]vector.Vector3, vertCount)
	circlePoints[0] = vector.Vector3Right()

	angleIncrement := (math.Pi * 2) / float64(sides)

	for i := 1; i < sides+1; i++ {
		rot := modeling.UnitQuaternionFromTheta(angleIncrement*float64(i), vector.Vector3Up())
		circlePoints[i] = rot.Rotate(vector.Vector3Right())
	}

	pointDirections := directionsOfExtrusionPoints(points)

	// Vertices and normals ===================================================
	for i, p := range points {

		dir := pointDirections[i]

		for sideIndex := 0; sideIndex < vertCount; sideIndex++ {

			point := circlePoints[sideIndex]

			angleVector := dir.Cross(vector.Vector3Up())
			if angleVector != vector.Vector3Zero() {
				angleDot := dir.Angle(vector.Vector3Up())
				// log.Print(angleVector, angleDot)
				rot := modeling.UnitQuaternionFromTheta(angleDot, angleVector)
				point = rot.Rotate(point)
			}

			// rot := mesh.UnitQuaternionFromTheta(angleIncrement*float64(i), dir)
			vertices = append(vertices, point.MultByConstant(p.Thickness).Add(p.Point))
			normals = append(normals, point)
		}
	}

	// UVs ====================================================================
	for i, p := range points {

		var dirA vector.Vector2
		var dirB vector.Vector2

		if i == 0 {
			dirA = points[0].UvPoint
			dirB = points[1].UvPoint
		} else {
			dirA = points[i-1].UvPoint
			dirB = p.UvPoint
		}

		dir := dirB.Sub(dirA).Normalized()
		perp := vector.NewVector2(dir.Y(), -dir.X()).
			MultByConstant(p.UvThickness / 2.)

		// log.Print(perp)
		for sideIndex := 0; sideIndex < vertCount; sideIndex++ {
			percentUsed := ((float64(sideIndex) / float64(sides)) * 2) - 1.
			uvPoint := p.UvPoint.Add(perp.MultByConstant(percentUsed))
			// log.Print(percentUsed, uvPoint)
			uvs = append(uvs, uvPoint)
		}
	}

	// Triangles ==============================================================
	tris := make([]int, 0)

	for pathIndex := range points {
		bottom := pathIndex * vertCount
		top := (pathIndex + 1) * vertCount
		if pathIndex == len(points)-1 {
			if closed {
				top = 0
			} else {
				continue
			}
		}
		for sideIndex := 0; sideIndex < sides; sideIndex++ {
			topRight := top + sideIndex
			bottomRight := bottom + sideIndex

			topLeft := topRight + 1
			bottomLeft := bottomRight + 1

			tris = append(
				tris,

				bottomLeft,
				topLeft,
				topRight,

				bottomLeft,
				topRight,
				bottomRight,
			)
		}
	}

	return modeling.NewMesh(tris, vertices, normals, [][]vector.Vector2{uvs})
}

func ClosedCircleWithConstantThickness(sides int, thickness float64, path []vector.Vector3) modeling.Mesh {
	points := make([]ExtrusionPoint, len(path))
	for i, p := range path {
		points[i] = ExtrusionPoint{
			Point:     p,
			Thickness: thickness,
		}
	}
	return polygon(sides, points, true)
}

func CircleWithConstantThickness(sides int, thickness float64, path []vector.Vector3) modeling.Mesh {
	points := make([]ExtrusionPoint, len(path))
	for i, p := range path {
		points[i] = ExtrusionPoint{
			Point:     p,
			Thickness: thickness,
		}
	}
	return polygon(sides, points, false)
}

func CircleWithThickness(sides int, thickness []float64, path []vector.Vector3) modeling.Mesh {
	points := make([]ExtrusionPoint, len(path))
	for i, p := range path {
		points[i] = ExtrusionPoint{
			Point:     p,
			Thickness: thickness[i],
		}
	}
	return polygon(sides, points, false)
}

func ClosedCircleWithThickness(sides int, thickness []float64, path []vector.Vector3) modeling.Mesh {
	points := make([]ExtrusionPoint, len(path))
	for i, p := range path {
		points[i] = ExtrusionPoint{
			Point:     p,
			Thickness: thickness[i],
		}
	}
	return polygon(sides, points, true)
}

func Polygon(sides int, points []ExtrusionPoint) modeling.Mesh {
	return polygon(sides, points, false)
}
