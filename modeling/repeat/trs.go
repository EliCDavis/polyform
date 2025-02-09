package repeat

import (
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
)

func TRS(input, transforms []trs.TRS) []trs.TRS {
	result := make([]trs.TRS, 0, len(transforms)*len(input))
	for _, transform := range transforms {
		for _, i := range input {
			result = append(result, i.Multiply(transform))
		}
	}
	return result
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
