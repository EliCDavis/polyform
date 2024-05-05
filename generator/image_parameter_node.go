package generator

import (
	"bytes"
	"flag"
	"image"
	_ "image/jpeg"
	"image/png"
	"os"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/nodes"
)

type ImageParameterNode struct {
	Name         string
	DefaultValue image.Image
	CLI          *CliParameterNodeConfig[string]

	subs           []nodes.Alertable
	version        int
	appliedProfile image.Image
}

func (in *ImageParameterNode) Node() nodes.Node {
	return in
}

func (in ImageParameterNode) Port() string {
	return "Out"
}

func (vn ImageParameterNode) SetInput(input string, output nodes.Output) {
	panic("input can not be set")
}

func (pn *ImageParameterNode) DisplayName() string {
	return pn.Name
}

func (pn *ImageParameterNode) ApplyMessage(msg []byte) (bool, error) {
	val, _, err := image.Decode(bytes.NewBuffer(msg))
	if err != nil {
		return false, err
	}

	pn.version++
	pn.appliedProfile = val

	for _, s := range pn.subs {
		s.Alert(pn.version, nodes.Processed)
	}

	return true, nil
}

func (pn *ImageParameterNode) ToMessage() []byte {
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

func (pn *ImageParameterNode) Value() image.Image {
	if pn.appliedProfile != nil {
		return pn.appliedProfile
	}

	if pn.CLI != nil && pn.CLI.value != nil && *pn.CLI.value != "" {
		f, err := os.Open(*pn.CLI.value)
		if err != nil {
			return nil
		}

		pn.appliedProfile, _, err = image.Decode(f)
		if err != nil {
			return nil
		}

		return pn.appliedProfile
	}
	return pn.DefaultValue
}

func (pn *ImageParameterNode) Schema() ParameterSchema {
	return ParameterSchemaBase{
		Name: pn.Name,
		Type: "image.Image",
	}
}

func (pn *ImageParameterNode) AddSubscription(a nodes.Alertable) {
	if pn.subs == nil {
		pn.subs = make([]nodes.Alertable, 0, 1)
	}

	pn.subs = append(pn.subs, a)
}

func (pn *ImageParameterNode) Dependencies() []nodes.NodeDependency {
	return nil
}

func (pn *ImageParameterNode) State() nodes.NodeState {
	return nodes.Processed
}

type ImageNodeOutput struct {
	Val *ImageParameterNode
}

func (sno ImageNodeOutput) Value() image.Image {
	return sno.Val.Value()
}

func (sno ImageNodeOutput) Node() nodes.Node {
	return sno.Val
}

func (sno ImageNodeOutput) Port() string {
	return "Out"
}

func (tn *ImageParameterNode) Outputs() []nodes.Output {
	return []nodes.Output{
		{
			Type: "image.Image",
			NodeOutput: ImageNodeOutput{
				Val: tn,
			},
		},
	}
}

func (tn *ImageParameterNode) Out() ImageNodeOutput {
	return ImageNodeOutput{
		Val: tn,
	}
}

func (tn ImageParameterNode) Inputs() []nodes.Input {
	return []nodes.Input{}
}

func (pn ImageParameterNode) Version() int {
	return pn.version
}

func (pn ImageParameterNode) initializeForCLI(set *flag.FlagSet) {
	if pn.CLI == nil {
		return
	}
	pn.CLI.value = set.String(pn.CLI.FlagName, "", pn.CLI.Usage)
}

// CUSTOM JTF Serialization ===================================================

type imageNodeGraphSchema struct {
	Name         string                          `json:"name"`
	CurrentValue *jbtf.Png                       `json:"currentValue"`
	DefaultValue *jbtf.Png                       `json:"defaultValue"`
	CLI          *CliParameterNodeConfig[string] `json:"cli"`
}

func (pn *ImageParameterNode) ToJSON(encoder *jbtf.Encoder) ([]byte, error) {
	schema := imageNodeGraphSchema{
		Name: pn.Name,
		CLI:  pn.CLI,
	}

	if pn.Value() != nil {
		schema.CurrentValue = &jbtf.Png{
			Image: pn.Value(),
		}
	}

	if schema.DefaultValue != nil {
		schema.DefaultValue = &jbtf.Png{
			Image: pn.DefaultValue,
		}
	}

	return encoder.Marshal(schema)
}

func (pn *ImageParameterNode) FromJSON(decoder jbtf.Decoder, body []byte) (err error) {
	gn, err := jbtf.Decode[imageNodeGraphSchema](decoder, body)
	if err != nil {
		return
	}

	pn.Name = gn.Name
	pn.CLI = gn.CLI

	if gn.DefaultValue != nil {
		pn.DefaultValue = gn.DefaultValue.Image
	}
	if gn.CurrentValue != nil {
		pn.appliedProfile = gn.CurrentValue.Image
	}
	return
}
