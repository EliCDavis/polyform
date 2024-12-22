package ply_test

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
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
}

func TestToMeshLittleEndian(t *testing.T) {
	bunny, err := ply.Load("../../test-models/stanford-bunny.ply")

	assert.NoError(t, err)
	assert.NotNil(t, bunny)
	assert.Equal(t, 69451, bunny.PrimitiveCount())
	assert.Equal(t, 35947, bunny.AttributeLength())
}

func TestToMeshLittleEndianTextured(t *testing.T) {
	covid, err := ply.Load("../../test-models/covid.ply")

	assert.NoError(t, err)
	assert.NotNil(t, covid)
	assert.Equal(t, 67960, covid.PrimitiveCount())
	assert.Equal(t, 203880, covid.AttributeLength())
	assert.True(t, covid.HasFloat2Attribute(modeling.TexCoordAttribute))
}

func TestMeshReader_Binary_SimplePointCloud(t *testing.T) {
	// ARRANGE ================================================================
	reader := ply.MeshReader{
		AttributeElement: ply.VertexElementName,
		Properties: []ply.PropertyReader{
			&ply.Vector3PropertyReader{
				ModelAttribute: modeling.PositionAttribute,
				PlyPropertyX:   "x",
				PlyPropertyY:   "y",
				PlyPropertyZ:   "z",
			},
		},
	}

	header := ply.Header{
		Format: ply.BinaryLittleEndian,
		Elements: []ply.Element{{
			Name:  ply.VertexElementName,
			Count: 3,
			Properties: []ply.Property{
				ply.ScalarProperty{PropertyName: "x", Type: ply.Float},
				ply.ScalarProperty{PropertyName: "y", Type: ply.Float},
				ply.ScalarProperty{PropertyName: "z", Type: ply.Float},
			},
		}},
	}

	inputData := []vector3.Serializable[float32]{
		{X: 1, Y: 2, Z: 3},
		{X: 4, Y: 5, Z: 6},
		{X: 7, Y: 8, Z: 9},
	}

	buf := &bytes.Buffer{}
	header.Write(buf)
	binary.Write(buf, binary.LittleEndian, inputData)

	// ACT ====================================================================
	mesh, err := reader.Read(buf)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.True(t, mesh.HasFloat3Attribute(modeling.PositionAttribute), "mesh should have position attribute")

	positionData := mesh.Float3Attribute(modeling.PositionAttribute)
	assert.Equal(t, 3, positionData.Len())
	assert.Equal(t, inputData[0].Immutable().ToFloat64(), positionData.At(0))
	assert.Equal(t, inputData[1].Immutable().ToFloat64(), positionData.At(1))
	assert.Equal(t, inputData[2].Immutable().ToFloat64(), positionData.At(2))
}

