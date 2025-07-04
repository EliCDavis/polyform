package generator

import (
	"io"
)

func WriteMermaid(a App, out io.Writer) error {
	a.initGraphInstance()
	return a.Graph.WriteMermaid(out)
}
