package parameter

import (
	"bytes"
	"flag"
	"image"
	_ "image/jpeg"
	"image/png"
	"os"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
)

type Image struct {
	Name         string
	DefaultValue image.Image
	CLI          *CliConfig[string]

	subs           []nodes.Alertable
	version        int
	appliedProfile image.Image
}

func (in *Image) Node() nodes.Node {
	return in
}

func (in Image) Port() string {
	return "Out"
}

func (vn Image) SetInput(input string, output nodes.Output) {
	panic("input can not be set")
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
	pn.appliedProfile = val

	for _, s := range pn.subs {
		s.Alert(pn.version, nodes.Processed)
	}

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
	if pn.appliedProfile != nil {
		return pn.appliedProfile
	}

	if pn.CLI != nil && pn.CLI.value != nil && *pn.CLI.value != "" {
		f, err := os.Open(*pn.CLI.value)
		if err != nil {
			return nil
		}
		defer f.Close()

		pn.appliedProfile, _, err = image.Decode(f)
		if err != nil {
			return nil
		}

		return pn.appliedProfile
	}
	return pn.DefaultValue
}

func (pn *Image) Schema() schema.Parameter {
	return schema.ParameterBase{
		Name: pn.Name,
		Type: "image.Image",
	}
}

func (pn *Image) AddSubscription(a nodes.Alertable) {
	if pn.subs == nil {
		pn.subs = make([]nodes.Alertable, 0, 1)
	}

	pn.subs = append(pn.subs, a)
}

func (pn *Image) Dependencies() []nodes.NodeDependency {
	return nil
}

func (pn *Image) State() nodes.NodeState {
	return nodes.Processed
}

type ImageNodeOutput struct {
	Val *Image
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

func (tn *Image) Outputs() []nodes.Output {
	return []nodes.Output{
		{
			Type: "image.Image",
			NodeOutput: ImageNodeOutput{
				Val: tn,
			},
		},
	}
}

func (tn *Image) Out() ImageNodeOutput {
	return ImageNodeOutput{
		Val: tn,
	}
}

func (tn Image) Inputs() []nodes.Input {
	return []nodes.Input{}
}

func (pn Image) Version() int {
	return pn.version
}

func (pn Image) InitializeForCLI(set *flag.FlagSet) {
	if pn.CLI == nil {
		return
	}
	pn.CLI.value = set.String(pn.CLI.FlagName, "", pn.CLI.Usage)
}

// CUSTOM JTF Serialization ===================================================

type imageNodeGraphSchema struct {
	Name         string             `json:"name"`
	CurrentValue *jbtf.Png          `json:"currentValue"`
	DefaultValue *jbtf.Png          `json:"defaultValue"`
	CLI          *CliConfig[string] `json:"cli"`
}

func (pn *Image) ToJSON(encoder *jbtf.Encoder) ([]byte, error) {
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

func (pn *Image) FromJSON(decoder jbtf.Decoder, body []byte) (err error) {
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
