package meshops

import (
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
)

type TransformNode struct {
	Mesh nodes.LiftedPort[modeling.Mesh] `description:"The mesh to transform. Empty mesh if unconnected."`
	TRS  nodes.LiftedPort[trs.TRS]       `description:"Translation, rotation and scale to apply. If unconnected, Mesh passes through unchanged."`
}

func (tnd TransformNode) Description() string {
	return "Applies a TRS to an entire mesh, rotating its normals along with its positions."
}

func (tnd TransformNode) Out(out *nodes.Lifted[modeling.Mesh]) {
	nodes.Zip2(
		out,
		nodes.LiftedOr(tnd.Mesh, modeling.EmptyMesh(modeling.TriangleTopology)),
		nodes.LiftedOr(tnd.TRS, trs.Identity()),
		func(mesh modeling.Mesh, transform trs.TRS) modeling.Mesh {
			if tnd.TRS == nil {
				return mesh
			}
			return mesh.ApplyTRS(transform)
		},
	)
}
