package primitives

import (
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

func UVSphere(radius float64, rows, columns int) modeling.Mesh {

	if rows < 3 {
		panic(fmt.Errorf("invalid row count (%d) for uv sphere", rows))
	}

	if columns < 3 {
		panic(fmt.Errorf("invalid columns count (%d) for uv sphere", columns))
	}

	positions := make([]vector.Vector3, 0)

	// add top vertex
	v0 := vector.NewVector3(0, radius, 0)
	positions = append(positions, v0)

	// generate vertices per stack / slice
	for i := 0; i < columns-1; i++ {
		phi := math.Pi * float64(i+1) / float64(columns)
		for j := 0; j < rows; j++ {
			theta := 2.0 * math.Pi * float64(j) / float64(rows)
			x := math.Sin(phi) * math.Cos(theta)
			y := math.Cos(phi)
			z := math.Sin(phi) * math.Sin(theta)
			positions = append(positions, vector.NewVector3(x, y, z).MultByConstant(radius))
		}
	}

	// add bottom vertex
	v1 := vector.NewVector3(0, -radius, 0)
	positions = append(positions, v1)

	// add top / bottom triangles
	tris := make([]int, 0)
	for i := 0; i < rows; i++ {
		i0 := i + 1
		i1 := (i+1)%rows + 1
		tris = append(tris, 0, i1, i0)

		i0 = i + rows*(columns-2) + 1
		i1 = (i+1)%rows + rows*(columns-2) + 1
		tris = append(tris, 1, i0, i1)
	}

	// add quads per stack / slice
	for j := 0; j < columns-2; j++ {
		j0 := j*rows + 1
		j1 := (j+1)*rows + 1
		for i := 0; i < rows; i++ {
			i0 := j0 + i
			i1 := j0 + (i+1)%rows
			i2 := j1 + (i+1)%rows
			i3 := j1 + i
			// mesh.add_quad(Vertex(i0), Vertex(i1),
			// 	Vertex(i2), Vertex(i3))

			tris = append(
				tris,
				i0, i1, i2,
				i0, i2, i3,
			)
		}
	}
	return modeling.NewMesh(
		tris,
		map[string][]vector.Vector3{
			modeling.PositionAttribute: positions,
			modeling.NormalAttribute:   vector.Vector3Array(positions).Normalized(),
		},
		nil,
		nil,
		nil,
	)
}
