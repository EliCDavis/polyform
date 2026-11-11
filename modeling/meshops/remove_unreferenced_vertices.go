package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
)

type RemovedUnreferencedVerticesTransformer struct{}

func (fnt RemovedUnreferencedVerticesTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	return RemovedUnreferencedVertices(m), nil
}

func RemovedUnreferencedVertices(m modeling.Mesh) modeling.Mesh {
	originalIndices := m.Indices()

	used := make([]bool, m.AttributeLength())
	for i := 0; i < originalIndices.Len(); i++ {
		used[originalIndices.At(i)] = true
	}

	remapped := make([]int, len(used))
	kept := make([]int, 0)
	for i, u := range used {
		if u {
			remapped[i] = len(kept)
			kept = append(kept, i)
		}
	}

	finalIndices := make([]int, originalIndices.Len())
	for i := range finalIndices {
		finalIndices[i] = remapped[originalIndices.At(i)]
	}

	v4, v3, v2, v1 := gatherVertices(m, kept)
	return modeling.
		NewMesh(m.Topology(), finalIndices).
		SetFloat4Data(v4).
		SetFloat3Data(v3).
		SetFloat2Data(v2).
		SetFloat1Data(v1)
}
