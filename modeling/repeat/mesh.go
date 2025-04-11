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

type MeshNode = nodes.Struct[MeshNodeData]

type MeshNodeData struct {
	Mesh       nodes.Output[modeling.Mesh]
	Transforms nodes.Output[[]trs.TRS]
}

func (rnd MeshNodeData) Description() string {
	return "Duplicates and transforms the input mesh for every TRS provided"
}

func (rnd MeshNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	if rnd.Mesh == nil {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.TriangleTopology))
	}
	mesh := rnd.Mesh.Value()

	if rnd.Transforms == nil {
		return nodes.NewStructOutput(modeling.EmptyMesh(mesh.Topology()))
	}
	return nodes.NewStructOutput(Mesh(mesh, rnd.Transforms.Value()))
}
