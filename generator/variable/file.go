package variable

import (
	"encoding/json"
	"fmt"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
)

type FileVariable struct {
	value   []byte
	version int
	info    Info
}

func (tv *FileVariable) SetValue(v []byte) {
	tv.value = v
	tv.version++
}

func (tv *FileVariable) GetValue() []byte {
	return tv.value
}

func (tv *FileVariable) Version() int {
	return tv.version
}

func (tv *FileVariable) Info() Info {
	return tv.info
}

func (tv *FileVariable) setInfo(i Info) error {
	if tv.info != nil {
		return fmt.Errorf("already assigned info")
	}
	tv.info = i
	return nil
}

func (tv *FileVariable) currentValue() any {
	return tv.value
}

func (tv *FileVariable) currentVersion() int {
	return tv.version
}

func (tv *FileVariable) ApplyMessage(msg []byte) (bool, error) {
	tv.version++
	tv.value = msg
	return true, nil
}

func (tv *FileVariable) applyProfile(profile json.RawMessage) error {
	tv.version++
	tv.value = profile
	return nil
}

func (tv FileVariable) ToMessage() []byte {
	return tv.value
}

func (tv *FileVariable) NodeReference() nodes.Node {
	return &VariableReferenceNode[[]byte]{
		variable: tv,
	}
}

// func (tv FileVariable) MarshalJSON() ([]byte, error) {
// 	var t T
// 	return json.Marshal(typedVariableSchema[T]{
// 		variableSchemaBase: variableSchemaBase{
// 			Type: refutil.GetTypeName(t),
// 		},
// 		Value: tv.value,
// 	})
// }

type fileDetails struct {
	Size int `json:"size"`
}

func (tv FileVariable) runtimeSchema() schema.RuntimeVariable {
	return schema.RuntimeVariable{
		Description: tv.info.Description(),
		Type:        "file", // refutil.GetTypeName(tv.value),
		Value: fileDetails{
			Size: len(tv.value),
		},
	}
}

type fileNodeGraphSchema struct {
	Type  string `json:"type"`
	Value *jbtf.Bytes
}

func (tv FileVariable) toPersistantJSON(encoder *jbtf.Encoder) ([]byte, error) {
	schema := fileNodeGraphSchema{
		Type: "file",
	}

	if tv.value != nil {
		schema.Value = &jbtf.Bytes{
			Data: tv.value,
		}
	}

	return encoder.Marshal(schema)
}

func (tv *FileVariable) fromPersistantJSON(decoder jbtf.Decoder, body []byte) error {
	gn, err := jbtf.Decode[fileNodeGraphSchema](decoder, body)
	if err != nil {
		return err
	}
	if gn.Value != nil {
		tv.value = gn.Value.Data
	}
	return nil
}
