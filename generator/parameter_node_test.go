package generator_test

import (
	"testing"
	"time"

	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/EliCDavis/polyform/generator"
	"github.com/stretchr/testify/assert"
)

func TestParameterNodeSwaggerProperty(t *testing.T) {
	tests := map[string]struct {
		input      generator.SwaggerParameter
		propType   swagger.PropertyType
		propFormat swagger.PropertyFormat
	}{
		"basic string parameter": {
			input:      &generator.ParameterNode[string]{},
			propType:   swagger.StringPropertyType,
			propFormat: "",
		},
		"date string parameter": {
			input:      &generator.ParameterNode[time.Time]{},
			propType:   swagger.StringPropertyType,
			propFormat: swagger.DateTimePropertyFormat,
		},
		"float64 parameter": {
			input:      &generator.ParameterNode[float64]{},
			propType:   swagger.NumberPropertyType,
			propFormat: swagger.DoublePropertyFormat,
		},
		"float32 parameter": {
			input:      &generator.ParameterNode[float32]{},
			propType:   swagger.NumberPropertyType,
			propFormat: swagger.FloatPropertyFormat,
		},
		"int parameter": {
			input:    &generator.ParameterNode[int]{},
			propType: swagger.IntegerPropertyType,
		},
		"int32 parameter": {
			input:      &generator.ParameterNode[int32]{},
			propType:   swagger.IntegerPropertyType,
			propFormat: swagger.Int32PropertyFormat,
		},
		"int64 parameter": {
			input:      &generator.ParameterNode[int64]{},
			propType:   swagger.IntegerPropertyType,
			propFormat: swagger.Int64PropertyFormat,
		},
		"bool parameter": {
			input:    &generator.ParameterNode[bool]{},
			propType: swagger.BooleanPropertyType,
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			prop := test.input.SwaggerProperty()
			assert.Equal(t, test.propFormat, prop.Format)
			assert.Equal(t, test.propType, prop.Type)
		})
	}
}
