package repeat

import (
	"fmt"

	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
)

type Transformation struct {
	Initial        trs.TRS
	Transformation trs.TRS
	Samples        int
}

func (t Transformation) TRS() ([]trs.TRS, error) {
	results := make([]trs.TRS, t.Samples)

	previous := t.Initial
	for i := range t.Samples {
		composed, err := trs.FromMatrix(t.Transformation.Multiply(previous))
		if err != nil {
			return nil, fmt.Errorf("sample %d: %w", i, err)
		}
		results[i] = composed
		previous = results[i]
	}

	return results, nil
}

type TransformationNode struct {
	Initial        nodes.Output[trs.TRS] `description:"Starting transform (the first copy's placement before any repetition is applied). Defaults to identity."`
	Transformation nodes.Output[trs.TRS] `description:"The transform applied on top of the previous copy to produce the next one. Applied once for copy 1, twice for copy 2, and so on (cumulative, not reset each time). Defaults to identity (no change between copies)."`
	Samples        nodes.Output[int]     `description:"How many copies to produce. Defaults to 0."`
}

func (rnd TransformationNode) Description() string {
	return "Produces Samples transforms by repeatedly applying Transformation on top of the previous result, starting from Initial (Cumulative). Each copy's placement depends on all the ones before it."
}

func (rnd TransformationNode) Out(out *nodes.StructOutput[[]trs.TRS]) {
	result, err := Transformation{
		Initial:        nodes.TryGetOutputValue(out, rnd.Initial, trs.Identity()),
		Transformation: nodes.TryGetOutputValue(out, rnd.Transformation, trs.Identity()),
		Samples:        nodes.TryGetOutputValue(out, rnd.Samples, 0),
	}.TRS()
	if err != nil {
		out.CaptureError(err)
		return
	}
	out.Set(result)
}
