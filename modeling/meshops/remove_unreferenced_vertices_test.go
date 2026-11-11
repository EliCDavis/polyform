package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
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

func TestRemoveUnreferencedVerticesRemapsTriangles(t *testing.T) {
	m := modeling.NewTriangleMesh([]int{4, 2, 0, 0, 2, 5}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
			vector3.New(0., 0, 0), vector3.New(1., 0, 0), vector3.New(2., 0, 0),
			vector3.New(3., 0, 0), vector3.New(4., 0, 0), vector3.New(5., 0, 0),
		}).
		SetFloat2Attribute(modeling.TexCoordAttribute, []vector2.Float64{
			vector2.New(0., 0), vector2.New(1., 0), vector2.New(2., 0),
			vector2.New(3., 0), vector2.New(4., 0), vector2.New(5., 0),
		}).
		SetFloat1Attribute("weight", []float64{0, 1, 2, 3, 4, 5})

	cleaned := meshops.RemovedUnreferencedVertices(m)

	assert.Equal(t, modeling.TriangleTopology, cleaned.Topology())
	assert.Equal(t, 4, cleaned.AttributeLength())
	indices := cleaned.Indices()
	got := make([]int, indices.Len())
	for i := range got {
		got[i] = indices.At(i)
	}
	assert.Equal(t, []int{2, 1, 0, 0, 1, 3}, got)
	for k, original := range []float64{0, 2, 4, 5} {
		assert.Equal(t, original, cleaned.Float3Attribute(modeling.PositionAttribute).At(k).X())
		assert.Equal(t, original, cleaned.Float2Attribute(modeling.TexCoordAttribute).At(k).X())
		assert.Equal(t, original, cleaned.Float1Attribute("weight").At(k))
	}
}

func TestRemoveUnreferencedVerticesWithNoIndicesDropsEveryAttribute(t *testing.T) {
	m := modeling.NewMesh(modeling.PointTopology, []int{}).SetFloat1Attribute("test", []float64{1, 2})

	cleaned := meshops.RemovedUnreferencedVertices(m)

	assert.Equal(t, 0, cleaned.AttributeLength())
	assert.False(t, cleaned.HasFloat1Attribute("test"))
}
