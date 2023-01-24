package extrude

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// TODO: Pretty sure this breaks for paths that have multiple points in the
// same direction.
func makeShape(shape []vector2.Float64, path []vector3.Float64, close bool) modeling.Mesh {
	if len(path) < 2 {
		panic("Can not extrude a path with less than 2 points")
	}

	vertices := make([]vector3.Float64, 0, len(path)*len(shape))
	normals := make([]vector3.Float64, 0, len(path)*len(shape))
	for i, p := range path {
		var dir vector3.Float64
		var per vector3.Float64

		if i == 0 {
			dir = path[1].Sub(path[0])
		} else if i == len(path)-1 {
			dir = path[i].Sub(path[i-1])
		} else {
			dir = path[i+1].Sub(path[i]).Add(path[i].Sub(path[i-1]))
		}

		if len(path) == 2 {
			per = dir.Perpendicular()
		} else if i == 0 {
			per = path[i+1].Sub(path[i]).Cross(path[i].Sub(path[len(path)-1]))
		} else if i == len(path)-1 {
			per = path[0].Sub(path[i]).Cross(path[i].Sub(path[i-1]))
		} else {
			per = path[i+1].Sub(path[i]).Cross(path[i].Sub(path[i-1]))
		}

		verts, norms := ProjectFace(p, dir.Normalized(), per.Normalized(), shape)
		vertices = append(vertices, verts...)
		normals = append(normals, norms...)
	}

	sides := len(shape)

	tris := make([]int, 0)

	for pathIndex := range path {
		bottom := pathIndex * sides
		top := (pathIndex + 1) * sides
		if pathIndex == len(path)-1 {
			if close {
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

	return modeling.NewMesh(
		tris,
		map[string][]vector3.Float64{
			modeling.PositionAttribute: vertices,
			modeling.NormalAttribute:   normals,
		},
		nil,
		nil,
		nil,
	)
}

func Shape(shape []vector2.Float64, path []vector3.Float64) modeling.Mesh {
	return makeShape(shape, path, false)
}

func ClosedShape(shape []vector2.Float64, path []vector3.Float64) modeling.Mesh {
	return makeShape(shape, path, true)
}
