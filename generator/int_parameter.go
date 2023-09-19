package generator

import (
	"encoding/json"
	"flag"
)

type IntCliParameterConfig struct {
	FlagName string
	Usage    string
	value    *int
}

type IntParameter struct {
	Name           string
	DefaultValue   int
	appliedProfile *int
	CLI            *IntCliParameterConfig
}

func (fp *IntParameter) Reset() {
	fp.appliedProfile = nil
}

func (ip *IntParameter) ApplyJsonMessage(msg json.RawMessage) error {
	num := 0
	err := json.Unmarshal(msg, &num)
	if err != nil {
		return err
	}
	ip.appliedProfile = &num
	return nil
}

func (ip IntParameter) Schema() ParameterSchema {
	return IntParameterSchema{
		ParameterSchemaBase: ParameterSchemaBase{
			Name: ip.Name,
			Type: "Int",
		},
		CurrentValue: ip.Value(),
		DefaultValue: ip.DefaultValue,
	}
}

func (ip IntParameter) DisplayName() string {
	return ip.Name
}

func (ip IntParameter) Value() int {
	if ip.appliedProfile != nil {
		return *ip.appliedProfile
	}

	if ip.CLI != nil && ip.CLI.value != nil {
		return *ip.CLI.value
	}
	return ip.DefaultValue
}

func (ip IntParameter) initializeForCLI(set *flag.FlagSet) {
	if ip.CLI == nil {
		return
	}

	ip.CLI.value = set.Int(
		ip.CLI.FlagName,
		ip.DefaultValue,
		ip.CLI.Usage,
	)
}
