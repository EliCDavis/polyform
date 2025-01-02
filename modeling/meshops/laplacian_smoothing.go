package meshops

import (
	"fmt"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type LaplacianSmoothTransformer struct {
	Attribute       string
	Iterations      int
	SmoothingFactor float64
}

func (lst LaplacianSmoothTransformer) attribute() string {
	return lst.Attribute
}

func (lst LaplacianSmoothTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(lst, modeling.PositionAttribute)

	if err = RequireV3Attribute(m, attribute); err != nil {
		return
	}

	return LaplacianSmooth(m, attribute, lst.Iterations, lst.SmoothingFactor), nil
}

func LaplacianSmooth(m modeling.Mesh, attribute string, iterations int, smoothingFactor float64) modeling.Mesh {
	if !m.HasFloat3Attribute(attribute) {
		panic(fmt.Errorf("attempting to apply laplacian smoothing to a mesh without the attribute: %s", attribute))
	}

	lut := m.VertexNeighborTable()

	oldVertices := m.Float3Attribute(attribute)
	vertices := make([]vector3.Float64, oldVertices.Len())
	for i := range vertices {
		vertices[i] = oldVertices.At(i)
	}

	for i := 0; i < iterations; i++ {
		for vi, vertex := range vertices {
			var sum vector3.Float64

			for vn := range lut.Lookup(vi) {
				sum = sum.Add(vertices[vn])
			}

			vertices[vi] = vertex.Add(
				sum.
					DivByConstant(float64(lut.Count(vi))).
					Sub(vertex).
					Scale(smoothingFactor))
		}
	}

	return m.SetFloat3Attribute(attribute, vertices)
}

func LaplacianSmoothAlongAxis(m modeling.Mesh, attribute string, iterations int, smoothingFactor float64, axis vector3.Float64) modeling.Mesh {
	if !m.HasFloat3Attribute(attribute) {
		panic(fmt.Errorf("attempting to apply laplacian smoothing to a mesh without the attribute: %s", attribute))
	}

	cleanedAxis := axis.Normalized().Abs()

	lut := m.VertexNeighborTable()

	oldVertices := m.Float3Attribute(attribute)
	vertices := make([]vector3.Float64, oldVertices.Len())
	for i := range vertices {
		vertices[i] = oldVertices.At(i)
	}

	for i := 0; i < iterations; i++ {
		for vi, vertex := range vertices {
			var sum vector3.Float64

			for vn := range lut.Lookup(vi) {
				sum = sum.Add(vertices[vn])
			}

			vertices[vi] = vertex.Add(
				sum.
					DivByConstant(float64(lut.Count(vi))).
					Sub(vertex).
					Scale(smoothingFactor).
					MultByVector(cleanedAxis))
		}
	}

	return m.SetFloat3Attribute(attribute, vertices)
}

type LaplacianSmoothNode = nodes.Struct[modeling.Mesh, LaplacianSmoothNodeData]

type LaplacianSmoothNodeData struct {
	Mesh            nodes.NodeOutput[modeling.Mesh]
	Attribute       nodes.NodeOutput[string]
	Iterations      nodes.NodeOutput[int]
	SmoothingFactor nodes.NodeOutput[float64]
}

func (lp LaplacianSmoothNodeData) Process() (modeling.Mesh, error) {
	atrr := modeling.PositionAttribute
	if lp.Attribute != nil {
		atrr = lp.Attribute.Value()
	}

	return LaplacianSmooth(
		lp.Mesh.Value(),
		atrr,
		lp.Iterations.Value(),
		lp.SmoothingFactor.Value(),
	), nil
}
