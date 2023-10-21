package generator

import (
	"encoding/json"
	"flag"
)

type BoolCliParameterConfig struct {
	FlagName string
	Usage    string
	value    *bool
}

type BoolParameter struct {
	Name           string
	DefaultValue   bool
	appliedProfile *bool
	CLI            *BoolCliParameterConfig
}

func (fp *BoolParameter) Reset() {
	fp.appliedProfile = nil
}

func (ip *BoolParameter) ApplyJsonMessage(msg json.RawMessage) (bool, error) {
	b := false
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

func (ip BoolParameter) Schema() ParameterSchema {
	return BoolParameterSchema{
		ParameterSchemaBase: ParameterSchemaBase{
			Name: ip.Name,
			Type: "Bool",
		},
		CurrentValue: ip.Value(),
		DefaultValue: ip.DefaultValue,
	}
}

func (ip BoolParameter) DisplayName() string {
	return ip.Name
}

func (ip BoolParameter) Value() bool {
	if ip.appliedProfile != nil {
		return *ip.appliedProfile
	}

	if ip.CLI != nil && ip.CLI.value != nil {
		return *ip.CLI.value
	}
	return ip.DefaultValue
}

func (ip BoolParameter) initializeForCLI(set *flag.FlagSet) {
	if ip.CLI == nil {
		return
	}

	ip.CLI.value = set.Bool(
		ip.CLI.FlagName,
		ip.DefaultValue,
		ip.CLI.Usage,
	)
}
