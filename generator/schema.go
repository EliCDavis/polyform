package generator

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
