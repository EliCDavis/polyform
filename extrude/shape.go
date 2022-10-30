package extrude

import (
	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/vector"
)

// TODO: Pretty sure this breaks for paths that have multiple points in the
// same direction.
func Shape(shape []vector.Vector2, path []vector.Vector3) mesh.Mesh {
	if len(path) < 2 {
		panic("Can not extrude a path with less than 2 points")
	}

	vertices := make([]vector.Vector3, 0, len(path)*len(shape))
	normals := make([]vector.Vector3, 0, len(path)*len(shape))
	for i, p := range path {
		var dir vector.Vector3
		var per vector.Vector3

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
			continue
		}

		// if pathIndex > 0 {
		// 	continue
		// }

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
