package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type FlatNormalsTransformer struct{}

func (fnt FlatNormalsTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	if err = RequireTopology(m, modeling.TriangleTopology); err != nil {
		return
	}

	if err = RequireV3Attribute(m, modeling.PositionAttribute); err != nil {
		return
	}

	return FlatNormals(m), nil
}

func FlatNormals(m modeling.Mesh) modeling.Mesh {
	check(RequireTopology(m, modeling.TriangleTopology))
	check(RequireV3Attribute(m, modeling.PositionAttribute))

	vertices := m.Float3Attribute(modeling.PositionAttribute)
	normals := make([]vector3.Float64, vertices.Len())
	for i := range normals {
		normals[i] = vector3.One[float64]()
	}

	tris := m.Indices()
	for triIndex := 0; triIndex < tris.Len(); triIndex += 3 {
		p1 := tris.At(triIndex)
		p2 := tris.At(triIndex + 1)
		p3 := tris.At(triIndex + 2)

		// normalize(cross(B-A, C-A))
		normalized := vertices.At(p2).Sub(vertices.At(p1)).Cross(vertices.At(p3).Sub(vertices.At(p1))).Normalized()
		normals[p1] = normalized
		normals[p2] = normalized
		normals[p3] = normalized
	}

	for i, n := range normals {
		normals[i] = n.Normalized()
	}

	return m.SetFloat3Attribute(modeling.NormalAttribute, normals)
}

type FlatNormalsNode = nodes.Struct[FlatNormalsNodeData]

type FlatNormalsNodeData struct {
	Mesh nodes.Output[modeling.Mesh]
}

func (fnnd FlatNormalsNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	if fnnd.Mesh == nil {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.TriangleTopology))
	}

	mesh := fnnd.Mesh.Value()

	if !mesh.HasFloat3Attribute(modeling.PositionAttribute) || mesh.Topology() != modeling.TriangleTopology {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.TriangleTopology))
	}

	return nodes.NewStructOutput(FlatNormals(mesh))
}
