package obj_test

import (
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ReadOBJ_NoTexture(t *testing.T) {
	// ARRANGE ================================================================
	objString := `# cube.obj
#
	
o cube
	
v  0.0  0.0  0.0
v  0.0  0.0  1.0
v  0.0  1.0  0.0
v  0.0  1.0  1.0
v  1.0  0.0  0.0
v  1.0  0.0  1.0
v  1.0  1.0  0.0
v  1.0  1.0  1.0

vn  0.0  0.0  1.0
vn  0.0  0.0 -1.0
vn  0.0  1.0  0.0
vn  0.0 -1.0  0.0
vn  1.0  0.0  0.0
vn -1.0  0.0  0.0
	
f  1//  7//  5//
f  1//  3//  7// 
f  1//  4//  3// 
f  1//6  2//6  4//6 
f  3//3  8//3  7//3 
f  3//3  4//3  8//3 
f  5//5  7//5  8//5 
f  5//5  8//5  6//5 
f  1//4  5//4  6//4 
f  1//4  6//4  2//4 
f  2//1  6//1  8//1 
f  2//1  8//1  4//1 
`

	// ACT ====================================================================
	scene, matReferences, err := obj.ReadMesh(strings.NewReader(objString))

	// ASSERT =================================================================
	require.NoError(t, err)
	assert.Len(t, scene.Objects, 1)
	assert.Equal(t, scene.Objects[0].Name, "cube")
	assert.Len(t, scene.Objects[0].Entries, 1)

	square := scene.Objects[0].Entries[0]
	assert.Nil(t, square.Material)
	assert.Len(t, matReferences, 0)
	assert.Equal(t, 12, square.Mesh.PrimitiveCount())

	squareIndices := square.Mesh.Indices()
	assert.Equal(t, 0, squareIndices.At(0))
	assert.Equal(t, 1, squareIndices.At(1))
	assert.Equal(t, 2, squareIndices.At(2))

	assert.Equal(t, 0, squareIndices.At(3))
	assert.Equal(t, 3, squareIndices.At(4))
	assert.Equal(t, 1, squareIndices.At(5))
}

func Test_ReadOBJ_MultipleGroups(t *testing.T) {
	// ARRANGE ================================================================
	objString := `# cube.obj
#
	
	
v  0.0  0.0  0.0
v  0.0  0.0  1.0
v  0.0  1.0  0.0
v  0.0  1.0  1.0
v  1.0  0.0  0.0
v  1.0  0.0  1.0
v  1.0  1.0  0.0
v  1.0  1.0  1.0

vn  0.0  0.0  1.0
vn  0.0  0.0 -1.0
vn  0.0  1.0  0.0
vn  0.0 -1.0  0.0
vn  1.0  0.0  0.0
vn -1.0  0.0  0.0

o side 1
f  1//  7//  5//
f  1//  3//  7// 
o side 2
f  1//6  4//6  3//6 
f  1//6  2//6  4//6 
o side 3
f  3//3  8//3  7//3 
f  3//3  4//3  8//3 
o side 4
f  5//5  7//5  8//5 
f  5//5  8//5  6//5 
o side 5
f  1//4  5//4  6//4 
f  1//4  6//4  2//4 
o side 6
f  2//1  6//1  8//1 
f  2//1  8//1  4//1 
`

	// ACT ====================================================================
	scene, matReferences, err := obj.ReadMesh(strings.NewReader(objString))

	// ASSERT =================================================================
	require.NoError(t, err)
	assert.Len(t, matReferences, 0)
	assert.Len(t, scene.Objects, 6)
	assert.Equal(t, "side 1", scene.Objects[0].Name)
	assert.Equal(t, "side 2", scene.Objects[1].Name)
	assert.Equal(t, "side 3", scene.Objects[2].Name)
	assert.Equal(t, "side 4", scene.Objects[3].Name)
	assert.Equal(t, "side 5", scene.Objects[4].Name)
	assert.Equal(t, "side 6", scene.Objects[5].Name)

	assert.Equal(t, 2, scene.Objects[0].Entries[0].Mesh.PrimitiveCount())
	assert.Equal(t, 2, scene.Objects[1].Entries[0].Mesh.PrimitiveCount())
	assert.Equal(t, 2, scene.Objects[2].Entries[0].Mesh.PrimitiveCount())
	assert.Equal(t, 2, scene.Objects[3].Entries[0].Mesh.PrimitiveCount())
	assert.Equal(t, 2, scene.Objects[4].Entries[0].Mesh.PrimitiveCount())
	assert.Equal(t, 2, scene.Objects[5].Entries[0].Mesh.PrimitiveCount())
}

func Test_ReadOBJ_SimpleSquare_NoNormalOrTextures(t *testing.T) {
	// ARRANGE ================================================================
	objString := `	
v  0.0  0.0  0.0
v  0.0  1.0  0.0
v  0.0  1.0  1.0
v  0.0  0.0  1.0

	
f  1 2 3
f  2 3 4
`

	// ACT ====================================================================
	scene, matReferences, err := obj.ReadMesh(strings.NewReader(objString))

	// ASSERT =================================================================
	require.NoError(t, err)
	square := scene.Objects[0].Entries[0].Mesh
	assert.Equal(t, 2, square.PrimitiveCount())
	assert.Len(t, matReferences, 0)
	assert.Equal(t, "", scene.Objects[0].Name)
	assert.Len(t, scene.Objects, 1)

	squareIndices := square.Indices()
	assert.Equal(t, 0, squareIndices.At(0))
	assert.Equal(t, 1, squareIndices.At(1))
	assert.Equal(t, 2, squareIndices.At(2))

	assert.Equal(t, 1, squareIndices.At(3))
	assert.Equal(t, 2, squareIndices.At(4))
	assert.Equal(t, 3, squareIndices.At(5))

	vertices := square.Float3Attribute(modeling.PositionAttribute)
	assert.Equal(t, vertices.Len(), 4)
	assert.Equal(t, vector3.New(0.0, 0.0, 0.0), vertices.At(0))
	assert.Equal(t, vector3.New(0.0, 1.0, 0.0), vertices.At(1))
	assert.Equal(t, vector3.New(0.0, 1.0, 1.0), vertices.At(2))
	assert.Equal(t, vector3.New(0.0, 0.0, 1.0), vertices.At(3))
}

func Test_ReadOBJ_SimpleSquare(t *testing.T) {
	// ARRANGE ================================================================
	objString := `	

mtllib test.mtl 

v  0.0  0.0  0.0
v  0.0  1.0  0.0
v  0.0  1.0  1.0
v  0.0  0.0  1.0

vn  0.0  0.0  0.0
vn  0.0  1.0  0.0
vn  0.0  1.0  1.0
vn  0.0  0.0  1.0

vt  0.0  0.0
vt  0.0  1.0
vt  0.0  1.0
vt  0.0  0.0

usemtl red 
f  1/1/1 2/2/2 3/3/3

usemtl green
f  2/2/2 3/3/3 4/4/4

usemtl red 
f  1/1/1 2/2/2 3/3/3
`

	// ACT ====================================================================
	scene, matReferences, err := obj.ReadMesh(strings.NewReader(objString))

	// ASSERT =================================================================
	require.NoError(t, err)
	require.Len(t, scene.Objects, 1)
	assert.Equal(t, "", scene.Objects[0].Name)
	require.Len(t, scene.Objects[0].Entries, 3)

	require.Len(t, matReferences, 1)
	assert.Equal(t, "test.mtl", matReferences[0])

	for _, entry := range scene.Objects[0].Entries {

		square := entry.Mesh

		assert.Equal(t, 1, square.PrimitiveCount())
		squareIndices := square.Indices()
		assert.Equal(t, 0, squareIndices.At(0))
		assert.Equal(t, 1, squareIndices.At(1))
		assert.Equal(t, 2, squareIndices.At(2))

		// vertices := square.Float3Attribute(modeling.PositionAttribute)
		// assert.Equal(t, vertices.Len(), 4)
		// assert.Equal(t, vector3.New(0.0, 0.0, 0.0), vertices.At(0))
		// assert.Equal(t, vector3.New(0.0, 1.0, 0.0), vertices.At(1))
		// assert.Equal(t, vector3.New(0.0, 1.0, 1.0), vertices.At(2))
		// assert.Equal(t, vector3.New(0.0, 0.0, 1.0), vertices.At(3))

		// normals := square.Float3Attribute(modeling.NormalAttribute)
		// assert.Equal(t, normals.Len(), 4)
		// assert.Equal(t, vector3.New(0.0, 0.0, 0.0), normals.At(0))
		// assert.Equal(t, vector3.New(0.0, 1.0, 0.0), normals.At(1))
		// assert.Equal(t, vector3.New(0.0, 1.0, 1.0), normals.At(2))
		// assert.Equal(t, vector3.New(0.0, 0.0, 1.0), normals.At(3))

		// uvs := square.Float2Attribute(modeling.TexCoordAttribute)
		// assert.Equal(t, uvs.Len(), 4)
		// assert.Equal(t, vector2.New(0.0, 0.0), uvs.At(0))
		// assert.Equal(t, vector2.New(0.0, 1.0), uvs.At(1))
		// assert.Equal(t, vector2.New(0.0, 1.0), uvs.At(2))
		// assert.Equal(t, vector2.New(0.0, 0.0), uvs.At(3))
	}

	require.NotNil(t, scene.Objects[0].Entries[0].Material)
	assert.Equal(t, "red", scene.Objects[0].Entries[0].Material.Name)

	require.NotNil(t, scene.Objects[0].Entries[1].Material)
	assert.Equal(t, "green", scene.Objects[0].Entries[1].Material.Name)

	require.NotNil(t, scene.Objects[0].Entries[2].Material)
	assert.Equal(t, "red", scene.Objects[0].Entries[2].Material.Name)

	if scene.Objects[0].Entries[0].Material != scene.Objects[0].Entries[2].Material {
		t.Error("mesh materials don't reference same underlying material")
	}
}
