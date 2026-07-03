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

func (pb ParameterBase) DisplayName() string {
	return pb.Name
}

func (pb ParameterBase) ValueType() string {
	return pb.Type
}
