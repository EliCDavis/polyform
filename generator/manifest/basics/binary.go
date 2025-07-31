package basics

import (
	"io"

	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/nodes"
)

type BinaryNode = nodes.Struct[BinaryNodeData]

type Binary struct {
	Data []byte
}

func (ba Binary) Write(w io.Writer) error {
	_, err := w.Write(ba.Data)
	return err
}

func (Binary) Mime() string {
	return "application/octet-stream"
}

type BinaryNodeData struct {
	In nodes.Output[[]byte]
}

func (pn BinaryNodeData) Out() nodes.StructOutput[manifest.Artifact] {
	out := nodes.StructOutput[manifest.Artifact]{}
	out.Set(Binary{Data: nodes.TryGetOutputValue(&out, pn.In, []byte{})})
	return out
}
