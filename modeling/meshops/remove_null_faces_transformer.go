package meshops

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
)

type RemoveNullFaces3DTransformer struct {
	Attribute string
	MinArea   float64
}

func (rnft RemoveNullFaces3DTransformer) attribute() string {
	return rnft.Attribute
}

func (rnft RemoveNullFaces3DTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(rnft, modeling.PositionAttribute)

	if err = RequireV3Attribute(m, attribute); err != nil {
		return
	}

	if err = RequireTopology(m, modeling.TriangleTopology); err != nil {
		return
	}

	return RemoveNullFaces3D(m, attribute, rnft.MinArea), nil
}

func RemoveNullFaces3D(m modeling.Mesh, attribute string, minArea float64) modeling.Mesh {
	if err := RequireTopology(m, modeling.TriangleTopology); err != nil {
		panic(err)
	}

	if err := RequireV3Attribute(m, attribute); err != nil {
		panic(err)
	}

	indices := m.Indices()
	trisToKeep := make([]int, 0)
	for i := 0; i < m.PrimitiveCount(); i++ {
		tri := m.Tri(i)
		area := tri.Area3D(attribute)
		if !math.IsNaN(area) && area > minArea {
			trisToKeep = append(
				trisToKeep,
				tri.P1(),
				tri.P2(),
				tri.P3(),
			)
		}
	}

	// nothing to remove, just return the mesh passed in
	if len(trisToKeep) == indices.Len() {
		return m
	}

	return RemovedUnreferencedVertices(m.SetIndices(trisToKeep))
}
