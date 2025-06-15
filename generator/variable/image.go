package variable

import (
	"bytes"
	"fmt"
	"image"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"

	_ "image/jpeg"
	"image/png"
	_ "image/png"
)

type ImageVariable struct {
	value   image.Image
	version int
	info    Info
}

func (tv *ImageVariable) SetValue(v image.Image) {
	tv.value = v
	tv.version++
}

func (tv *ImageVariable) GetValue() image.Image {
	return tv.value
}

func (tv *ImageVariable) Version() int {
	return tv.version
}

func (tv *ImageVariable) Info() Info {
	return tv.info
}

func (tv *ImageVariable) setInfo(i Info) error {
	if tv.info != nil {
		return fmt.Errorf("already assigned info")
	}
	tv.info = i
	return nil
}

func (tv *ImageVariable) currentValue() any {
	return tv.value
}

func (tv *ImageVariable) currentVersion() int {
	return tv.version
}

func (tv *ImageVariable) ApplyMessage(msg []byte) (bool, error) {
	if len(msg) == 0 {
		changed := tv.value == nil
		tv.value = nil
		tv.version++
		return changed, nil
	}

	img, _, err := image.Decode(bytes.NewReader(msg))
	if err != nil {
		return false, err
	}

	// if pn.appliedProfile != nil && val == *pn.appliedProfile {
	// 	return false, nil
	// }

	tv.version++
	tv.value = img
	return true, nil
}

func (tv ImageVariable) ToMessage() []byte {
	out := bytes.Buffer{}
	err := png.Encode(&out, tv.value)
	if err != nil {
		panic(err)
	}
	return out.Bytes()
}

func (tv *ImageVariable) NodeReference() nodes.Node {
	return &VariableReferenceNode[image.Image]{
		variable: tv,
	}
}

// func (tv ImageVariable) MarshalJSON() ([]byte, error) {
// 	var t T
// 	return json.Marshal(typedVariableSchema[T]{
// 		variableSchemaBase: variableSchemaBase{
// 			Type: refutil.GetTypeName(t),
// 		},
// 		Value: tv.value,
// 	})
// }

func (tv ImageVariable) runtimeSchema() schema.RuntimeVariable {
	return schema.RuntimeVariable{
		Description: tv.info.Description(),
		Type:        "image.Image", // refutil.GetTypeName(tv.value),
		Value:       tv.value,
	}
}

type imageNodeGraphSchema struct {
	Type  string `json:"type"`
	Value *jbtf.Png
}

func (tv ImageVariable) toPersistantJSON(encoder *jbtf.Encoder) ([]byte, error) {
	schema := imageNodeGraphSchema{
		Type: "image.Image",
	}

	if tv.value != nil {
		schema.Value = &jbtf.Png{
			Image: tv.value,
		}
	}

	return encoder.Marshal(schema)
}

func (tv *ImageVariable) fromPersistantJSON(decoder jbtf.Decoder, body []byte) error {
	gn, err := jbtf.Decode[imageNodeGraphSchema](decoder, body)
	if err != nil {
		return err
	}
	if gn.Value != nil {
		tv.value = gn.Value.Image
	}
	return nil
}
