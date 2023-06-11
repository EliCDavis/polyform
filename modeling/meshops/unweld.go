package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type UnweldTransformer struct{}

func (fnt UnweldTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	return Unweld(m), nil
}

// Unweld duplicates all vertex data such that no two primitive indices share
// any one vertex
func Unweld(m modeling.Mesh) modeling.Mesh {
	unweldedV4Data := make(map[string][]vector4.Float64)
	for _, attr := range m.Float4Attributes() {
		unweldedV4Data[attr] = make([]vector4.Float64, 0)
	}

	unweldedV3Data := make(map[string][]vector3.Float64)
	for _, attr := range m.Float3Attributes() {
		unweldedV3Data[attr] = make([]vector3.Float64, 0)
	}

	unweldedV2Data := make(map[string][]vector2.Float64)
	for _, attr := range m.Float2Attributes() {
		unweldedV2Data[attr] = make([]vector2.Float64, 0)
	}

	unweldedV1Data := make(map[string][]float64)
	for _, attr := range m.Float1Attributes() {
		unweldedV1Data[attr] = make([]float64, 0)
	}

	originalIndices := m.Indices()
	indices := make([]int, originalIndices.Len())
	for i := 0; i < len(indices); i++ {
		indices[i] = i
		for _, atr := range m.Float4Attributes() {
			data := m.Float4Attribute(atr)
			unweldedV4Data[atr] = append(unweldedV4Data[atr], data.At(originalIndices.At(i)))
		}

		for _, atr := range m.Float3Attributes() {
			data := m.Float3Attribute(atr)
			unweldedV3Data[atr] = append(unweldedV3Data[atr], data.At(originalIndices.At(i)))
		}

		for _, atr := range m.Float2Attributes() {
			data := m.Float2Attribute(atr)
			unweldedV2Data[atr] = append(unweldedV2Data[atr], data.At(originalIndices.At(i)))
		}

		for _, atr := range m.Float1Attributes() {
			data := m.Float1Attribute(atr)
			unweldedV1Data[atr] = append(unweldedV1Data[atr], data.At(originalIndices.At(i)))
		}
	}

	return modeling.
		NewMesh(m.Topology(), indices).
		SetFloat4Data(unweldedV4Data).
		SetFloat3Data(unweldedV3Data).
		SetFloat2Data(unweldedV2Data).
		SetFloat1Data(unweldedV1Data).
		SetMaterials(m.Materials())
}
