package artifact

import (
	"io"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
)

type Gltf struct {
	Scene gltf.PolyformScene
}

func (Gltf) Mime() string {
	return "model/gltf-binary"
}

func (ga Gltf) Write(w io.Writer) error {
	return gltf.WriteBinary(ga.Scene, w)
}

type GltfNode = nodes.Struct[generator.Artifact, GltfNodeData]

type GltfNodeData struct {
	In nodes.NodeOutput[gltf.PolyformScene]
}

func (pn GltfNodeData) Process() (generator.Artifact, error) {
	return Gltf{Scene: pn.In.Value()}, nil
}

func NewGltfNode(bytesNode nodes.NodeOutput[gltf.PolyformScene]) nodes.NodeOutput[generator.Artifact] {
	return (&GltfNode{
		Data: GltfNodeData{
			In: bytesNode,
		},
	}).Out()
}
