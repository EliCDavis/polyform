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

type LaplacianSmoothNode struct {
	Mesh            nodes.Output[modeling.Mesh]
	Attribute       nodes.Output[string]
	Iterations      nodes.Output[int]
	SmoothingFactor nodes.Output[float64]
}

func (lp LaplacianSmoothNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if lp.Mesh == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	out.Set(LaplacianSmooth(
		nodes.GetOutputValue(out, lp.Mesh),
		nodes.TryGetOutputValue(out, lp.Attribute, modeling.PositionAttribute),
		nodes.TryGetOutputValue(out, lp.Iterations, 10),
		nodes.TryGetOutputValue(out, lp.SmoothingFactor, 0.1),
	))
}
