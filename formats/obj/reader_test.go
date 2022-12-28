package obj_test

import (
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
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
	square, matReferences, err := obj.ReadMesh(strings.NewReader(objString))
	squareView := square.View()

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Len(t, matReferences, 0)
	assert.Equal(t, 12, square.TriCount())

	assert.Equal(t, 0, squareView.Indices[0])
	assert.Equal(t, 1, squareView.Indices[1])
	assert.Equal(t, 2, squareView.Indices[2])

	assert.Equal(t, 0, squareView.Indices[3])
	assert.Equal(t, 3, squareView.Indices[4])
	assert.Equal(t, 1, squareView.Indices[5])
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
	square, matReferences, err := obj.ReadMesh(strings.NewReader(objString))
	squareView := square.View()

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, 2, square.TriCount())
	assert.Len(t, matReferences, 0)

	assert.Equal(t, 0, squareView.Indices[0])
	assert.Equal(t, 1, squareView.Indices[1])
	assert.Equal(t, 2, squareView.Indices[2])

	assert.Equal(t, 1, squareView.Indices[3])
	assert.Equal(t, 2, squareView.Indices[4])
	assert.Equal(t, 3, squareView.Indices[5])

	vertices := squareView.Float3Data[modeling.PositionAttribute]
	assert.Len(t, vertices, 4)
	assert.Equal(t, vector.NewVector3(0.0, 0.0, 0.0), vertices[0])
	assert.Equal(t, vector.NewVector3(0.0, 1.0, 0.0), vertices[1])
	assert.Equal(t, vector.NewVector3(0.0, 1.0, 1.0), vertices[2])
	assert.Equal(t, vector.NewVector3(0.0, 0.0, 1.0), vertices[3])
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
	square, matReferences, err := obj.ReadMesh(strings.NewReader(objString))
	squareView := square.View()

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, 3, square.TriCount())
	if assert.Len(t, matReferences, 1) {
		assert.Equal(t, "test.mtl", matReferences[0])
	}

	assert.Equal(t, 0, squareView.Indices[0])
	assert.Equal(t, 1, squareView.Indices[1])
	assert.Equal(t, 2, squareView.Indices[2])

	assert.Equal(t, 1, squareView.Indices[3])
	assert.Equal(t, 2, squareView.Indices[4])
	assert.Equal(t, 3, squareView.Indices[5])

	vertices := squareView.Float3Data[modeling.PositionAttribute]
	assert.Len(t, vertices, 4)
	assert.Equal(t, vector.NewVector3(0.0, 0.0, 0.0), vertices[0])
	assert.Equal(t, vector.NewVector3(0.0, 1.0, 0.0), vertices[1])
	assert.Equal(t, vector.NewVector3(0.0, 1.0, 1.0), vertices[2])
	assert.Equal(t, vector.NewVector3(0.0, 0.0, 1.0), vertices[3])

	normals := squareView.Float3Data[modeling.NormalAttribute]
	assert.Len(t, normals, 4)
	assert.Equal(t, vector.NewVector3(0.0, 0.0, 0.0), normals[0])
	assert.Equal(t, vector.NewVector3(0.0, 1.0, 0.0), normals[1])
	assert.Equal(t, vector.NewVector3(0.0, 1.0, 1.0), normals[2])
	assert.Equal(t, vector.NewVector3(0.0, 0.0, 1.0), normals[3])

	uvs := squareView.Float2Data[modeling.TexCoordAttribute]
	assert.Len(t, uvs, 4)
	assert.Equal(t, vector.NewVector2(0.0, 0.0), uvs[0])
	assert.Equal(t, vector.NewVector2(0.0, 1.0), uvs[1])
	assert.Equal(t, vector.NewVector2(0.0, 1.0), uvs[2])
	assert.Equal(t, vector.NewVector2(0.0, 0.0), uvs[3])

	if assert.Len(t, square.Materials(), 3) {
		assert.Equal(t, 1, square.Materials()[0].NumOfTris)
		assert.Equal(t, 1, square.Materials()[1].NumOfTris)
		assert.Equal(t, 1, square.Materials()[2].NumOfTris)

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
