package modeling_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
	"github.com/stretchr/testify/assert"
)

func Test_SplitOnUniqueMaterials_Simple(t *testing.T) {
	// ARRANGE ================================================================
	m := modeling.NewMesh(
		[]int{
			0, 1, 2,
			3, 4, 5,
		},
		map[string][]vector.Vector3{
			modeling.PositionAttribute: []vector.Vector3{
				vector.NewVector3(0, 0, 0),
				vector.NewVector3(0, 1, 0),
				vector.NewVector3(1, 1, 0),
				vector.NewVector3(0, 0, 0),
				vector.NewVector3(1, 1, 0),
				vector.NewVector3(1, 0, 0),
			},
			modeling.NormalAttribute: []vector.Vector3{
				vector.NewVector3(0, 0, 0),
				vector.NewVector3(0, 1, 0),
				vector.NewVector3(1, 1, 0),
				vector.NewVector3(0, 0, 0),
				vector.NewVector3(1, 1, 0),
				vector.NewVector3(1, 0, 0),
			},
		},
		map[string][]vector.Vector2{
			modeling.TexCoordAttribute: []vector.Vector2{
				vector.NewVector2(0, 0),
				vector.NewVector2(0, 1),
				vector.NewVector2(1, 1),
				vector.NewVector2(0, 0),
				vector.NewVector2(1, 1),
				vector.NewVector2(1, 0),
			},
		},
		nil,
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
	if assert.Len(t, v1.Indices, 3) {
		assert.Equal(t, 1, v1.Indices[1])
		assert.Equal(t, 0, v1.Indices[0])
		assert.Equal(t, 2, v1.Indices[2])
	}

	v1Verts := v1.Float3Data[modeling.PositionAttribute]
	if assert.Len(t, v1Verts, 3) {
		assert.Equal(t, vector.NewVector3(0, 0, 0), v1Verts[0])
		assert.Equal(t, vector.NewVector3(0, 1, 0), v1Verts[1])
		assert.Equal(t, vector.NewVector3(1, 1, 0), v1Verts[2])
	}

	v1UVs := v1.Float2Data[modeling.TexCoordAttribute]
	if assert.Len(t, v1UVs, 3) {
		assert.Equal(t, vector.NewVector2(0, 0), v1UVs[0])
		assert.Equal(t, vector.NewVector2(0, 1), v1UVs[1])
		assert.Equal(t, vector.NewVector2(1, 1), v1UVs[2])
	}

	v2 := meshes[1].View()
	if assert.Len(t, v2.Indices, 3) {
		assert.Equal(t, 0, v2.Indices[0])
		assert.Equal(t, 1, v2.Indices[1])
		assert.Equal(t, 2, v2.Indices[2])
	}

	v2Verts := v2.Float3Data[modeling.PositionAttribute]
	if assert.Len(t, v2Verts, 3) {
		assert.Equal(t, vector.NewVector3(0, 0, 0), v2Verts[0])
		assert.Equal(t, vector.NewVector3(1, 1, 0), v2Verts[1])
		assert.Equal(t, vector.NewVector3(1, 0, 0), v2Verts[2])
	}

	v2UVs := v2.Float2Data[modeling.TexCoordAttribute]
	if assert.Len(t, v2UVs, 3) {
		assert.Equal(t, vector.NewVector2(0, 0), v2UVs[0])
		assert.Equal(t, vector.NewVector2(1, 1), v2UVs[1])
		assert.Equal(t, vector.NewVector2(1, 0), v2UVs[2])
	}
}
