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

type IONode struct {
	In nodes.Output[io.Reader]
}

func (pn IONode) Description() string {
	return "Writes whatever a reader produces out as a file."
}

func (pn IONode) Out(out *nodes.StructOutput[manifest.Artifact]) {
	out.Set(IO{Reader: nodes.TryGetOutputValue(out, pn.In, nil)})
}
