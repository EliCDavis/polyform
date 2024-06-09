package potree_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/polyform/formats/potree"
	"github.com/stretchr/testify/assert"
)

var metadataJson = []byte(`{
	"version": "2.0",
	"name": "heidentor",
	"description": "",
	"points": 25836417,
	"projection": "",
	"hierarchy": {
		"firstChunkSize": 18502, 
		"stepSize": 4, 
		"depth": 7
	},
	"offset": [-8.0960000000000001, -4.7999999999999998, 1.6870000000000001],
	"scale": [0.001, 0.001, 0.001],
	"spacing": 0.12154687500000001,
	"boundingBox": {
		"min": [-8.0960000000000001, -4.7999999999999998, 1.6870000000000001], 
		"max": [7.4620000000000015, 10.758000000000003, 17.245000000000001]
	},
	"encoding": "BROTLI",
	"attributes": [
		{
			"name": "position",
			"description": "",
			"size": 12,
			"numElements": 3,
			"elementSize": 4,
			"type": "int32",
			"min": [-8.0960000000000001, -4.7990000000000004, 1.6879999999999999],
			"max": [0.45300000000000001, 10.758000000000001, 15.781000000000001]
		},{
			"name": "intensity",
			"description": "",
			"size": 2,
			"numElements": 1,
			"elementSize": 2,
			"type": "uint16",
			"min": [0],
			"max": [0]
		},{
			"name": "return number",
			"description": "",
			"size": 1,
			"numElements": 1,
			"elementSize": 1,
			"type": "uint8",
			"min": [0],
			"max": [0]
		},{
			"name": "number of returns",
			"description": "",
			"size": 1,
			"numElements": 1,
			"elementSize": 1,
			"type": "uint8",
			"min": [0],
			"max": [0]
		},{
			"name": "classification",
			"description": "",
			"size": 1,
			"numElements": 1,
			"elementSize": 1,
			"type": "uint8",
			"min": [0],
			"max": [0]
		},{
			"name": "scan angle rank",
			"description": "",
			"size": 1,
			"numElements": 1,
			"elementSize": 1,
			"type": "uint8",
			"min": [0],
			"max": [0]
		},{
			"name": "user data",
			"description": "",
			"size": 1,
			"numElements": 1,
			"elementSize": 1,
			"type": "uint8",
			"min": [0],
			"max": [0]
		},{
			"name": "point source id",
			"description": "",
			"size": 2,
			"numElements": 1,
			"elementSize": 2,
			"type": "uint16",
			"min": [0],
			"max": [0]
		},{
			"name": "rgb",
			"description": "",
			"size": 6,
			"numElements": 3,
			"elementSize": 2,
			"type": "uint16",
			"min": [0, 0, 0],
			"max": [65280, 65280, 65280]
		}
	]
}`)

func TestReadMetadata(t *testing.T) {
	// Dataset from potree.org
	// https://potree.org/potree/examples/vr_heidentor.html

	// ACT ====================================================================
	metadata, err := potree.ReadMetadata(bytes.NewReader(metadataJson))

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, "2.0", metadata.Version)
	assert.Equal(t, "heidentor", metadata.Name)
	assert.Equal(t, "", metadata.Description)
	assert.Equal(t, int64(25836417), metadata.Points)
	assert.Equal(t, "", metadata.Projection)
	assert.Equal(t, "BROTLI", metadata.Encoding)

	// Hierarchy
	assert.Equal(t, uint64(18502), metadata.Hierarchy.FirstChunkSize)
	assert.Equal(t, 4, metadata.Hierarchy.StepSize)
	assert.Equal(t, 7, metadata.Hierarchy.Depth)

	// Attributes
	assert.Len(t, metadata.Attributes, 9)
	assert.Equal(t, 27, metadata.BytesPerPoint())
	assert.Equal(t, 0, metadata.AttributeOffset("position"))
	assert.Equal(t, 12, metadata.AttributeOffset("intensity"))
	assert.Equal(t, -1, metadata.AttributeOffset("blah"))
}

func TestMetadataReadHierarchy(t *testing.T) {
	// ARRANGE ================================================================
	metadata, err := potree.ReadMetadata(bytes.NewReader(metadataJson))
	assert.NoError(t, err)

	// ACT ====================================================================
	tree, err := metadata.LoadHierarchy("hierarchy.bin")

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.NotNil(t, tree)
	assert.Equal(t, 7, tree.Height())
	assert.Equal(t, 8881, tree.DescendentCount())
	assert.Equal(t, uint64(25836417), tree.PointCount())
	assert.Equal(t, 20640, tree.MaxPointCount())
}
