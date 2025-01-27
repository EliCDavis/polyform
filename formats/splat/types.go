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

type ArtifactNode = nodes.Struct[artifact.Artifact, ArtifactNodeData]

type ArtifactNodeData struct {
	In nodes.NodeOutput[modeling.Mesh]
}

func (pn ArtifactNodeData) Description() string {
	return "Mkkellogg's SPLAT format for their three.js Gaussian Splat Viewer"
}

func (pn ArtifactNodeData) Process() (artifact.Artifact, error) {
	return Splat{Mesh: pn.In.Value()}, nil
}

func NewSplatNode(meshNode nodes.NodeOutput[modeling.Mesh]) nodes.NodeOutput[artifact.Artifact] {
	return (&ArtifactNode{
		Data: ArtifactNodeData{
			In: meshNode,
		},
	}).Out()
}
