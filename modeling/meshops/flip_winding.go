package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
)

type FlipTriangleWindingTransformer struct {
	Attribute string
}

func (cat FlipTriangleWindingTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	if err = requireTopology(m, modeling.TriangleTopology); err != nil {
		return
	}

	return FlipTriangleWinding(m), nil
}

func FlipTriangleWinding(m modeling.Mesh) modeling.Mesh {
	if err := requireTopology(m, modeling.TriangleTopology); err != nil {
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
