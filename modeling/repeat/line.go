package repeat

import (
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type Line struct {
	// Start of the line
	Start vector3.Float64

	// End of the line
	End vector3.Float64

	// How many TRS matrices to produce
	Samples int

	// If true, the start and end points are not included in the resulting
	// array of TRS values
	Exclusive bool
}

func (l Line) TRS() []trs.TRS {
	if l.Samples == 0 {
		return make([]trs.TRS, 0)
	}

	if l.Samples == 1 {
		return []trs.TRS{trs.Position(vector3.Midpoint(l.Start, l.End))}
	}

	values := make([]trs.TRS, 0, l.Samples)
	if !l.Exclusive {
		values = append(values, trs.Position(l.Start))
	}

	inbetweenSamples := l.Samples
	if !l.Exclusive {
		inbetweenSamples -= 2
	}

	dir := l.End.Sub(l.Start)
	inc := dir.DivByConstant(float64(inbetweenSamples + 1))
	for i := 0; i < inbetweenSamples; i++ {
		values = append(values, trs.Position(l.Start.Add(inc.Scale(float64(i+1)))))
	}

	if !l.Exclusive {
		values = append(values, trs.Position(l.End))
	}

	return values
}

type LineNode = nodes.Struct[LineNodeData]

type LineNodeData struct {
	Start     nodes.Output[vector3.Float64]
	End       nodes.Output[vector3.Float64]
	Samples   nodes.Output[int]
	Exclusive nodes.Output[bool]
}

func (r LineNodeData) Out(out *nodes.StructOutput[[]trs.TRS]) {
	line := Line{
		Start:     nodes.TryGetOutputValue(out, r.Start, vector3.Zero[float64]()),
		End:       nodes.TryGetOutputValue(out, r.End, vector3.Zero[float64]()),
		Samples:   max(nodes.TryGetOutputValue(out, r.Samples, 0), 0),
		Exclusive: nodes.TryGetOutputValue(out, r.Exclusive, false),
	}
	out.Set(line.TRS())
}
