package meshops_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
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

type weldTestVertex struct {
	position vector3.Float64
	facing   vector3.Float64
	tangent  vector4.Float64
	patch    vector2.Float64
	weight   float64
}

func weldTestVertices(m modeling.Mesh) []weldTestVertex {
	vertices := make([]weldTestVertex, m.AttributeLength())
	for i := range vertices {
		vertices[i] = weldTestVertex{
			position: m.Float3Attribute(modeling.PositionAttribute).At(i),
			facing:   m.Float3Attribute("Facing").At(i),
			tangent:  m.Float4Attribute("Tangent").At(i),
			patch:    m.Float2Attribute("Patch").At(i),
			weight:   m.Float1Attribute("Weight").At(i),
		}
	}
	return vertices
}

func cubeWithPerFaceData() modeling.Mesh {
	cube := meshops.Unweld(primitives.Cube{Width: 2, Height: 2, Depth: 2}.Welded())
	n := cube.AttributeLength()
	facing := make([]vector3.Float64, n)
	tangent := make([]vector4.Float64, n)
	patch := make([]vector2.Float64, n)
	weight := make([]float64, n)
	for f := 0; f < cube.PrimitiveCount(); f++ {
		tri := cube.Tri(f)
		normal := tri.Normal(modeling.PositionAttribute).Normalized()
		for _, v := range []int{tri.P1(), tri.P2(), tri.P3()} {
			facing[v] = normal
			tangent[v] = vector4.New(normal.Y(), normal.Z(), normal.X(), 1.)
			patch[v] = vector2.New(normal.X()+normal.Y(), normal.Z())
			weight[v] = 2*normal.X() + 3*normal.Y() + 5*normal.Z()
		}
	}
	return cube.
		SetFloat3Attribute("Facing", facing).
		SetFloat4Attribute("Tangent", tangent).
		SetFloat2Attribute("Patch", patch).
		SetFloat1Attribute("Weight", weight)
}

func TestWeldIdenticalKeepsEveryVertexsData(t *testing.T) {
	cube := cubeWithPerFaceData()
	require.Equal(t, 36, cube.AttributeLength())

	joined := meshops.WeldIdentical(cube, modeling.PositionAttribute, 0)

	assert.Equal(t, 24, joined.AttributeLength(), "four corners per face, none shared between faces")
	assert.Equal(t, cube.PrimitiveCount(), joined.PrimitiveCount())
	assert.ElementsMatch(t, uniqueWeldTestVertices(cube), weldTestVertices(joined))
}

func uniqueWeldTestVertices(m modeling.Mesh) []weldTestVertex {
	seen := map[weldTestVertex]bool{}
	var unique []weldTestVertex
	for _, v := range weldTestVertices(m) {
		if !seen[v] {
			seen[v] = true
			unique = append(unique, v)
		}
	}
	return unique
}

func TestWeldIdenticalMergesNearPointsOnlyWhenTheirDataMatches(t *testing.T) {
	positions := []vector3.Float64{
		vector3.New(0.4999, 0., 0.), vector3.New(1., 0., 0.), vector3.New(0., 1., 0.),
		vector3.New(0.5001, 0., 0.), vector3.New(0., 1., 0.), vector3.New(0., 0., 1.),
	}
	m := modeling.NewTriangleMesh([]int{0, 1, 2, 3, 4, 5}).
		SetFloat3Attribute(modeling.PositionAttribute, positions)

	same := meshops.WeldIdentical(m.SetFloat1Attribute("Weight", []float64{1, 2, 3, 1, 3, 4}), modeling.PositionAttribute, 0.001)
	different := meshops.WeldIdentical(m.SetFloat1Attribute("Weight", []float64{1, 2, 3, 9, 3, 4}), modeling.PositionAttribute, 0.001)

	assert.Equal(t, 4, same.AttributeLength())
	assert.Equal(t, 5, different.AttributeLength())
}

func TestWeldIdenticalNeverMergesNaN(t *testing.T) {
	m := modeling.NewTriangleMesh([]int{0, 1, 2, 3, 4, 5}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
			vector3.New(0., 0., 0.), vector3.New(1., 0., 0.), vector3.New(0., 1., 0.),
			vector3.New(0., 0., 0.), vector3.New(0., 1., 0.), vector3.New(0., 0., 1.),
		}).
		SetFloat1Attribute("Weight", []float64{math.NaN(), 0, 0, math.NaN(), 0, 0})

	joined := meshops.WeldIdentical(m, modeling.PositionAttribute, 0)

	assert.Equal(t, 5, joined.AttributeLength())
}

func TestWeldNodeMatchAttributesKeepsPerFaceData(t *testing.T) {
	cube := cubeWithPerFaceData()
	node := &nodes.Struct[meshops.WeldNode]{
		Data: meshops.WeldNode{
			Mesh:            nodes.ConstOutput[modeling.Mesh]{Val: cube},
			MatchAttributes: nodes.ConstOutput[bool]{Val: true},
		},
	}

	out := nodes.GetNodeOutputPort[modeling.Mesh](node, "Out").Value()

	assert.Equal(t, 24, out.AttributeLength())
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
