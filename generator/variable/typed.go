package variable

import (
	"encoding/json"
	"fmt"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

type TypeVariable[T any] struct {
	value   T
	version int
	info    Info
}

func (tv *TypeVariable[T]) SetValue(v T) {
	tv.value = v
	tv.version++
}

func (tv *TypeVariable[T]) GetValue() T {
	return tv.value
}

func (tv *TypeVariable[T]) Version() int {
	return tv.version
}

func (tv *TypeVariable[T]) currentValue() any {
	return tv.value
}

func (tv *TypeVariable[T]) currentVersion() int {
	return tv.version
}

func (tv *TypeVariable[T]) Info() Info {
	return tv.info
}

func (tv *TypeVariable[T]) setInfo(i Info) error {
	if tv.info != nil {
		return fmt.Errorf("already assigned info")
	}
	tv.info = i
	return nil
}

func (tv *TypeVariable[T]) ApplyMessage(msg []byte) (bool, error) {
	var val T
	err := json.Unmarshal(msg, &val)
	if err != nil {
		return false, err
	}

	tv.version++
	tv.value = val
	return true, nil
}

func (tv *TypeVariable[T]) applyProfile(profile json.RawMessage) error {
	var val T
	err := json.Unmarshal(profile, &val)
	if err != nil {
		return err
	}

	tv.version++
	tv.value = val
	return nil
}

func (tv *TypeVariable[T]) getProfile() json.RawMessage {
	return tv.ToMessage()
}

func (tv TypeVariable[T]) ToMessage() []byte {
	data, err := json.Marshal(tv.value)
	if err != nil {
		panic(err)
	}
	return data
}

func (tv *TypeVariable[T]) NodeReference() nodes.Node {
	return &VariableReferenceNode[T]{
		variable: tv,
	}
}

func (tv TypeVariable[T]) MarshalJSON() ([]byte, error) {
	var t T
	return json.Marshal(typedVariableSchema[T]{
		variableSchemaBase: variableSchemaBase{
			Type: refutil.GetTypeName(t),
		},
		Value: tv.value,
	})
}

func (tv TypeVariable[T]) runtimeSchema() schema.RuntimeVariable {
	var t T
	return schema.RuntimeVariable{
		Description: tv.info.Description(),
		Type:        refutil.GetTypeName(t),
		Value:       tv.value,
	}
}

func (tv TypeVariable[T]) toPersistantJSON(encoder *jbtf.Encoder) ([]byte, error) {
	return json.Marshal(tv)
}

type typedVariableSchema[T any] struct {
	variableSchemaBase
	Value T             `json:"value"`
	CLI   *cliConfig[T] `json:"cli,omitempty"`
}

func (tv *TypeVariable[T]) fromPersistantJSON(decoder jbtf.Decoder, body []byte) error {
	vsb := &typedVariableSchema[T]{}
	err := json.Unmarshal(body, vsb)
	if err != nil {
		return err
	}
	tv.value = vsb.Value
	return nil
}

func (tv *TypeVariable[T]) SwaggerProperty() swagger.Property {

	prop := swagger.Property{}

	var t T
	switch refutil.GetTypeName(t) {
	case "string":
		prop.Type = swagger.StringPropertyType

	// case Value[time.Time]:
	// 	prop.Type = swagger.StringPropertyType
	// 	prop.Format = swagger.DateTimePropertyFormat

	case "float64":
		prop.Type = swagger.NumberPropertyType
		prop.Format = swagger.DoublePropertyFormat

	case "float32":
		prop.Type = swagger.NumberPropertyType
		prop.Format = swagger.FloatPropertyFormat

	case "bool":
		prop.Type = swagger.BooleanPropertyType

	case "int":
		prop.Type = swagger.IntegerPropertyType

	case "int64":
		prop.Type = swagger.IntegerPropertyType
		prop.Format = swagger.Int64PropertyFormat

	case "int32":
		prop.Type = swagger.IntegerPropertyType
		prop.Format = swagger.Int32PropertyFormat

	case "vector3.Vector[float64]":
		prop.Ref = "#/definitions/Float3"

	case "vector2.Vector[float64]":
		prop.Ref = "#/definitions/Float2"

	case "vector3.Vector[int]":
		prop.Ref = "#/definitions/Int3"

	case "vector2.Vector[int]":
		prop.Ref = "#/definitions/Int2"

	case "geometry.AABB":
		prop.Ref = "#/definitions/AABB"

	case "coloring.WebColor":
		prop.Type = swagger.StringPropertyType

	case "[]vector3.Vector[float64]":
		prop.Type = swagger.ArrayPropertyType
		prop.Items = map[string]any{
			"$ref": "#/definitions/Vector3",
		}
	}

	return prop
}
