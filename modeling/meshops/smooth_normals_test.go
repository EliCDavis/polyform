package meshops_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/csg"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmoothNormalsAtABoxCornerIgnoreFaceSizeAndTriangulation(t *testing.T) {
	mesh := modeling.NewTriangleMesh([]int{
		0, 1, 2,
		0, 2, 3,
		0, 4, 5,
		0, 6, 1,
	}).SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
		vector3.New(0., 0., 0.),
		vector3.New(3., 0., 0.),
		vector3.New(1., 1., 0.),
		vector3.New(0., 1., 0.),
		vector3.New(0., 2., 0.),
		vector3.New(0., 0., 1.),
		vector3.New(0., 0., 5.),
	})

	normal := meshops.SmoothNormals(mesh).Float3Attribute(modeling.NormalAttribute).At(0)

	want := vector3.New(1., 1., 1.).Normalized()
	assert.InDelta(t, want.X(), normal.X(), 1e-9)
	assert.InDelta(t, want.Y(), normal.Y(), 1e-9)
	assert.InDelta(t, want.Z(), normal.Z(), 1e-9)
}

func TestSmoothNormalsImplicitWeldAtAnUnweldedBoxCornerIgnoreFaceSizeAndTriangulation(t *testing.T) {
	mesh := modeling.NewTriangleMesh([]int{
		0, 1, 2,
		0, 2, 3,
		4, 5, 6,
		7, 8, 9,
	}).SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
		vector3.New(0., 0., 0.),
		vector3.New(3., 0., 0.),
		vector3.New(1., 1., 0.),
		vector3.New(0., 1., 0.),
		vector3.New(0., 0., 0.),
		vector3.New(0., 2., 0.),
		vector3.New(0., 0., 1.),
		vector3.New(0., 0., 0.),
		vector3.New(0., 0., 5.),
		vector3.New(3., 0., 0.),
	})

	normals := meshops.SmoothNormalsImplicitWeld(mesh, 1e-6).Float3Attribute(modeling.NormalAttribute)

	want := vector3.New(1., 1., 1.).Normalized()
	for _, corner := range []int{0, 4, 7} {
		normal := normals.At(corner)
		assert.InDelta(t, want.X(), normal.X(), 1e-9)
		assert.InDelta(t, want.Y(), normal.Y(), 1e-9)
		assert.InDelta(t, want.Z(), normal.Z(), 1e-9)
	}
}

func assertFlatFaces(t *testing.T, m modeling.Mesh) {
	t.Helper()
	normals := m.Float3Attribute(modeling.NormalAttribute)
	for f := 0; f < m.PrimitiveCount(); f++ {
		tri := m.Tri(f)
		want := tri.Normal(modeling.PositionAttribute).Normalized()
		for _, v := range []int{tri.P1(), tri.P2(), tri.P3()} {
			assert.InDelta(t, 0., normals.At(v).Distance(want), 1e-9, "face %d vertex %d", f, v)
		}
	}
}

func TestSmoothNormalsByAngleSplitsAWeldedCubeAlongItsEdges(t *testing.T) {
	cube := primitives.Cube{Width: 2, Height: 2, Depth: 2}.Welded()
	require.Equal(t, 8, cube.AttributeLength())

	smoothed := meshops.SmoothNormalsByAngle(cube, 45)

	assert.Equal(t, 24, smoothed.AttributeLength())
	assert.Equal(t, cube.PrimitiveCount(), smoothed.PrimitiveCount())
	assertFlatFaces(t, smoothed)
}

func TestSmoothNormalsByAngleNeverMergesVertices(t *testing.T) {
	unwelded := meshops.Unweld(primitives.Cube{Width: 2, Height: 2, Depth: 2}.Welded())

	smoothed := meshops.SmoothNormalsByAngle(unwelded, 45)

	assert.Equal(t, unwelded.AttributeLength(), smoothed.AttributeLength())
	assertFlatFaces(t, smoothed)
}

func TestSmoothNormalsByAngleSmoothsAcrossVerticesThatOnlyShareAPosition(t *testing.T) {
	unwelded := meshops.Unweld(primitives.Cube{Width: 2, Height: 2, Depth: 2}.Welded())

	smoothed := meshops.SmoothNormalsByAngle(unwelded, 180)

	positions := smoothed.Float3Attribute(modeling.PositionAttribute)
	normals := smoothed.Float3Attribute(modeling.NormalAttribute)
	for i := 0; i < smoothed.AttributeLength(); i++ {
		assert.InDelta(t, 0., normals.At(i).Distance(positions.At(i).Normalized()), 1e-9)
	}
}

func TestSmoothNormalsByAngleLeavesACurvedSurfaceWhole(t *testing.T) {
	sphere := marching.March(sdf.Sphere(vector3.Zero[float64](), 1), geometry.NewAABB(vector3.Zero[float64](), vector3.Fill(2.6)), 1./16, 0)

	smoothed := meshops.SmoothNormalsByAngle(sphere, 45)

	require.Equal(t, sphere.AttributeLength(), smoothed.AttributeLength())
	want := meshops.SmoothNormals(sphere).Float3Attribute(modeling.NormalAttribute)
	got := smoothed.Float3Attribute(modeling.NormalAttribute)
	for i := 0; i < want.Len(); i++ {
		assert.InDelta(t, 0., got.At(i).Distance(want.At(i)), 1e-9)
	}
}

