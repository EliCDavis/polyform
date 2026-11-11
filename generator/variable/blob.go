package variable

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/EliCDavis/jbtf"
)

// Encoder writes equal byte content to one buffer view per save. Given a
// blobs map, it keeps the content there instead and references it by hash.
type Encoder struct {
	jbtf  *jbtf.Encoder
	views map[string]int
	blobs map[string][]byte
}

func NewEncoder(encoder *jbtf.Encoder, blobs map[string][]byte) Encoder {
	return Encoder{jbtf: encoder, views: make(map[string]int), blobs: blobs}
}

func (e Encoder) Marshal(v any) ([]byte, error) {
	return e.jbtf.Marshal(v)
}

type Decoder struct {
	jbtf  jbtf.Decoder
	blobs map[string][]byte
}

func NewDecoder(decoder jbtf.Decoder, blobs map[string][]byte) Decoder {
	return Decoder{jbtf: decoder, blobs: blobs}
}

func (d Decoder) blob(hash string) ([]byte, error) {
	data, ok := d.blobs[hash]
	if !ok {
		return nil, fmt.Errorf("file content %s is not held alongside this graph", hash)
	}
	return data, nil
}

type bytesView struct {
	Value  *jbtf.Bytes
	SHA256 string `json:"sha256,omitempty"`
}

type bytesViewRef struct {
	Value int `json:"$Value"`
}

type contentRef struct {
	Value  *int   `json:"$Value,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
}

func (e Encoder) reference(hash string, data []byte) (contentRef, error) {
	if e.blobs != nil {
		e.blobs[hash] = data
		return contentRef{SHA256: hash}, nil
	}
	index, err := e.view(hash, data)
	if err != nil {
		return contentRef{}, err
	}
	return contentRef{Value: &index}, nil
}

func (e Encoder) view(hash string, data []byte) (int, error) {
	if index, ok := e.views[hash]; ok {
		return index, nil
	}
	encoded, err := e.jbtf.Marshal(bytesView{Value: &jbtf.Bytes{Data: data}})
	if err != nil {
		return 0, err
	}
	var ref bytesViewRef
	if err := json.Unmarshal(encoded, &ref); err != nil {
		return 0, err
	}
	e.views[hash] = ref.Value
	return ref.Value, nil
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

type blobRef struct {
	SHA256 string `json:"sha256"`
}

func parseBlobRef(raw json.RawMessage) (string, bool) {
	var ref blobRef
	if json.Unmarshal(raw, &ref) != nil || ref.SHA256 == "" {
		return "", false
	}
	return ref.SHA256, true
}

func encodeBlobRef(hash string) json.RawMessage {
	raw, _ := json.Marshal(blobRef{SHA256: hash})
	return raw
}

func isNull(raw json.RawMessage) bool {
	var v any
	return json.Unmarshal(raw, &v) == nil && v == nil
}

func decodeInlineBytes(raw json.RawMessage) ([]byte, error) {
	var data []byte
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("file profile entry is neither a content reference nor base64 bytes: %w", err)
	}
	return data, nil
}
