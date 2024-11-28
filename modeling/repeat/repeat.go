package repeat

import (
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type Node = nodes.StructNode[modeling.Mesh, NodeData]

type NodeData struct {
	Mesh     nodes.NodeOutput[modeling.Mesh]
	Position nodes.NodeOutput[[]vector3.Float64]
	Rotation nodes.NodeOutput[[]quaternion.Quaternion]
	Scale    nodes.NodeOutput[[]vector3.Float64]
}

func (rnd NodeData) Process() (modeling.Mesh, error) {
	if rnd.Mesh == nil || rnd.Position == nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}

	mesh := rnd.Mesh.Value()
	positions := rnd.Position.Value()
	var rotations []quaternion.Quaternion
	var scales []vector3.Float64

	if rnd.Rotation != nil {
		rotations = rnd.Rotation.Value()
	}

	if rnd.Scale != nil {
		scales = rnd.Scale.Value()
	}

	result := modeling.EmptyMesh(modeling.TriangleTopology)
	for i, p := range positions {
		s := vector3.One[float64]()
		r := quaternion.Identity()

		if i < len(rotations) {
			r = rotations[i]
		}

		if i < len(scales) {
			s = scales[i]
		}

		result = result.Append(mesh.Scale(s).Rotate(r).Translate(p))
	}

	return result, nil
}
