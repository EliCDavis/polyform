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
	Input      nodes.Output[[]trs.TRS] `description:"The base set of transforms to duplicate. Empty output if unconnected."`
	Transforms nodes.Output[[]trs.TRS] `description:"One outer transform per group. Input is copied once per entry here and each copy is further transformed by it. If unconnected, Input passes through unchanged."`
}

func (rnd TRSNode) Description() string {
	return "Combines two transform lists by cross product: every transform in Input is duplicated once per entry in Transforms, with that entry applied on top — producing len(Input) * len(Transforms) results."
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