func TestMeshReader_Binary_EverythingPointCloud(t *testing.T) {
	// ARRANGE ================================================================
	header := ply.Header{
		Format: ply.BinaryBigEndian,
		Elements: []ply.Element{{
			Name:  ply.VertexElementName,
			Count: 2,
			Properties: []ply.Property{
				ply.ScalarProperty{PropertyName: "x", Type: ply.Double},
				ply.ScalarProperty{PropertyName: "nx", Type: ply.Float},
				ply.ScalarProperty{PropertyName: "r", Type: ply.UChar},
				ply.ScalarProperty{PropertyName: "y", Type: ply.Double},
				ply.ScalarProperty{PropertyName: "ny", Type: ply.Float},
				ply.ScalarProperty{PropertyName: "g", Type: ply.UChar},
				ply.ScalarProperty{PropertyName: "z", Type: ply.Double},
				ply.ScalarProperty{PropertyName: "nz", Type: ply.Float},
				ply.ScalarProperty{PropertyName: "b", Type: ply.UChar},
				ply.ScalarProperty{PropertyName: "a", Type: ply.UChar},
				ply.ScalarProperty{PropertyName: "s", Type: ply.Float},
				ply.ScalarProperty{PropertyName: "t", Type: ply.Float},
				ply.ScalarProperty{PropertyName: "opacity", Type: ply.Float},
				ply.ScalarProperty{PropertyName: "unknown", Type: ply.Float},
			},
		}},
	}

	type Point struct {
		x       float64
		nx      float32
		r       byte
		y       float64
		ny      float32
		g       byte
		z       float64
		nz      float32
		b       byte
		a       byte
		s       float32
		t       float32
		opacity float32
		unknown float32
	}

	inputData := []Point{
		{x: 1, y: 2, z: 3, nx: 1, ny: 2, nz: 3, r: 1, g: 2, b: 3, a: 255, s: 10, t: 20, opacity: 10, unknown: 70},
		{x: 4, y: 5, z: 6, nx: 4, ny: 5, nz: 6, r: 4, g: 5, b: 6, a: 255, s: 30, t: 40, opacity: 20, unknown: 80},
	}

	buf := &bytes.Buffer{}
	header.Write(buf)
	binary.Write(buf, binary.BigEndian, inputData)

	// ACT ====================================================================
	mesh, err := ply.ReadMesh(buf)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.True(t, mesh.HasFloat3Attribute(modeling.PositionAttribute), "mesh should have position attribute")
	assert.True(t, mesh.HasFloat3Attribute(modeling.NormalAttribute), "mesh should have normal attribute")
	assert.True(t, mesh.HasFloat4Attribute(modeling.ColorAttribute), "mesh should have color attribute")
	assert.True(t, mesh.HasFloat2Attribute(modeling.TexCoordAttribute), "mesh should have texcoord attribute")
	assert.True(t, mesh.HasFloat1Attribute(modeling.OpacityAttribute), "mesh should have opacity attribute")
	assert.True(t, mesh.HasFloat1Attribute("unknown"), "mesh should have unknown attribute")

	positionData := mesh.Float3Attribute(modeling.PositionAttribute)
	assert.Equal(t, 2, positionData.Len())
	assert.Equal(t, vector3.New(1., 2., 3.), positionData.At(0))
	assert.Equal(t, vector3.New(4., 5., 6.), positionData.At(1))

	normalData := mesh.Float3Attribute(modeling.NormalAttribute)
	assert.Equal(t, 2, normalData.Len())
	assert.Equal(t, vector3.New(1., 2., 3.), normalData.At(0))
	assert.Equal(t, vector3.New(4., 5., 6.), normalData.At(1))

	colorData := mesh.Float4Attribute(modeling.ColorAttribute)
	assert.Equal(t, 2, colorData.Len())
	assert.Equal(t, vector4.New(1./255., 2./255., 3./255., 1.), colorData.At(0))
	assert.Equal(t, vector4.New(4./255., 5./255., 6./255., 1.), colorData.At(1))

	texData := mesh.Float2Attribute(modeling.TexCoordAttribute)
	assert.Equal(t, 2, texData.Len())
	assert.Equal(t, vector2.New(10., 20.), texData.At(0))
	assert.Equal(t, vector2.New(30., 40.), texData.At(1))

	opacityData := mesh.Float1Attribute(modeling.OpacityAttribute)
	assert.Equal(t, 2, opacityData.Len())
	assert.Equal(t, 10., opacityData.At(0))
	assert.Equal(t, 20., opacityData.At(1))

	unkownData := mesh.Float1Attribute("unknown")
	assert.Equal(t, 2, unkownData.Len())
	assert.Equal(t, 70., unkownData.At(0))
	assert.Equal(t, 80., unkownData.At(1))
}

