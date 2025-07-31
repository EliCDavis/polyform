package repeat

import (
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
)

type Transformation struct {
	Initial        trs.TRS
	Transformation trs.TRS
	Samples        int
}

func (t Transformation) TRS() []trs.TRS {
	results := make([]trs.TRS, t.Samples)

	previous := t.Initial
	for i := range t.Samples {
		results[i] = t.Transformation.Multiply(previous)
		previous = results[i]
	}

	return results
}

type TransformationNode = nodes.Struct[TransformationNodeData]

type TransformationNodeData struct {
	Initial        nodes.Output[trs.TRS]
	Transformation nodes.Output[trs.TRS]
	Samples        nodes.Output[int]
}

func (rnd TransformationNodeData) Out() nodes.StructOutput[[]trs.TRS] {
	out := nodes.StructOutput[[]trs.TRS]{}
	out.Set(Transformation{
		Initial:        nodes.TryGetOutputValue(&out, rnd.Initial, trs.Identity()),
		Transformation: nodes.TryGetOutputValue(&out, rnd.Transformation, trs.Identity()),
		Samples:        nodes.TryGetOutputValue(&out, rnd.Samples, 0),
	}.TRS())
	return out
}
