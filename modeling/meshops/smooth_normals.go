package meshops

import (
	"fmt"
	"math"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type SmoothNormalsTransformer struct{}

func (smt SmoothNormalsTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	if err = RequireTopology(m, modeling.TriangleTopology); err != nil {
		return
	}

	if err = RequireV3Attribute(m, modeling.PositionAttribute); err != nil {
		return
	}

	return SmoothNormals(m), nil
}

func SmoothNormals(m modeling.Mesh) modeling.Mesh {
	check(RequireTopology(m, modeling.TriangleTopology))
	check(RequireV3Attribute(m, modeling.PositionAttribute))

	vertices := m.Float3Attribute(modeling.PositionAttribute)
	normals := make([]vector3.Float64, vertices.Len())

	tris := m.Indices()
	for triIndex := 0; triIndex < tris.Len(); triIndex += 3 {
		p1 := tris.At(triIndex)
		p2 := tris.At(triIndex + 1)
		p3 := tris.At(triIndex + 2)
		v1, v2, v3 := vertices.At(p1), vertices.At(p2), vertices.At(p3)

		normal := m.Tri(triIndex / 3).Normal(modeling.PositionAttribute)
		if normal == vector3.Zero[float64]() {
			continue
		}

		a1 := cornerAngle(v2.Sub(v1), v3.Sub(v1))
		a2 := cornerAngle(v3.Sub(v2), v1.Sub(v2))
		a3 := math.Pi - a1 - a2

		normals[p1] = normals[p1].Add(normal.Scale(a1))
		normals[p2] = normals[p2].Add(normal.Scale(a2))
		normals[p3] = normals[p3].Add(normal.Scale(a3))
	}

	zero := vector3.Zero[float64]()
	for i, n := range normals {
		if n == zero {
			continue
		}
		normals[i] = n.Normalized()
	}

	return m.SetFloat3Attribute(modeling.NormalAttribute, normals)
}

func cornerAngle(a, b vector3.Float64) float64 {
	return math.Atan2(a.Cross(b).Length(), a.Dot(b))
}

type cornerFans []int

func (f cornerFans) find(c int) int {
	for f[c] != c {
		f[c] = f[f[c]]
		c = f[c]
	}
	return c
}

func (f cornerFans) union(a, b int) {
	f[f.find(a)] = f.find(b)
}

func SmoothNormalsByAngle(m modeling.Mesh, smoothingAngle float64) modeling.Mesh {
	check(RequireTopology(m, modeling.TriangleTopology))
	check(RequireV3Attribute(m, modeling.PositionAttribute))

	positions := m.Float3Attribute(modeling.PositionAttribute)
	weld := geometry.NewPointWelder3D(1e-9 * m.BoundingBox(modeling.PositionAttribute).Size().Length())
	at := make([]int, positions.Len())
	for v := range at {
		at[v] = weld.Index(positions.At(v))
	}

	tris := m.Indices()
	corners := tris.Len() - tris.Len()%3
	faceNormals := make([]vector3.Float64, corners/3)
	valid := make([]bool, corners/3)
	angles := make([]float64, corners)
	for f := range faceNormals {
		v1, v2, v3 := positions.At(tris.At(3*f)), positions.At(tris.At(3*f+1)), positions.At(tris.At(3*f+2))
		normal := m.Tri(f).Normal(modeling.PositionAttribute)
		if normal == vector3.Zero[float64]() {
			continue
		}
		valid[f] = true
		faceNormals[f] = normal
		angles[3*f] = cornerAngle(v2.Sub(v1), v3.Sub(v1))
		angles[3*f+1] = cornerAngle(v3.Sub(v2), v1.Sub(v2))
		angles[3*f+2] = math.Pi - angles[3*f] - angles[3*f+1]
	}

	next := func(c int) int { return c - c%3 + (c+1)%3 }
	cornerAt := func(f, position int) int {
		for c := 3 * f; c < 3*f+3; c++ {
			if at[tris.At(c)] == position {
				return c
			}
		}
		return -1
	}

	edges := make(map[[2]int][]int)
	for c := range corners {
		if !valid[c/3] {
			continue
		}
		a, b := at[tris.At(c)], at[tris.At(next(c))]
		if a != b {
			key := [2]int{min(a, b), max(a, b)}
			edges[key] = append(edges[key], c)
		}
	}

	fans := make(cornerFans, corners)
	for c := range fans {
		fans[c] = c
	}
	smooth := math.Cos(smoothingAngle * math.Pi / 180)
	for _, sides := range edges {
		if len(sides) != 2 {
			continue
		}
		c1, c2 := sides[0], sides[1]
		if faceNormals[c1/3].Dot(faceNormals[c2/3]) <= smooth {
			continue
		}
		fans.union(c1, cornerAt(c2/3, at[tris.At(c1)]))
		fans.union(next(c1), cornerAt(c2/3, at[tris.At(next(c1))]))
	}

	fanSums := make([]vector3.Float64, corners)
	positionSums := make(map[int]vector3.Float64)
	for c := range corners {
		if !valid[c/3] {
			continue
		}
		weighted := faceNormals[c/3].Scale(angles[c])
		fanSums[fans.find(c)] = fanSums[fans.find(c)].Add(weighted)
		positionSums[at[tris.At(c)]] = positionSums[at[tris.At(c)]].Add(weighted)
	}

	type corner struct{ vertex, fan int }
	vertexCount := positions.Len()
	assigned := make(map[corner]int)
	claimed := make([]bool, vertexCount)
	order := make([]int, 0, vertexCount)
	normals := make([]vector3.Float64, 0, vertexCount)
	for v := range vertexCount {
		order = append(order, v)
		normals = append(normals, vector3.Zero[float64]())
	}
	indices := make([]int, corners)
	for c := range corners {
		key := corner{vertex: tris.At(c), fan: -1}
		sum := positionSums[at[key.vertex]]
		if valid[c/3] {
			key.fan = fans.find(c)
			sum = fanSums[key.fan]
		}

		index, ok := assigned[key]
		if !ok {
			index = key.vertex
			if claimed[key.vertex] {
				index = len(order)
				order = append(order, key.vertex)
				normals = append(normals, vector3.Zero[float64]())
			}
			claimed[key.vertex] = true
			assigned[key] = index
			if length := sum.Length(); length > 0 {
				normals[index] = sum.Scale(1 / length)
			}
		}
		indices[c] = index
	}

	return modeling.
		NewMesh(modeling.TriangleTopology, indices).
		SetFloat4Data(gatherByIndex(order, m.Float4Attributes(),
			func(s string) *iter.ArrayIterator[vector4.Float64] { return m.Float4Attribute(s) })).
		SetFloat3Data(gatherByIndex(order, m.Float3Attributes(),
			func(s string) *iter.ArrayIterator[vector3.Float64] { return m.Float3Attribute(s) })).
		SetFloat2Data(gatherByIndex(order, m.Float2Attributes(),
			func(s string) *iter.ArrayIterator[vector2.Float64] { return m.Float2Attribute(s) })).
		SetFloat1Data(gatherByIndex(order, m.Float1Attributes(),
			func(s string) *iter.ArrayIterator[float64] { return m.Float1Attribute(s) })).
		SetFloat3Attribute(modeling.NormalAttribute, normals)
}

type SmoothNormalsImplicitWeldTransformer struct {
	Distance float64
}

func (smt SmoothNormalsImplicitWeldTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	if err = RequireTopology(m, modeling.TriangleTopology); err != nil {
		return
	}

	if err = RequireV3Attribute(m, modeling.PositionAttribute); err != nil {
		return
	}

	return SmoothNormalsImplicitWeld(m, smt.Distance), nil
}

func SmoothNormalsImplicitWeld(m modeling.Mesh, distance float64) modeling.Mesh {
	if distance < 0 {
		panic(fmt.Errorf("weld distance can not be negative, recieved: %f", distance))
	}

	check(RequireTopology(m, modeling.TriangleTopology))
	check(RequireV3Attribute(m, modeling.PositionAttribute))

	tree := m.ToPointCloud().OctTree()

	vertices := m.Float3Attribute(modeling.PositionAttribute)
	normals := make([]vector3.Float64, vertices.Len())
	for i := range normals {
		normals[i] = vector3.Zero[float64]()
	}

	tris := m.Indices()
	for triIndex := 0; triIndex < tris.Len(); triIndex += 3 {
		p1 := tris.At(triIndex)
		p2 := tris.At(triIndex + 1)
		p3 := tris.At(triIndex + 2)
		v1, v2, v3 := vertices.At(p1), vertices.At(p2), vertices.At(p3)

		normal := m.Tri(triIndex / 3).Normal(modeling.PositionAttribute)
		if normal == vector3.Zero[float64]() {
			continue
		}

		a1 := cornerAngle(v2.Sub(v1), v3.Sub(v1))
		a2 := cornerAngle(v3.Sub(v2), v1.Sub(v2))
		a3 := math.Pi - a1 - a2

		for _, weld := range tree.ElementsWithinRange(v1, distance) {
			normals[weld] = normals[weld].Add(normal.Scale(a1))
		}
		for _, weld := range tree.ElementsWithinRange(v2, distance) {
			normals[weld] = normals[weld].Add(normal.Scale(a2))
		}
		for _, weld := range tree.ElementsWithinRange(v3, distance) {
			normals[weld] = normals[weld].Add(normal.Scale(a3))
		}
	}

	zero := vector3.Zero[float64]()
	for i, n := range normals {
		if n == zero {
			continue
		}
		normals[i] = n.Normalized()
	}

	return m.SetFloat3Attribute(modeling.NormalAttribute, normals)
}

type SmoothNormalsNode struct {
	Mesh           nodes.Output[modeling.Mesh] `description:"The mesh to recompute normals for."`
	SmoothingAngle nodes.Output[float64]       `description:"Faces meeting at this many degrees or more keep a hard edge, and vertices sharing a position smooth together. Unset smooths each vertex over every face that uses it."`
}

func (snn SmoothNormalsNode) Description() string {
	return "Recomputes smooth per-vertex normals by averaging the face normals of every triangle touching each vertex, each weighted by its corner angle there. Separate vertices at the same position are not merged."
}

func (snn SmoothNormalsNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if snn.Mesh == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}
	mesh := nodes.GetOutputValue(out, snn.Mesh)
	if !mesh.HasFloat3Attribute(modeling.PositionAttribute) {
		out.Set(mesh)
		out.CaptureError(fmt.Errorf("can't calculate normals without position data"))
		return
	}

	if snn.SmoothingAngle != nil {
		out.Set(SmoothNormalsByAngle(mesh, nodes.GetOutputValue(out, snn.SmoothingAngle)))
		return
	}
	out.Set(SmoothNormals(mesh))
}

type SmoothNormalsImplicitWeldNode struct {
	Mesh     nodes.Output[modeling.Mesh] `description:"The mesh to recompute normals for."`
	Distance nodes.Output[float64]       `description:"Vertices within this distance of each other have their face normals averaged together, even if their positions aren't exactly identical. Must stay below the thinnest feature's thickness. Defaults to 0.01."`
}

func (snn SmoothNormalsImplicitWeldNode) Description() string {
	return "Recomputes smooth per-vertex normals, welding together normals from any vertices within Distance of each other, not just exactly-identical positions. Each face is weighted by its corner angle."
}

func (snn SmoothNormalsImplicitWeldNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if snn.Mesh == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}
	mesh := nodes.GetOutputValue(out, snn.Mesh)
	if !mesh.HasFloat3Attribute(modeling.PositionAttribute) {
		out.Set(mesh)
		out.CaptureError(fmt.Errorf("can't calculate normals without position data"))
		return
	}

	out.Set(SmoothNormalsImplicitWeld(
		mesh,
		max(0, nodes.TryGetOutputValue(out, snn.Distance, 0.01)),
	))
}
