package extrude

import (
	"math"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/vector"
)

func clamp(f, min, max float64) float64 {
	return math.Max(math.Min(f, max), min)
}

// angle in radians
func angle(from, to vector.Vector3) float64 {
	// sqrt(a) * sqrt(b) = sqrt(a * b) -- valid for real numbers
	denominator := math.Sqrt(from.SquaredLength() * to.SquaredLength())
	if denominator < 1e-15 {
		return 0.
	}
	dot := clamp(from.Dot(to)/denominator, -1., 1.)
	return (math.Acos(dot))
}

// TODO: Pretty sure this breaks for paths that have multiple points in the
// same direction.
func circle(sides int, points []ExtrusionPoint, closed bool) mesh.Mesh {
	vertCount := sides + 1
	vertices := make([]vector.Vector3, 0, len(points)*vertCount)
	uvs := make([]vector.Vector2, 0, len(points)*vertCount)
	normals := make([]vector.Vector3, 0, len(points)*vertCount)

	circlePoints := make([]vector.Vector3, vertCount)
	circlePoints[0] = vector.Vector3Right()

	angleIncrement := (math.Pi * 2) / float64(sides)

	for i := 1; i < sides+1; i++ {
		rot := mesh.UnitQuaternionFromTheta(angleIncrement*float64(i), vector.Vector3Up())
		circlePoints[i] = rot.Rotate(vector.Vector3Right())
	}

	// Vertices and normals ===================================================
	for i, p := range points {

		var dirA vector.Vector3
		var dirB vector.Vector3

		if i == 0 {
			dirA = points[0].Point.MultByConstant(points[0].Thickness)
			dirB = points[1].Point.MultByConstant(points[1].Thickness)
		} else if i == len(points)-1 {
			dirA = points[i-1].Point.MultByConstant(points[i-1].Thickness)
			dirB = points[i].Point.MultByConstant(points[i].Thickness)
		} else {
			dirA = points[i].Point.Sub(points[i-1].Point).MultByConstant(points[i].Thickness)
			dirB = points[i+1].Point.Sub(points[i].Point).MultByConstant(points[i+1].Thickness)
		}

		dir := dirB.Sub(dirA).Normalized()
		// log.Print(i, dirA, dirB, dir)

		for sideIndex := 0; sideIndex < vertCount; sideIndex++ {

			point := circlePoints[sideIndex]

			if dir.Cross(vector.Vector3Up()) != vector.Vector3Zero() {
				angleVector := dir.Cross(vector.Vector3Up())
				angleDot := angle(dir, vector.Vector3Up())
				// log.Print(angleVector, angleDot)
				rot := mesh.UnitQuaternionFromTheta(angleDot, angleVector)
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

			topLeft := top + sideIndex
			bottomLeft := bottom + sideIndex

			topRight := topLeft + 1
			bottomRight := bottomLeft + 1

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

	return mesh.NewMesh(tris, vertices, normals, [][]vector.Vector2{uvs})
}

func ClosedCircleWithConstantThickness(sides int, thickness float64, path []vector.Vector3) mesh.Mesh {
	points := make([]ExtrusionPoint, len(path))
	for i, p := range path {
		points[i] = ExtrusionPoint{
			Point:     p,
			Thickness: thickness,
		}
	}
	return circle(sides, points, true)
}

func CircleWithConstantThickness(sides int, thickness float64, path []vector.Vector3) mesh.Mesh {
	points := make([]ExtrusionPoint, len(path))
	for i, p := range path {
		points[i] = ExtrusionPoint{
			Point:     p,
			Thickness: thickness,
		}
	}
	return circle(sides, points, false)
}

func CircleWithThickness(sides int, thickness []float64, path []vector.Vector3) mesh.Mesh {
	points := make([]ExtrusionPoint, len(path))
	for i, p := range path {
		points[i] = ExtrusionPoint{
			Point:     p,
			Thickness: thickness[i],
		}
	}
	return circle(sides, points, false)
}

func ClosedCircleWithThickness(sides int, thickness []float64, path []vector.Vector3) mesh.Mesh {
	points := make([]ExtrusionPoint, len(path))
	for i, p := range path {
		points[i] = ExtrusionPoint{
			Point:     p,
			Thickness: thickness[i],
		}
	}
	return circle(sides, points, true)
}

func Circle(sides int, points []ExtrusionPoint) mesh.Mesh {
	return circle(sides, points, false)
}
