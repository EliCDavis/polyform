package meshops

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type SmoothNormalsTransformer struct{}

func (smt SmoothNormalsTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	if err = requireTopology(m, modeling.TriangleTopology); err != nil {
		return
	}

	if err = requireV3Attribute(m, modeling.PositionAttribute); err != nil {
		return
	}

	return SmoothNormals(m), nil
}

func SmoothNormals(m modeling.Mesh) modeling.Mesh {
	check(requireTopology(m, modeling.TriangleTopology))
	check(requireV3Attribute(m, modeling.PositionAttribute))

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
		// normalize(cross(B-A, C-A))
		normalized := vertices.At(p2).Sub(vertices.At(p1)).Cross(vertices.At(p3).Sub(vertices.At(p1)))

		// This occurs whenever the given tri is actually just a line
		if math.IsNaN(normalized.X()) {
			continue
		}

		normals[p1] = normals[p1].Add(normalized)
		normals[p2] = normals[p2].Add(normalized)
		normals[p3] = normals[p3].Add(normalized)
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
	nodes.StructData[modeling.Mesh]

	Mesh nodes.NodeOutput[modeling.Mesh]
}

func (snn SmoothNormalsNode) Process() (modeling.Mesh, error) {
	return SmoothNormals(snn.Mesh.Data()), nil
}

func (snn *SmoothNormalsNode) SmoothedMesh() nodes.NodeOutput[modeling.Mesh] {
	return &nodes.StructNodeOutput[modeling.Mesh]{
		Definition: snn,
	}
}
