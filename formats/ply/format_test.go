package ply_test

import (
	"testing"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/stretchr/testify/assert"
)

func TestFormat_String(t *testing.T) {
	assert.Equal(t, "ASCII", ply.ASCII.String())
	assert.Equal(t, "Binary Big Endian", ply.BinaryBigEndian.String())
	assert.Equal(t, "Binary Little Endian", ply.BinaryLittleEndian.String())

	assert.PanicsWithError(t, "unrecognized format Test", func() {
		ply.Format("Test").String()
	})
}
