package primitives

import (
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func QuadSphere(radius float64, cube Cube, welded bool) modeling.Mesh {
	var m modeling.Mesh
	if welded {
		m = cube.Welded()
	} else {
		m = cube.UnweldedQuads()
	}
	return m.ModifyFloat3Attribute(modeling.PositionAttribute, func(i int, v vector3.Float64) vector3.Float64 {
		return v.Normalized().Scale(radius)
	})
}

func UVSphere(radius float64, rows, columns int) modeling.Mesh {
	if columns < 3 {
		panic(fmt.Errorf("invalid row count (%d) for uv sphere", columns))
	}

	if rows < 2 {
		panic(fmt.Errorf("invalid columns count (%d) for uv sphere", rows))
	}

	positions := make([]vector3.Float64, 0)

	// add top vertex
	v0 := vector3.New(0, radius, 0)
	positions = append(positions, v0)

	// generate vertices per stack / slice
	for i := 0; i < rows-1; i++ {
		phi := math.Pi * float64(i+1) / float64(rows)
		for j := 0; j < columns; j++ {
			theta := 2.0 * math.Pi * float64(j) / float64(columns)
			x := math.Sin(phi) * math.Cos(theta)
			y := math.Cos(phi)
			z := math.Sin(phi) * math.Sin(theta)
			positions = append(positions, vector3.New(x, y, z).Scale(radius))
		}
	}

	// add bottom vertex
	v1i := len(positions)
	v1 := vector3.New(0, -radius, 0)
	positions = append(positions, v1)

	// add top / bottom triangles
	tris := make([]int, 0)
	for i := 0; i < columns; i++ {
		i0 := i + 1
		i1 := (i+1)%columns + 1
		tris = append(tris, 0, i1, i0)

		i0 = i + columns*(rows-2) + 1
		i1 = (i+1)%columns + columns*(rows-2) + 1
		tris = append(tris, v1i, i0, i1)
	}

	// add quads per stack / slice
	for j := 0; j < rows-2; j++ {
		j0 := j*columns + 1
		j1 := (j+1)*columns + 1
		for i := 0; i < columns; i++ {
			i0 := j0 + i
			i1 := j0 + (i+1)%columns
			i2 := j1 + (i+1)%columns
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
	return modeling.NewTriangleMesh(tris).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: positions,
			modeling.NormalAttribute:   vector3.Array[float64](positions).Normalized(),
		})
}

func UVSphereUnwelded(radius float64, rows, columns int) modeling.Mesh {
	if columns < 3 {
		panic(fmt.Errorf("invalid row count (%d) for uv sphere", columns))
	}

	if rows < 2 {
		panic(fmt.Errorf("invalid columns count (%d) for uv sphere", rows))
	}

	calculatedPositions := make([]vector3.Float64, 0)

	// add top vertex
	v0 := vector3.New(0, radius, 0)
	calculatedPositions = append(calculatedPositions, v0)

	// generate vertices per stack / slice
	for i := 0; i < rows-1; i++ {
		phi := math.Pi * float64(i+1) / float64(rows)
		for j := 0; j < columns; j++ {
			theta := 2.0 * math.Pi * float64(j) / float64(columns)
			x := math.Sin(phi) * math.Cos(theta)
			y := math.Cos(phi)
			z := math.Sin(phi) * math.Sin(theta)
			calculatedPositions = append(calculatedPositions, vector3.New(x, y, z).Scale(radius))
		}
	}

	// add bottom vertex
	v1i := len(calculatedPositions)
	v1 := vector3.New(0, -radius, 0)
	calculatedPositions = append(calculatedPositions, v1)
	finalVerts := make([]vector3.Float64, 0)

	// add top / bottom triangles
	tris := make([]int, 0)
	for i := 0; i < columns; i++ {
		i0 := i + 1
		i1 := (i+1)%columns + 1
		finalVerts = append(
			finalVerts,
			calculatedPositions[0],
			calculatedPositions[i1],
			calculatedPositions[i0],
		)
		a1 := len(finalVerts) - 3
		a2 := len(finalVerts) - 2
		a3 := len(finalVerts) - 1
		tris = append(tris, a1, a2, a3)

		i0 = i + columns*(rows-2) + 1
		i1 = (i+1)%columns + columns*(rows-2) + 1
		// tris = append(tris, v1i, i0, i1)

		finalVerts = append(
			finalVerts,
			calculatedPositions[v1i],
			calculatedPositions[i0],
			calculatedPositions[i1],
		)
		a1 = len(finalVerts) - 3
		a2 = len(finalVerts) - 2
		a3 = len(finalVerts) - 1
		tris = append(tris, a1, a2, a3)
	}

	// add quads per stack / slice
	for j := 0; j < rows-2; j++ {
		j0 := j*columns + 1
		j1 := (j+1)*columns + 1
		for i := 0; i < columns; i++ {
			i0 := j0 + i
			i1 := j0 + (i+1)%columns
			i2 := j1 + (i+1)%columns
			i3 := j1 + i
			// mesh.add_quad(Vertex(i0), Vertex(i1),
			// 	Vertex(i2), Vertex(i3))

			a := calculatedPositions[i0]
			b := calculatedPositions[i1]
			c := calculatedPositions[i2]
			d := calculatedPositions[i3]
			finalVerts = append(finalVerts, a, b, c, d)

			i0 = len(finalVerts) - 4
			i1 = len(finalVerts) - 3
			i2 = len(finalVerts) - 2
			i3 = len(finalVerts) - 1

			tris = append(
				tris,
				i0, i1, i2,
				i0, i2, i3,
			)
		}
	}
	return modeling.NewTriangleMesh(tris).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: finalVerts,
		})
}

type UvSphereNode struct {
	Radius  nodes.Output[float64]
	Rows    nodes.Output[int]
	Columns nodes.Output[int]
	Weld    nodes.Output[bool]
}

func (c UvSphereNode) Description() string {
	return "A spherical mesh that is created by starting with a square grid, turning it into a cylinder, and then squeezing the top and bottom. It is the simplest way to create a sphere, but it has a poor vertex distribution."
}

func (c UvSphereNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	radius := nodes.TryGetOutputValue(out, c.Radius, .5)
	rows := max(nodes.TryGetOutputValue(out, c.Rows, 10), 2)
	columns := max(nodes.TryGetOutputValue(out, c.Columns, 10), 3)

	if nodes.TryGetOutputValue(out, c.Weld, false) {
		out.Set(UVSphere(radius, rows, columns))
	} else {
		out.Set(UVSphereUnwelded(radius, rows, columns))
	}
}

type QuadSphereNode struct {
	Radius     nodes.Output[float64]
	Weld       nodes.Output[bool]
	Resolution nodes.Output[int]
	UVs        nodes.Output[CubeUVs]
}

func (c QuadSphereNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	strip := &StripUVs{
		Start: vector2.New(0, 0.5),
		End:   vector2.New(1, 0.5),
		Width: 1.,
	}

	out.Set(QuadSphere(
		nodes.TryGetOutputValue(out, c.Radius, .5),
		Cube{
			Height:     1,
			Width:      1,
			Depth:      1,
			Dimensions: max(1, nodes.TryGetOutputValue(out, c.Resolution, 10)),
			UVs: nodes.TryGetOutputReference(out, c.UVs, &CubeUVs{
				Top:    strip,
				Bottom: strip,
				Left:   strip,
				Right:  strip,
				Front:  strip,
				Back:   strip,
			}),
		},
		nodes.TryGetOutputValue(out, c.Weld, false),
	))
}
