package spz_test

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"testing"

	"github.com/EliCDavis/polyform/formats/spz"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadStoresSHAsOneFloat3PerCoefficient(t *testing.T) {
	raw := &bytes.Buffer{}
	require.NoError(t, binary.Write(raw, binary.LittleEndian, spz.Header{
		Magic:          0x5053474e,
		Version:        2,
		NumPoints:      1,
		ShDegree:       1,
		FractionalBits: 12,
	}))
	raw.Write(make([]byte, 9)) // position
	raw.Write([]byte{255})     // alpha
	raw.Write([]byte{0, 0, 0}) // color
	raw.Write([]byte{0, 0, 0}) // scale
	raw.Write([]byte{0, 0, 0}) // rotation
	raw.Write([]byte{128, 192, 64, 0, 255, 128, 160, 96, 128})

	compressed := &bytes.Buffer{}
	zw := gzip.NewWriter(compressed)
	_, err := zw.Write(raw.Bytes())
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	cloud, err := spz.Read(compressed)
	require.NoError(t, err)

	mesh := cloud.Mesh
	assert.Equal(t, 3, modeling.SHCoefficientCount(mesh))
	assert.Equal(t, vector3.New(0., 0.5, -0.5), mesh.Float3Attribute(modeling.SHAttribute(0)).At(0))
	assert.Equal(t, vector3.New(-1., 127./128., 0.), mesh.Float3Attribute(modeling.SHAttribute(1)).At(0))
	assert.Equal(t, vector3.New(0.25, -0.25, 0.), mesh.Float3Attribute(modeling.SHAttribute(2)).At(0))
}
