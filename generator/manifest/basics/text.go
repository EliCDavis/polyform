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

type TextManifest struct {
	Artifact TextArtifact
	Name     string
}

func (ta TextManifest) Main() string {
	if ta.Name == "" {
		return "text.txt"
	}
	return ta.Name
}

func (ta TextManifest) Artifacts() map[string]manifest.Entry {
	return map[string]manifest.Entry{
		ta.Main(): {
			Artifact: ta.Artifact,
		},
	}
}

// ============================================================================

type TextNode = nodes.Struct[TextNodeData]

type TextNodeData struct {
	In   nodes.Output[string]
	Name nodes.Output[string]
}

func (tand TextNodeData) Out() nodes.StructOutput[manifest.Manifest] {
	return nodes.NewStructOutput[manifest.Manifest](TextManifest{
		Name:     nodes.TryGetOutputValue(tand.Name, "text") + ".txt",
		Artifact: TextArtifact{Data: nodes.TryGetOutputValue(tand.In, "")},
	})
}
