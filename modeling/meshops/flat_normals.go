package meshops

import (
	"errors"

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

type FlatNormalsNode struct {
	Mesh nodes.Output[modeling.Mesh]
}

func (fnnd FlatNormalsNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
	if fnnd.Mesh == nil {
		return
	}

	mesh := nodes.GetOutputValue(out, fnnd.Mesh)
	if !mesh.HasFloat3Attribute(modeling.PositionAttribute) {
		out.CaptureError(errors.New("can't calculate normals without normal data"))
		return
	}

	if mesh.Topology() != modeling.TriangleTopology {
		out.CaptureError(errors.New("calculating normals requires triangle topology"))
		return
	}

	out.Set(FlatNormals(mesh))
}
