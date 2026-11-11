package ply_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func splatWithSH(coefficients int) modeling.Mesh {
	v3 := map[string][]vector3.Float64{
		modeling.PositionAttribute: {vector3.New(1., 2., 3.), vector3.New(4., 5., 6.)},
		modeling.ScaleAttribute:    {vector3.New(1., 1., 1.), vector3.New(2., 2., 2.)},
		modeling.FDCAttribute:      {vector3.New(0.5, 0.25, 0.125), vector3.New(1., 2., 3.)},
	}
	for j := 0; j < coefficients; j++ {
		f := float64(j)
		v3[modeling.SHAttribute(j)] = []vector3.Float64{
			vector3.New(f, f+0.25, f+0.5),
			vector3.New(-f, -f-0.25, -f-0.5),
		}
	}
	return modeling.NewPointCloud(
		map[string][]vector4.Float64{
			modeling.RotationAttribute: {vector4.New(1., 0., 0., 0.), vector4.New(0., 1., 0., 0.)},
		},
		v3,
		nil,
		map[string][]float64{modeling.OpacityAttribute: {0.5, 1}},
	)
}

func shPropertyNames(header ply.Header) []string {
	names := make([]string, 0)
	for _, prop := range header.Elements[0].Properties {
		if len(prop.Name()) > 7 && prop.Name()[:7] == "f_rest_" {
			names = append(names, prop.Name())
		}
	}
	return names
}

func TestArtifactWriteReadRoundTripsSH(t *testing.T) {
	original := splatWithSH(15)

	buf := &bytes.Buffer{}
	require.NoError(t, ply.Artifact{Mesh: original}.Write(buf))

	header, err := ply.ReadHeader(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	assert.Len(t, shPropertyNames(header), 45)

	mesh, err := ply.ReadMesh(buf)
	require.NoError(t, err)

	assert.Equal(t, 15, modeling.SHCoefficientCount(*mesh))
	assert.False(t, mesh.HasFloat1Attribute("f_rest_0"))
	for j := 0; j < 15; j++ {
		want := original.Float3Attribute(modeling.SHAttribute(j))
		got := mesh.Float3Attribute(modeling.SHAttribute(j))
		assert.Equal(t, want.At(0), got.At(0), "coefficient %d", j)
		assert.Equal(t, want.At(1), got.At(1), "coefficient %d", j)
	}
	assert.Equal(t, vector3.New(0.5, 0.25, 0.125), mesh.Float3Attribute(modeling.FDCAttribute).At(0))
	assert.Equal(t, vector4.New(0., 1., 0., 0.), mesh.Float4Attribute(modeling.RotationAttribute).At(1))
}

func TestWriteASCIIRoundTripsSH(t *testing.T) {
	original := splatWithSH(8)

	buf := &bytes.Buffer{}
	require.NoError(t, ply.Write(buf, original, ply.ASCII, ""))

	mesh, err := ply.ReadMesh(buf)
	require.NoError(t, err)

	assert.Equal(t, 8, modeling.SHCoefficientCount(*mesh))
	for j := 0; j < 8; j++ {
		want := original.Float3Attribute(modeling.SHAttribute(j))
		got := mesh.Float3Attribute(modeling.SHAttribute(j))
		assert.Equal(t, want.At(0), got.At(0), "coefficient %d", j)
		assert.Equal(t, want.At(1), got.At(1), "coefficient %d", j)
	}
}

func TestWriteEmitsOnlyThePresentBands(t *testing.T) {
	buf := &bytes.Buffer{}
	require.NoError(t, ply.Write(buf, splatWithSH(3), ply.BinaryLittleEndian, ""))

	header, err := ply.ReadHeader(buf)
	require.NoError(t, err)

	names := shPropertyNames(header)
	assert.Len(t, names, 9)
	for i, name := range names {
		assert.Equal(t, fmt.Sprintf("f_rest_%d", i), name)
	}
	for _, prop := range header.Elements[0].Properties {
		assert.NotContains(t, prop.Name(), "SH_")
	}
}

func TestWriteWithoutSHEmitsNoRestProperties(t *testing.T) {
	buf := &bytes.Buffer{}
	require.NoError(t, ply.Write(buf, splatWithSH(0), ply.BinaryLittleEndian, ""))

	header, err := ply.ReadHeader(buf)
	require.NoError(t, err)
	assert.Empty(t, shPropertyNames(header))
}

func TestReadSHIsChannelMajor(t *testing.T) {
	props := []ply.Property{
		ply.ScalarProperty{PropertyName: "x", Type: ply.Float},
		ply.ScalarProperty{PropertyName: "y", Type: ply.Float},
		ply.ScalarProperty{PropertyName: "z", Type: ply.Float},
	}
	for i := 0; i < 9; i++ {
		props = append(props, ply.ScalarProperty{PropertyName: fmt.Sprintf("f_rest_%d", i), Type: ply.Float})
	}
	header := ply.Header{
		Format: ply.BinaryLittleEndian,
		Elements: []ply.Element{{
			Name:       ply.VertexElementName,
			Count:      1,
			Properties: props,
		}},
	}

	buf := &bytes.Buffer{}
	require.NoError(t, header.Write(buf))
	row := []float32{0, 0, 0}
	for i := 0; i < 9; i++ {
		row = append(row, float32(i)*10)
	}
	require.NoError(t, binary.Write(buf, binary.LittleEndian, row))

	mesh, err := ply.ReadMesh(buf)
	require.NoError(t, err)

	assert.Equal(t, 3, modeling.SHCoefficientCount(*mesh))
	assert.Equal(t, vector3.New(0., 30., 60.), mesh.Float3Attribute(modeling.SHAttribute(0)).At(0))
	assert.Equal(t, vector3.New(10., 40., 70.), mesh.Float3Attribute(modeling.SHAttribute(1)).At(0))
	assert.Equal(t, vector3.New(20., 50., 80.), mesh.Float3Attribute(modeling.SHAttribute(2)).At(0))
	assert.False(t, mesh.HasFloat1Attribute("f_rest_0"))
}

func TestReadIncompleteRestPropertiesFallBackToScalars(t *testing.T) {
	header := ply.Header{
		Format: ply.BinaryLittleEndian,
		Elements: []ply.Element{{
			Name:  ply.VertexElementName,
			Count: 1,
			Properties: []ply.Property{
				ply.ScalarProperty{PropertyName: "x", Type: ply.Float},
				ply.ScalarProperty{PropertyName: "y", Type: ply.Float},
				ply.ScalarProperty{PropertyName: "z", Type: ply.Float},
				ply.ScalarProperty{PropertyName: "f_rest_0", Type: ply.Float},
				ply.ScalarProperty{PropertyName: "f_rest_1", Type: ply.Float},
			},
		}},
	}

	buf := &bytes.Buffer{}
	require.NoError(t, header.Write(buf))
	require.NoError(t, binary.Write(buf, binary.LittleEndian, []float32{0, 0, 0, 7, 8}))

	mesh, err := ply.ReadMesh(buf)
	require.NoError(t, err)

	assert.Equal(t, 0, modeling.SHCoefficientCount(*mesh))
	assert.Equal(t, 7., mesh.Float1Attribute("f_rest_0").At(0))
	assert.Equal(t, 8., mesh.Float1Attribute("f_rest_1").At(0))
}
