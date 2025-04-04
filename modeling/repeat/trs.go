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

type TRSNode = nodes.Struct[TRSNodeData]

type TRSNodeData struct {
	Input      nodes.Output[[]trs.TRS]
	Transforms nodes.Output[[]trs.TRS]
}

func (rnd TRSNodeData) Description() string {
	return "Duplicates the input transforms and transforms it for every TRS provided"
}

func (rnd TRSNodeData) Out() nodes.StructOutput[[]trs.TRS] {
	if rnd.Input == nil {
		return nodes.NewStructOutput(make([]trs.TRS, 0))
	}
	mesh := rnd.Input.Value()

	if rnd.Transforms == nil {
		return nodes.NewStructOutput(mesh)
	}
	transforms := rnd.Transforms.Value()

	return nodes.NewStructOutput(TRS(mesh, transforms))
}
