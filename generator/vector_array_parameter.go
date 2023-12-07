package generator

import (
	"encoding/json"
	"flag"

	"github.com/EliCDavis/vector/vector3"
)

type VectorArrayCliParameterConfig struct {
	FlagName string
	Usage    string
	value    []vector3.Float64
}

type VectorArrayParameter struct {
	Name           string
	DefaultValue   []vector3.Float64
	appliedProfile []vector3.Float64
	CLI            *VectorArrayCliParameterConfig
}

func (fp *VectorArrayParameter) Reset() {
	fp.appliedProfile = nil
}

func (ip *VectorArrayParameter) ApplyJsonMessage(msg json.RawMessage) (bool, error) {
	num := make([]vector3.Float64, 0)
	err := json.Unmarshal(msg, &num)
	if err != nil {
		return false, err
	}

	// if ip.appliedProfile != nil && ip.appliedProfile == num {
	// 	return false, nil
	// }

	ip.appliedProfile = num
	return true, nil
}

func (ip VectorArrayParameter) Schema() ParameterSchema {
	return VectorArrayParameterSchema{
		ParameterSchemaBase: ParameterSchemaBase{
			Name: ip.Name,
			Type: "VectorArray",
		},
		CurrentValue: ip.Value(),
		DefaultValue: ip.DefaultValue,
	}
}

func (ip VectorArrayParameter) DisplayName() string {
	return ip.Name
}

func (ip VectorArrayParameter) Value() []vector3.Float64 {
	if ip.appliedProfile != nil {
		return ip.appliedProfile
	}

	if ip.CLI != nil && ip.CLI.value != nil {
		return ip.CLI.value
	}
	return ip.DefaultValue
}

func (ip VectorArrayParameter) initializeForCLI(set *flag.FlagSet) {
	if ip.CLI == nil {
		return
	}

	panic("unimplemented")
	// ip.CLI.value = set.VectorArray(
	// 	ip.CLI.FlagName,
	// 	ip.DefaultValue,
	// 	ip.CLI.Usage,
	// )
}
