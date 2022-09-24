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

func MeshFromView(view MeshView) Mesh {
	return Mesh{
		vertices:  view.Vertices,
		triangles: view.Triangles,
		normals:   view.Normals,
		uv:        view.UV,
	}
}

// View exposes the underlying data to be modified. Using this breaks the
// immutable design of the system, but required for some mesh processing.
//
// Modifying the data stored at the indices in the mesh found here will
// directly update the mesh, and side steps any type of validation we could
// have done previously.
func (m Mesh) View() MeshView {
	return MeshView{
		Vertices:  m.vertices,
		Triangles: m.triangles,
		Normals:   m.normals,
		UV:        m.uv,
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

func (m Mesh) CalculateSmoothNormals() Mesh {
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
