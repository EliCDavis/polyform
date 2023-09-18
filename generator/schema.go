package generator

import "encoding/json"

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
}
type IntParameterSchema struct {
	ParameterSchemaBase
	DefaultValue int `json:"defaultValue"`
}
