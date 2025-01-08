package artifact

import (
	"io"

	"github.com/EliCDavis/polyform/nodes"
)

type IO struct {
	Reader io.Reader
}

func (ga IO) Write(w io.Writer) error {
	_, err := io.Copy(w, ga.Reader)
	return err
}

func (IO) Mime() string {
	return "application/octet-stream"
}

type IONode = nodes.Struct[Artifact, IONodeData]

type IONodeData struct {
	In nodes.NodeOutput[io.Reader]
}

func (pn IONodeData) Process() (Artifact, error) {
	return IO{Reader: pn.In.Value()}, nil
}

func NewIONode(readerNode nodes.NodeOutput[io.Reader]) nodes.NodeOutput[Artifact] {
	return (&IONode{
		Data: IONodeData{
			In: readerNode,
		},
	}).Out()
}
