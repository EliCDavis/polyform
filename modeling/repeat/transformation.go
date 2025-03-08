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
	for i := 0; i < t.Samples; i++ {
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
	return nodes.NewStructOutput(Transformation{
		Initial:        nodes.TryGetOutputValue(rnd.Initial, trs.Identity()),
		Transformation: nodes.TryGetOutputValue(rnd.Transformation, trs.Identity()),
		Samples:        nodes.TryGetOutputValue(rnd.Samples, 0),
	}.TRS())
}
