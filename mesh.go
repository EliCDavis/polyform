package mesh

import (
	"fmt"
	"io"
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
// Modifying the data stored at the indices in the mesh found here will
// directly update the mesh, and side steps any type of validation we could
// have done previously.
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

func (m Mesh) WriteObj(out io.Writer) error {
	for _, v := range m.vertices {
		_, err := fmt.Fprintf(out, "v %f %f %f\n", v.X(), v.Y(), v.Z())
		if err != nil {
			return err
		}
	}

	for _, uvChannel := range m.uv {
		for _, uv := range uvChannel {
			_, err := fmt.Fprintf(out, "vt %f %f\n", uv.X(), uv.Y())
			if err != nil {
				return err
			}
		}
	}

	for _, n := range m.normals {
		_, err := fmt.Fprintf(out, "vn %f %f %f\n", n.X(), n.Y(), n.Z())
		if err != nil {
			return err
		}
	}

	if len(m.normals) > 0 && len(m.uv) > 0 {
		for triIndex := 0; triIndex < len(m.triangles); triIndex += 3 {
			p1 := m.triangles[triIndex] + 1
			p2 := m.triangles[triIndex+1] + 1
			p3 := m.triangles[triIndex+2] + 1
			_, err := fmt.Fprintf(out, "f %d/%d/%d %d/%d/%d %d/%d/%d\n", p1, p1, p1, p2, p2, p2, p3, p3, p3)
			if err != nil {
				return err
			}
		}
	} else if len(m.normals) > 0 {
		for triIndex := 0; triIndex < len(m.triangles); triIndex += 3 {
			p1 := m.triangles[triIndex] + 1
			p2 := m.triangles[triIndex+1] + 1
			p3 := m.triangles[triIndex+2] + 1
			_, err := fmt.Fprintf(out, "f %d//%d %d//%d %d//%d\n", p1, p1, p2, p2, p3, p3)
			if err != nil {
				return err
			}
		}
	} else if len(m.uv) > 0 {
		for triIndex := 0; triIndex < len(m.triangles); triIndex += 3 {
			p1 := m.triangles[triIndex] + 1
			p2 := m.triangles[triIndex+1] + 1
			p3 := m.triangles[triIndex+2] + 1
			_, err := fmt.Fprintf(out, "f %d/%d %d/%d %d/%d\n", p1, p1, p2, p2, p3, p3)
			if err != nil {
				return err
			}
		}
	} else {
		for triIndex := 0; triIndex < len(m.triangles); triIndex += 3 {
			p1 := m.triangles[triIndex] + 1
			p2 := m.triangles[triIndex+1] + 1
			p3 := m.triangles[triIndex+2] + 1
			_, err := fmt.Fprintf(out, "f %d %d %d\n", p1, p2, p3)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
