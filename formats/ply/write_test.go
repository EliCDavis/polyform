package ply_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
)

func TestWriteASCII(t *testing.T) {
	plyData := `ply
format ascii 1.0
comment Created with github.com/EliCDavis/polyform
element vertex 8
property float x
property float y
property float z
property float nx
property float ny
property float nz
element face 12
property list uchar int vertex_indices
end_header
-0.5 -0.5 -0.5 -0.5773502691896258 -0.5773502691896258 -0.5773502691896258
-0.5 -0.5 0.5 -0.5773502691896258 -0.5773502691896258 0.5773502691896258
-0.5 0.5 -0.5 -0.5773502691896258 0.5773502691896258 -0.5773502691896258
-0.5 0.5 0.5 -0.5773502691896258 0.5773502691896258 0.5773502691896258
0.5 -0.5 -0.5 0.5773502691896258 -0.5773502691896258 -0.5773502691896258
0.5 -0.5 0.5 0.5773502691896258 -0.5773502691896258 0.5773502691896258
0.5 0.5 -0.5 0.5773502691896258 0.5773502691896258 -0.5773502691896258
0.5 0.5 0.5 0.5773502691896258 0.5773502691896258 0.5773502691896258
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
	cube := primitives.Cube{
		Height: 1,
		Width:  1,
		Depth:  1,
	}.Welded()

	buf := bytes.Buffer{}
	err := ply.Write(&buf, cube, ply.ASCII)

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
0 0 0
0 1 0
1 1 0
3 0 1 2 6 0 0 0 1 1 1
`

	imgName := "tri.png"
	tri := modeling.NewTriangleMesh([]int{0, 1, 2}).
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
	err := ply.Write(&buf, tri, ply.ASCII)

	assert.NoError(t, err)
	assert.Equal(t, plyData, buf.String())
}

func TestWriteASCII_PointCloud(t *testing.T) {
	// ARRANGE ================================================================
	writer := ply.MeshWriter{
		Format: ply.ASCII,
		Properties: []ply.PropertyWriter{
			ply.Vector1PropertyWriter{
				ModelAttribute: "1-uchar",
				PlyProperty:    "1-uchar",
				Type:           ply.UChar,
			},
			ply.Vector1PropertyWriter{
				ModelAttribute: "1-float",
				PlyProperty:    "1-float",
				Type:           ply.Float,
			},
			ply.Vector2PropertyWriter{
				ModelAttribute: "2-uchar",
				PlyPropertyX:   "2-uchar-x",
				PlyPropertyY:   "2-uchar-y",
				Type:           ply.UChar,
			},
			ply.Vector2PropertyWriter{
				ModelAttribute: "2-float",
				PlyPropertyX:   "2-float-x",
				PlyPropertyY:   "2-float-y",
				Type:           ply.Float,
			},
			ply.Vector3PropertyWriter{
				ModelAttribute: "3-uchar",
				PlyPropertyX:   "3-uchar-x",
				PlyPropertyY:   "3-uchar-y",
				PlyPropertyZ:   "3-uchar-z",
				Type:           ply.UChar,
			},
			ply.Vector3PropertyWriter{
				ModelAttribute: "3-float",
				PlyPropertyX:   "3-float-x",
				PlyPropertyY:   "3-float-y",
				PlyPropertyZ:   "3-float-z",
				Type:           ply.Float,
			},
			ply.Vector4PropertyWriter{
				ModelAttribute: "4-uchar",
				PlyPropertyX:   "4-uchar-x",
				PlyPropertyY:   "4-uchar-y",
				PlyPropertyZ:   "4-uchar-z",
				PlyPropertyW:   "4-uchar-w",
				Type:           ply.UChar,
			},
			ply.Vector4PropertyWriter{
				ModelAttribute: "4-float",
				PlyPropertyX:   "4-float-x",
				PlyPropertyY:   "4-float-y",
				PlyPropertyZ:   "4-float-z",
				PlyPropertyW:   "4-float-w",
				Type:           ply.Float,
			},
		},
	}

	mesh := modeling.NewPointCloud(
		map[string][]vector4.Float64{
			"4-uchar": {vector4.New(1/255., 2/255., 3/255., 4/255.)},
			"4-float": {vector4.New(1.1, 2.2, 3.3, 4.4)},
		},
		map[string][]vector3.Float64{
			"3-uchar": {vector3.New(5/255., 6/255., 7/255.)},
			"3-float": {vector3.New(5.5, 6.6, 7.7)},
		},
		map[string][]vector2.Float64{
			"2-uchar": {vector2.New(8/255., 9/255.)},
			"2-float": {vector2.New(8.8, 9.9)},
		},
		map[string][]float64{
			"1-uchar": {10 / 255.},
			"1-float": {10.1},
		},
		nil,
	)

	// ACT ====================================================================
	buf := bytes.Buffer{}
	err := writer.Write(mesh, &buf)

	// ASSERT =================================================================
	plyData := `ply
format ascii 1.0
comment Created with github.com/EliCDavis/polyform
element vertex 1
property uchar 1-uchar
property float 1-float
property uchar 2-uchar-x
property uchar 2-uchar-y
property float 2-float-x
property float 2-float-y
property uchar 3-uchar-x
property uchar 3-uchar-y
property uchar 3-uchar-z
property float 3-float-x
property float 3-float-y
property float 3-float-z
property uchar 4-uchar-x
property uchar 4-uchar-y
property uchar 4-uchar-z
property uchar 4-uchar-w
property float 4-float-x
property float 4-float-y
property float 4-float-z
property float 4-float-w
end_header
10 10.1 8 9 8.8 9.9 5 6 7 5.5 6.6 7.7 1 2 3 4 1.1 2.2 3.3 4.4
`
	assert.NoError(t, err)
	assert.Equal(t, plyData, buf.String())
}

