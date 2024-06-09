package ply_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/modeling"
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
obj_info bs stuff 
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
	assert.Equal(t, 12, ply.PrimitiveCount())
	assert.Equal(t, 8, ply.AttributeLength())
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
