package variant_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/variant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalDimension(t *testing.T) {
	tests := map[string]struct {
		json      string
		wantCount int
		wantValue string
	}{
		"discrete": {
			json:      `{"type":"discrete","data":{"values":[1,2,3]}}`,
			wantCount: 3,
			wantValue: `1`,
		},
		"numeric range": {
			json:      `{"type":"numericRange","data":{"min":0,"max":10,"samples":5}}`,
			wantCount: 5,
			wantValue: `0`,
		},
		"vector2 range": {
			json:      `{"type":"vector2Range","data":{"min":{"x":0,"y":0},"max":{"x":1,"y":1},"samples":{"x":2,"y":2}}}`,
			wantCount: 4,
			wantValue: `{"x":0,"y":0}`,
		},
		"vector3 range": {
			json:      `{"type":"vector3Range","data":{"min":{"x":0,"y":0,"z":0},"max":{"x":1,"y":1,"z":1},"samples":{"x":2,"y":2,"z":2}}}`,
			wantCount: 8,
			wantValue: `{"x":0,"y":0,"z":0}`,
		},
		"rgb range": {
			json:      `{"type":"rgbRange","data":{"min":{"x":0,"y":0,"z":0},"max":{"x":1,"y":1,"z":1},"samples":{"x":2,"y":2,"z":2}}}`,
			wantCount: 8,
			wantValue: `"#000000"`,
		},
		"hsv range": {
			json:      `{"type":"hsvRange","data":{"min":{"x":0,"y":1,"z":1},"max":{"x":0,"y":1,"z":1},"samples":{"x":1,"y":1,"z":1}}}`,
			wantCount: 1,
			wantValue: `"#ff0000"`,
		},
		"combination": {
			json:      `{"type":"combination","data":{"dimensions":[{"type":"discrete","data":{"values":[1]}},{"type":"numericRange","data":{"min":0,"max":1,"samples":2}}]}}`,
			wantCount: 3,
			wantValue: `1`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			dim, err := variant.UnmarshalDimension("Path", []byte(test.json))
			require.NoError(t, err)
			assert.Equal(t, "Path", dim.Path())
			assert.Equal(t, test.wantCount, dim.Count())

			value, err := dim.Value(0)
			require.NoError(t, err)
			assert.JSONEq(t, test.wantValue, string(value))
		})
	}
}

func TestUnmarshalDimensionRejectsUnknownType(t *testing.T) {
	_, err := variant.UnmarshalDimension("x", []byte(`{"type":"not-a-real-type"}`))
	assert.Error(t, err)
}

func TestUnmarshalDimensionRejectsMalformedEnvelope(t *testing.T) {
	_, err := variant.UnmarshalDimension("x", []byte(`not json`))
	assert.Error(t, err)
}

func TestUnmarshalDimensionRejectsMalformedData(t *testing.T) {
	_, err := variant.UnmarshalDimension("x", []byte(`{"type":"numericRange","data":"not an object"}`))
	assert.Error(t, err)
}
