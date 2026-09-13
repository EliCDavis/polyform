// Jacobson, Kavan and Sorkine-Hornung, "Robust Inside-Outside Segmentation
// using Generalized Winding Numbers", SIGGRAPH 2013.
//
//	https://igl.ethz.ch/projects/winding-number/
//
// Barill, Dickson, Schmidt, Levin and Jacobson, "Fast Winding Numbers for
// Soups and Clouds", SIGGRAPH 2018.
//
//	https://dl.acm.org/doi/10.1145/3197517.3201337
package winding

import (
	"errors"
	"fmt"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

type Sampler interface {
	Number(p vector3.Float64) float64
	Inside(p vector3.Float64) bool
}

func FromMesh(m modeling.Mesh) (Sampler, error) {
	if m.Topology() != modeling.TriangleTopology {
		return nil, fmt.Errorf("need a triangle mesh, got %v", m.Topology())
	}
	if !m.HasFloat3Attribute(modeling.PositionAttribute) {
		return nil, errors.New("mesh carries no position attribute")
	}

	indices := m.Indices()
	positions := m.Float3Attribute(modeling.PositionAttribute)

	tris := make([][3]vector3.Float64, 0, indices.Len()/3)
	for i := 0; i+2 < indices.Len(); i += 3 {
		tris = append(tris, [3]vector3.Float64{
			positions.At(indices.At(i)),
			positions.At(indices.At(i + 1)),
			positions.At(indices.At(i + 2)),
		})
	}

	return &TriangleSampler{tris: tris}, nil
}
