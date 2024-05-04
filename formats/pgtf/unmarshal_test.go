package pgtf_test

import (
	"testing"

	"github.com/EliCDavis/polyform/formats/pgtf"
	"github.com/stretchr/testify/assert"
)

func TestParseJsonWithBuffers(t *testing.T) {

	// ARRANGE ================================================================
	type TestStruct struct {
		A string
		B *pgtfSerializableStruct
	}

	json := []byte(`{
		"A": "Something",
		"$B": 0  
	}`)

	buffers := []pgtf.Buffer{
		{
			ByteLength: 4,
			URI:        "data:application/octet-stream;base64,OTAAAA==",
		},
	}

	bufferViews := []pgtf.BufferView{
		{
			Buffer:     0,
			ByteOffset: 0,
			ByteLength: 4,
		},
	}

	// ACT ====================================================================
	result, err := pgtf.ParseJsonUsingBuffers[TestStruct](buffers, bufferViews, json)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, "Something", result.A)
	assert.Equal(t, int32(12345), result.B.data)
}
