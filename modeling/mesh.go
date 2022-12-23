package modeling

import (
	"fmt"
	"math"

	"github.com/EliCDavis/vector"
)

type Mesh struct {
	vertices  []vector.Vector3
	indices   []int
	normals   []vector.Vector3
	uv        [][]vector.Vector2
	materials []MeshMaterial
	topology  Topology
}

func EmptyMesh() Mesh {
	return Mesh{
		vertices:  make([]vector.Vector3, 0),
		indices:   make([]int, 0),
		normals:   make([]vector.Vector3, 0),
		uv:        make([][]vector.Vector2, 0),
		materials: make([]MeshMaterial, 0),
		topology:  Triangle,
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
		indices:   triangles,
		normals:   normals,
		uv:        uvs,
		materials: []MeshMaterial{{len(triangles) / 3, nil}},
		topology:  Triangle,
	}
}

func NewMeshWithMaterials(
	triangles []int,
	vertices []vector.Vector3,
	normals []vector.Vector3,
	uvs [][]vector.Vector2,
	materials []MeshMaterial,
) Mesh {
	return Mesh{
		vertices:  vertices,
		indices:   triangles,
		normals:   normals,
		uv:        uvs,
		materials: materials,
		topology:  Triangle,
	}
}

func MeshFromView(view MeshView) Mesh {
	return Mesh{
		vertices: view.Vertices,
		indices:  view.Triangles,
		normals:  view.Normals,
		uv:       view.UVs,
		topology: Triangle,
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
		Triangles: m.indices,
		Normals:   m.normals,
		UVs:       m.uv,
	}
}

func (m Mesh) Materials() []MeshMaterial {
	return m.materials
}

func (m Mesh) SetMaterial(mat Material) Mesh {
	return NewMeshWithMaterials(m.indices, m.vertices, m.normals, m.uv, []MeshMaterial{{NumOfTris: len(m.indices) / 3, Material: &mat}})
}

func (m Mesh) Tri(i int) Tri {
	return Tri{
		mesh:          &m,
		startingIndex: i * 3,
	}
}

func (m Mesh) TriCount() int {
	return len(m.indices) / 3
}

func (m Mesh) hasUVs() bool {
	return len(m.uv) > 0 && len(m.uv[0]) > 0
}

func (m Mesh) Append(other Mesh) Mesh {
	finalTris := append(m.indices, other.indices...)
	finalMaterials := append(m.materials, other.materials...)

	vertexCountShift := len(m.vertices)
	for i := len(m.indices); i < len(finalTris); i++ {
		finalTris[i] += vertexCountShift
	}
	finalVerts := append(m.vertices, other.vertices...)

	finalNormals := make([]vector.Vector3, 0, len(finalVerts))

	// Fill 2nd "half" of normals
	if len(m.normals) != 0 && len(other.normals) == 0 {
		finalNormals = append(finalNormals, m.normals...)
		for i := 0; i < len(other.vertices); i++ {
			finalNormals = append(finalNormals, vector.Vector3Zero())
		}
	} else if len(m.normals) == 0 && len(other.normals) != 0 {
		for i := 0; i < len(m.vertices); i++ {
			finalNormals = append(finalNormals, vector.Vector3Zero())
		}
		finalNormals = append(finalNormals, other.normals...)
	} else {
		finalNormals = append(finalNormals, m.normals...)
		finalNormals = append(finalNormals, other.normals...)
	}

	finalUVs := make([]vector.Vector2, 0, len(finalVerts))

	// Fill 2nd "half" of UVs
	if m.hasUVs() && !other.hasUVs() {
		finalUVs = append(finalUVs, m.uv[0]...)
		for i := 0; i < len(other.vertices); i++ {
			finalUVs = append(finalUVs, vector.Vector2Zero())
		}
	} else if !m.hasUVs() && other.hasUVs() {
		for i := 0; i < len(m.vertices); i++ {
			finalUVs = append(finalUVs, vector.Vector2Zero())
		}
		finalUVs = append(finalUVs, other.uv[0]...)
	} else {
		finalUVs = append(finalUVs, m.uv[0]...)
		finalUVs = append(finalUVs, other.uv[0]...)
	}

	return NewMeshWithMaterials(finalTris, finalVerts, finalNormals, [][]vector.Vector2{finalUVs}, finalMaterials)
}

