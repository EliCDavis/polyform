package primitives

import (
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

const minHemisphereColumns = 3
const minHemisphereRows = 2

type Hemisphere struct {
	Radius float64
	Height float64
	Capped bool
}

func (h Hemisphere) UV(rows, columns int) modeling.Mesh {
	if columns < minHemisphereColumns {
		panic(fmt.Errorf("invalid row count (%d < %d) for uv sphere", columns, minHemisphereColumns))
	}

	if rows < minHemisphereRows {
		panic(fmt.Errorf("invalid columns count (%d < %d) for uv sphere", rows, minHemisphereRows))
	}

	positions := make([]vector3.Float64, 0)

	// add bottom vertex
	positions = append(positions, vector3.Zero[float64]())

	// generate vertices per stack / slice
	for i := 0; i < rows-1; i++ {
		phi := (-math.Pi * float64(i)) / float64(rows)
		for j := range columns {
			theta := 2.0 * math.Pi * (float64(j) / float64(columns))

			ugh := (phi / 2) + (math.Pi / 2)

			x := math.Sin(ugh) * math.Cos(theta)
			y := math.Cos(ugh)
			z := math.Sin(ugh) * math.Sin(theta)
			positions = append(positions, vector3.New(x*h.Radius, y*h.Height, z*h.Radius))
		}
	}

	// add top vertex
	v1i := len(positions)
	v1 := vector3.New(0, h.Height, 0)
	positions = append(positions, v1)

	// add top / bottom triangles
	tris := make([]int, 0)
	for i := range columns {

		i0 := i + columns*(rows-2) + 1
		i1 := (i+1)%columns + columns*(rows-2) + 1
		tris = append(tris, v1i, i1, i0)

		if h.Capped {
			i0 = i + 1
			i1 = (i+1)%columns + 1
			tris = append(tris, 0, i0, i1)
		}
	}

	// add quads per stack / slice
	for j := 0; j < rows-2; j++ {
		j0 := j*columns + 1
		j1 := (j+1)*columns + 1
		for i := range columns {
			i0 := j0 + i
			i1 := j0 + (i+1)%columns
			i2 := j1 + (i+1)%columns
			i3 := j1 + i
			// mesh.add_quad(Vertex(i0), Vertex(i1),
			// 	Vertex(i2), Vertex(i3))

			tris = append(
				tris,
				i0, i2, i1,
				i0, i3, i2,
			)
		}
	}

	return modeling.NewTriangleMesh(tris).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: positions,
		})
}

type HemisphereNode struct {
	Rows    nodes.Output[int]
	Columns nodes.Output[int]
	Radius  nodes.Output[float64]
	Height  nodes.Output[float64]
	Capped  nodes.Output[bool]
}

func (hnd HemisphereNode) Out(out *nodes.StructOutput[modeling.Mesh]) {

	radius := nodes.TryGetOutputValue(out, hnd.Radius, 0.5)

	hemi := Hemisphere{
		Radius: radius,
		Height: nodes.TryGetOutputValue(out, hnd.Height, radius),
		Capped: nodes.TryGetOutputValue(out, hnd.Capped, true),
	}

	out.Set(hemi.UV(
		max(minHemisphereRows, nodes.TryGetOutputValue(out, hnd.Rows, 20)),
		max(minHemisphereColumns, nodes.TryGetOutputValue(out, hnd.Columns, 20)),
	))
}
