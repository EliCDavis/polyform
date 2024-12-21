package ply_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
)

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
	}

	inputData := []Point{
		{x: 1, y: 2, z: 3, nx: 1, ny: 2, nz: 3, r: 1, g: 2, b: 3, a: 255, s: 10, t: 20, opacity: 10},
		{x: 4, y: 5, z: 6, nx: 4, ny: 5, nz: 6, r: 4, g: 5, b: 6, a: 255, s: 30, t: 40, opacity: 20},
	}

	buf := &bytes.Buffer{}
	header.Write(buf)
	binary.Write(buf, binary.BigEndian, inputData)

	// ACT ====================================================================
	mesh, err := ply.ReadMesh2(buf)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.True(t, mesh.HasFloat3Attribute(modeling.PositionAttribute), "mesh should have position attribute")
	assert.True(t, mesh.HasFloat3Attribute(modeling.NormalAttribute), "mesh should have normal attribute")
	assert.True(t, mesh.HasFloat4Attribute(modeling.ColorAttribute), "mesh should have color attribute")
	assert.True(t, mesh.HasFloat2Attribute(modeling.TexCoordAttribute), "mesh should have texcoord attribute")
	assert.True(t, mesh.HasFloat1Attribute(modeling.OpacityAttribute), "mesh should have opacity attribute")

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
}
