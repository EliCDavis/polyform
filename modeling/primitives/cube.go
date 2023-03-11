package primitives

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
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
	verts := []vector3.Float64{
		// bottom, back, left
		vector3.New(-.5, -.5, -.5),
		// bottom, front, left
		vector3.New(-.5, -.5, .5),
		// top, back, left
		vector3.New(-.5, .5, -.5),
		// top, front, left
		vector3.New(-.5, .5, .5),

		// bottom, back, right
		vector3.New(.5, -.5, -.5),
		// bottom, front, right
		vector3.New(.5, -.5, .5),
		// top, back, right
		vector3.New(.5, .5, -.5),
		// top, front, right
		vector3.New(.5, .5, .5),
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
	).
		SetFloat3Attribute(modeling.PositionAttribute, verts).
		SetFloat3Attribute(modeling.NormalAttribute, vector3.Array[float64](verts).Normalized())
}
