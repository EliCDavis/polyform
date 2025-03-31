package basics

import (
	"io"

	"github.com/EliCDavis/polyform/generator/artifact"
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

type TextNode = nodes.Struct[TextNodeData]

type TextNodeData struct {
	In nodes.Output[string]
}

func (tand TextNodeData) Out() nodes.StructOutput[artifact.Artifact] {
	return nodes.NewStructOutput[artifact.Artifact](Text{Data: tand.In.Value()})
}