func (m Mesh) Translate(v vector.Vector3) Mesh {
	finalVerts := make([]vector.Vector3, len(m.vertices))
	for i := 0; i < len(finalVerts); i++ {
		finalVerts[i] = m.vertices[i].Add(v)
	}
	return NewMeshWithMaterials(m.indices, finalVerts, m.normals, m.uv, m.materials)
}

func (m Mesh) Rotate(q Quaternion) Mesh {
	finalVerts := make([]vector.Vector3, len(m.vertices))
	for i := 0; i < len(finalVerts); i++ {
		finalVerts[i] = q.Rotate(m.vertices[i])
	}

	finalNormals := make([]vector.Vector3, len(m.normals))
	for i := 0; i < len(finalNormals); i++ {
		finalNormals[i] = q.Rotate(m.normals[i])
	}

	return NewMeshWithMaterials(m.indices, finalVerts, finalNormals, m.uv, m.materials)
}

func (m Mesh) Scale(origin, amount vector.Vector3) Mesh {
	scaledVerts := make([]vector.Vector3, len(m.vertices))

	for i, v := range m.vertices {
		scaledVerts[i] = origin.Add(v.Sub(origin).MultByVector(amount))
	}

	return NewMeshWithMaterials(m.indices, scaledVerts, m.normals, m.uv, m.materials)
}

func (m Mesh) ModifyVertices(f func(v vector.Vector3) vector.Vector3) Mesh {
	modified := make([]vector.Vector3, len(m.vertices))

	for i, v := range m.vertices {
		modified[i] = f(v)
	}

	return NewMeshWithMaterials(m.indices, modified, m.normals, m.uv, m.materials)
}

func (m Mesh) ModifyUVs(f func(v vector.Vector3, uv vector.Vector2) vector.Vector2) Mesh {
	modified := make([]vector.Vector2, len(m.uv[0]))

	for i, uv := range m.uv[0] {
		modified[i] = f(m.vertices[i], uv)
	}

	return NewMeshWithMaterials(m.indices, m.vertices, m.normals, [][]vector.Vector2{modified}, m.materials)
}

func (m Mesh) CalculateFlatNormals() Mesh {
	m.requireTopology(Triangle)

	normals := make([]vector.Vector3, len(m.vertices))
	for i := range normals {
		normals[i] = vector.Vector3One()
	}

	verts := m.vertices
	tris := m.indices
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
		indices:   m.indices,
		uv:        m.uv,
		materials: m.materials,
		topology:  m.topology,
	}
}

