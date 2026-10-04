package repeat

import (
	"fmt"

	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
)

func TRS(input, transforms []trs.TRS) ([]trs.TRS, error) {
	result := make([]trs.TRS, 0, len(transforms)*len(input))
	for ti, transform := range transforms {
		for ii, i := range input {
			composed, err := trs.FromMatrix(i.Multiply(transform))
			if err != nil {
				return nil, fmt.Errorf("input %d against transform %d: %w", ii, ti, err)
			}
			result = append(result, composed)
		}
	}
	return result, nil
}

type TRSNode struct {
	Input      nodes.Output[[]trs.TRS] `description:"The parent transforms. Empty output if unconnected."`
	Transforms nodes.Output[[]trs.TRS] `description:"Applied inside each Input transform's own frame, so a rotation here turns each copy about its Input transform, not about the origin. To arrange a whole group around the origin, swap the two ports. If unconnected, Input passes through unchanged."`
}

func (rnd TRSNode) Description() string {
	return "Every pairing of the two lists, as Input[i] * Transforms[j]: Input is the parent and Transforms is applied in its local frame. Results are grouped by Transforms entry, len(Input) * len(Transforms) in all."
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
	result, err := TRS(mesh, transforms)
	if err != nil {
		out.CaptureError(err)
		return
	}
	out.Set(result)
}
