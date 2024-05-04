package pgtf_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/formats/pgtf"
	"github.com/stretchr/testify/assert"
)

func TestEncoder(t *testing.T) {
	type testStruct struct {
		A json.RawMessage
		B json.RawMessage
	}

	encoder := &pgtf.Encoder{}

	ts := testStruct{}

	// ACT ====================================================================
	aData, aErr := encoder.Marshal(&embeddedBinaryStruct{Serializable: &pgtfSerializableStruct{data: 12345}})
	bData, bErr := encoder.Marshal(&embeddedBinaryStruct{Serializable: &pgtfSerializableStruct{data: 67890}})

	ts.A = aData
	ts.B = bData

	finalJson, jErr := encoder.ToPgtf(ts)

	// ASSERT =================================================================

	assert.NoError(t, aErr)
	assert.NoError(t, bErr)
	assert.NoError(t, jErr)

	assert.Equal(t, `{
	"buffers": [
		{
			"byteLength": 8,
			"uri": "data:application/octet-stream;base64,OTAAADIJAQA="
		}
	],
	"bufferViews": [
		{
			"buffer": 0,
			"byteLength": 4
		},
		{
			"buffer": 0,
			"byteOffset": 4,
			"byteLength": 4
		}
	],
	"data": {
		"A": {
			"$Serializable": 0,
			"Basic": {
				"A": 0,
				"B": false,
				"C": ""
			}
		},
		"B": {
			"$Serializable": 1,
			"Basic": {
				"A": 0,
				"B": false,
				"C": ""
			}
		}
	}
}`, string(finalJson))
}

func TestEncoder_MultipleBuffers(t *testing.T) {
	type testStruct struct {
		A json.RawMessage
		B json.RawMessage
	}

	encoder := &pgtf.Encoder{}

	ts := testStruct{}

	// ACT ====================================================================
	aData, aErr := encoder.Marshal(&embeddedBinaryStruct{Serializable: &pgtfSerializableStruct{data: 12345}})
	encoder.StartNewBuffer()
	bData, bErr := encoder.Marshal(&embeddedBinaryStruct{Serializable: &pgtfSerializableStruct{data: 67890}})

	ts.A = aData
	ts.B = bData

	finalJson, jErr := encoder.ToPgtf(ts)

	// ASSERT =================================================================

	assert.NoError(t, aErr)
	assert.NoError(t, bErr)
	assert.NoError(t, jErr)

	assert.Equal(t, `{
	"buffers": [
		{
			"byteLength": 4,
			"uri": "data:application/octet-stream;base64,OTAAAA=="
		},
		{
			"byteLength": 4,
			"uri": "data:application/octet-stream;base64,MgkBAA=="
		}
	],
	"bufferViews": [
		{
			"buffer": 0,
			"byteLength": 4
		},
		{
			"buffer": 1,
			"byteLength": 4
		}
	],
	"data": {
		"A": {
			"$Serializable": 0,
			"Basic": {
				"A": 0,
				"B": false,
				"C": ""
			}
		},
		"B": {
			"$Serializable": 1,
			"Basic": {
				"A": 0,
				"B": false,
				"C": ""
			}
		}
	}
}`, string(finalJson))
}
