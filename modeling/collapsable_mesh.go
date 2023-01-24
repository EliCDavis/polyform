package modeling

import (
	"fmt"
	"sort"

	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type CollapsableMesh struct {
	vertices        []vector3.Float64
	triangles       []int
	normals         []vector3.Float64
	uv              []vector2.Float64
	verticesRemoved []int
	trisRemoved     int

	// Doubly linked vertex to vertex lookup
	v2vLUT VertexLUT

	// One way lookup from vertex index to triangle indexes that reference it
	v2tLUT VertexLUT
}

func NewCollapsableMesh(m Mesh) CollapsableMesh {
	v2vLUT := VertexLUT{}
	v2tLUT := VertexLUT{}
	for triI := 0; triI < len(m.indices); triI += 3 {
		v2vLUT.Link(m.indices[triI], m.indices[triI+1])
		v2vLUT.Link(m.indices[triI+1], m.indices[triI+2])
		v2vLUT.Link(m.indices[triI+2], m.indices[triI])
		v2tLUT.AddLookup(m.indices[triI], triI)
		v2tLUT.AddLookup(m.indices[triI+1], triI+1)
		v2tLUT.AddLookup(m.indices[triI+2], triI+2)
	}

	meshVerts, ok := m.v3Data[PositionAttribute]
	if !ok {
		panic("expected float3 position attribute")
	}
	verts := make([]vector3.Float64, len(meshVerts))
	copy(verts, meshVerts)

	triangles := make([]int, len(m.indices))
	copy(triangles, m.indices)

	meshNormals, ok := m.v3Data[NormalAttribute]
	if !ok {
		panic("expected float3 normal attribute")
	}
	normals := make([]vector3.Float64, len(meshNormals))
	copy(normals, meshNormals)

	meshUVs, ok := m.v2Data[TexCoordAttribute]
	if !ok {
		panic("expected float3 normal attribute")
	}
	uvs := make([]vector2.Float64, len(meshUVs))
	copy(uvs, meshUVs)

	return CollapsableMesh{
		vertices:    verts,
		triangles:   triangles,
		normals:     normals,
		uv:          uvs,
		v2vLUT:      v2vLUT,
		v2tLUT:      v2tLUT,
		trisRemoved: 0,
	}
}

func (cm CollapsableMesh) ToMesh() Mesh {
	finalTris := make([]int, len(cm.triangles)-(cm.trisRemoved*3))

	sort.Ints(cm.verticesRemoved)
	vertRemovedIndex := 0
	shift := make([]int, len(cm.vertices))
	currentShift := 0
	finalVerts := make([]vector3.Float64, 0)
	finalNormals := make([]vector3.Float64, 0)
	finalUVs := make([]vector2.Float64, 0)
	for i := 0; i < len(cm.vertices); i++ {
		if vertRemovedIndex < len(cm.verticesRemoved) {
			if i == cm.verticesRemoved[vertRemovedIndex] {
				currentShift++
				vertRemovedIndex++
			} else {
				finalVerts = append(finalVerts, cm.vertices[i])
				if len(cm.normals) > 0 {
					finalNormals = append(finalNormals, cm.normals[i])
				}
				if len(cm.uv) > 0 {
					finalUVs = append(finalUVs, cm.uv[i])
				}
			}
		}
		shift[i] = currentShift
	}

	finalTriIndex := 0
	for triI := 0; triI < len(cm.triangles); triI += 3 {
		if cm.triangles[triI] == -1 {
			continue
		}
		finalTris[finalTriIndex] = cm.triangles[triI] - shift[cm.triangles[triI]]
		finalTris[finalTriIndex+1] = cm.triangles[triI+1] - shift[cm.triangles[triI+1]]
		finalTris[finalTriIndex+2] = cm.triangles[triI+2] - shift[cm.triangles[triI+2]]
		finalTriIndex += 3
	}

	return NewMesh(
		finalTris,
		map[string][]vector3.Float64{
			PositionAttribute: finalVerts,
			NormalAttribute:   finalNormals,
		},
		map[string][]vector2.Float64{
			TexCoordAttribute: finalUVs,
		},
		nil,
		nil,
	)
}

func (cm CollapsableMesh) validTri(triIndex int) bool {
	if cm.triangles[triIndex] == cm.triangles[triIndex+1] {
		return false
	}

	if cm.triangles[triIndex] == cm.triangles[triIndex+2] {
		return false
	}

	if cm.triangles[triIndex+1] == cm.triangles[triIndex+2] {
		return false
	}

	return true
}

func (cm *CollapsableMesh) checkAndInvalidateTri(triIndex int) {
	if cm.triangles[triIndex] == -1 {
		return
	}

	if cm.validTri(triIndex) {
		return
	}
	cm.trisRemoved++

	cm.v2tLUT.RemoveLookup(cm.triangles[triIndex], triIndex)
	cm.v2tLUT.RemoveLookup(cm.triangles[triIndex+1], triIndex+1)
	cm.v2tLUT.RemoveLookup(cm.triangles[triIndex+2], triIndex+2)

	cm.v2vLUT.RemoveLink(cm.triangles[triIndex], cm.triangles[triIndex+1])
	cm.v2vLUT.RemoveLink(cm.triangles[triIndex+1], cm.triangles[triIndex+2])
	cm.v2vLUT.RemoveLink(cm.triangles[triIndex+2], cm.triangles[triIndex])

	cm.triangles[triIndex] = -1
	cm.triangles[triIndex+1] = -1
	cm.triangles[triIndex+2] = -1
}

func (cm *CollapsableMesh) CollapseTri(tri int) {
	triIndex := tri * 3
	if triIndex > len(cm.triangles) {
		panic(fmt.Sprintf("collapsing tri %d does not correspond to any triangle in mesh, outside bounds", tri))
	}

	// This has already been collapsed by a previous operation. Nothing to do!
	if cm.triangles[triIndex] == -1 {
		return
	}

	newVertIndex := len(cm.vertices)
	cm.vertices = append(
		cm.vertices,
		cm.vertices[cm.triangles[triIndex+0]].
			Add(cm.vertices[cm.triangles[triIndex+1]]).
			Add(cm.vertices[cm.triangles[triIndex+2]]).
			DivByConstant(3.0))

	if len(cm.normals) > 0 {
		cm.normals = append(
			cm.normals,
			cm.normals[cm.triangles[triIndex+0]].
				Add(cm.normals[cm.triangles[triIndex+1]]).
				Add(cm.normals[cm.triangles[triIndex+2]]).
				DivByConstant(3.0).Normalized())
	}
	if len(cm.uv) > 0 {
		cm.uv = append(cm.uv, cm.uv[cm.triangles[triIndex+0]].
			Add(cm.uv[cm.triangles[triIndex+1]]).
			Add(cm.uv[cm.triangles[triIndex+2]]).
			DivByConstant(3.0))
	}

	// ========== Vertex to Vertex LUT Updates ================================
	// Link up all vertices that used to reference this triangle to the newly
	// created vertex.
	for vn := range cm.v2vLUT.Lookup(cm.triangles[triIndex]) {
		cm.v2vLUT.Link(vn, newVertIndex)
	}
	for vn := range cm.v2vLUT.Lookup(cm.triangles[triIndex+1]) {
		cm.v2vLUT.Link(vn, newVertIndex)
	}
	for vn := range cm.v2vLUT.Lookup(cm.triangles[triIndex+2]) {
		cm.v2vLUT.Link(vn, newVertIndex)
	}

	// Remove all references to the old triangle's vertices.
	cm.v2vLUT.RemoveVertex(cm.triangles[triIndex])
	cm.v2vLUT.RemoveVertex(cm.triangles[triIndex+1])
	cm.v2vLUT.RemoveVertex(cm.triangles[triIndex+2])

	cm.verticesRemoved = append(
		cm.verticesRemoved,
		cm.triangles[triIndex],
		cm.triangles[triIndex+1],
		cm.triangles[triIndex+2],
	)

	// ========== Vertex to Triangle LUT Updates ==============================
	// Update all triangles that references the old triangle's vertices to now
	// look at this one.
	triIndicesUpdated := make([]int, 0)
	for t := range cm.v2tLUT.Remove(cm.triangles[triIndex]) {
		cm.triangles[t] = newVertIndex
		triIndicesUpdated = append(triIndicesUpdated, t)
	}
	for t := range cm.v2tLUT.Remove(cm.triangles[triIndex+1]) {
		cm.triangles[t] = newVertIndex
		triIndicesUpdated = append(triIndicesUpdated, t)
	}
	for t := range cm.v2tLUT.Remove(cm.triangles[triIndex+2]) {
		cm.triangles[t] = newVertIndex
		triIndicesUpdated = append(triIndicesUpdated, t)
	}

	// Look for any triangles we just invalidated by collapsing this one. This
	// will apply to any triangles that shared a side with the triangle that
	// was just collapsed. There might be some smart way to use this fact to
	// quickly do this computation. For the time being, brute force!
	for _, t := range triIndicesUpdated {
		cm.checkAndInvalidateTri(t - (t % 3))
	}

	// cm.triangles[triIndex] = -1
	// cm.triangles[triIndex+1] = -1
	// cm.triangles[triIndex+2] = -1
	// cm.trisRemoved++
}
