package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeldClosesAnUnweldedCube(t *testing.T) {
	split := primitives.Cube{Width: 2, Height: 2, Depth: 2}.UnweldedQuads()
	require.NotEmpty(t, split.BoundaryEdges(), "sanity: an unwelded cube is split along its seams")

	joined := meshops.Weld(split, modeling.PositionAttribute, 0.0001)

	assert.Equal(t, 8, joined.AttributeLength(), "a cube has eight distinct corners")
	assert.Equal(t, split.PrimitiveCount(), joined.PrimitiveCount(), "welding should not lose faces")
	assert.Empty(t, joined.BoundaryEdges())
}

// The reason this does not snap to a grid: 0.4999 and 0.5001 are a hair apart
// but sit either side of every round decimal between them.
func TestWeldMergesAcrossARoundNumber(t *testing.T) {
	m := modeling.NewTriangleMesh([]int{0, 1, 2, 3, 4, 5}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
			vector3.New(0.4999, 0., 0.),
			vector3.New(1., 0., 0.),
			vector3.New(0., 1., 0.),

			vector3.New(0.5001, 0., 0.),
			vector3.New(0., 1., 0.),
			vector3.New(0., 0., 1.),
		})

	joined := meshops.Weld(m, modeling.PositionAttribute, 0.001)

	assert.Equal(t, 4, joined.AttributeLength(),
		"the two near-identical points and the two shared corners should each become one")
	assert.Equal(t, 2, joined.PrimitiveCount())
}

func TestWeldLeavesDistinctPointsAlone(t *testing.T) {
	sphere := primitives.UVSphere(1, 12, 16)
	joined := meshops.Weld(sphere, modeling.PositionAttribute, 1e-9)

	assert.Equal(t, sphere.AttributeLength(), joined.AttributeLength())
	assert.Equal(t, sphere.PrimitiveCount(), joined.PrimitiveCount())
}

func TestWeldDropsFacesCollapsedToALine(t *testing.T) {
	m := modeling.NewTriangleMesh([]int{0, 1, 2, 3, 4, 5}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
			// Two corners within tolerance, so this one becomes a line.
			vector3.New(0., 0., 0.),
			vector3.New(0.00001, 0., 0.),
			vector3.New(0., 1., 0.),

			vector3.New(0., 0., 0.),
			vector3.New(1., 0., 0.),
			vector3.New(0., 0., 1.),
		})

	joined := meshops.Weld(m, modeling.PositionAttribute, 0.001)

	assert.Equal(t, 1, joined.PrimitiveCount(), "the collapsed face should be dropped")
	assert.Equal(t, 3, joined.AttributeLength(), "and its unused vertices with it")
}

func TestWeldCarriesEveryAttribute(t *testing.T) {
	split := primitives.Cube{Width: 2, Height: 2, Depth: 2}.UnweldedQuads()
	require.Contains(t, split.Float3Attributes(), modeling.NormalAttribute,
		"sanity: this fixture should carry normals to check they survive")

	joined := meshops.Weld(split, modeling.PositionAttribute, 0.0001)

	assert.ElementsMatch(t, split.Float3Attributes(), joined.Float3Attributes())
	assert.ElementsMatch(t, split.Float2Attributes(), joined.Float2Attributes())
	for _, attr := range joined.Float3Attributes() {
		assert.Equal(t, joined.AttributeLength(), joined.Float3Attribute(attr).Len(), attr)
	}
	for _, attr := range joined.Float2Attributes() {
		assert.Equal(t, joined.AttributeLength(), joined.Float2Attribute(attr).Len(), attr)
	}
}
