package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/stretchr/testify/assert"
)

func TestFlipTriangleWinding(t *testing.T) {
	// ARRANGE ================================================================
	mesh := modeling.NewTriangleMesh([]int{
		0, 1, 2,
		2, 3, 4,
	})
	op := meshops.FlipTriangleWindingTransformer{}

	// ACT ====================================================================
	flippedMesh := mesh.Transform(op)
	indices := flippedMesh.Indices()

	// ASSERT =================================================================
	assert.Equal(t, 1, indices.At(0))
	assert.Equal(t, 0, indices.At(1))
	assert.Equal(t, 2, indices.At(2))

	assert.Equal(t, 3, indices.At(3))
	assert.Equal(t, 2, indices.At(4))
	assert.Equal(t, 4, indices.At(5))
}