func TestSmoothNormalsByAngleKeepsEveryAttributeOnSplitVertices(t *testing.T) {
	cube := primitives.Cube{Width: 2, Height: 2, Depth: 2}.Welded()
	weights := make([]float64, cube.AttributeLength())
	for i := range weights {
		weights[i] = float64(i)
	}
	cube = cube.SetFloat1Attribute("Weight", weights)

	smoothed := meshops.SmoothNormalsByAngle(cube, 45)

	for f := 0; f < cube.PrimitiveCount(); f++ {
		before, after := cube.Tri(f), smoothed.Tri(f)
		assert.Equal(t, cube.Float1Attribute("Weight").At(before.P1()), smoothed.Float1Attribute("Weight").At(after.P1()))
		assert.Equal(t, cube.Float3Attribute(modeling.PositionAttribute).At(before.P2()), smoothed.Float3Attribute(modeling.PositionAttribute).At(after.P2()))
	}
}

func TestSmoothNormalsByAngleGivesNoNaNToADegenerateFace(t *testing.T) {
	mesh := modeling.NewTriangleMesh([]int{0, 1, 2, 3, 4, 5}).SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
		vector3.New(0., 0., 0.), vector3.New(1., 0., 0.), vector3.New(0., 1., 0.),
		vector3.New(0., 0., 0.), vector3.New(1., 0., 0.), vector3.New(2., 0., 0.),
	})

	normals := meshops.SmoothNormalsByAngle(mesh, 45).Float3Attribute(modeling.NormalAttribute)

	for i := 0; i < normals.Len(); i++ {
		assert.False(t, normals.At(i).ContainsNaN(), "vertex %d", i)
	}
	assert.InDelta(t, 0., normals.At(3).Distance(vector3.New(0., 0., 1.)), 1e-9)
}

func TestSmoothNormalsNodeKeepsTodaysOutputWithoutAnAngle(t *testing.T) {
	cube := primitives.Cube{Width: 2, Height: 2, Depth: 2}.Welded()
	plain := &nodes.Struct[meshops.SmoothNormalsNode]{
		Data: meshops.SmoothNormalsNode{Mesh: nodes.ConstOutput[modeling.Mesh]{Val: cube}},
	}
	angled := &nodes.Struct[meshops.SmoothNormalsNode]{
		Data: meshops.SmoothNormalsNode{
			Mesh:           nodes.ConstOutput[modeling.Mesh]{Val: cube},
			SmoothingAngle: nodes.ConstOutput[float64]{Val: 45},
		},
	}

	assert.Equal(t, 8, nodes.GetNodeOutputPort[modeling.Mesh](plain, "Out").Value().AttributeLength())
	assert.Equal(t, 24, nodes.GetNodeOutputPort[modeling.Mesh](angled, "Out").Value().AttributeLength())
}

func TestSmoothNormalsByAngleThenWeldIdenticalKeepsACSGCutCrisp(t *testing.T) {
	center := vector3.New(0., 0., 1.)
	box := primitives.Cube{Width: 2, Height: 2, Depth: 2}.UnweldedQuads()
	ball := primitives.UVSphere(0.6, 24, 32).Translate(center)
	carved, err := csg.Subtract(box, ball)
	require.NoError(t, err)

	smoothed := meshops.SmoothNormalsByAngle(carved, 45)
	welded := meshops.WeldIdentical(smoothed, modeling.PositionAttribute, 0)

	assert.Less(t, welded.AttributeLength(), carved.AttributeLength()/2)
	positions := welded.Float3Attribute(modeling.PositionAttribute)
	normals := welded.Float3Attribute(modeling.NormalAttribute)
	topFaces, dimpleFaces := 0, 0
	for f := 0; f < welded.PrimitiveCount(); f++ {
		tri := welded.Tri(f)
		corners := []int{tri.P1(), tri.P2(), tri.P3()}
		onTop, onBall := true, true
		for _, v := range corners {
			onTop = onTop && math.Abs(positions.At(v).Z()-1) < 1e-9
			onBall = onBall && math.Abs(positions.At(v).Distance(center)-0.6) < 0.01
		}
		switch {
		case onTop:
			topFaces++
			for _, v := range corners {
				assert.InDelta(t, 0., normals.At(v).Distance(vector3.New(0., 0., 1.)), 1e-9, "flat top at %v", positions.At(v))
			}
		case onBall:
			dimpleFaces++
			for _, v := range corners {
				if positions.At(v).Z() > 1-1e-6 {
					continue
				}
				inward := center.Sub(positions.At(v)).Normalized()
				assert.Greater(t, normals.At(v).Dot(inward), math.Cos(1*math.Pi/180), "dimple at %v", positions.At(v))
			}
		}
	}
	assert.NotZero(t, topFaces)
	assert.NotZero(t, dimpleFaces)
}
