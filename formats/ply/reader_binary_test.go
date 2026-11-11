package ply_test

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"testing"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func splatCloud(n int) modeling.Mesh {
	v3 := map[string][]vector3.Float64{
		modeling.PositionAttribute: make([]vector3.Float64, n),
		modeling.ScaleAttribute:    make([]vector3.Float64, n),
		modeling.FDCAttribute:      make([]vector3.Float64, n),
	}
	for j := range 15 {
		v3[modeling.SHAttribute(j)] = make([]vector3.Float64, n)
	}
	rotation := make([]vector4.Float64, n)
	opacity := make([]float64, n)
	for i := range n {
		f := float64(i)
		v3[modeling.PositionAttribute][i] = vector3.New(f, f*0.5, -f)
		v3[modeling.ScaleAttribute][i] = vector3.New(-5., 0.25, 0.125)
		v3[modeling.FDCAttribute][i] = vector3.New(0.5, 0.25, f*0.125)
		for j := range 15 {
			v3[modeling.SHAttribute(j)][i] = vector3.New(f+float64(j), float64(j)*0.5, -0.25)
		}
		rotation[i] = vector4.New(1., 0.5, 0.25, 0.125)
		opacity[i] = f * 0.0625
	}
	return modeling.NewPointCloud(
		map[string][]vector4.Float64{modeling.RotationAttribute: rotation},
		v3,
		nil,
		map[string][]float64{modeling.OpacityAttribute: opacity},
	)
}

func TestBinaryRoundTripSpansManyChunks(t *testing.T) {
	original := splatCloud(50_000)
	buf := &bytes.Buffer{}
	require.NoError(t, ply.Artifact{Mesh: original}.Write(buf))

	read, err := ply.ReadMesh(buf)

	require.NoError(t, err)
	require.Equal(t, original.AttributeLength(), read.AttributeLength())
	for _, attr := range original.Float3Attributes() {
		want, got := original.Float3Attribute(attr), read.Float3Attribute(attr)
		for i := 0; i < want.Len(); i++ {
			if want.At(i) != got.At(i) {
				t.Fatalf("%s differs at %d: %v vs %v", attr, i, want.At(i), got.At(i))
			}
		}
	}
	for i := 0; i < original.AttributeLength(); i++ {
		if original.Float4Attribute(modeling.RotationAttribute).At(i) != read.Float4Attribute(modeling.RotationAttribute).At(i) {
			t.Fatalf("rotation differs at %d", i)
		}
		if original.Float1Attribute(modeling.OpacityAttribute).At(i) != read.Float1Attribute(modeling.OpacityAttribute).At(i) {
			t.Fatalf("opacity differs at %d", i)
		}
	}
}

func TestBinaryReadOfTruncatedVertexDataFails(t *testing.T) {
	buf := &bytes.Buffer{}
	require.NoError(t, ply.Artifact{Mesh: splatCloud(20_000)}.Write(buf))
	truncated := buf.Bytes()[:buf.Len()-1000]

	_, err := ply.ReadMesh(bytes.NewReader(truncated))

	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
}

func TestBinaryReadUnsupportedScalarTypePanicsOnTheCaller(t *testing.T) {
	header := ply.Header{
		Format: ply.BinaryLittleEndian,
		Elements: []ply.Element{{
			Name:  ply.VertexElementName,
			Count: 2,
			Properties: []ply.Property{
				ply.ScalarProperty{PropertyName: "x", Type: ply.Short},
				ply.ScalarProperty{PropertyName: "y", Type: ply.Short},
				ply.ScalarProperty{PropertyName: "z", Type: ply.Short},
			},
		}},
	}
	buf := &bytes.Buffer{}
	require.NoError(t, header.Write(buf))
	require.NoError(t, binary.Write(buf, binary.LittleEndian, []int16{1, 2, 3, 4, 5, 6}))

	assert.PanicsWithError(t, "unimplemented short", func() {
		_, _ = ply.ReadMesh(buf)
	})
}

func TestBinaryWriteUnsupportedScalarTypePanicsOnTheCaller(t *testing.T) {
	writer := ply.MeshWriter{
		Format: ply.BinaryLittleEndian,
		Properties: []ply.PropertyWriter{ply.Vector3PropertyWriter{
			ModelAttribute: modeling.PositionAttribute,
			Type:           ply.Short,
			PlyPropertyX:   "x",
			PlyPropertyY:   "y",
			PlyPropertyZ:   "z",
		}},
	}
	mesh := modeling.NewPointCloud(nil, map[string][]vector3.Float64{
		modeling.PositionAttribute: {vector3.New(1., 2, 3)},
	}, nil, nil)

	assert.PanicsWithError(t, "unimplemented short", func() {
		_ = writer.Write(mesh, "", io.Discard)
	})
}

func TestBinaryWriteBigEndianMatchesEncoding(t *testing.T) {
	mesh := modeling.NewPointCloud(nil, map[string][]vector3.Float64{
		modeling.PositionAttribute: {vector3.New(1., 2, 3), vector3.New(-4., 5.5, 6)},
	}, nil, nil)
	buf := &bytes.Buffer{}

	require.NoError(t, ply.Write(buf, mesh, ply.BinaryBigEndian, ""))

	body := buf.Bytes()[buf.Len()-24:]
	for i, want := range []float32{1, 2, 3, -4, 5.5, 6} {
		assert.Equal(t, want, math.Float32frombits(binary.BigEndian.Uint32(body[i*4:])))
	}
}

func BenchmarkReadMeshBinary(b *testing.B) {
	buf := &bytes.Buffer{}
	if err := (ply.Artifact{Mesh: splatCloud(1_000_000)}).Write(buf); err != nil {
		b.Fatal(err)
	}
	data := buf.Bytes()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for range b.N {
		if _, err := ply.ReadMesh(bytes.NewReader(data)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteASCII(b *testing.B) {
	mesh := splatCloud(200_000)
	b.ResetTimer()
	for range b.N {
		if err := ply.Write(io.Discard, mesh, ply.ASCII, ""); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteBinary(b *testing.B) {
	artifact := ply.Artifact{Mesh: splatCloud(1_000_000)}
	b.ResetTimer()
	for range b.N {
		if err := artifact.Write(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}
