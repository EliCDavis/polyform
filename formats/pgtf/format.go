package pgtf

import "io"

type format = typedFormat[any]

type buffer struct {
	ByteLength int    `json:"byteLength"`    // The length of the buffer in bytes.
	URI        string `json:"uri,omitempty"` // The URI (or IRI) of the buffer.  Relative paths are relative to the current glTF asset.  Instead of referencing an external file, this field **MAY** contain a `data:`-URI.
}

type bufferView struct {
	Buffer     int `json:"buffer"`               // The index of the buffer
	ByteOffset int `json:"byteOffset,omitempty"` // The offset into the buffer in bytes.
	ByteLength int `json:"byteLength"`           // The length of the bufferView in bytes.
}

type typedFormat[T any] struct {
	Buffers     []buffer     `json:"buffers,omitempty"`
	BufferViews []bufferView `json:"bufferViews,omitempty"`
	Data        T            `json:"data"`
}

type PgtfSerializable interface {
	Deserialize(io.Reader) error
	Serialize(io.Writer) error
}
