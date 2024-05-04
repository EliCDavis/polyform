package pgtf

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
)

type Encoder struct {
	currentBuffer int
	buffers       []*bytes.Buffer
	bufferViews   []BufferView
}

func (e *Encoder) StartNewBuffer() {
	if len(e.buffers) == 0 {
		// We've never even initialized yet, but they expect us to be on buffer
		// 2 so, add 2.
		e.buffers = append(e.buffers, &bytes.Buffer{}, &bytes.Buffer{})
	} else {
		e.buffers = append(e.buffers, &bytes.Buffer{})
	}
	e.currentBuffer++
}

func (e *Encoder) Marshal(v any) ([]byte, error) {
	if len(e.buffers) == 0 {
		e.buffers = append(e.buffers, &bytes.Buffer{})
	}
	curBuf := e.buffers[e.currentBuffer]

	if isStruct(v) {
		jsonInterprettedData, err := toJsonMap(v)
		if err != nil {
			return nil, err
		}
		data := make(map[string]any)
		newBufs, err := buildBufferView(curBuf, data, jsonInterprettedData, v, e.bufferViews, e.currentBuffer)
		if err != nil {
			return nil, err
		}
		e.bufferViews = newBufs
		return json.MarshalIndent(data, "", "\t")
	}
	return json.MarshalIndent(v, "", "\t")
}

func (e *Encoder) ToPgtf(v any) ([]byte, error) {
	structure := &schema{
		Data:        v,
		Buffers:     make([]Buffer, 0),
		BufferViews: make([]BufferView, 0),
	}

	if len(e.buffers) == 0 {
		e.buffers = append(e.buffers, &bytes.Buffer{})
	}
	curBuf := e.buffers[e.currentBuffer]

	if isStruct(v) {
		jsonInterprettedData, err := toJsonMap(v)
		if err != nil {
			return nil, err
		}
		data := make(map[string]any)
		views, err := buildBufferView(curBuf, data, jsonInterprettedData, v, e.bufferViews, e.currentBuffer)
		if err != nil {
			return nil, err
		}
		structure.Data = data
		structure.BufferViews = views
	}

	structure.Buffers = make([]Buffer, len(e.buffers))
	for i, buf := range e.buffers {
		bufData := buf.Bytes()
		structure.Buffers[i] = Buffer{
			ByteLength: len(bufData),
			URI:        "data:application/octet-stream;base64," + base64.StdEncoding.EncodeToString(bufData),
		}
	}

	return json.MarshalIndent(structure, "", "\t")

}
