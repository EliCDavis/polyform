package repeat

import (
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
)

func Mesh(mesh modeling.Mesh, transforms []trs.TRS) modeling.Mesh {
	result := modeling.EmptyMesh(mesh.Topology())
	for _, transform := range transforms {
		result = result.Append(mesh.ApplyTRS(transform))
	}
	return result
}

type MeshNode = nodes.Struct[modeling.Mesh, MeshNodeData]

type MeshNodeData struct {
	Mesh       nodes.NodeOutput[modeling.Mesh]
	Transforms nodes.NodeOutput[[]trs.TRS]
}

func (rnd MeshNodeData) Description() string {
	return "Duplicates and transforms the input mesh for every TRS provided"
}

func (rnd MeshNodeData) Process() (modeling.Mesh, error) {
	if rnd.Mesh == nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}
	mesh := rnd.Mesh.Value()

	if rnd.Transforms == nil {
		return modeling.EmptyMesh(mesh.Topology()), nil
	}
	transforms := rnd.Transforms.Value()

	return Mesh(mesh, transforms), nil
}
