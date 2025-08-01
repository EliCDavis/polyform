package basics

import (
	"io"

	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/nodes"
)

type IO struct {
	Reader io.Reader
}

func (ga IO) Write(w io.Writer) error {
	if ga.Reader == nil {
		return nil
	}
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

func (pn IONodeData) Out(out *nodes.StructOutput[manifest.Artifact]) {
	out.Set(IO{Reader: nodes.TryGetOutputValue(out, pn.In, nil)})
}