func TestMeshReader_ASCII_EverythingPointCloud(t *testing.T) {
	// ARRANGE ================================================================
	plyFile := `ply
format ascii 1.0
element vertex 2
property double x
property float32 nx
property uchar r
property double y
property float32 ny
property uchar g
property double z
property float32 nz
property uchar b
property uchar a
property float32 s
property float32 t
property uchar opacity
property float32 unknown
end_header
1 1 1 2 2 2 3 3 3 255 10 20 10 70
4 4 4 5 5 5 6 6 6 255 30 40 20 80
`

	// ACT ====================================================================
	mesh, err := ply.ReadMesh(strings.NewReader(plyFile))

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.True(t, mesh.HasFloat3Attribute(modeling.PositionAttribute), "mesh should have position attribute")
	assert.True(t, mesh.HasFloat3Attribute(modeling.NormalAttribute), "mesh should have normal attribute")
	assert.True(t, mesh.HasFloat4Attribute(modeling.ColorAttribute), "mesh should have color attribute")
	assert.True(t, mesh.HasFloat2Attribute(modeling.TexCoordAttribute), "mesh should have texcoord attribute")
	assert.True(t, mesh.HasFloat1Attribute(modeling.OpacityAttribute), "mesh should have opacity attribute")
	assert.True(t, mesh.HasFloat1Attribute("unknown"), "mesh should have unknown attribute")

	positionData := mesh.Float3Attribute(modeling.PositionAttribute)
	assert.Equal(t, 2, positionData.Len())
	assert.Equal(t, vector3.New(1., 2., 3.), positionData.At(0))
	assert.Equal(t, vector3.New(4., 5., 6.), positionData.At(1))

	normalData := mesh.Float3Attribute(modeling.NormalAttribute)
	assert.Equal(t, 2, normalData.Len())
	assert.Equal(t, vector3.New(1., 2., 3.), normalData.At(0))
	assert.Equal(t, vector3.New(4., 5., 6.), normalData.At(1))

	colorData := mesh.Float4Attribute(modeling.ColorAttribute)
	assert.Equal(t, 2, colorData.Len())
	assert.Equal(t, vector4.New(1./255., 2./255., 3./255., 1.), colorData.At(0))
	assert.Equal(t, vector4.New(4./255., 5./255., 6./255., 1.), colorData.At(1))

	texData := mesh.Float2Attribute(modeling.TexCoordAttribute)
	assert.Equal(t, 2, texData.Len())
	assert.Equal(t, vector2.New(10., 20.), texData.At(0))
	assert.Equal(t, vector2.New(30., 40.), texData.At(1))

	opacityData := mesh.Float1Attribute(modeling.OpacityAttribute)
	assert.Equal(t, 2, opacityData.Len())
	assert.Equal(t, 10., opacityData.At(0))
	assert.Equal(t, 20., opacityData.At(1))

	unkownData := mesh.Float1Attribute("unknown")
	assert.Equal(t, 2, unkownData.Len())
	assert.Equal(t, 70., unkownData.At(0))
	assert.Equal(t, 80., unkownData.At(1))
}

func TestMeshReader_ASCII_Vector4FallsbackToVector3WhenWMissing(t *testing.T) {
	// ARRANGE ================================================================
	plyFile := `ply
format ascii 1.0
element vertex 2
property uchar r
property uchar g
property uchar b
end_header
1 2 3
4 5 6
`

	// ACT ====================================================================
	mesh, err := ply.ReadMesh(strings.NewReader(plyFile))

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.True(t, mesh.HasFloat3Attribute(modeling.ColorAttribute), "mesh should have color attribute")

	colorData := mesh.Float3Attribute(modeling.ColorAttribute)
	assert.Equal(t, 2, colorData.Len())
	assert.Equal(t, vector3.New(1./255., 2./255., 3./255.), colorData.At(0))
	assert.Equal(t, vector3.New(4./255., 5./255., 6./255.), colorData.At(1))

}

func TestMeshReader_Binary_Vector4FallsbackToVector3WhenWMissing(t *testing.T) {
	// ARRANGE ================================================================
	header := ply.Header{
		Format: ply.BinaryBigEndian,
		Elements: []ply.Element{{
			Name:  ply.VertexElementName,
			Count: 2,
			Properties: []ply.Property{
				ply.ScalarProperty{PropertyName: "r", Type: ply.UChar},
				ply.ScalarProperty{PropertyName: "g", Type: ply.UChar},
				ply.ScalarProperty{PropertyName: "b", Type: ply.UChar},
			},
		}},
	}

	type Point struct {
		r byte
		g byte
		b byte
	}

	inputData := []Point{
		{r: 1, g: 2, b: 3},
		{r: 4, g: 5, b: 6},
	}

	buf := &bytes.Buffer{}
	header.Write(buf)
	binary.Write(buf, binary.BigEndian, inputData)

	// ACT ====================================================================
	mesh, err := ply.ReadMesh(buf)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.True(t, mesh.HasFloat3Attribute(modeling.ColorAttribute), "mesh should have color attribute")

	colorData := mesh.Float3Attribute(modeling.ColorAttribute)
	assert.Equal(t, 2, colorData.Len())
	assert.Equal(t, vector3.New(1./255., 2./255., 3./255.), colorData.At(0))
	assert.Equal(t, vector3.New(4./255., 5./255., 6./255.), colorData.At(1))
}
