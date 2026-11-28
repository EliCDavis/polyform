package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func PlaceOnGround(m modeling.Mesh) modeling.Mesh {
	if m.AttributeLength() == 0 || !m.HasFloat3Attribute(modeling.PositionAttribute) {
		return m
	}
	lowest := m.BoundingBox(modeling.PositionAttribute).Min().Y()
	return TranslateAttribute3D(m, modeling.PositionAttribute, vector3.New(0, -lowest, 0))
}

type PlaceOnGroundNode struct {
	Mesh nodes.LiftedPort[modeling.Mesh]
}

func (PlaceOnGroundNode) Description() string {
	return "Moves a mesh up or down so its lowest point sits at y = 0."
}

func (n PlaceOnGroundNode) Out(out *nodes.Lifted[modeling.Mesh]) {
	nodes.Zip1(out, nodes.LiftedOr(n.Mesh, modeling.EmptyMesh(modeling.TriangleTopology)), PlaceOnGround)
}
