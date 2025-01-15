package parameter

import (
	"flag"
	"io"
	"os"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
)

type File struct {
	Name         string
	DefaultValue []byte
	CLI          *CliConfig[string]

	version        int
	appliedProfile []byte
}

func (in *File) SetName(name string) {
	in.Name = name
}

func (in *File) Node() nodes.Node {
	return in
}

func (in File) Port() string {
	return "Out"
}

func (vn File) SetInput(input string, output nodes.Output) {
	panic("input can not be set")
}

func (pn *File) DisplayName() string {
	return pn.Name
}

func (pn *File) ApplyMessage(msg []byte) (bool, error) {
	pn.version++
	pn.appliedProfile = msg
	return true, nil
}

func (pn *File) ToMessage() []byte {
	return pn.Value()
}

func (pn *File) Value() []byte {
	if pn.appliedProfile != nil {
		return pn.appliedProfile
	}

	if pn.CLI != nil && pn.CLI.value != nil && *pn.CLI.value != "" {
		f, err := os.Open(*pn.CLI.value)
		if err != nil {
			panic(err)
			return nil
		}
		defer f.Close()

		pn.appliedProfile, err = io.ReadAll(f)
		if err != nil {
			return nil
		}

		return pn.appliedProfile
	}
	return pn.DefaultValue
}

func (pn *File) Schema() schema.Parameter {
	return schema.ParameterBase{
		Name: pn.Name,
		Type: "[]uint8",
	}
}

func (pn *File) Dependencies() []nodes.NodeDependency {
	return nil
}

func (pn *File) State() nodes.NodeState {
	return nodes.Processed
}

type FileNodeOutput struct {
	Val *File
}

func (sno FileNodeOutput) Value() []byte {
	return sno.Val.Value()
}

func (sno FileNodeOutput) Node() nodes.Node {
	return sno.Val
}

func (sno FileNodeOutput) Port() string {
	return "Out"
}

func (tn *File) Outputs() []nodes.Output {
	return []nodes.Output{
		{
			Type: "[]uint8",
			NodeOutput: FileNodeOutput{
				Val: tn,
			},
		},
	}
}

func (tn *File) Out() FileNodeOutput {
	return FileNodeOutput{
		Val: tn,
	}
}

func (tn File) Inputs() []nodes.Input {
	return []nodes.Input{}
}

func (pn File) Version() int {
	return pn.version
}

func (pn File) InitializeForCLI(set *flag.FlagSet) {
	if pn.CLI == nil {
		return
	}
	pn.CLI.value = set.String(pn.CLI.FlagName, "", pn.CLI.Usage)
}

// CUSTOM JTF Serialization ===================================================

type fileNodeGraphSchema struct {
	Name         string             `json:"name"`
	CurrentValue *jbtf.Bytes        `json:"currentValue"`
	DefaultValue *jbtf.Bytes        `json:"defaultValue"`
	CLI          *CliConfig[string] `json:"cli"`
}

func (pn *File) ToJSON(encoder *jbtf.Encoder) ([]byte, error) {
	schema := fileNodeGraphSchema{
		Name: pn.Name,
		CLI:  pn.CLI,
	}

	if pn.Value() != nil {
		schema.CurrentValue = &jbtf.Bytes{
			Data: pn.Value(),
		}
	}

	if schema.DefaultValue != nil {
		schema.DefaultValue = &jbtf.Bytes{
			Data: pn.DefaultValue,
		}
	}

	return encoder.Marshal(schema)
}

func (pn *File) FromJSON(decoder jbtf.Decoder, body []byte) (err error) {
	gn, err := jbtf.Decode[fileNodeGraphSchema](decoder, body)
	if err != nil {
		return
	}

	pn.Name = gn.Name
	pn.CLI = gn.CLI

	if gn.DefaultValue != nil {
		pn.DefaultValue = gn.DefaultValue.Data
	}
	if gn.CurrentValue != nil {
		pn.appliedProfile = gn.CurrentValue.Data
	}
	return
}