func TestWriteBinary_PointCloud(t *testing.T) {
	// ARRANGE ================================================================
	writer := ply.MeshWriter{
		Format: ply.BinaryLittleEndian,
		Properties: []ply.PropertyWriter{
			ply.Vector1PropertyWriter{
				ModelAttribute: "1-uchar",
				PlyProperty:    "1-uchar",
				Type:           ply.UChar,
			},
			ply.Vector1PropertyWriter{
				ModelAttribute: "1-float",
				PlyProperty:    "1-float",
				Type:           ply.Float,
			},
			ply.Vector2PropertyWriter{
				ModelAttribute: "2-uchar",
				PlyPropertyX:   "2-uchar-x",
				PlyPropertyY:   "2-uchar-y",
				Type:           ply.UChar,
			},
			ply.Vector2PropertyWriter{
				ModelAttribute: "2-float",
				PlyPropertyX:   "2-float-x",
				PlyPropertyY:   "2-float-y",
				Type:           ply.Float,
			},
			ply.Vector3PropertyWriter{
				ModelAttribute: "3-uchar",
				PlyPropertyX:   "3-uchar-x",
				PlyPropertyY:   "3-uchar-y",
				PlyPropertyZ:   "3-uchar-z",
				Type:           ply.UChar,
			},
			ply.Vector3PropertyWriter{
				ModelAttribute: "3-float",
				PlyPropertyX:   "3-float-x",
				PlyPropertyY:   "3-float-y",
				PlyPropertyZ:   "3-float-z",
				Type:           ply.Float,
			},
			ply.Vector4PropertyWriter{
				ModelAttribute: "4-uchar",
				PlyPropertyX:   "4-uchar-x",
				PlyPropertyY:   "4-uchar-y",
				PlyPropertyZ:   "4-uchar-z",
				PlyPropertyW:   "4-uchar-w",
				Type:           ply.UChar,
			},
			ply.Vector4PropertyWriter{
				ModelAttribute: "4-float",
				PlyPropertyX:   "4-float-x",
				PlyPropertyY:   "4-float-y",
				PlyPropertyZ:   "4-float-z",
				PlyPropertyW:   "4-float-w",
				Type:           ply.Float,
			},
		},
	}

	type point struct {
		Char1  byte
		Float1 float32

		Char2X  byte
		Char2Y  byte
		Float2X float32
		Float2Y float32

		Char3X  byte
		Char3Y  byte
		Char3Z  byte
		Float3X float32
		Float3Y float32
		Float3Z float32

		Char4X  byte
		Char4Y  byte
		Char4Z  byte
		Char4W  byte
		Float4X float32
		Float4Y float32
		Float4Z float32
		Float4W float32
	}

	mesh := modeling.NewPointCloud(
		map[string][]vector4.Float64{
			"4-uchar": {vector4.New(1/255., 2/255., 3/255., 4/255.)},
			"4-float": {vector4.New(1.1, 2.2, 3.3, 4.4)},
		},
		map[string][]vector3.Float64{
			"3-uchar": {vector3.New(5/255., 6/255., 7/255.)},
			"3-float": {vector3.New(5.5, 6.6, 7.7)},
		},
		map[string][]vector2.Float64{
			"2-uchar": {vector2.New(8/255., 9/255.)},
			"2-float": {vector2.New(8.8, 9.9)},
		},
		map[string][]float64{
			"1-uchar": {10 / 255.},
			"1-float": {10.1},
		},
		nil,
	)

	// ACT ====================================================================
	buf := &bytes.Buffer{}
	err := writer.Write(mesh, buf)
	buf = bytes.NewBuffer(buf.Bytes())
	ply.ReadHeader(buf)

	// ASSERT =================================================================
	pt := &point{}
	binary.Read(buf, binary.LittleEndian, pt)
	assert.NoError(t, err)

	assert.Equal(t, uint8(10), pt.Char1)
	assert.Equal(t, uint8(8), pt.Char2X)
	assert.Equal(t, uint8(9), pt.Char2Y)
	assert.Equal(t, uint8(5), pt.Char3X)
	assert.Equal(t, uint8(6), pt.Char3Y)
	assert.Equal(t, uint8(7), pt.Char3Z)
	assert.Equal(t, uint8(1), pt.Char4X)
	assert.Equal(t, uint8(2), pt.Char4Y)
	assert.Equal(t, uint8(3), pt.Char4Z)
	assert.Equal(t, uint8(4), pt.Char4W)

	assert.Equal(t, float32(10.1), pt.Float1)
	assert.Equal(t, float32(8.8), pt.Float2X)
	assert.Equal(t, float32(9.9), pt.Float2Y)
	assert.Equal(t, float32(5.5), pt.Float3X)
	assert.Equal(t, float32(6.6), pt.Float3Y)
	assert.Equal(t, float32(7.7), pt.Float3Z)
	assert.Equal(t, float32(1.1), pt.Float4X)
	assert.Equal(t, float32(2.2), pt.Float4Y)
	assert.Equal(t, float32(3.3), pt.Float4Z)
	assert.Equal(t, float32(4.4), pt.Float4W)
}
func TestWrite_IncludeAllUnspecied(t *testing.T) {
	// ARRANGE ================================================================
	writer := ply.MeshWriter{
		Format:                     ply.BinaryLittleEndian,
		WriteUnspecifiedProperties: true,
	}

	mesh := modeling.NewPointCloud(
		map[string][]vector4.Float64{
			"4-uchar": {vector4.New(1/255., 2/255., 3/255., 4/255.)},
			"4-float": {vector4.New(1.1, 2.2, 3.3, 4.4)},
		},
		map[string][]vector3.Float64{
			"3-uchar": {vector3.New(5/255., 6/255., 7/255.)},
			"3-float": {vector3.New(5.5, 6.6, 7.7)},
		},
		map[string][]vector2.Float64{
			"2-uchar": {vector2.New(8/255., 9/255.)},
			"2-float": {vector2.New(8.8, 9.9)},
		},
		map[string][]float64{
			"1-uchar": {10 / 255.},
			"1-float": {10.1},
		},
		nil,
	)

	// ACT ====================================================================
	buf := &bytes.Buffer{}
	err := writer.Write(mesh, buf)
	buf = bytes.NewBuffer(buf.Bytes())
	meshBack, readErr := ply.ReadMesh(buf)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.NoError(t, readErr)

	assert.True(t, meshBack.HasFloat1Attribute("4-uchar_0"), "mesh should have 4-uchar_0")
	assert.True(t, meshBack.HasFloat1Attribute("4-uchar_1"), "mesh should have 4-uchar_1")
	assert.True(t, meshBack.HasFloat1Attribute("4-uchar_2"), "mesh should have 4-uchar_2")
	assert.True(t, meshBack.HasFloat1Attribute("4-uchar_3"), "mesh should have 4-uchar_3")
	assert.True(t, meshBack.HasFloat1Attribute("4-float_0"), "mesh should have 4-float_0")
	assert.True(t, meshBack.HasFloat1Attribute("4-float_1"), "mesh should have 4-float_1")
	assert.True(t, meshBack.HasFloat1Attribute("4-float_2"), "mesh should have 4-float_2")
	assert.True(t, meshBack.HasFloat1Attribute("4-float_3"), "mesh should have 4-float_3")
	assert.True(t, meshBack.HasFloat1Attribute("3-uchar_0"), "mesh should have 3-uchar_0")
	assert.True(t, meshBack.HasFloat1Attribute("3-uchar_1"), "mesh should have 3-uchar_1")
	assert.True(t, meshBack.HasFloat1Attribute("3-uchar_2"), "mesh should have 3-uchar_2")
	assert.True(t, meshBack.HasFloat1Attribute("3-float_0"), "mesh should have 3-float_0")
	assert.True(t, meshBack.HasFloat1Attribute("3-float_1"), "mesh should have 3-float_1")
	assert.True(t, meshBack.HasFloat1Attribute("3-float_2"), "mesh should have 3-float_2")
	assert.True(t, meshBack.HasFloat1Attribute("2-uchar_0"), "mesh should have 2-uchar_0")
	assert.True(t, meshBack.HasFloat1Attribute("2-uchar_1"), "mesh should have 2-uchar_1")
	assert.True(t, meshBack.HasFloat1Attribute("2-float_0"), "mesh should have 2-float_0")
	assert.True(t, meshBack.HasFloat1Attribute("2-float_1"), "mesh should have 2-float_1")
	assert.True(t, meshBack.HasFloat1Attribute("1-uchar"), "mesh should have 1-uchar")
	assert.True(t, meshBack.HasFloat1Attribute("1-float"), "mesh should have 1-float")
}

