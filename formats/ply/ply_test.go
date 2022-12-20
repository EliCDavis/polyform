package ply_test

import (
	"strings"
	"testing"

	"github.com/EliCDavis/mesh/formats/ply"
	"github.com/stretchr/testify/assert"
)

func TestToMeshThrowsWithBadMagicNumber(t *testing.T) {
	plyData := `test
format ascii 1.0
`
	ply, err := ply.ToMesh(strings.NewReader(plyData))

	assert.EqualError(t, err, "unrecognized magic number: 'test' (expected 'ply')")
	assert.Nil(t, ply)
}

func TestToMeshThrowsWithBadFormatLine(t *testing.T) {
	plyData := `ply
trash
`
	ply, err := ply.ToMesh(strings.NewReader(plyData))

	assert.EqualError(t, err, "unrecognized format line")
	assert.Nil(t, ply)
}

func TestToMeshThrowsWithBadFormatVersion(t *testing.T) {
	plyData := `ply
format ascii 1.2
`
	ply, err := ply.ToMesh(strings.NewReader(plyData))

	assert.EqualError(t, err, "unrecognized version format: 1.2")
	assert.Nil(t, ply)
}

func TestToMeshThrowsWithUnknownFormatType(t *testing.T) {
	plyData := `ply
format bad 1.0
`
	ply, err := ply.ToMesh(strings.NewReader(plyData))

	assert.EqualError(t, err, "unrecognized format: bad")
	assert.Nil(t, ply)
}

func TestToMeshThrowsNoFormatLine(t *testing.T) {
	plyData := `ply
bad ascii 1.0
`
	ply, err := ply.ToMesh(strings.NewReader(plyData))

	assert.EqualError(t, err, "expected format line, received bad")
	assert.Nil(t, ply)
}

func TestToMesh(t *testing.T) {
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
	ply, err := ply.ToMesh(strings.NewReader(plyData))

	assert.NoError(t, err)
	assert.NotNil(t, ply)
}
