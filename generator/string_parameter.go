package generator

import (
	"encoding/json"
	"flag"
)

type StringCliParameterConfig struct {
	FlagName string
	Usage    string
	value    *string
}

type StringParameter struct {
	Name           string
	DefaultValue   string
	appliedProfile *string
	CLI            *StringCliParameterConfig
}

func (fp *StringParameter) Reset() {
	fp.appliedProfile = nil
}

func (ip *StringParameter) ApplyJsonMessage(msg json.RawMessage) (bool, error) {
	b := ""
	err := json.Unmarshal(msg, &b)
	if err != nil {
		return false, err
	}

	if ip.appliedProfile != nil && *ip.appliedProfile == b {
		return false, nil
	}

	ip.appliedProfile = &b
	return true, nil
}

func (ip StringParameter) Schema() ParameterSchema {
	return StringParameterSchema{
		ParameterSchemaBase: ParameterSchemaBase{
			Name: ip.Name,
			Type: "String",
		},
		CurrentValue: ip.Value(),
		DefaultValue: ip.DefaultValue,
	}
}

func (ip StringParameter) DisplayName() string {
	return ip.Name
}

func (ip StringParameter) Value() string {
	if ip.appliedProfile != nil {
		return *ip.appliedProfile
	}

	if ip.CLI != nil && ip.CLI.value != nil {
		return *ip.CLI.value
	}
	return ip.DefaultValue
}

func (ip StringParameter) initializeForCLI(set *flag.FlagSet) {
	if ip.CLI == nil {
		return
	}

	ip.CLI.value = set.String(
		ip.CLI.FlagName,
		ip.DefaultValue,
		ip.CLI.Usage,
	)
}