func TestWriteBinary_Tri(t *testing.T) {
	// ARRANGE ================================================================
	writer := ply.MeshWriter{
		Format: ply.BinaryLittleEndian,
		Properties: []ply.PropertyWriter{
			ply.Vector3PropertyWriter{
				ModelAttribute: modeling.PositionAttribute,
				PlyPropertyX:   "x",
				PlyPropertyY:   "y",
				PlyPropertyZ:   "z",
				Type:           ply.Float,
			},
		},
	}

	type point struct {
		X float32
		Y float32
		Z float32
	}

	type triangle struct {
		IndiceCount byte
		Indices     [3]uint32
		// UvCount     byte
		// Uvs         [6]float32
	}

	mesh := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: {
				vector3.New(1., 2., 3.),
				vector3.New(4., 5., 6.),
				vector3.New(7., 8., 9.),
			},
		})

	// ACT ====================================================================
	buf := &bytes.Buffer{}
	err := writer.Write(mesh, buf)
	assert.NoError(t, err)

	buf = bytes.NewBuffer(buf.Bytes())
	_, err = ply.ReadHeader(buf)
	assert.NoError(t, err)

	pt := make([]point, 3)
	err = binary.Read(buf, binary.LittleEndian, pt)
	assert.NoError(t, err)

	tri := &triangle{}
	err = binary.Read(buf, binary.LittleEndian, tri)
	assert.NoError(t, err)

	// ASSERT =================================================================
	assert.NoError(t, err)

	assert.Equal(t, float32(1.), pt[0].X)
	assert.Equal(t, float32(2.), pt[0].Y)
	assert.Equal(t, float32(3.), pt[0].Z)

	assert.Equal(t, float32(4.), pt[1].X)
	assert.Equal(t, float32(5.), pt[1].Y)
	assert.Equal(t, float32(6.), pt[1].Z)

	assert.Equal(t, float32(7.), pt[2].X)
	assert.Equal(t, float32(8.), pt[2].Y)
	assert.Equal(t, float32(9.), pt[2].Z)

	assert.Equal(t, byte(3), tri.IndiceCount)
	assert.Equal(t, uint32(0), tri.Indices[0])
	assert.Equal(t, uint32(1), tri.Indices[1])
	assert.Equal(t, uint32(2), tri.Indices[2])
}

