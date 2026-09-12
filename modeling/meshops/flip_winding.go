package meshops

import (
	"fmt"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type FlipTriangleWindingTransformer struct {
	Attribute string
}

func (cat FlipTriangleWindingTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	if err = RequireTopology(m, modeling.TriangleTopology); err != nil {
		return
	}

	return FlipTriangleWinding(m), nil
}

func FlipTriangleWinding(m modeling.Mesh) modeling.Mesh {
	if err := RequireTopology(m, modeling.TriangleTopology); err != nil {
		panic(err)
	}

	tris := m.Indices()
	finalTris := make([]int, tris.Len())
	for triIndex := 0; triIndex < tris.Len(); triIndex += 3 {
		finalTris[triIndex+1] = tris.At(triIndex)
		finalTris[triIndex] = tris.At(triIndex + 1)
		finalTris[triIndex+2] = tris.At(triIndex + 2)
	}

	flipped := m.SetIndices(finalTris)

	if !m.HasFloat3Attribute(modeling.NormalAttribute) {
		return flipped
	}
	normals := m.Float3Attribute(modeling.NormalAttribute)
	negated := make([]vector3.Float64, normals.Len())
	for i := 0; i < normals.Len(); i++ {
		negated[i] = normals.At(i).Scale(-1)
	}
	return flipped.SetFloat3Attribute(modeling.NormalAttribute, negated)
}

type FlipTriangleWindingNode struct {
	Mesh nodes.Output[modeling.Mesh]
}

func (n FlipTriangleWindingNode) Description() string {
	return "Reverses every triangle's winding, flipping which side is the front face. Negates normals too, so the mesh shades from its new front."
}

func (n FlipTriangleWindingNode) Flipped(out *nodes.StructOutput[modeling.Mesh]) {
	out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
	if n.Mesh == nil {
		return
	}

	mesh := nodes.GetOutputValue(out, n.Mesh)
	if mesh.Topology() != modeling.TriangleTopology {
		out.CaptureError(fmt.Errorf("Cant flip triangles of a non triangle mesh"))
		return
	}
	out.Set(FlipTriangleWinding(mesh))
}
