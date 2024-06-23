package artifact

import (
	"io"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
)

type BinaryNode = nodes.StructNode[generator.Artifact, BinaryNodeData]

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
	In nodes.NodeOutput[[]byte]
}

func (pn BinaryNodeData) Process() (generator.Artifact, error) {
	return Binary{Data: pn.In.Value()}, nil
}

func NewBinaryNode(bytesNode nodes.NodeOutput[[]byte]) nodes.NodeOutput[generator.Artifact] {
	return (&BinaryNode{
		Data: BinaryNodeData{
			In: bytesNode,
		},
	}).Out()
}
