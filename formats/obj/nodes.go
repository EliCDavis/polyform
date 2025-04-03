package obj

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

	refutil.RegisterType[ArtifactNode](factory)
	refutil.RegisterType[ReadNode](factory)

	generator.RegisterTypes(factory)
}

type Artifact struct {
	Scene        Scene
	MaterialFile string
}

func (sa Artifact) Write(w io.Writer) error {
	return Write(sa.Scene, sa.MaterialFile, w)
}

func (Artifact) Mime() string {
	return "model/obj"
}

type ArtifactNode = nodes.Struct[ArtifactNodeData]

type ArtifactNodeData struct {
	Scene        nodes.Output[Scene]
	MaterialFile nodes.Output[string]
}

func (pn ArtifactNodeData) Out() nodes.StructOutput[artifact.Artifact] {
	if pn.Scene == nil {
		return nodes.NewStructOutput[artifact.Artifact](Artifact{})
	}

	return nodes.NewStructOutput[artifact.Artifact](Artifact{
		Scene:        pn.Scene.Value(),
		MaterialFile: nodes.TryGetOutputValue(pn.MaterialFile, ""),
	})
}

type ReadNode = nodes.Struct[ReadNodeData]

type ReadNodeData struct {
	In nodes.Output[[]byte]
}

func (pn ReadNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	if pn.In == nil {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.TriangleTopology))
	}

	data := pn.In.Value()

	scene, _, err := ReadMesh(bytes.NewReader(data))
	if err != nil {
		output := nodes.NewStructOutput(modeling.EmptyMesh(modeling.TriangleTopology))
		output.LogError(err)
		return output
	}
	return nodes.NewStructOutput(scene.ToMesh())
}
