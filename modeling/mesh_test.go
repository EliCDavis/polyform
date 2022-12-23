package modeling_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
	"github.com/stretchr/testify/assert"
)

func Test_SplitOnUniqueMaterials_Simple(t *testing.T) {
	// ARRANGE ================================================================
	m := modeling.NewMeshWithMaterials(
		[]int{
			0, 1, 2,
			3, 4, 5,
		},
		[]vector.Vector3{
			vector.NewVector3(0, 0, 0),
			vector.NewVector3(0, 1, 0),
			vector.NewVector3(1, 1, 0),
			vector.NewVector3(0, 0, 0),
			vector.NewVector3(1, 1, 0),
			vector.NewVector3(1, 0, 0),
		},
		[]vector.Vector3{
			vector.NewVector3(0, 0, 0),
			vector.NewVector3(0, 1, 0),
			vector.NewVector3(1, 1, 0),
			vector.NewVector3(0, 0, 0),
			vector.NewVector3(1, 1, 0),
			vector.NewVector3(1, 0, 0),
		},
		[][]vector.Vector2{
			{
				vector.NewVector2(0, 0),
				vector.NewVector2(0, 1),
				vector.NewVector2(1, 1),
				vector.NewVector2(0, 0),
				vector.NewVector2(1, 1),
				vector.NewVector2(1, 0),
			},
		},
		[]modeling.MeshMaterial{
			{
				NumOfTris: 1,
				Material: &modeling.Material{
					Name: "red",
				},
			},
			{
				NumOfTris: 1,
				Material: &modeling.Material{
					Name: "blue",
				},
			},
		},
	)

	// ACT ====================================================================
	meshes := m.SplitOnUniqueMaterials()

	// ASSERT =================================================================
	if !assert.Len(t, meshes, 2) {
		return
	}

	v1 := meshes[0].View()
	if assert.Len(t, v1.Triangles, 3) {
		assert.Equal(t, 0, v1.Triangles[0])
		assert.Equal(t, 1, v1.Triangles[1])
		assert.Equal(t, 2, v1.Triangles[2])
	}
	if assert.Len(t, v1.Vertices, 3) {
		assert.Equal(t, vector.NewVector3(0, 0, 0), v1.Vertices[0])
		assert.Equal(t, vector.NewVector3(0, 1, 0), v1.Vertices[1])
		assert.Equal(t, vector.NewVector3(1, 1, 0), v1.Vertices[2])
	}
	if assert.Len(t, v1.UVs[0], 3) {
		assert.Equal(t, vector.NewVector2(0, 0), v1.UVs[0][0])
		assert.Equal(t, vector.NewVector2(0, 1), v1.UVs[0][1])
		assert.Equal(t, vector.NewVector2(1, 1), v1.UVs[0][2])
	}

	v2 := meshes[1].View()
	if assert.Len(t, v2.Triangles, 3) {
		assert.Equal(t, 0, v2.Triangles[0])
		assert.Equal(t, 1, v2.Triangles[1])
		assert.Equal(t, 2, v2.Triangles[2])
	}
	if assert.Len(t, v2.Vertices, 3) {
		assert.Equal(t, vector.NewVector3(0, 0, 0), v2.Vertices[0])
		assert.Equal(t, vector.NewVector3(1, 1, 0), v2.Vertices[1])
		assert.Equal(t, vector.NewVector3(1, 0, 0), v2.Vertices[2])
	}
	if assert.Len(t, v2.UVs[0], 3) {
		assert.Equal(t, vector.NewVector2(0, 0), v2.UVs[0][0])
		assert.Equal(t, vector.NewVector2(1, 1), v2.UVs[0][1])
		assert.Equal(t, vector.NewVector2(1, 0), v2.UVs[0][2])
	}
}