func (m Mesh) RemoveDegenerateTriangles(sideLength float64) Mesh {
	removedSomething := true

	finalMesh := m
	for removedSomething {
		removedSomething = false
		cm := NewCollapsableMesh(finalMesh)

		for triI := 0; triI < len(finalMesh.indices); triI += 3 {
			if finalMesh.vertices[finalMesh.indices[triI]].Distance(finalMesh.vertices[finalMesh.indices[triI+1]]) < sideLength {
				cm.CollapseTri(triI / 3)
				removedSomething = true
				continue
			}
			if finalMesh.vertices[finalMesh.indices[triI+1]].Distance(finalMesh.vertices[finalMesh.indices[triI+2]]) < sideLength {
				cm.CollapseTri(triI / 3)
				removedSomething = true
				continue
			}
			if finalMesh.vertices[finalMesh.indices[triI+2]].Distance(finalMesh.vertices[finalMesh.indices[triI]]) < sideLength {
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
	for triI := 0; triI < len(m.indices); triI += 3 {
		v1 := Vector3ToInt(m.vertices[m.indices[triI+0]], decimalPlace)
		v2 := Vector3ToInt(m.vertices[m.indices[triI+1]], decimalPlace)
		v3 := Vector3ToInt(m.vertices[m.indices[triI+2]], decimalPlace)

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
		indices:  newTris,
		vertices: finalVerts,
		normals:  finalNormals,
		uv:       [][]vector.Vector2{finalUVs},
		topology: m.topology,
	}
}

func (m Mesh) VertexNeighborTable() VertexLUT {
	table := VertexLUT{}
	for triI := 0; triI < len(m.indices); triI += 3 {
		p1 := m.indices[triI]
		p2 := m.indices[triI+1]
		p3 := m.indices[triI+2]

		table.AddLookup(p1, p2)
		table.AddLookup(p1, p3)

		table.AddLookup(p2, p1)
		table.AddLookup(p2, p3)

		table.AddLookup(p3, p1)
		table.AddLookup(p3, p2)
	}
	return table
}

func (m Mesh) requireTopology(t Topology) {
	if m.topology != t {
		panic(fmt.Errorf("can not perform operation for a mesh with a topology of %s, requires %s topology", m.topology.String(), t.String()))
	}
}

func (m Mesh) SmoothLaplacian(iterations int, smoothingFactor float64) Mesh {
	m.requireTopology(Triangle)

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
		indices:   m.indices,
		uv:        m.uv,
		materials: m.materials,
		topology:  m.topology,
	}
}

func (m Mesh) CalculateSmoothNormals() Mesh {
	m.requireTopology(Triangle)

	normals := make([]vector.Vector3, len(m.vertices))
	for i := range normals {
		normals[i] = vector.Vector3Zero()
	}

	verts := m.vertices
	tris := m.indices
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
		indices:   m.indices,
		uv:        m.uv,
		materials: m.materials,
		topology:  m.topology,
	}
}

func (m Mesh) RemoveUnusedIndices() Mesh {
	finalTris := make([]int, len(m.indices))
	finalVerts := make([]vector.Vector3, 0)
	finalNormals := make([]vector.Vector3, 0)
	finalUVs := make([]vector.Vector2, 0)

	used := make([]bool, len(m.vertices))
	for _, t := range m.indices {
		used[t] = true
	}

	shiftBy := make([]int, len(m.vertices))
	skipped := 0
	for i, v := range m.vertices {
		if used[i] {
			finalVerts = append(finalVerts, v)
		} else {
			skipped++
		}
		shiftBy[i] = skipped
	}

	if len(m.normals) > 0 {
		for i, n := range m.normals {
			if used[i] {
				finalNormals = append(finalNormals, n)
			}
		}
	}

	if len(m.uv) > 0 && len(m.uv[0]) > 0 {
		for i, n := range m.uv[0] {
			if used[i] {
				finalUVs = append(finalUVs, n)
			}
		}
	}

	for triI := 0; triI < len(finalTris); triI++ {
		finalTris[triI] = m.indices[triI] - shiftBy[m.indices[triI]]
	}

	return Mesh{
		indices:   finalTris,
		vertices:  finalVerts,
		normals:   finalNormals,
		uv:        [][]vector.Vector2{finalUVs},
		materials: m.materials,
		topology:  m.topology,
	}
}

// SplitOnUniqueMaterials generates a mesh per material,
func (m Mesh) SplitOnUniqueMaterials() []Mesh {
	if len(m.materials) < 2 {
		return []Mesh{m}
	}

	workingMeshes := make(map[*Material]*Mesh)

	curMatIndex := 0
	trisFromOtherMats := 0

	workingMeshes[m.materials[curMatIndex].Material] = &Mesh{
		vertices: m.vertices,
		normals:  m.normals,
		uv:       m.uv,
		materials: []MeshMaterial{
			{
				NumOfTris: 0,
				Material:  m.materials[curMatIndex].Material,
			},
		},
	}

	for triStart := 0; triStart < len(m.indices); triStart += 3 {
		if m.materials[curMatIndex].NumOfTris+trisFromOtherMats <= triStart/3 {
			trisFromOtherMats += m.materials[curMatIndex].NumOfTris
			curMatIndex++
			if _, ok := workingMeshes[m.materials[curMatIndex].Material]; !ok {
				workingMeshes[m.materials[curMatIndex].Material] = &Mesh{
					vertices: m.vertices,
					normals:  m.normals,
					uv:       m.uv,
					materials: []MeshMaterial{
						{
							NumOfTris: 0,
							Material:  m.materials[curMatIndex].Material,
						},
					},
					topology: m.topology,
				}
			}
		}
		mesh := workingMeshes[m.materials[curMatIndex].Material]
		mesh.indices = append(
			mesh.indices,
			m.indices[triStart],
			m.indices[triStart+1],
			m.indices[triStart+2],
		)
		mesh.materials[0].NumOfTris += 1
	}

	finalMeshes := make([]Mesh, 0, len(workingMeshes))
	for _, m := range workingMeshes {
		finalMeshes = append(finalMeshes, m.RemoveUnusedIndices())
	}
	return finalMeshes
}
