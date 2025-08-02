package splat

import (
	"io"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[ManifestNode](factory)

	generator.RegisterTypes(factory)
}

type Splat struct {
	Mesh modeling.Mesh
}

func (sa Splat) Write(w io.Writer) error {
	return Write(w, sa.Mesh)
}

func (Splat) Mime() string {
	return "application/octet-stream"
}

type ManifestNode = nodes.Struct[ManifestNodeData]

type ManifestNodeData struct {
	In nodes.Output[modeling.Mesh]
}

func (pn ManifestNodeData) Description() string {
	return "Mkkellogg's SPLAT format for their three.js Gaussian Splat Viewer"
}

func (pn ManifestNodeData) Out(out *nodes.StructOutput[manifest.Manifest]) {
	entry := manifest.Entry{
		Artifact: Splat{Mesh: nodes.TryGetOutputValue(out, pn.In, modeling.EmptyPointcloud())},
	}
	out.Set(manifest.SingleEntryManifest("model.splat", entry))
}
