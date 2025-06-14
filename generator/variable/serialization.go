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
	jc.Variable, err = DeserializeVariableJSON(b)
	return
}

func (jc JsonContainer) MarshalJSON() ([]byte, error) {
	return json.Marshal(jc.Variable)
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
	return &TypeVariable[T]{value: vsb.Value}, nil
}

func deserialiseImageVariable(msg []byte) (Variable, error) {
	iv := &ImageVariable{}
	_, err := iv.ApplyMessage(msg)
	return iv, err
}

type variableSchemaBase struct {
	Type string `json:"type"`
}

func DeserializeVariableJSON(msg []byte) (Variable, error) {
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

	case "vector2.vector[float64]":
		return deserialiseTypedVariableSchema[vector2.Float64](msg)

	case "vector2.vector[int]":
		return deserialiseTypedVariableSchema[vector2.Int](msg)

	case "vector3.vector[float64]":
		return deserialiseTypedVariableSchema[vector3.Float64](msg)

	case "vector3.vector[int]":
		return deserialiseTypedVariableSchema[vector3.Int](msg)

	case "[]vector3.vector[float64]":
		return deserialiseTypedVariableSchema[[]vector3.Float64](msg)

	case "geometry.aabb":
		return deserialiseTypedVariableSchema[geometry.AABB](msg)

	case "coloring.webcolor":
		return deserialiseTypedVariableSchema[coloring.WebColor](msg)

	case "image.image":
		return deserialiseImageVariable(nil)

	default:
		return nil, fmt.Errorf("unrecognized variable type: %q", vsb.Type)
	}
}
