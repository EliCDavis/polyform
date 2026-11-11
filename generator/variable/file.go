package variable

import (
	"encoding/json"
	"fmt"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
)

type FileVariable struct {
	value   []byte
	hash    string
	blobs   map[string][]byte
	version int
	info    Info
}

func (tv *FileVariable) SetValue(v []byte) {
	tv.set(v, hashBytes(v))
}

func (tv *FileVariable) set(data []byte, hash string) {
	tv.value = data
	tv.hash = hash
	tv.version++
	nodes.Touch()
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
	hash := hashBytes(msg)
	if tv.value != nil && hash == tv.hash {
		return false, nil
	}
	tv.set(msg, hash)
	return true, nil
}

func (tv *FileVariable) retain(hash string, data []byte) []byte {
	if existing, ok := tv.blobs[hash]; ok {
		return existing
	}
	if tv.blobs == nil {
		tv.blobs = make(map[string][]byte)
	}
	tv.blobs[hash] = data
	return data
}

func (tv *FileVariable) applyProfile(profile json.RawMessage) error {
	if isNull(profile) {
		tv.set(nil, "")
		return nil
	}
	if hash, ok := parseBlobRef(profile); ok {
		data, ok := tv.blobs[hash]
		if !ok {
			return fmt.Errorf("profile references file content %s, which is not loaded", hash)
		}
		tv.set(data, hash)
		return nil
	}
	data, err := decodeInlineBytes(profile)
	if err != nil {
		return err
	}
	tv.SetValue(data)
	return nil
}

func (tv *FileVariable) getProfile() json.RawMessage {
	if tv.value == nil {
		return json.RawMessage("null")
	}
	tv.retain(tv.hash, tv.value)
	return encodeBlobRef(tv.hash)
}

// Keeps only the content the current value and the given profile entries name.
func (tv *FileVariable) pruneBlobs(referenced []json.RawMessage) {
	keep := make(map[string][]byte, len(referenced)+1)
	if tv.value != nil {
		keep[tv.hash] = tv.value
	}
	for _, raw := range referenced {
		if hash, ok := parseBlobRef(raw); ok {
			if data, ok := tv.blobs[hash]; ok {
				keep[hash] = data
			}
		}
	}
	tv.blobs = keep
}

func (tv *FileVariable) persistProfile(encoder Encoder, profile json.RawMessage) (json.RawMessage, error) {
	if isNull(profile) {
		return profile, nil
	}
	var hash string
	var data []byte
	if ref, ok := parseBlobRef(profile); ok {
		blob, loaded := tv.blobs[ref]
		if !loaded {
			return nil, fmt.Errorf("profile references file content %s, which is not loaded", ref)
		}
		hash, data = ref, blob
	} else {
		inline, err := decodeInlineBytes(profile)
		if err != nil {
			return nil, err
		}
		hash, data = hashBytes(inline), inline
	}
	ref, err := encoder.reference(hash, data)
	if err != nil {
		return nil, err
	}
	return json.Marshal(ref)
}

func (tv *FileVariable) loadProfile(decoder Decoder, profile json.RawMessage) (json.RawMessage, error) {
	if isNull(profile) {
		return profile, nil
	}
	view, err := jbtf.Decode[bytesView](decoder.jbtf, profile)
	if err != nil {
		return nil, err
	}
	if view.SHA256 != "" {
		data, err := decoder.blob(view.SHA256)
		if err != nil {
			return nil, err
		}
		tv.retain(view.SHA256, data)
		return encodeBlobRef(view.SHA256), nil
	}
	if view.Value == nil {
		return nil, fmt.Errorf("file profile entry holds no buffer reference")
	}
	hash := hashBytes(view.Value.Data)
	tv.retain(hash, view.Value.Data)
	return encodeBlobRef(hash), nil
}

func (tv FileVariable) ToMessage() []byte {
	return tv.value
}

func (tv *FileVariable) NodeReference() nodes.Node {
	return &VariableReferenceNode[[]byte]{
		variable: tv,
	}
}

type fileDetails struct {
	Size int `json:"size"`
}

func (tv FileVariable) runtimeSchema() schema.Variable {
	v := schema.Variable{
		Type: "file",
		Value: fileDetails{
			Size: len(tv.value),
		},
	}
	if tv.info != nil {
		v.Description = tv.info.Description()
	}
	return v
}

type fileNodeGraphSchema struct {
	Type   string `json:"type"`
	Value  *jbtf.Bytes
	SHA256 string `json:"sha256"`
}

type fileNodeGraphRef struct {
	Type string `json:"type"`
	contentRef
}

func (tv FileVariable) toPersistantJSON(encoder Encoder) ([]byte, error) {
	saved := fileNodeGraphRef{Type: "file"}
	if tv.value != nil {
		ref, err := encoder.reference(tv.hash, tv.value)
		if err != nil {
			return nil, err
		}
		saved.contentRef = ref
	}
	return json.MarshalIndent(saved, "", "\t")
}

func (tv *FileVariable) fromPersistantJSON(decoder Decoder, body []byte) error {
	gn, err := jbtf.Decode[fileNodeGraphSchema](decoder.jbtf, body)
	if err != nil {
		return err
	}
	switch {
	case gn.SHA256 != "":
		data, err := decoder.blob(gn.SHA256)
		if err != nil {
			return err
		}
		tv.value, tv.hash = data, gn.SHA256
	case gn.Value != nil:
		tv.value, tv.hash = gn.Value.Data, hashBytes(gn.Value.Data)
	default:
		return nil
	}
	tv.retain(tv.hash, tv.value)
	return nil
}

func (tv *FileVariable) SwaggerProperty() swagger.Property {
	return swagger.Property{
		Type:        swagger.StringPropertyType,
		Format:      swagger.BinaryPropertyFormat,
		Description: tv.info.Description(),
	}
}
