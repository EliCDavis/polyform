package ply_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestToMeshThrowsWithBadMagicNumber(t *testing.T) {
	plyData := `test
format ascii 1.0
`
	ply, err := ply.ReadMesh(strings.NewReader(plyData))

	assert.EqualError(t, err, "unrecognized magic number: 'test' (expected 'ply')")
	assert.Nil(t, ply)
}

func TestToMeshThrowsWithBadFormatLine(t *testing.T) {
	plyData := `ply
trash
`
	ply, err := ply.ReadMesh(strings.NewReader(plyData))

	assert.EqualError(t, err, "unrecognized format line")
	assert.Nil(t, ply)
}

func TestToMeshThrowsWithBadFormatVersion(t *testing.T) {
	plyData := `ply
format ascii 1.2
`
	ply, err := ply.ReadMesh(strings.NewReader(plyData))

	assert.EqualError(t, err, "unrecognized version format: 1.2")
	assert.Nil(t, ply)
}

func TestToMeshThrowsWithUnknownFormatType(t *testing.T) {
	plyData := `ply
format bad 1.0
`
	ply, err := ply.ReadMesh(strings.NewReader(plyData))

	assert.EqualError(t, err, "unrecognized format: bad")
	assert.Nil(t, ply)
}

func TestToMeshThrowsNoFormatLine(t *testing.T) {
	plyData := `ply
bad ascii 1.0
`
	ply, err := ply.ReadMesh(strings.NewReader(plyData))

	assert.EqualError(t, err, "expected format line, received bad")
	assert.Nil(t, ply)
}

func TestToMeshASCII(t *testing.T) {
	plyData := `ply
format ascii 1.0
comment made by anonymous
comment this file is a cube
element vertex 8
property float32 x
property float32 y
property float32 z
element face 6
property list uint8 int32 vertex_index
end_header
0 0 0
0 0 1
0 1 1
0 1 0
1 0 0
1 0 1
1 1 1
1 1 0
4 0 1 2 3
4 7 6 5 4
4 0 4 5 1
4 1 5 6 2
4 2 6 7 3
4 3 7 4 0
`
	ply, err := ply.ReadMesh(strings.NewReader(plyData))

	assert.NoError(t, err)
	assert.NotNil(t, ply)
	assert.Equal(t, ply.PrimitiveCount(), 12)
	assert.Equal(t, ply.AttributeLength(), 8)
}

func TestToMeshASCIIWithTextureCords(t *testing.T) {
	plyData := `ply
format ascii 1.0
comment TextureFile tri.png
comment this file is a tesselated cube
element vertex 8
property float32 x
property float32 y
property float32 z
element face 12
property list uint8 int32 vertex_index
property list uchar float texcoord
end_header
0 0 0
0 0 1
0 1 1
0 1 0
1 0 0
1 0 1
1 1 1
1 1 0
3 0 1 2 6 0.1 0.8 0.2 0.1 0.8 0.2
3 1 2 3 6 0.1 0.8 0.2 0.1 0.8 0.2
3 7 6 5 6 0.1 0.8 0.2 0.1 0.8 0.2
3 6 5 4 6 0.1 0.8 0.2 0.1 0.8 0.2
3 0 4 5 6 0.1 0.8 0.2 0.1 0.8 0.2
3 4 5 1 6 0.1 0.8 0.2 0.1 0.8 0.2
3 1 5 6 6 0.1 0.8 0.2 0.1 0.8 0.2
3 5 6 2 6 0.1 0.8 0.2 0.1 0.8 0.2
3 2 6 7 6 0.1 0.8 0.2 0.1 0.8 0.2
3 6 7 3 6 0.1 0.8 0.2 0.1 0.8 0.2
3 3 7 4 6 0.1 0.8 0.2 0.1 0.8 0.2
3 7 4 0 6 0.1 0.8 0.2 0.1 0.8 0.2
`
	ply, err := ply.ReadMesh(strings.NewReader(plyData))

	assert.NoError(t, err)
	assert.NotNil(t, ply)
	assert.Equal(t, 12, ply.PrimitiveCount())
	assert.Equal(t, 36, ply.AttributeLength())
	assert.True(t, ply.HasFloat2Attribute(modeling.TexCoordAttribute))
	if assert.Len(t, ply.Materials(), 1) && assert.NotNil(t, ply.Materials()[0].Material) {
		if assert.NotNil(t, ply.Materials()[0].Material.ColorTextureURI) {
			assert.Equal(t, "tri.png", *ply.Materials()[0].Material.ColorTextureURI)
		}
	}
}

