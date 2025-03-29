package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
)

type CombineNode = nodes.Struct[CombineNodeData]

type CombineNodeData struct {
	A nodes.Output[modeling.Mesh]
	B nodes.Output[modeling.Mesh]
}

func (cnd CombineNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	if cnd.A == nil && cnd.B == nil {
		return nodes.NewStructOutput(modeling.EmptyMesh(modeling.TriangleTopology))
	}

	if cnd.A == nil {
		return nodes.NewStructOutput(cnd.B.Value())
	}

	if cnd.B == nil {
		return nodes.NewStructOutput(cnd.A.Value())
	}

	return nodes.NewStructOutput(cnd.A.Value().Append(cnd.B.Value()))
}
