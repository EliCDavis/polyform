package artifact

import (
	"io"

	"github.com/EliCDavis/polyform/formats/splat"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
)

type Splat struct {
	Mesh modeling.Mesh
}

func (sa Splat) Write(w io.Writer) error {
	return splat.Write(w, sa.Mesh)
}

func (Splat) Mime() string {
	return "application/octet-stream"
}

type SplatNode = nodes.StructNode[generator.Artifact, SplatNodeData]

type SplatNodeData struct {
	In nodes.NodeOutput[modeling.Mesh]
}

func (pn SplatNodeData) Process() (generator.Artifact, error) {
	return Splat{Mesh: pn.In.Value()}, nil
}

func NewSplatNode(meshNode nodes.NodeOutput[modeling.Mesh]) nodes.NodeOutput[generator.Artifact] {
	return (&SplatNode{
		Data: SplatNodeData{
			In: meshNode,
		},
	}).Out()
}
