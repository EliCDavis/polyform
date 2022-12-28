package modeling

import "github.com/EliCDavis/vector"

type MeshView struct {
	Float3Data map[string][]vector.Vector3
	Float2Data map[string][]vector.Vector2
	Float1Data map[string][]float64
	Indices    []int
}
