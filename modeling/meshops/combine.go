package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
)

type CombineNode = nodes.Struct[CombineNodeData]

type CombineNodeData struct {
	Meshes []nodes.Output[modeling.Mesh]
}

func (cnd CombineNodeData) Out(out *nodes.StructOutput[modeling.Mesh]) {
	fallback := modeling.EmptyMesh(modeling.TriangleTopology)

	meshes := nodes.GetOutputValues(out, cnd.Meshes)
	if len(meshes) == 0 {
		out.Set(fallback)
		return
	}

	result := meshes[0]
	for i := 1; i < len(meshes); i++ {
		result = result.Append(meshes[i])
	}

	out.Set(result)
}
