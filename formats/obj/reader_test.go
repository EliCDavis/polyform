package obj_test

import (
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func Test_ReadOBJ_NoTexture(t *testing.T) {
	// ARRANGE ================================================================
	objString := `# cube.obj
#
	
g cube
	
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
	
f  1//2  7//2  5//2
f  1//2  3//2  7//2 
f  1//6  4//6  3//6 
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
	contents, matReferences, err := obj.ReadMesh(strings.NewReader(objString))
	square := contents[0].Mesh

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Len(t, contents, 1)
	assert.Equal(t, "cube", contents[0].Name)
	assert.Len(t, matReferences, 0)
	assert.Equal(t, 12, square.PrimitiveCount())

	squareIndices := square.Indices()
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

g side 1
f  1//2  7//2  5//2
f  1//2  3//2  7//2 
g side 2
f  1//6  4//6  3//6 
f  1//6  2//6  4//6 
g side 3
f  3//3  8//3  7//3 
f  3//3  4//3  8//3 
g side 4
f  5//5  7//5  8//5 
f  5//5  8//5  6//5 
g side 5
f  1//4  5//4  6//4 
f  1//4  6//4  2//4 
g side 6
f  2//1  6//1  8//1 
f  2//1  8//1  4//1 
`

	// ACT ====================================================================
	contents, matReferences, err := obj.ReadMesh(strings.NewReader(objString))

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Len(t, matReferences, 0)
	assert.Len(t, contents, 6)
	assert.Equal(t, "side 1", contents[0].Name)
	assert.Equal(t, "side 2", contents[1].Name)
	assert.Equal(t, "side 3", contents[2].Name)
	assert.Equal(t, "side 4", contents[3].Name)
	assert.Equal(t, "side 5", contents[4].Name)
	assert.Equal(t, "side 6", contents[5].Name)

	assert.Equal(t, 2, contents[0].Mesh.PrimitiveCount())
	assert.Equal(t, 2, contents[1].Mesh.PrimitiveCount())
	assert.Equal(t, 2, contents[2].Mesh.PrimitiveCount())
	assert.Equal(t, 2, contents[3].Mesh.PrimitiveCount())
	assert.Equal(t, 2, contents[4].Mesh.PrimitiveCount())
	assert.Equal(t, 2, contents[5].Mesh.PrimitiveCount())
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
	contents, matReferences, err := obj.ReadMesh(strings.NewReader(objString))
	square := contents[0].Mesh

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, 2, square.PrimitiveCount())
	assert.Len(t, matReferences, 0)
	assert.Equal(t, "", contents[0].Name)
	assert.Len(t, contents, 1)

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
	contents, matReferences, err := obj.ReadMesh(strings.NewReader(objString))
	square := contents[0].Mesh

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, 3, square.PrimitiveCount())
	if assert.Len(t, matReferences, 1) {
		assert.Equal(t, "test.mtl", matReferences[0])
	}
	assert.Equal(t, "", contents[0].Name)
	assert.Len(t, contents, 1)

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

	normals := square.Float3Attribute(modeling.NormalAttribute)
	assert.Equal(t, normals.Len(), 4)
	assert.Equal(t, vector3.New(0.0, 0.0, 0.0), normals.At(0))
	assert.Equal(t, vector3.New(0.0, 1.0, 0.0), normals.At(1))
	assert.Equal(t, vector3.New(0.0, 1.0, 1.0), normals.At(2))
	assert.Equal(t, vector3.New(0.0, 0.0, 1.0), normals.At(3))

	uvs := square.Float2Attribute(modeling.TexCoordAttribute)
	assert.Equal(t, uvs.Len(), 4)
	assert.Equal(t, vector2.New(0.0, 0.0), uvs.At(0))
	assert.Equal(t, vector2.New(0.0, 1.0), uvs.At(1))
	assert.Equal(t, vector2.New(0.0, 1.0), uvs.At(2))
	assert.Equal(t, vector2.New(0.0, 0.0), uvs.At(3))

	if assert.Len(t, square.Materials(), 3) {
		assert.Equal(t, 1, square.Materials()[0].PrimitiveCount)
		assert.Equal(t, 1, square.Materials()[1].PrimitiveCount)
		assert.Equal(t, 1, square.Materials()[2].PrimitiveCount)

		if assert.NotNil(t, square.Materials()[0].Material) {
			assert.Equal(t, "red", square.Materials()[0].Material.Name)
		}

		if assert.NotNil(t, square.Materials()[1].Material) {
			assert.Equal(t, "green", square.Materials()[1].Material.Name)
		}

		if assert.NotNil(t, square.Materials()[2].Material) {
			assert.Equal(t, "red", square.Materials()[2].Material.Name)
		}

		if square.Materials()[0].Material != square.Materials()[2].Material {
			t.Error("mesh materials don't reference same underlying material")
		}
	}
}
