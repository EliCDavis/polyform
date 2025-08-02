package basics

import (
	"io"

	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/nodes"
)

type TextArtifact struct {
	Data string
}

func (ta TextArtifact) Write(w io.Writer) error {
	_, err := w.Write([]byte(ta.Data))
	return err
}

func (TextArtifact) Mime() string {
	return "text/plain"
}

// ============================================================================

type TextNode struct {
	In   nodes.Output[string]
	Name nodes.Output[string]
}

func (tand TextNode) Out(out *nodes.StructOutput[manifest.Manifest]) {
	name := nodes.TryGetOutputValue(out, tand.Name, "text.txt")
	entry := manifest.Entry{Artifact: TextArtifact{Data: nodes.TryGetOutputValue(out, tand.In, "")}}
	out.Set(manifest.SingleEntryManifest(name, entry))
}
