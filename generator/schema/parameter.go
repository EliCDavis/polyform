package schema

type Parameter interface {
	ValueType() string
	DisplayName() string
}

type ParameterBase struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

func (gps ParameterBase) DisplayName() string {
	return gps.Name
}

func (gps ParameterBase) ValueType() string {
	return gps.Type
}