func TestToMeshLittleEndian(t *testing.T) {
	data, err := os.ReadFile("../../test-models/stanford-bunny.ply")
	if !assert.NoError(t, err) {
		return
	}

	bunny, err := ply.ReadMesh(bytes.NewBuffer(data))

	assert.NoError(t, err)
	assert.NotNil(t, bunny)
	assert.Equal(t, 69451, bunny.PrimitiveCount())
	assert.Equal(t, 35947, bunny.AttributeLength())
}

func TestToMeshLittleEndianTextured(t *testing.T) {
	data, err := os.ReadFile("../../test-models/covid.ply")
	if !assert.NoError(t, err) {
		return
	}

	covid, err := ply.ReadMesh(bytes.NewBuffer(data))

	assert.NoError(t, err)
	assert.NotNil(t, covid)
	assert.Equal(t, 67960, covid.PrimitiveCount())
	assert.Equal(t, 203880, covid.AttributeLength())
	assert.True(t, covid.HasFloat2Attribute(modeling.TexCoordAttribute))
}

func TestWriteASCII(t *testing.T) {
	plyData := `ply
format ascii 1.0
comment Created with github.com/EliCDavis/polyform
element vertex 8
property float nx
property float ny
property float nz
property float x
property float y
property float z
element face 12
property list uchar int vertex_indices
end_header
-0.577350 -0.577350 -0.577350 -0.500000 -0.500000 -0.500000
-0.577350 -0.577350 0.577350 -0.500000 -0.500000 0.500000
-0.577350 0.577350 -0.577350 -0.500000 0.500000 -0.500000
-0.577350 0.577350 0.577350 -0.500000 0.500000 0.500000
0.577350 -0.577350 -0.577350 0.500000 -0.500000 -0.500000
0.577350 -0.577350 0.577350 0.500000 -0.500000 0.500000
0.577350 0.577350 -0.577350 0.500000 0.500000 -0.500000
0.577350 0.577350 0.577350 0.500000 0.500000 0.500000
3 0 2 6
3 0 6 4
3 1 3 2
3 1 2 0
3 4 6 7
3 4 7 5
3 2 3 7
3 2 7 6
3 1 0 4
3 1 4 5
3 5 7 3
3 5 3 1
`
	cube := primitives.Cube()

	buf := bytes.Buffer{}
	err := ply.WriteASCII(&buf, cube)

	assert.NoError(t, err)
	assert.Equal(t, plyData, buf.String())
}

func TestWriteASCIIWithTextureData(t *testing.T) {
	plyData := `ply
format ascii 1.0
comment TextureFile tri.png
comment Created with github.com/EliCDavis/polyform
element vertex 3
property float x
property float y
property float z
element face 1
property list uchar int vertex_indices
property list uchar float texcoord
end_header
0.000000 0.000000 0.000000
0.000000 1.000000 0.000000
1.000000 1.000000 0.000000
3 0 1 2 6 0.000000 0.000000 0.000000 1.000000 1.000000 1.000000
`

	imgName := "tri.png"
	tri := modeling.NewMesh([]int{0, 1, 2}).
		SetFloat3Data(map[string][]vector3.Vector[float64]{
			modeling.PositionAttribute: []vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 1., 0.),
			},
		}).
		SetFloat2Data(map[string][]vector2.Vector[float64]{
			modeling.TexCoordAttribute: []vector2.Float64{
				vector2.New(0., 0.),
				vector2.New(0., 1.),
				vector2.New(1., 1.),
			},
		}).
		SetMaterials([]modeling.MeshMaterial{
			{
				PrimitiveCount: 1,
				Material: &modeling.Material{
					Name:            "example",
					ColorTextureURI: &imgName,
				},
			},
		})

	buf := bytes.Buffer{}
	err := ply.WriteASCII(&buf, tri)

	assert.NoError(t, err)
	assert.Equal(t, plyData, buf.String())
}
