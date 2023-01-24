package modeling

import (
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type MeshView struct {
	Float3Data map[string][]vector3.Float64
	Float2Data map[string][]vector2.Float64
	Float1Data map[string][]float64
	Indices    []int
}
