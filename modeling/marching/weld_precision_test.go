package marching_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func boundaryEdgeCount(m modeling.Mesh) int {
	indices := m.Indices()
	directed := make(map[[2]int]int)
	for i := 0; i+3 <= indices.Len(); i += 3 {
		a, b, c := indices.At(i), indices.At(i+1), indices.At(i+2)
		directed[[2]int{a, b}]++
		directed[[2]int{b, c}]++
		directed[[2]int{c, a}]++
	}

	unpaired := 0
	for edge, count := range directed {
		if directed[[2]int{edge[1], edge[0]}] != count {
			unpaired++
		}
	}
	return unpaired
}

// Radius shrinks with resolution so every case marches the same voxel grid and
// must therefore produce the same mesh.
func marchSphereOnFixedGrid(cubesPerUnit float64) modeling.Mesh {
	const voxelsAcross = 30

	canvas := marching.NewMarchingCanvas(cubesPerUnit)
	canvas.AddField(marching.Sphere(
		vector3.Zero[float64](),
		voxelsAcross/(2*cubesPerUnit),
		1.,
	))
	return canvas.March(0)
}

func TestMarchingCanvasProducesGeometry(t *testing.T) {
	mesh := marchSphereOnFixedGrid(10)

	require.Greater(t, mesh.PrimitiveCount(), 0)
	assert.Zero(t, boundaryEdgeCount(mesh))
}

func TestMarchingCanvasWeldIsResolutionIndependent(t *testing.T) {
	coarse := marchSphereOnFixedGrid(10)
	require.Greater(t, coarse.PrimitiveCount(), 0)

	for _, cubesPerUnit := range []float64{100, 1000, 4000} {
		fine := marchSphereOnFixedGrid(cubesPerUnit)

		assert.Equal(t, coarse.PrimitiveCount(), fine.PrimitiveCount(),
			"%v cubes per unit", cubesPerUnit)
		assert.Equal(t, coarse.AttributeLength(), fine.AttributeLength(),
			"%v cubes per unit", cubesPerUnit)
		assert.Zero(t, boundaryEdgeCount(fine), "%v cubes per unit", cubesPerUnit)
	}
}
