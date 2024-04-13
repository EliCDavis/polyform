package generator

import (
	"bytes"
	"flag"
	"image"
	"image/png"
	"os"

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

func (tn ImageParameterNode) Outputs() []nodes.Output {
	return []nodes.Output{
		{
			Name: "Data",
			Type: "image.Image",
		},
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
