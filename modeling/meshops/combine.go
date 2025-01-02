package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
)

type CombineNode = nodes.Struct[modeling.Mesh, CombineNodeData]

type CombineNodeData struct {
	A nodes.NodeOutput[modeling.Mesh]
	B nodes.NodeOutput[modeling.Mesh]
}

func (cnd CombineNodeData) Process() (modeling.Mesh, error) {
	if cnd.A == nil && cnd.B == nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}

	if cnd.A == nil {
		return cnd.B.Value(), nil
	}

	if cnd.B == nil {
		return cnd.A.Value(), nil
	}

	return cnd.A.Value().Append(cnd.B.Value()), nil
}
