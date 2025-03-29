package basics

import (
	"io"

	"github.com/EliCDavis/polyform/generator/artifact"
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

type IONode = nodes.Struct[IONodeData]

type IONodeData struct {
	In nodes.Output[io.Reader]
}

func (pn IONodeData) Out() nodes.StructOutput[artifact.Artifact] {
	return nodes.NewStructOutput[artifact.Artifact](IO{Reader: pn.In.Value()})
}
