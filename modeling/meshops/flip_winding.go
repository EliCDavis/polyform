package meshops

import (
	"fmt"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
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

	return m.SetIndices(finalTris)
}

type FlipTriangleWindingNode struct {
	Mesh nodes.Output[modeling.Mesh]
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
