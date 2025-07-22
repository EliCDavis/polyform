package variable_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestParameterNodeSwaggerProperty(t *testing.T) {
	tests := map[string]struct {
		input      variable.Variable
		propType   swagger.PropertyType
		propFormat swagger.PropertyFormat
		ref        any
		items      any
	}{
		"basic string parameter": {
			input:      &variable.TypeVariable[string]{},
			propType:   swagger.StringPropertyType,
			propFormat: "",
		},
		"float64 parameter": {
			input:      &variable.TypeVariable[float64]{},
			propType:   swagger.NumberPropertyType,
			propFormat: swagger.DoublePropertyFormat,
		},
		"float32 parameter": {
			input:      &variable.TypeVariable[float32]{},
			propType:   swagger.NumberPropertyType,
			propFormat: swagger.FloatPropertyFormat,
		},
		"int parameter": {
			input:    &variable.TypeVariable[int]{},
			propType: swagger.IntegerPropertyType,
		},
		"int32 parameter": {
			input:      &variable.TypeVariable[int32]{},
			propType:   swagger.IntegerPropertyType,
			propFormat: swagger.Int32PropertyFormat,
		},
		"int64 parameter": {
			input:      &variable.TypeVariable[int64]{},
			propType:   swagger.IntegerPropertyType,
			propFormat: swagger.Int64PropertyFormat,
		},
		"bool parameter": {
			input:    &variable.TypeVariable[bool]{},
			propType: swagger.BooleanPropertyType,
		},
		"vector2 parameter": {
			input: &variable.TypeVariable[vector2.Float64]{},
			ref:   "#/definitions/Float2",
		},
		"vector3 parameter": {
			input: &variable.TypeVariable[vector3.Float64]{},
			ref:   "#/definitions/Float3",
		},
		"aabb parameter": {
			input: &variable.TypeVariable[geometry.AABB]{},
			ref:   "#/definitions/AABB",
		},
		"color": {
			input:      &variable.TypeVariable[coloring.WebColor]{},
			propType:   swagger.StringPropertyType,
			propFormat: "color",
		},
		"vector3 array parameter": {
			input:    &variable.TypeVariable[[]vector3.Float64]{},
			propType: swagger.ArrayPropertyType,
			items: map[string]any{
				"$ref": "#/definitions/Vector3",
			},
		},
		"vector2 array parameter": {
			input:    &variable.TypeVariable[[]vector2.Float64]{},
			propType: swagger.ArrayPropertyType,
			items: map[string]any{
				"$ref": "#/definitions/Vector2",
			},
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			prop := test.input.SwaggerProperty()
			assert.Equal(t, test.propFormat, prop.Format)
			assert.Equal(t, test.propType, prop.Type)
			assert.Equal(t, test.ref, prop.Ref)
			assert.Equal(t, test.items, prop.Items)
		})
	}
}
