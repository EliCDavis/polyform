package generator

import (
	"encoding/json"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/vector/vector3"
)

type Profile struct {
	Parameters    json.RawMessage    `json:"parameters"`
	SubGenerators map[string]Profile `json:"subGenerators"`
}

type GeneratorSchema struct {
	Parameters    GroupParameterSchema       `json:"parameters"`
	Producers     []string                   `json:"producers"`
	SubGenerators map[string]GeneratorSchema `json:"subGenerators"`
}

type ParameterSchema interface {
	ValueType() string
	DisplayName() string
}

type ParameterSchemaBase struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (gps ParameterSchemaBase) DisplayName() string {
	return gps.Name
}

func (gps ParameterSchemaBase) ValueType() string {
	return gps.Type
}

type GroupParameterSchema struct {
	ParameterSchemaBase
	Parameters []ParameterSchema `json:"parameters"`
}

type FloatParameterSchema struct {
	ParameterSchemaBase
	DefaultValue float64 `json:"defaultValue"`
	CurrentValue float64 `json:"currentValue"`
}

type IntParameterSchema struct {
	ParameterSchemaBase
	DefaultValue int `json:"defaultValue"`
	CurrentValue int `json:"currentValue"`
}

type BoolParameterSchema struct {
	ParameterSchemaBase
	DefaultValue bool `json:"defaultValue"`
	CurrentValue bool `json:"currentValue"`
}

type StringParameterSchema struct {
	ParameterSchemaBase
	DefaultValue string `json:"defaultValue"`
	CurrentValue string `json:"currentValue"`
}

type ColorParameterSchema struct {
	ParameterSchemaBase
	DefaultValue coloring.WebColor `json:"defaultValue"`
	CurrentValue coloring.WebColor `json:"currentValue"`
}

type VectorArrayParameterSchema struct {
	ParameterSchemaBase
	DefaultValue []vector3.Float64 `json:"defaultValue"`
	CurrentValue []vector3.Float64 `json:"currentValue"`
}
