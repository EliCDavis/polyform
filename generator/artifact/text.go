package artifact

import (
	"io"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
)

type Text struct {
	Data string
}

func (ta Text) Write(w io.Writer) error {
	_, err := w.Write([]byte(ta.Data))
	return err
}

func (Text) Mime() string {
	return "text/plain"
}

type TextNode = nodes.StructNode[generator.Artifact, TextNodeData]

type TextNodeData struct {
	In nodes.NodeOutput[string]
}

func (tand TextNodeData) Process() (generator.Artifact, error) {
	return Text{Data: tand.In.Value()}, nil
}

func NewTextNode(textNode nodes.NodeOutput[string]) nodes.NodeOutput[generator.Artifact] {
	return (&TextNode{
		Data: TextNodeData{
			In: textNode,
		},
	}).Out()
}
