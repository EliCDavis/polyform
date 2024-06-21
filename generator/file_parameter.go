package generator

import (
	"flag"
	"io"
	"os"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/nodes"
)

type FileParameterNode struct {
	Name         string
	DefaultValue []byte
	CLI          *CliParameterNodeConfig[string]

	version        int
	appliedProfile []byte
}

func (in *FileParameterNode) Node() nodes.Node {
	return in
}

func (in FileParameterNode) Port() string {
	return "Out"
}

func (vn FileParameterNode) SetInput(input string, output nodes.Output) {
	panic("input can not be set")
}

func (pn *FileParameterNode) DisplayName() string {
	return pn.Name
}

func (pn *FileParameterNode) ApplyMessage(msg []byte) (bool, error) {
	pn.version++
	pn.appliedProfile = msg
	return true, nil
}

func (pn *FileParameterNode) ToMessage() []byte {
	return pn.Value()
}

func (pn *FileParameterNode) Value() []byte {
	if pn.appliedProfile != nil {
		return pn.appliedProfile
	}

	if pn.CLI != nil && pn.CLI.value != nil && *pn.CLI.value != "" {
		f, err := os.Open(*pn.CLI.value)
		if err != nil {
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

func (pn *FileParameterNode) Schema() ParameterSchema {
	return ParameterSchemaBase{
		Name: pn.Name,
		Type: "[]uint8",
	}
}

func (pn *FileParameterNode) Dependencies() []nodes.NodeDependency {
	return nil
}

func (pn *FileParameterNode) State() nodes.NodeState {
	return nodes.Processed
}

type FileNodeOutput struct {
	Val *FileParameterNode
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

func (tn *FileParameterNode) Outputs() []nodes.Output {
	return []nodes.Output{
		{
			Type: "[]uint8",
			NodeOutput: FileNodeOutput{
				Val: tn,
			},
		},
	}
}

func (tn *FileParameterNode) Out() FileNodeOutput {
	return FileNodeOutput{
		Val: tn,
	}
}

func (tn FileParameterNode) Inputs() []nodes.Input {
	return []nodes.Input{}
}

func (pn FileParameterNode) Version() int {
	return pn.version
}

func (pn FileParameterNode) initializeForCLI(set *flag.FlagSet) {
	if pn.CLI == nil {
		return
	}
	pn.CLI.value = set.String(pn.CLI.FlagName, "", pn.CLI.Usage)
}

// CUSTOM JTF Serialization ===================================================

type fileNodeGraphSchema struct {
	Name         string                          `json:"name"`
	CurrentValue *jbtf.Bytes                     `json:"currentValue"`
	DefaultValue *jbtf.Bytes                     `json:"defaultValue"`
	CLI          *CliParameterNodeConfig[string] `json:"cli"`
}

func (pn *FileParameterNode) ToJSON(encoder *jbtf.Encoder) ([]byte, error) {
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

func (pn *FileParameterNode) FromJSON(decoder jbtf.Decoder, body []byte) (err error) {
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
