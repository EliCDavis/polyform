package mesh

import (
	"math"

	"github.com/EliCDavis/vector"
)

type Mesh struct {
	vertices  []vector.Vector3
	triangles []int
	normals   []vector.Vector3
	uv        [][]vector.Vector2
}

func EmptyMesh() Mesh {
	return Mesh{
		vertices:  make([]vector.Vector3, 0),
		triangles: make([]int, 0),
		normals:   make([]vector.Vector3, 0),
		uv:        make([][]vector.Vector2, 0),
	}
}

func NewMesh(
	triangles []int,
	vertices []vector.Vector3,
	normals []vector.Vector3,
	uvs [][]vector.Vector2,
) Mesh {
	return Mesh{
		vertices:  vertices,
		triangles: triangles,
		normals:   normals,
		uv:        uvs,
	}
}

func MeshFromView(view MeshView) Mesh {
	return Mesh{
		vertices:  view.Vertices,
		triangles: view.Triangles,
		normals:   view.Normals,
		uv:        view.UVs,
	}
}

// View exposes the underlying data to be modified. Using this breaks the
// immutable design of the system, but required for some mesh processing.
//
// Modifying the data stored in the mesh found here will directly update the
// mesh, and side-steps any type of validation we could have done previously.
//
// If you make changes to this view, assume the mesh and all ancestors of said
// mesh have just become garbage.
func (m Mesh) View() MeshView {
	return MeshView{
		Vertices:  m.vertices,
		Triangles: m.triangles,
		Normals:   m.normals,
		UVs:       m.uv,
	}
}

func (m Mesh) Tri(i int) Tri {
	return Tri{
		mesh:          m,
		startingIndex: i * 3,
	}
}

func (m Mesh) TriCount() int {
	return len(m.triangles) / 3
}

func (m Mesh) Scale(origin, amount vector.Vector3) Mesh {
	scaledVerts := make([]vector.Vector3, len(m.vertices))

	for i, v := range m.vertices {
		scaledVerts[i] = origin.Add(v.Sub(origin).MultByVector(amount))
	}

	return Mesh{
		vertices:  scaledVerts,
		normals:   m.normals,
		triangles: m.triangles,
		uv:        m.uv,
	}
}

func (m Mesh) CalculateFlatNormals() Mesh {
	normals := make([]vector.Vector3, len(m.vertices))
	for i := range normals {
		normals[i] = vector.Vector3One()
	}

	verts := m.vertices
	tris := m.triangles
	for triIndex := 0; triIndex < len(tris); triIndex += 3 {
		p1 := tris[triIndex]
		p2 := tris[triIndex+1]
		p3 := tris[triIndex+2]
		// normalize(cross(B-A, C-A))
		normalized := verts[p2].Sub(verts[p1]).Cross(verts[p3].Sub(verts[p1])).Normalized()
		normals[p1] = normalized
		normals[p2] = normalized
		normals[p3] = normalized
	}

	for i, n := range normals {
		normals[i] = n.Normalized()
	}

	return Mesh{
		vertices:  m.vertices,
		normals:   normals,
		triangles: m.triangles,
		uv:        m.uv,
	}
}

func (m Mesh) RemoveDegenerateTriangles(sideLength float64) Mesh {
	removedSomething := true

	finalMesh := m
	for removedSomething {
		removedSomething = false
		cm := NewCollapsableMesh(finalMesh)

		for triI := 0; triI < len(finalMesh.triangles); triI += 3 {
			if finalMesh.vertices[finalMesh.triangles[triI]].Distance(finalMesh.vertices[finalMesh.triangles[triI+1]]) < sideLength {
				cm.CollapseTri(triI / 3)
				removedSomething = true
				continue
			}
			if finalMesh.vertices[finalMesh.triangles[triI+1]].Distance(finalMesh.vertices[finalMesh.triangles[triI+2]]) < sideLength {
				cm.CollapseTri(triI / 3)
				removedSomething = true
				continue
			}
			if finalMesh.vertices[finalMesh.triangles[triI+2]].Distance(finalMesh.vertices[finalMesh.triangles[triI]]) < sideLength {
				removedSomething = true
				cm.CollapseTri(triI / 3)
			}
		}
		if removedSomething {
			finalMesh = cm.ToMesh()
		}
	}

	return finalMesh
}

