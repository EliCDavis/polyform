package extrude

import (
	"log"
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
func circle(sides int, thickness []float64, path []vector.Vector3, closed bool) mesh.Mesh {
	if len(thickness) != len(path) {
		panic("thickness count must match path count")
	}

	if len(path) < 2 {
		panic("Can not extrude a path with less than 2 points")
	}

	vertices := make([]vector.Vector3, 0, len(path)*sides)
	normals := make([]vector.Vector3, 0, len(path)*sides)

	circlePoints := make([]vector.Vector3, sides)
	circlePoints[0] = vector.Vector3Right()

	angleIncrement := (math.Pi * 2) / float64(sides)

	for i := 1; i < sides; i++ {
		rot := mesh.UnitQuaternionFromTheta(angleIncrement*float64(i), vector.Vector3Up())
		circlePoints[i] = rot.Rotate(vector.Vector3Right())
	}

	for i, p := range path {
		var dir vector.Vector3

		if i == 0 {
			dir = path[1].Sub(path[0])
		} else if i == len(path)-1 {
			dir = path[i].Sub(path[i-1])
		} else {
			dir = path[i+1].Sub(path[i]).Add(path[i].Sub(path[i-1]))
		}

		dir = dir.Normalized()
		log.Println(dir)

		for sideIndex := 0; sideIndex < sides; sideIndex++ {

			point := circlePoints[sideIndex]

			if dir.Sub(vector.Vector3Up()).SquaredLength() > 0.00000001 {
				angleVector := dir.Cross(vector.Vector3Up())
				angleDot := angle(dir, vector.Vector3Up())
				rot := mesh.UnitQuaternionFromTheta(angleDot, angleVector)
				point = rot.Rotate(point)
			}

			// rot := mesh.UnitQuaternionFromTheta(angleIncrement*float64(i), dir)
			vertices = append(vertices, point.MultByConstant(thickness[i]).Add(p))
			normals = append(normals, point)
		}
	}

	tris := make([]int, 0)

	for pathIndex := range path {
		bottom := pathIndex * sides
		top := (pathIndex + 1) * sides
		if pathIndex == len(path)-1 {
			if closed {
				top = 0
			} else {
				continue
			}
		}
		for sideIndex := 0; sideIndex < sides; sideIndex++ {
			topRight := top + sideIndex
			bottomRight := bottom + sideIndex

			topLeft := topRight - 1
			bottomLeft := bottomRight - 1
			if sideIndex == 0 {
				topLeft = top + sides - 1
				bottomLeft = bottom + sides - 1
			}

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

	return mesh.NewMesh(tris, vertices, normals, nil)
}

func ClosedCircle(sides int, thickness float64, path []vector.Vector3) mesh.Mesh {
	continousThickness := make([]float64, len(path))
	for i := range path {
		continousThickness[i] = thickness
	}
	return circle(sides, continousThickness, path, true)
}

func Circle(sides int, thickness float64, path []vector.Vector3) mesh.Mesh {
	continousThickness := make([]float64, len(path))
	for i := range path {
		continousThickness[i] = thickness
	}
	return circle(sides, continousThickness, path, false)
}

func CircleWithThickness(sides int, thickness []float64, path []vector.Vector3) mesh.Mesh {
	return circle(sides, thickness, path, false)
}

func ClosedCircleWithThickness(sides int, thickness []float64, path []vector.Vector3) mesh.Mesh {
	return circle(sides, thickness, path, true)
}
