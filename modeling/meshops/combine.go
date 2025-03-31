package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
)

type CombineNode = nodes.Struct[CombineNodeData]

type CombineNodeData struct {
	Meshes []nodes.Output[modeling.Mesh]
}

func (cnd CombineNodeData) Out() nodes.StructOutput[modeling.Mesh] {

	fallback := modeling.EmptyMesh(modeling.TriangleTopology)

	if len(cnd.Meshes) == 0 {
		return nodes.NewStructOutput(fallback)
	}

	if len(cnd.Meshes) == 1 {
		return nodes.NewStructOutput(nodes.TryGetOutputValue(cnd.Meshes[0], fallback))
	}

	result := nodes.TryGetOutputValue(cnd.Meshes[0], fallback)
	for i := 1; i < len(cnd.Meshes); i++ {
		if cnd.Meshes[i] == nil {
			continue
		}
		result = result.Append(cnd.Meshes[i].Value())
	}

	return nodes.NewStructOutput(result)
}
