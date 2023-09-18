package generator

import (
	"encoding/json"
	"flag"
)

type FloatCliParameterConfig struct {
	FlagName string
	Usage    string
	value    *float64
}

type FloatParameter struct {
	Name           string
	DefaultValue   float64
	appliedProfile *float64
	CLI            *FloatCliParameterConfig
}

func (fp *FloatParameter) ApplyJsonMessage(msg json.RawMessage) error {
	num := 0.
	err := json.Unmarshal(msg, &num)
	if err != nil {
		return err
	}
	fp.appliedProfile = &num
	return nil
}

func (fp FloatParameter) Schema() ParameterSchema {
	return FloatParameterSchema{
		ParameterSchemaBase: ParameterSchemaBase{
			Name: fp.Name,
			Type: "Float",
		},
		DefaultValue: fp.DefaultValue,
	}
}

func (fp FloatParameter) DisplayName() string {
	return fp.Name
}

func (fp FloatParameter) Value() float64 {
	if fp.appliedProfile != nil {
		return *fp.appliedProfile
	}

	if fp.CLI != nil && fp.CLI.value != nil {
		return *fp.CLI.value
	}
	return fp.DefaultValue
}

func (fp FloatParameter) initializeForCLI(set *flag.FlagSet) {
	if fp.CLI == nil {
		return
	}

	fp.CLI.value = set.Float64(
		fp.CLI.FlagName,
		fp.DefaultValue,
		fp.CLI.Usage,
	)
}