func (m Mesh) WeldByVertices(decimalPlace int) Mesh {
	// =================== Finding unique vertices ============================
	vertILU := make(map[VectorInt]int)
	vertIToOriginalLU := make(map[int]int)

	// Mapping from rounded vector to whether or not it get's used by a triangle
	// in the resulting mesh
	vertLUUsed := make(map[VectorInt]bool)

	// count of unique vertices once rounded
	uniqueVertCount := 0

	for vi, v := range m.vertices {
		vInt := Vector3ToInt(v, decimalPlace)

		if _, ok := vertILU[vInt]; !ok {
			vertILU[vInt] = uniqueVertCount
			vertLUUsed[vInt] = false
			vertIToOriginalLU[uniqueVertCount] = vi
			uniqueVertCount++
		}
	}

	// Building tris from unique vertices
	newTris := make([]int, 0)
	for triI := 0; triI < len(m.triangles); triI += 3 {
		v1 := Vector3ToInt(m.vertices[m.triangles[triI+0]], decimalPlace)
		v2 := Vector3ToInt(m.vertices[m.triangles[triI+1]], decimalPlace)
		v3 := Vector3ToInt(m.vertices[m.triangles[triI+2]], decimalPlace)

		if v1 == v2 {
			continue
		}

		if v1 == v3 {
			continue
		}

		if v2 == v3 {
			continue
		}

		vertLUUsed[v1] = true
		vertLUUsed[v2] = true
		vertLUUsed[v3] = true
		newTris = append(newTris, vertILU[v1], vertILU[v2], vertILU[v3])
	}

	finalVerts := make([]vector.Vector3, 0)
	finalNormals := make([]vector.Vector3, 0)
	finalUVs := make([]vector.Vector2, 0)
	shiftBy := make([]int, uniqueVertCount)
	curShift := 0
	for vertIndex := 0; vertIndex < uniqueVertCount; vertIndex++ {

		originalIndex := vertIToOriginalLU[vertIndex]
		v := m.vertices[originalIndex]
		vi := Vector3ToInt(v, decimalPlace)
		if vertLUUsed[vi] {
			finalVerts = append(finalVerts, v)
			if len(m.normals) > 0 {
				finalNormals = append(finalNormals, m.normals[originalIndex])
			}

			if len(m.uv) > 0 && len(m.uv[0]) > 0 {
				finalUVs = append(finalUVs, m.uv[0][originalIndex])
			}
		} else {
			// Not used, need to shift triangles who's points point to vertices that come after this unsed one
			curShift++
		}
		shiftBy[vertIndex] = curShift
	}

	// Shift all the triangles appropriately since we just removed a bunch of vertices no longer used
	for triI := 0; triI < len(newTris); triI++ {
		newTris[triI] -= shiftBy[newTris[triI]]
	}

	return Mesh{
		triangles: newTris,
		vertices:  finalVerts,
		normals:   finalNormals,
		uv:        [][]vector.Vector2{finalUVs},
	}
}

func (m Mesh) VertexNeighborTable() VertexLUT {
	table := VertexLUT{}
	for triI := 0; triI < len(m.triangles); triI += 3 {
		p1 := m.triangles[triI]
		p2 := m.triangles[triI+1]
		p3 := m.triangles[triI+2]

		table.AddLookup(p1, p2)
		table.AddLookup(p1, p3)

		table.AddLookup(p2, p1)
		table.AddLookup(p2, p3)

		table.AddLookup(p3, p1)
		table.AddLookup(p3, p2)
	}
	return table
}

func (m Mesh) SmoothLaplacian(iterations int, smoothingFactor float64) Mesh {
	lut := m.VertexNeighborTable()

	vertices := make([]vector.Vector3, len(m.vertices))
	for i := range vertices {
		vertices[i] = m.vertices[i]
	}

	for i := 0; i < iterations; i++ {
		for vi, vertex := range vertices {
			vs := vector.Vector3Zero()

			for vn := range lut.Lookup(vi) {
				vs = vs.Add(vertices[vn])
			}

			vertices[vi] = vertex.Add(
				vs.
					DivByConstant(float64(lut.Count(vi))).
					Sub(vertex).
					MultByConstant(smoothingFactor))
		}
	}

	return Mesh{
		vertices:  vertices,
		normals:   m.normals,
		triangles: m.triangles,
		uv:        m.uv,
	}
}

func (m Mesh) CalculateSmoothNormals() Mesh {
	normals := make([]vector.Vector3, len(m.vertices))
	for i := range normals {
		normals[i] = vector.Vector3Zero()
	}

	verts := m.vertices
	tris := m.triangles
	for triIndex := 0; triIndex < len(tris); triIndex += 3 {
		p1 := tris[triIndex]
		p2 := tris[triIndex+1]
		p3 := tris[triIndex+2]
		// normalize(cross(B-A, C-A))
		normalized := verts[p2].Sub(verts[p1]).Cross(verts[p3].Sub(verts[p1]))

		// This occurs whenever the given tri is actually just a line
		if math.IsNaN(normalized.X()) {
			continue
		}

		normals[p1] = normals[p1].Add(normalized)
		normals[p2] = normals[p2].Add(normalized)
		normals[p3] = normals[p3].Add(normalized)
	}

	for i, n := range normals {
		normals[i] = n.Normalized()
	}

	return Mesh{
		vertices:  m.vertices,
		normals:   normals,
		triangles: m.triangles,
		uv:        m.uv,
	}
}
