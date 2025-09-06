package variable

import (
	"encoding/json"

	"github.com/EliCDavis/jbtf"
)

type cliConfig[T any] struct {
	FlagName string `json:"flagName"`
	Usage    string `json:"usage"`
	value    *T
}

type variableSchemaBase struct {
	Type string `json:"type"`
}

func DeserializePersistantVariableJSON(msg []byte, decoder jbtf.Decoder, variableFactory func(string) (Variable, error)) (Variable, error) {
	vsb := &variableSchemaBase{}
	err := json.Unmarshal(msg, vsb)
	if err != nil {
		return nil, err
	}

	v, err := variableFactory(vsb.Type)
	if err != nil {
		return nil, err
	}

	return v, v.fromPersistantJSON(decoder, msg)
}

// func CreateVariable(variableType string) (Variable, error) {
// 	switch strings.ToLower(variableType) {
// 	case "float64":
// 		return &TypeVariable[float64]{}, nil

// 	case "string":
// 		return &TypeVariable[string]{}, nil

// 	case "int":
// 		return &TypeVariable[int]{}, nil

// 	case "bool":
// 		return &TypeVariable[bool]{}, nil

// 	case "vector2.vector[float64]":
// 		return &TypeVariable[vector2.Float64]{}, nil

// 	case "vector2.vector[int]":
// 		return &TypeVariable[vector2.Int]{}, nil

// 	case "vector3.vector[float64]":
// 		return &TypeVariable[vector3.Float64]{}, nil

// 	case "vector3.vector[int]":
// 		return &TypeVariable[vector3.Int]{}, nil

// 	case "[]vector3.vector[float64]":
// 		return &TypeVariable[[]vector3.Float64]{}, nil

// 	case "geometry.aabb":
// 		return &TypeVariable[geometry.AABB]{}, nil

// 	case "coloring.Color":
// 		return &TypeVariable[coloring.Color]{}, nil

// 	case "image.image":
// 		return &ImageVariable{}, nil

// 	case "file":
// 		return &FileVariable{}, nil

// 	default:
// 		return nil, fmt.Errorf("unrecognized variable type: %q", variableType)
// 	}
// }
