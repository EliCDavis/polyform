package primitives

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

func Cylinder(sides int, height, radius float64) modeling.Mesh {
	halfHeight := height / 2.

	angleIncrement := (1.0 / float64(sides)) * 2.0 * math.Pi
	vertices := make([]vector.Vector3, (sides*2)+2)
	normals := make([]vector.Vector3, (sides*2)+2)
	for sideIndex := 0; sideIndex < sides; sideIndex++ {
		angle := angleIncrement * float64(sideIndex)
		vertices[sideIndex*2] = vector.NewVector3(math.Cos(angle)*radius, halfHeight, math.Sin(angle)*radius)
		vertices[(sideIndex*2)+1] = vector.NewVector3(math.Cos(angle)*radius, -halfHeight, math.Sin(angle)*radius)

		normals[sideIndex*2] = vector.NewVector3(math.Cos(angle), .1, math.Sin(angle)).Normalized()
		normals[(sideIndex*2)+1] = vector.NewVector3(math.Cos(angle), -.1, math.Sin(angle)).Normalized()
	}

	topMiddleVert := sides * 2
	bottomMiddleVert := (sides * 2) + 1
	vertices[topMiddleVert] = vector.NewVector3(0, halfHeight, 0)
	vertices[bottomMiddleVert] = vector.NewVector3(0, -halfHeight, 0)
	normals[topMiddleVert] = vector.NewVector3(0, 1, 0)
	normals[bottomMiddleVert] = vector.NewVector3(0, -1, 0)

	tris := make([]int, 0, sides*4*3)
	for sideIndex := 1; sideIndex < sides; sideIndex++ {
		topLeft := (sideIndex - 1) * 2
		topRight := (sideIndex) * 2
		bottomLeft := topLeft + 1
		bottomRight := topRight + 1
		tris = append(
			tris,

			topLeft,
			topMiddleVert,
			topRight,

			bottomLeft,
			bottomRight,
			bottomMiddleVert,

			bottomLeft,
			topLeft,
			topRight,

			bottomLeft,
			topRight,
			bottomRight,
		)
	}

	tris = append(
		tris,

		(sides*2)-2,
		topMiddleVert,
		0,

		1,
		bottomMiddleVert,
		(sides*2)-1,

		(sides*2)-2,
		0,
		1,

		(sides*2)-1,
		(sides*2)-2,
		1,
	)

	return modeling.NewMesh(
		tris,
		vertices,
		normals,
		nil,
	)
}
