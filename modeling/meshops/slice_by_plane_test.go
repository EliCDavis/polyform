package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sliceTestTriangle() modeling.Mesh {
	positions := []vector3.Float64{
		vector3.New(0., 0., 0.),
		vector3.New(2., 0., 0.),
		vector3.New(0., 2., 0.),
	}
	colors := make([]vector3.Float64, len(positions))
	uvs := make([]vector2.Float64, len(positions))
	weights := make([]float64, len(positions))
	normals := make([]vector3.Float64, len(positions))
	for i, p := range positions {
		colors[i] = p.Scale(0.5)
		uvs[i] = vector2.New(p.X(), p.Y()).Scale(0.5)
		weights[i] = p.X() + p.Y()
		normals[i] = vector3.New(0., 0., 1.)
	}
	return modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(modeling.PositionAttribute, positions).
		SetFloat3Attribute(modeling.NormalAttribute, normals).
		SetFloat3Attribute(modeling.ColorAttribute, colors).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs).
		SetFloat1Attribute("Weight", weights)
}

func TestSliceByPlaneAboveIsTheSideTheNormalPoints(t *testing.T) {
	plane := geometry.NewPlane(vector3.New(0., 1., 0.), vector3.New(0., 1., 0.))

	above, below := meshops.SliceByPlaneWithAttribute(sliceTestTriangle(), plane, modeling.PositionAttribute)

	aboveBounds := above.BoundingBox(modeling.PositionAttribute)
	belowBounds := below.BoundingBox(modeling.PositionAttribute)
	assert.InDelta(t, 1., aboveBounds.Min().Y(), 1e-9)
	assert.InDelta(t, 2., aboveBounds.Max().Y(), 1e-9)
	assert.InDelta(t, 0., belowBounds.Min().Y(), 1e-9)
	assert.InDelta(t, 1., belowBounds.Max().Y(), 1e-9)
}

func TestSliceByPlaneInterpolatesEveryAttributeOntoTheCut(t *testing.T) {
	plane := geometry.NewPlane(vector3.New(0., 1., 0.), vector3.New(0., 1., 0.))

	above, below := meshops.SliceByPlaneWithAttribute(sliceTestTriangle(), plane, modeling.PositionAttribute)

	assert.Equal(t, 3, above.AttributeLength())
	assert.Equal(t, 4, below.AttributeLength())
	for _, piece := range []modeling.Mesh{above, below} {
		positions := piece.Float3Attribute(modeling.PositionAttribute)
		require.True(t, piece.HasFloat3Attribute(modeling.NormalAttribute))
		require.True(t, piece.HasFloat3Attribute(modeling.ColorAttribute))
		require.True(t, piece.HasFloat2Attribute(modeling.TexCoordAttribute))
		require.True(t, piece.HasFloat1Attribute("Weight"))
		for i := 0; i < positions.Len(); i++ {
			p := positions.At(i)
			assert.InDelta(t, 0., piece.Float3Attribute(modeling.ColorAttribute).At(i).Distance(p.Scale(0.5)), 1e-9)
			assert.InDelta(t, 0., piece.Float2Attribute(modeling.TexCoordAttribute).At(i).Distance(vector2.New(p.X(), p.Y()).Scale(0.5)), 1e-9)
			assert.InDelta(t, p.X()+p.Y(), piece.Float1Attribute("Weight").At(i), 1e-9)
			assert.InDelta(t, 0., piece.Float3Attribute(modeling.NormalAttribute).At(i).Distance(vector3.New(0., 0., 1.)), 1e-9)
		}
	}
}

func TestSliceByPlaneKeepsTheTriangleWinding(t *testing.T) {
	plane := geometry.NewPlane(vector3.New(0., 1., 0.), vector3.New(0., 1., 0.))

	above, below := meshops.SliceByPlaneWithAttribute(sliceTestTriangle(), plane, modeling.PositionAttribute)

	for _, piece := range []modeling.Mesh{above, below} {
		require.Greater(t, piece.PrimitiveCount(), 0)
		for i := 0; i < piece.PrimitiveCount(); i++ {
			assert.Greater(t, piece.Tri(i).Normal(modeling.PositionAttribute).Z(), 0.)
		}
	}
}

func TestSliceByPlaneOfAnEmptyMeshIsEmpty(t *testing.T) {
	plane := geometry.NewPlane(vector3.New(0., 1., 0.), vector3.New(0., 1., 0.))
	above, _ := meshops.SliceByPlaneWithAttribute(sliceTestTriangle(), geometry.NewPlane(vector3.New(0., 5., 0.), vector3.New(0., 1., 0.)), modeling.PositionAttribute)
	require.Zero(t, above.PrimitiveCount())

	var again, rest modeling.Mesh
	require.NotPanics(t, func() {
		again, rest = meshops.SliceByPlaneWithAttribute(above, plane, modeling.PositionAttribute)
	})
	assert.Zero(t, again.PrimitiveCount())
	assert.Zero(t, rest.PrimitiveCount())
}

func TestSliceByPlaneGivesAFaceOnThePlaneToTheSolidItBounds(t *testing.T) {
	cube := primitives.Cube{Width: 2, Height: 2, Depth: 2}.Welded()

	aboveTop, belowTop := meshops.SliceByPlaneWithAttribute(cube, geometry.NewPlane(vector3.New(0., 1., 0.), vector3.New(0., 1., 0.)), modeling.PositionAttribute)
	assert.Zero(t, aboveTop.PrimitiveCount())
	assert.Equal(t, cube.PrimitiveCount(), belowTop.PrimitiveCount())

	aboveBottom, belowBottom := meshops.SliceByPlaneWithAttribute(cube, geometry.NewPlane(vector3.New(0., -1., 0.), vector3.New(0., 1., 0.)), modeling.PositionAttribute)
	assert.Equal(t, cube.PrimitiveCount(), aboveBottom.PrimitiveCount())
	assert.Zero(t, belowBottom.PrimitiveCount())
}
