package meshops

import (
	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type RemovedUnreferencedVerticesTransformer struct{}

func (fnt RemovedUnreferencedVerticesTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	return RemovedUnreferencedVertices(m), nil
}

func removedUnreferenced[T any](used []bool, attributes []string, retriever func(string) *iter.ArrayIterator[T]) map[string][]T {
	finalData := make(map[string][]T)
	for _, attribute := range attributes {
		data := retriever(attribute)
		finalAtrVals := make([]T, 0)
		for i := 0; i < data.Len(); i++ {
			if used[i] {
				finalAtrVals = append(finalAtrVals, data.At(i))
			}
		}

		if len(finalAtrVals) == 0 {
			continue
		}

		finalData[attribute] = finalAtrVals
	}
	return finalData
}

func RemovedUnreferencedVertices(m modeling.Mesh) modeling.Mesh {
	originalIndices := m.Indices()

	used := make([]bool, m.AttributeLength())
	for i := 0; i < originalIndices.Len(); i++ {
		used[originalIndices.At(i)] = true
	}

	shiftBy := make([]int, m.AttributeLength())
	skipped := 0
	for i := range shiftBy {
		if !used[i] {
			skipped++
		}
		shiftBy[i] = skipped
	}

	finalV4Data := removedUnreferenced(used, m.Float4Attributes(), func(s string) *iter.ArrayIterator[vector4.Float64] { return m.Float4Attribute(s) })
	finalV3Data := removedUnreferenced(used, m.Float3Attributes(), func(s string) *iter.ArrayIterator[vector3.Float64] { return m.Float3Attribute(s) })
	finalV2Data := removedUnreferenced(used, m.Float2Attributes(), func(s string) *iter.ArrayIterator[vector2.Float64] { return m.Float2Attribute(s) })
	finalV1Data := removedUnreferenced(used, m.Float1Attributes(), func(s string) *iter.ArrayIterator[float64] { return m.Float1Attribute(s) })

	finalIndices := make([]int, originalIndices.Len())
	for triI := 0; triI < len(finalIndices); triI++ {
		finalIndices[triI] = originalIndices.At(triI) - shiftBy[originalIndices.At(triI)]
	}

	return modeling.
		NewMesh(m.Topology(), finalIndices).
		SetFloat4Data(finalV4Data).
		SetFloat3Data(finalV3Data).
		SetFloat2Data(finalV2Data).
		SetFloat1Data(finalV1Data).
		SetMaterials(m.Materials())
}
