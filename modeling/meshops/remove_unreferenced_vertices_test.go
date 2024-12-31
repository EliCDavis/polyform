package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/stretchr/testify/assert"
)

func TestRemoveUnreferencedVertices(t *testing.T) {
	// ARRANGE ================================================================
	m := modeling.NewMesh(modeling.PointTopology, []int{1}).SetFloat1Attribute("test", []float64{1, 2})

	// ACT ====================================================================
	cleanedMesh := meshops.RemovedUnreferencedVertices(m)

	// ASSERT =================================================================
	indices := cleanedMesh.Indices()
	assert.Equal(t, 1, indices.Len())
	assert.Equal(t, 0, indices.At(0))

	assert.Equal(t, 1, cleanedMesh.AttributeLength())
	assert.True(t, cleanedMesh.HasFloat1Attribute("test"))

	testAttr := cleanedMesh.Float1Attribute("test")
	assert.Equal(t, 1, testAttr.Len())
	assert.Equal(t, 2., testAttr.At(0))
}
