package pgtf

import (
	"bytes"
	"encoding/json"
	"reflect"
)

type Decoder struct {
	buffers     []*bytes.Reader
	bufferViews []BufferView
}

func NewDecoder(graphJSON []byte) (Decoder, error) {
	type Schema struct {
		Buffers     []Buffer     `json:"buffers"`
		BufferViews []BufferView `json:"bufferViews"`
	}

	s := &Schema{}

	if err := json.Unmarshal(graphJSON, s); err != nil {
		return Decoder{}, err
	}

	return Decoder{
		bufferViews: s.BufferViews,
		buffers:     rebuildBuffers(s.Buffers),
	}, nil
}

func Decode[T any](d Decoder, data []byte) (T, error) {
	var thingToParse T
	err := json.Unmarshal(data, &thingToParse)
	if err != nil {
		return thingToParse, err
	}

	rawData := make(map[string]any)
	err = json.Unmarshal(data, &rawData)
	if err != nil {
		return thingToParse, err
	}

	return thingToParse, interpretBufferView(reflect.ValueOf(&thingToParse), rawData, d.bufferViews, d.buffers)

}
