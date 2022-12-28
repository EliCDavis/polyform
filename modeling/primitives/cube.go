package primitives

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

/*
Builds a cube with no normals


The Cube Indices:

   3 ------- 7
 / |      /  |
2  |     6   |
|  1     |   5
| /      | /
0 ------ 4

*/
func Cube() modeling.Mesh {
	verts := []vector.Vector3{
		// bottom, back, left
		vector.NewVector3(-.5, -.5, -.5),
		// bottom, front, left
		vector.NewVector3(-.5, -.5, .5),
		// top, back, left
		vector.NewVector3(-.5, .5, -.5),
		// top, front, left
		vector.NewVector3(-.5, .5, .5),

		// bottom, back, right
		vector.NewVector3(.5, -.5, -.5),
		// bottom, front, right
		vector.NewVector3(.5, -.5, .5),
		// top, back, right
		vector.NewVector3(.5, .5, -.5),
		// top, front, right
		vector.NewVector3(.5, .5, .5),
	}
	return modeling.NewMesh(
		[]int{
			// Back
			0, 2, 6,
			0, 6, 4,

			// Left
			1, 3, 2,
			1, 2, 0,

			// Right
			4, 6, 7,
			4, 7, 5,

			// Top
			2, 3, 7,
			2, 7, 6,

			// Bottom
			1, 0, 4,
			1, 4, 5,

			// Front
			5, 7, 3,
			5, 3, 1,
		},
		map[string][]vector.Vector3{
			modeling.PositionAttribute: verts,
			modeling.NormalAttribute:   vector.Vector3Array(verts).Normalized(),
		},
		nil,
		nil,
		nil,
	)
}
