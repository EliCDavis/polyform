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
	Mesh       nodes.Output[modeling.Mesh]
	Transforms nodes.Output[[]trs.TRS]
}

func (rnd MeshNode) Description() string {
	return "Duplicates and transforms the input mesh for every TRS provided"
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
