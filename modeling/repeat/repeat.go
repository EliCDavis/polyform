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

func TRS(initialTransforms, transforms []trs.TRS) []trs.TRS {
	result := make([]trs.TRS, 0, len(transforms)*len(initialTransforms))
	for _, transform := range transforms {
		for _, i := range initialTransforms {
			result = append(result, i.Multiply(transform))
		}
	}
	return result
}

type MeshNode = nodes.Struct[modeling.Mesh, MeshNodeData]

type MeshNodeData struct {
	Mesh       nodes.NodeOutput[modeling.Mesh]
	Transforms nodes.NodeOutput[[]trs.TRS]
}

func (rnd MeshNodeData) Description() string {
	return "Duplicates the input mesh and transforms it for every TRS provided"
}

func (rnd MeshNodeData) Process() (modeling.Mesh, error) {
	if rnd.Mesh == nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}
	mesh := rnd.Mesh.Value()

	if rnd.Transforms == nil {
		return modeling.EmptyMesh(mesh.Topology()), nil
	}
	transforms := rnd.Transforms.Value()

	return Mesh(mesh, transforms), nil
}

type TRSNode = nodes.Struct[[]trs.TRS, TRSNodeData]

type TRSNodeData struct {
	Input      nodes.NodeOutput[[]trs.TRS]
	Transforms nodes.NodeOutput[[]trs.TRS]
}

func (rnd TRSNodeData) Description() string {
	return "Duplicates the input transforms and transforms it for every TRS provided"
}

func (rnd TRSNodeData) Process() ([]trs.TRS, error) {
	if rnd.Input == nil {
		return make([]trs.TRS, 0), nil
	}
	mesh := rnd.Input.Value()

	if rnd.Transforms == nil {
		return mesh, nil
	}
	transforms := rnd.Transforms.Value()

	return TRS(mesh, transforms), nil
}
