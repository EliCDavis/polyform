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

type TRSNode struct {
	Input      nodes.Output[[]trs.TRS]
	Transforms nodes.Output[[]trs.TRS]
}

func (rnd TRSNode) Description() string {
	return "Duplicates the input transforms and transforms it for every TRS provided"
}

func (rnd TRSNode) Out(out *nodes.StructOutput[[]trs.TRS]) {
	if rnd.Input == nil {
		out.Set(make([]trs.TRS, 0))
		return
	}

	mesh := nodes.GetOutputValue(out, rnd.Input)
	if rnd.Transforms == nil {
		out.Set(mesh)
		return
	}
	transforms := nodes.GetOutputValue(out, rnd.Transforms)
	out.Set(TRS(mesh, transforms))
}
