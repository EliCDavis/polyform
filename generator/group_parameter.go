package generator

import (
	"encoding/json"
	"flag"
	"fmt"
)

func findParam[T Parameter](params []Parameter, parameterName string) T {
	for _, param := range params {
		group, ok := param.(T)
		if !ok {
			continue
		}
		if group.DisplayName() == parameterName {
			return group
		}
	}
	panic(fmt.Errorf("group contains no parameter: %s ", parameterName))
}

type GroupParameter struct {
	Name       string      `json:"name"`
	Parameters []Parameter `json:"parameters"`
}

func (gp GroupParameter) ApplyJsonMessage(msg json.RawMessage) error {
	subData := make(map[string]json.RawMessage)

	err := json.Unmarshal(msg, &subData)
	if err != nil {
		return err
	}

	for key, data := range subData {
		for _, param := range gp.Parameters {
			if param.DisplayName() != key {
				continue
			}
			err := param.ApplyJsonMessage(data)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (gp GroupParameter) GroupParameterSchema() GroupParameterSchema {
	gps := GroupParameterSchema{
		ParameterSchemaBase: ParameterSchemaBase{
			Name: gp.Name,
			Type: "Group",
		},
		Parameters: make([]ParameterSchema, len(gp.Parameters)),
	}

	for i, p := range gp.Parameters {
		gps.Parameters[i] = p.Schema()
	}

	return gps
}

func (gp GroupParameter) Schema() ParameterSchema {
	return gp.GroupParameterSchema()
}

func (gp GroupParameter) DisplayName() string {
	return gp.Name
}

func (gp GroupParameter) initializeForCLI(set *flag.FlagSet) {

	for _, p := range gp.Parameters {
		p.initializeForCLI(set)
	}

}

func (gp GroupParameter) Group(parameterName string) GroupParameter {
	return findParam[GroupParameter](gp.Parameters, parameterName)
}

func (gp GroupParameter) Float64(parameterName string) float64 {
	return findParam[*FloatParameter](gp.Parameters, parameterName).Value()
}

func (gp GroupParameter) Int(parameterName string) int {
	return findParam[*IntParameter](gp.Parameters, parameterName).Value()
}
