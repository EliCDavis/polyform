package parameter

import (
	"bytes"
	"image"
	_ "image/jpeg"
	"image/png"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
)

type imageNodeOutput struct {
	Val *Image
}

func (sno imageNodeOutput) Name() string {
	return valueOutputPortName
}

func (sno imageNodeOutput) Value() image.Image {
	return sno.Val.Value()
}

func (sno imageNodeOutput) Node() nodes.Node {
	return sno.Val
}

func (sno imageNodeOutput) Port() string {
	return valueOutputPortName
}

func (sno imageNodeOutput) Version() int {
	return sno.Val.version
}

// ============================================================================

type Image struct {
	Name        string
	Description string

	version      int
	appliedValue image.Image
}

func (in *Image) SetName(name string) {
	in.Name = name
}

func (in *Image) SetDescription(description string) {
	in.Description = description
}

func (pn *Image) DisplayName() string {
	return pn.Name
}

func (pn *Image) ApplyMessage(msg []byte) (bool, error) {
	val, _, err := image.Decode(bytes.NewBuffer(msg))
	if err != nil {
		return false, err
	}

	pn.version++
	pn.appliedValue = val

	return true, nil
}

func (pn *Image) ToMessage() []byte {
	img := pn.Value()
	if img == nil {
		return nil
	}
	buf := bytes.Buffer{}
	err := png.Encode(&buf, img)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func (pn *Image) Value() image.Image {
	return pn.appliedValue
}

func (pn *Image) Schema() schema.Parameter {
	return schema.ParameterBase{
		Name: pn.Name,
		Type: "image.Image",
	}
}

func (tn *Image) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		valueOutputPortName: imageNodeOutput{
			Val: tn,
		},
	}
}

func (tn Image) Inputs() map[string]nodes.InputPort {
	return nil
}

// CUSTOM JTF Serialization ===================================================

type imageNodeGraphSchema struct {
	Name         string    `json:"name"`
	CurrentValue *jbtf.Png `json:"currentValue"`
}

func (pn *Image) ToJSON(encoder *jbtf.Encoder) ([]byte, error) {
	schema := imageNodeGraphSchema{
		Name: pn.Name,
	}

	if pn.Value() != nil {
		schema.CurrentValue = &jbtf.Png{
			Image: pn.Value(),
		}
	}

	return encoder.Marshal(schema)
}

func (pn *Image) FromJSON(decoder jbtf.Decoder, body []byte) (err error) {
	gn, err := jbtf.Decode[imageNodeGraphSchema](decoder, body)
	if err != nil {
		return
	}

	pn.Name = gn.Name

	if gn.CurrentValue != nil {
		pn.appliedValue = gn.CurrentValue.Image
	}
	return
}
