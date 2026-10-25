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
	Input      nodes.LiftedPort[[]trs.TRS] `description:"The parent transforms. Empty output if unconnected."`
	Transforms nodes.LiftedPort[[]trs.TRS] `description:"Applied inside each Input transform's own frame, so a rotation here turns each copy about its Input transform. If unconnected, Input passes through unchanged."`
}

func (rnd TRSNode) Description() string {
	return "Every pairing of the two lists, as Input[i] * Transforms[j]: Input is the parent and Transforms is applied in its local frame. Results are grouped by Transforms entry, len(Input) * len(Transforms) in all."
}

func (rnd TRSNode) Out(out *nodes.Lifted[[]trs.TRS]) {
	input := nodes.LiftedOr(rnd.Input, []trs.TRS{})
	if rnd.Transforms == nil {
		nodes.Zip1(out, input, func(parents []trs.TRS) []trs.TRS { return parents })
		return
	}

	nodes.Zip2(out, input, rnd.Transforms, func(parents, transforms []trs.TRS) []trs.TRS {
		result, err := TRS(parents, transforms)
		out.CaptureError(err)
		return result
	})
}
