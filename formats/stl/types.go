package stl

import (
	"bytes"
	"io"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[ReadNode](factory)
	refutil.RegisterType[ManifestNode](factory)
	generator.RegisterTypes(factory)
}

type ReadNode = nodes.Struct[ReadNodeData]

type ReadNodeData struct {
	Data nodes.Output[[]byte]
}

func (gad ReadNodeData) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if gad.Data == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	data := nodes.GetOutputValue(out, gad.Data)
	if len(data) == 0 {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	cloud, err := ReadMesh(bytes.NewReader(data))
	if err != nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		out.CaptureError(err)
		return
	}

	out.Set(*cloud)
}

// ============================================================================

type Artifact struct {
	Mesh modeling.Mesh
}

func (sa Artifact) Write(w io.Writer) error {
	return WriteMesh(w, sa.Mesh)
}

func (Artifact) Mime() string {
	return "application/sla"
}

type ManifestNode = nodes.Struct[ManifestNodeData]

type ManifestNodeData struct {
	Mesh nodes.Output[modeling.Mesh]
}

func (pn ManifestNodeData) Out(out *nodes.StructOutput[manifest.Manifest]) {
	entry := manifest.Entry{
		Artifact: Artifact{
			Mesh: nodes.TryGetOutputValue(out, pn.Mesh, modeling.EmptyMesh(modeling.TriangleTopology)),
		},
	}
	out.Set(manifest.SingleEntryManifest("model.stl", entry))
}
