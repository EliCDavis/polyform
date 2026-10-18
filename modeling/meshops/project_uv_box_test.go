package meshops_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Eight shared corners, so every corner sits on three faces facing different ways.
func weldedBox() modeling.Mesh {
	corners := []vector3.Float64{
		vector3.New(-1., -1., -1.), vector3.New(1., -1., -1.), vector3.New(1., 1., -1.), vector3.New(-1., 1., -1.),
		vector3.New(-1., -1., 1.), vector3.New(1., -1., 1.), vector3.New(1., 1., 1.), vector3.New(-1., 1., 1.),
	}
	indices := []int{
		4, 5, 6, 4, 6, 7, // +Z
		1, 0, 3, 1, 3, 2, // -Z
		5, 1, 2, 5, 2, 6, // +X
		0, 4, 7, 0, 7, 3, // -X
		7, 6, 2, 7, 2, 3, // +Y
		0, 1, 5, 0, 5, 4, // -Y
	}
	return modeling.NewTriangleMesh(indices).SetFloat3Attribute(modeling.PositionAttribute, corners)
}

func zeroAreaUVTriangles(t *testing.T, m modeling.Mesh) int {
	t.Helper()
	require.True(t, m.HasFloat2Attribute(modeling.TexCoordAttribute))
	zero := 0
	for i := 0; i < m.PrimitiveCount(); i++ {
		tri := m.Tri(i)
		a := tri.P1Vec2Attr(modeling.TexCoordAttribute)
		ab := tri.P2Vec2Attr(modeling.TexCoordAttribute).Sub(a)
		ac := tri.P3Vec2Attr(modeling.TexCoordAttribute).Sub(a)
		if math.Abs(ab.X()*ac.Y()-ab.Y()*ac.X()) < 1e-12 {
			zero++
		}
	}
	return zero
}

func TestBoxProjectUVLeavesNoFaceEdgeOn(t *testing.T) {
	box := weldedBox()
	flat := meshops.ProjectUV(box, vector3.Forward[float64](), vector2.One[float64](), vector2.Zero[float64](), 0)
	require.Equal(t, 8, zeroAreaUVTriangles(t, flat), "a planar projection flattens the four sides it looks along")

	projected, err := meshops.BoxProjectUV(box, vector2.New(2., 2.), vector2.Zero[float64]())
	require.NoError(t, err)

	assert.Equal(t, 12, projected.PrimitiveCount())
	assert.Equal(t, 24, projected.AttributeLength(), "each corner splits once per face direction it touches")
	assert.Zero(t, zeroAreaUVTriangles(t, projected))
}

func TestBoxProjectUVReadsEachFaceUnmirroredFromOutside(t *testing.T) {
	projected, err := meshops.BoxProjectUV(weldedBox(), vector2.One[float64](), vector2.Zero[float64]())
	require.NoError(t, err)

	for i := 0; i < projected.PrimitiveCount(); i++ {
		tri := projected.Tri(i)
		a, b, c := tri.P1Vec3Attr(modeling.PositionAttribute), tri.P2Vec3Attr(modeling.PositionAttribute), tri.P3Vec3Attr(modeling.PositionAttribute)
		ua := tri.P1Vec2Attr(modeling.TexCoordAttribute)
		ub := tri.P2Vec2Attr(modeling.TexCoordAttribute).Sub(ua)
		uc := tri.P3Vec2Attr(modeling.TexCoordAttribute).Sub(ua)
		require.NotZero(t, b.Sub(a).Cross(c.Sub(a)).Length())
		assert.Positive(t, ub.X()*uc.Y()-ub.Y()*uc.X(), "triangle %d winds the same way in UV as in space", i)
	}
}

func TestCylinderProjectUVWrapsTheSidesAndFlattensTheCaps(t *testing.T) {
	can := primitives.Cylinder{Sides: 24, Height: 2, Radius: 0.5}.ToMesh()
	require.Equal(t, modeling.TriangleTopology, can.Topology())

	projected, err := meshops.CylinderProjectUV(can, vector3.Zero[float64](), vector3.Up[float64](), 3, 2)
	require.NoError(t, err)
	assert.Zero(t, zeroAreaUVTriangles(t, projected))

	positions := projected.Float3Attribute(modeling.PositionAttribute)
	normals := projected.Float3Attribute(modeling.NormalAttribute)
	uvs := projected.Float2Attribute(modeling.TexCoordAttribute)
	maxU := 0.
	for i := 0; i < uvs.Len(); i++ {
		if math.Abs(normals.At(i).Y()) < 0.5 {
			maxU = math.Max(maxU, uvs.At(i).X())
			assert.InDelta(t, positions.At(i).Y()/2, uvs.At(i).Y(), 1e-9, "v runs along the axis, one repeat per Height")
		}
	}
	assert.InDelta(t, 3, maxU, 1e-9, "u reaches Around repeats at the seam")

	for i := 0; i < projected.PrimitiveCount(); i++ {
		tri := projected.Tri(i)
		us := []float64{tri.P1Vec2Attr(modeling.TexCoordAttribute).X(), tri.P2Vec2Attr(modeling.TexCoordAttribute).X(), tri.P3Vec2Attr(modeling.TexCoordAttribute).X()}
		spread := math.Max(us[0], math.Max(us[1], us[2])) - math.Min(us[0], math.Min(us[1], us[2]))
		assert.Less(t, spread, 1.5, "triangle %d does not stretch across the seam", i)
	}
}
