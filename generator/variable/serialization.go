package variable

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type JsonContainer struct {
	Variable Variable
}

func (jc *JsonContainer) UnmarshalJSON(b []byte) (err error) {
	jc.Variable, err = DeserializeVariable(b)
	return
}

func (tv JsonContainer) MarshalJSON() ([]byte, error) {
	return json.Marshal(tv.Variable)
}

// ============================================================================

type cliConfig[T any] struct {
	FlagName string `json:"flagName"`
	Usage    string `json:"usage"`
	value    *T
}

type typedVariableSchema[T any] struct {
	variableSchemaBase
	Value T             `json:"value"`
	CLI   *cliConfig[T] `json:"cli,omitempty"`
}

func deserialiseTypedVariableSchema[T any](msg json.RawMessage) (Variable, error) {
	vsb := &typedVariableSchema[T]{}
	err := json.Unmarshal(msg, vsb)
	if err != nil {
		return nil, err
	}
	return &TypeVariable[T]{
		name:        vsb.Name,
		value:       vsb.Value,
		description: vsb.Description,
	}, nil
}

type variableSchemaBase struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

func DeserializeVariable(msg json.RawMessage) (Variable, error) {
	vsb := &variableSchemaBase{}
	err := json.Unmarshal(msg, vsb)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(vsb.Type) {
	case "float64":
		return deserialiseTypedVariableSchema[float64](msg)

	case "string":
		return deserialiseTypedVariableSchema[string](msg)

	case "int":
		return deserialiseTypedVariableSchema[int](msg)

	case "bool":
		return deserialiseTypedVariableSchema[bool](msg)

	case "float2":
		return deserialiseTypedVariableSchema[vector2.Float64](msg)

	case "int2":
		return deserialiseTypedVariableSchema[vector2.Int](msg)

	case "float3":
		return deserialiseTypedVariableSchema[vector3.Float64](msg)

	case "int3":
		return deserialiseTypedVariableSchema[vector3.Int](msg)

	case "float3[]":
		return deserialiseTypedVariableSchema[[]vector3.Float64](msg)

	case "aabb":
		return deserialiseTypedVariableSchema[geometry.AABB](msg)

	case "color":
		return deserialiseTypedVariableSchema[coloring.WebColor](msg)

	default:
		return nil, fmt.Errorf("unrecognized variable type: %q", vsb.Type)
	}
}
