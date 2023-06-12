package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func Test_SplitOnUniqueMaterials_Simple(t *testing.T) {
	// ARRANGE ================================================================
	m := modeling.NewTriangleMesh(
		[]int{
			0, 1, 2,
			3, 4, 5,
		},
	).SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
		vector3.New(0., 0., 0.),
		vector3.New(0., 1., 0.),
		vector3.New(1., 1., 0.),

		vector3.New(0., 0., 0.),
		vector3.New(1., 1., 0.),
		vector3.New(1., 0., 0.),
	}).SetFloat3Attribute(modeling.NormalAttribute, []vector3.Float64{
		vector3.New(0., 0., 0.),
		vector3.New(0., 1., 0.),
		vector3.New(1., 1., 0.),

		vector3.New(0., 0., 0.),
		vector3.New(1., 1., 0.),
		vector3.New(1., 0., 0.),
	}).SetFloat2Attribute(modeling.TexCoordAttribute, []vector2.Float64{
		vector2.New(0., 0.),
		vector2.New(0., 1.),
		vector2.New(1., 1.),

		vector2.New(0., 0.),
		vector2.New(1., 1.),
		vector2.New(1., 0.),
	}).SetMaterials([]modeling.MeshMaterial{
		{
			PrimitiveCount: 1,
			Material: &modeling.Material{
				Name: "red",
			},
		},
		{
			PrimitiveCount: 1,
			Material: &modeling.Material{
				Name: "blue",
			},
		},
	})

	// ACT ====================================================================
	meshes := meshops.SplitOnUniqueMaterials(m)

	// ASSERT =================================================================
	if !assert.Len(t, meshes, 2) {
		return
	}

	mesh1 := meshes[0]
	mesh1Indices := mesh1.Indices()
	if assert.Equal(t, 3, mesh1Indices.Len()) {
		assert.Equal(t, 1, mesh1Indices.At(1))
		assert.Equal(t, 0, mesh1Indices.At(0))
		assert.Equal(t, 2, mesh1Indices.At(2))
	}

	v1Verts := mesh1.Float3Attribute(modeling.PositionAttribute)
	if assert.Equal(t, v1Verts.Len(), 3) {
		assert.Equal(t, vector3.New[float64](0, 0, 0), v1Verts.At(0))
		assert.Equal(t, vector3.New[float64](0, 1, 0), v1Verts.At(1))
		assert.Equal(t, vector3.New[float64](1, 1, 0), v1Verts.At(2))
	}

	v1UVs := mesh1.Float2Attribute(modeling.TexCoordAttribute)
	if assert.Equal(t, v1UVs.Len(), 3) {
		assert.Equal(t, vector2.New[float64](0, 0), v1UVs.At(0))
		assert.Equal(t, vector2.New[float64](0, 1), v1UVs.At(1))
		assert.Equal(t, vector2.New[float64](1, 1), v1UVs.At(2))
	}

	mesh2 := meshes[1]
	mesh2Indices := mesh1.Indices()
	if assert.Equal(t, mesh2Indices.Len(), 3) {
		assert.Equal(t, 0, mesh2Indices.At(0))
		assert.Equal(t, 1, mesh2Indices.At(1))
		assert.Equal(t, 2, mesh2Indices.At(2))
	}

	v2Verts := mesh2.Float3Attribute(modeling.PositionAttribute)
	if assert.Equal(t, v2Verts.Len(), 3) {
		assert.Equal(t, vector3.New[float64](0, 0, 0), v2Verts.At(0))
		assert.Equal(t, vector3.New[float64](1, 1, 0), v2Verts.At(1))
		assert.Equal(t, vector3.New[float64](1, 0, 0), v2Verts.At(2))
	}

	v2UVs := mesh2.Float2Attribute(modeling.TexCoordAttribute)
	if assert.Equal(t, v2UVs.Len(), 3) {
		assert.Equal(t, vector2.New[float64](0, 0), v2UVs.At(0))
		assert.Equal(t, vector2.New[float64](1, 1), v2UVs.At(1))
		assert.Equal(t, vector2.New[float64](1, 0), v2UVs.At(2))
	}
}
