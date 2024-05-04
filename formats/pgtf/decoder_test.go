package pgtf_test

import (
	"testing"

	"github.com/EliCDavis/polyform/formats/pgtf"
	"github.com/stretchr/testify/assert"
)

func TestDecoder(t *testing.T) {
	jsonData := `{
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
	}`

	decoder, err := pgtf.NewDecoder([]byte(jsonData))
	assert.NoError(t, err)

	v, err := pgtf.Decode[embeddedBinaryStruct](decoder, []byte(`{
		"$Serializable": 0,
		"Basic": {
			"A": 12,
			"B": true,
			"C": "yee haw"
		}
	}`))

	assert.NoError(t, err)
	assert.Equal(t, 12, v.Basic.A)
	assert.Equal(t, true, v.Basic.B)
	assert.Equal(t, "yee haw", v.Basic.C)
	assert.NotNil(t, v.Serializable)
	assert.Equal(t, int32(12345), v.Serializable.data)

	v, err = pgtf.Decode[embeddedBinaryStruct](decoder, []byte(`{
		"$Serializable": 1,
		"Basic": {
			"A": 10,
			"B": false,
			"C": "yee naw"
		}
	}`))

	assert.NoError(t, err)
	assert.Equal(t, 10, v.Basic.A)
	assert.Equal(t, false, v.Basic.B)
	assert.Equal(t, "yee naw", v.Basic.C)
	assert.NotNil(t, v.Serializable)
	assert.Equal(t, int32(67890), v.Serializable.data)

}
