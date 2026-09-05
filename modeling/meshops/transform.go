package meshops

import (
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
)

// TransformNode moves a whole mesh, the mesh-side counterpart to
// math/sdf's TransformNode. The attribute-scoped nodes next to this one
// (Rotate/Translate/Scale Attribute 3D) each move exactly one named
// attribute, so transforming a mesh correctly through them means wiring one
// per attribute and remembering that normals need the same rotation as
// positions - which is easy to forget, and errors out entirely on a mesh
// that has no normals to rotate.
type TransformNode struct {
	Mesh nodes.Output[modeling.Mesh] `description:"The mesh to transform. Empty mesh if unconnected."`
	TRS  nodes.Output[trs.TRS]       `description:"Translation, rotation and scale to apply. If unconnected, Mesh passes through unchanged."`
}

func (tnd TransformNode) Description() string {
	return "Applies a TRS to an entire mesh, rotating its normals along with its positions."
}

func (tnd TransformNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if tnd.Mesh == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	mesh := nodes.GetOutputValue(out, tnd.Mesh)
	if tnd.TRS == nil {
		out.Set(mesh)
		return
	}

	out.Set(mesh.ApplyTRS(nodes.GetOutputValue(out, tnd.TRS)))
}
