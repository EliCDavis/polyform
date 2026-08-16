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

type MeshNode struct {
	Mesh       nodes.Output[modeling.Mesh] `description:"The mesh to duplicate. Empty mesh if unconnected."`
	Transforms nodes.Output[[]trs.TRS]     `description:"One transform per copy to place. If unconnected, Mesh passes through unchanged (one copy, untransformed)."`
}

func (rnd MeshNode) Description() string {
	return "Duplicates Mesh once per entry in Transforms and bakes every copy into one combined mesh."
}

func (rnd MeshNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if rnd.Mesh == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	mesh := nodes.GetOutputValue(out, rnd.Mesh)
	if rnd.Transforms == nil {
		out.Set(mesh)
		return
	}

	out.Set(Mesh(mesh, nodes.GetOutputValue(out, rnd.Transforms)))
}
