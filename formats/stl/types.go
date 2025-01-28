package stl

import (
	"bytes"
	"io"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[ReadNode](factory)
	refutil.RegisterType[ArtifactNode](factory)
	generator.RegisterTypes(factory)
}

type ReadNode = nodes.Struct[modeling.Mesh, ReadNodeData]

type ReadNodeData struct {
	Data nodes.NodeOutput[[]byte]
}

func (gad ReadNodeData) Process() (modeling.Mesh, error) {
	if gad.Data == nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}

	data := gad.Data.Value()
	if len(data) == 0 {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}

	cloud, err := ReadMesh(bytes.NewReader(data))
	if err != nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), err
	}

	return *cloud, nil
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

type ArtifactNode = nodes.Struct[artifact.Artifact, ArtifactNodeData]

type ArtifactNodeData struct {
	In nodes.NodeOutput[modeling.Mesh]
}

func (pn ArtifactNodeData) Process() (artifact.Artifact, error) {
	return Artifact{Mesh: pn.In.Value()}, nil
}
