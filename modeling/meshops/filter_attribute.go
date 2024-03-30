package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type FilterFloat1Transformer struct {
	Attribute string
	Filter    func(v float64) bool
}

func (fft FilterFloat1Transformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	if err = RequireV1Attribute(m, fft.Attribute); err != nil {
		return
	}

	return FilterFloat1(m, fft.Attribute, fft.Filter), nil
}

func FilterFloat1(m modeling.Mesh, attribute string, filter func(v float64) bool) modeling.Mesh {
	check(RequireV1Attribute(m, attribute))

	vertices := m.Float1Attribute(attribute)
	verticeToKeep := make(map[int]struct{}, 0)

	for i := 0; i < vertices.Len(); i++ {
		if filter(vertices.At(i)) {
			verticeToKeep[i] = struct{}{}
		}
	}

	indices := m.Indices()
	finalIndices := make([]int, 0)
	for i := 0; i < indices.Len(); i++ {
		if _, ok := verticeToKeep[indices.At(i)]; ok {
			finalIndices = append(finalIndices, i)
		}
	}

	return RemovedUnreferencedVertices(m.SetIndices(finalIndices))
}

// FLOAT 2 ====================================================================

type FilterFloat2Transformer struct {
	Attribute string
	Filter    func(v vector2.Float64) bool
}

func (fft FilterFloat2Transformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	if err = RequireV2Attribute(m, fft.Attribute); err != nil {
		return
	}

	return FilterFloat2(m, fft.Attribute, fft.Filter), nil
}

func FilterFloat2(m modeling.Mesh, attribute string, filter func(v vector2.Float64) bool) modeling.Mesh {
	check(RequireV2Attribute(m, attribute))

	vertices := m.Float2Attribute(attribute)
	verticeToKeep := make(map[int]struct{}, 0)

	for i := 0; i < vertices.Len(); i++ {
		if filter(vertices.At(i)) {
			verticeToKeep[i] = struct{}{}
		}
	}

	indices := m.Indices()
	finalIndices := make([]int, 0)
	for i := 0; i < indices.Len(); i++ {
		if _, ok := verticeToKeep[indices.At(i)]; ok {
			finalIndices = append(finalIndices, i)
		}
	}

	return RemovedUnreferencedVertices(m.SetIndices(finalIndices))
}

// FLOAT 3 ====================================================================

type FilterFloat3Transformer struct {
	Attribute string
	Filter    func(v vector3.Float64) bool
}

func (fft FilterFloat3Transformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	if err = RequireV3Attribute(m, fft.Attribute); err != nil {
		return
	}

	return FilterFloat3(m, fft.Attribute, fft.Filter), nil
}

func FilterFloat3(m modeling.Mesh, attribute string, filter func(v vector3.Float64) bool) modeling.Mesh {
	check(RequireV3Attribute(m, attribute))

	vertices := m.Float3Attribute(attribute)
	verticeToKeep := make(map[int]struct{}, 0)

	for i := 0; i < vertices.Len(); i++ {
		if filter(vertices.At(i)) {
			verticeToKeep[i] = struct{}{}
		}
	}

	indices := m.Indices()
	finalIndices := make([]int, 0)
	for i := 0; i < indices.Len(); i++ {
		if _, ok := verticeToKeep[indices.At(i)]; ok {
			finalIndices = append(finalIndices, i)
		}
	}

	return RemovedUnreferencedVertices(m.SetIndices(finalIndices))
}

// FLOAT 4 ====================================================================

type FilterFloat4Transformer struct {
	Attribute string
	Filter    func(v vector4.Float64) bool
}

func (fft FilterFloat4Transformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	if err = RequireV4Attribute(m, fft.Attribute); err != nil {
		return
	}

	return FilterFloat4(m, fft.Attribute, fft.Filter), nil
}

func FilterFloat4(m modeling.Mesh, attribute string, filter func(v vector4.Float64) bool) modeling.Mesh {
	check(RequireV4Attribute(m, attribute))

	vertices := m.Float4Attribute(attribute)
	verticeToKeep := make(map[int]struct{}, 0)

	for i := 0; i < vertices.Len(); i++ {
		if filter(vertices.At(i)) {
			verticeToKeep[i] = struct{}{}
		}
	}

	indices := m.Indices()
	finalIndices := make([]int, 0)
	for i := 0; i < indices.Len(); i++ {
		if _, ok := verticeToKeep[indices.At(i)]; ok {
			finalIndices = append(finalIndices, i)
		}
	}

	return RemovedUnreferencedVertices(m.SetIndices(finalIndices))
}
