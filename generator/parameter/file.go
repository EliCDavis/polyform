package parameter

import (
	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
)

type fileNodeOutput struct {
	Val *File
}

func (sno fileNodeOutput) Value() []byte {
	return sno.Val.Value()
}

func (sno fileNodeOutput) Node() nodes.Node {
	return sno.Val
}

func (sno fileNodeOutput) Name() string {
	return valueOutputPortName
}

func (sno fileNodeOutput) Version() int {
	return sno.Val.version
}

// ============================================================================

type File struct {
	Name        string
	Description string

	version      int
	appliedValue []byte
}

func (in *File) SetName(name string) {
	in.Name = name
}

func (in *File) SetDescription(description string) {
	in.Description = description
}

func (pn *File) DisplayName() string {
	return pn.Name
}

func (pn *File) ApplyMessage(msg []byte) (bool, error) {
	pn.version++
	pn.appliedValue = msg
	return true, nil
}

func (pn *File) ToMessage() []byte {
	return pn.Value()
}

func (pn *File) Value() []byte {
	return pn.appliedValue
}

func (pn *File) Schema() schema.Parameter {
	return schema.ParameterBase{
		Name: pn.Name,
		Type: "[]uint8",
	}
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

func (tn *File) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		valueOutputPortName: fileNodeOutput{Val: tn},
	}
}

func (tn File) Inputs() map[string]nodes.InputPort {
	return nil
}

// CUSTOM JTF Serialization ===================================================

type fileNodeGraphSchema struct {
	Name         string      `json:"name"`
	CurrentValue *jbtf.Bytes `json:"currentValue"`
}

func (pn *File) ToJSON(encoder *jbtf.Encoder) ([]byte, error) {
	schema := fileNodeGraphSchema{
		Name: pn.Name,
	}

	if pn.Value() != nil {
		schema.CurrentValue = &jbtf.Bytes{
			Data: pn.Value(),
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

	if gn.CurrentValue != nil {
		pn.appliedValue = gn.CurrentValue.Data
	}
	return
}
