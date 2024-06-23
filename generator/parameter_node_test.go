package generator_test

import (
	"testing"
	"time"

	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/stretchr/testify/assert"
)

func TestParameterNodeSwaggerProperty(t *testing.T) {
	tests := map[string]struct {
		input      generator.SwaggerParameter
		propType   swagger.PropertyType
		propFormat swagger.PropertyFormat
	}{
		"basic string parameter": {
			input:      &parameter.Value[string]{},
			propType:   swagger.StringPropertyType,
			propFormat: "",
		},
		"date string parameter": {
			input:      &parameter.Value[time.Time]{},
			propType:   swagger.StringPropertyType,
			propFormat: swagger.DateTimePropertyFormat,
		},
		"float64 parameter": {
			input:      &parameter.Float64{},
			propType:   swagger.NumberPropertyType,
			propFormat: swagger.DoublePropertyFormat,
		},
		"float32 parameter": {
			input:      &parameter.Value[float32]{},
			propType:   swagger.NumberPropertyType,
			propFormat: swagger.FloatPropertyFormat,
		},
		"int parameter": {
			input:    &parameter.Int{},
			propType: swagger.IntegerPropertyType,
		},
		"int32 parameter": {
			input:      &parameter.Value[int32]{},
			propType:   swagger.IntegerPropertyType,
			propFormat: swagger.Int32PropertyFormat,
		},
		"int64 parameter": {
			input:      &parameter.Value[int64]{},
			propType:   swagger.IntegerPropertyType,
			propFormat: swagger.Int64PropertyFormat,
		},
		"bool parameter": {
			input:    &parameter.Bool{},
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
