package mesh

import "github.com/EliCDavis/vector"

type MeshView struct {
	Vertices  []vector.Vector3
	Triangles []int
	Normals   []vector.Vector3
	UVs       [][]vector.Vector2
}
