package splat

import (
	"io"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[ArtifactNode](factory)

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

type ArtifactNode = nodes.Struct[ArtifactNodeData]

type ArtifactNodeData struct {
	In nodes.Output[modeling.Mesh]
}

func (pn ArtifactNodeData) Description() string {
	return "Mkkellogg's SPLAT format for their three.js Gaussian Splat Viewer"
}

func (pn ArtifactNodeData) Out() nodes.StructOutput[artifact.Artifact] {
	return nodes.NewStructOutput[artifact.Artifact](Splat{Mesh: pn.In.Value()})
}