func TestWriteBinary_TriWithUVData(t *testing.T) {
	// ARRANGE ================================================================
	writer := ply.MeshWriter{
		Format: ply.BinaryLittleEndian,
		Properties: []ply.PropertyWriter{
			ply.Vector3PropertyWriter{
				ModelAttribute: modeling.PositionAttribute,
				PlyPropertyX:   "x",
				PlyPropertyY:   "y",
				PlyPropertyZ:   "z",
				Type:           ply.Float,
			},
		},
	}

	type point struct {
		X float32
		Y float32
		Z float32
	}

	type triangle struct {
		IndiceCount byte
		Indices     [3]uint32
		UvCount     byte
		Uvs         [6]float32
	}

	mesh := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: {
				vector3.New(1., 2., 3.),
				vector3.New(4., 5., 6.),
				vector3.New(7., 8., 9.),
			},
		}).
		SetFloat2Data(map[string][]vector2.Float64{
			modeling.TexCoordAttribute: {
				vector2.New(1., 2.),
				vector2.New(3., 4.),
				vector2.New(5., 6.),
			},
		})

	// ACT ====================================================================
	buf := &bytes.Buffer{}
	err := writer.Write(mesh, buf)
	assert.NoError(t, err)

	buf = bytes.NewBuffer(buf.Bytes())
	_, err = ply.ReadHeader(buf)
	assert.NoError(t, err)

	pt := make([]point, 3)
	err = binary.Read(buf, binary.LittleEndian, pt)
	assert.NoError(t, err)

	tri := &triangle{}
	err = binary.Read(buf, binary.LittleEndian, tri)
	assert.NoError(t, err)

	// ASSERT =================================================================
	assert.NoError(t, err)

	assert.Equal(t, float32(1.), pt[0].X)
	assert.Equal(t, float32(2.), pt[0].Y)
	assert.Equal(t, float32(3.), pt[0].Z)

	assert.Equal(t, float32(4.), pt[1].X)
	assert.Equal(t, float32(5.), pt[1].Y)
	assert.Equal(t, float32(6.), pt[1].Z)

	assert.Equal(t, float32(7.), pt[2].X)
	assert.Equal(t, float32(8.), pt[2].Y)
	assert.Equal(t, float32(9.), pt[2].Z)

	assert.Equal(t, byte(3), tri.IndiceCount)
	assert.Equal(t, uint32(0), tri.Indices[0])
	assert.Equal(t, uint32(1), tri.Indices[1])
	assert.Equal(t, uint32(2), tri.Indices[2])

	assert.Equal(t, byte(6), tri.UvCount)
	assert.Equal(t, float32(1.), tri.Uvs[0])
	assert.Equal(t, float32(2.), tri.Uvs[1])
	assert.Equal(t, float32(3.), tri.Uvs[2])
	assert.Equal(t, float32(4.), tri.Uvs[3])
	assert.Equal(t, float32(5.), tri.Uvs[4])
	assert.Equal(t, float32(6.), tri.Uvs[5])
}
