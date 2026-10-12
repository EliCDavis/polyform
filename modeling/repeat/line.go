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

	// If true, the line is split into Samples equal segments and a TRS is
	// placed at the center of each, so the start and end points are excluded
	Exclusive bool
}

func (l Line) TRS() []trs.TRS {
	if l.Samples == 0 {
		return make([]trs.TRS, 0)
	}

	if l.Exclusive {
		dir := l.End.Sub(l.Start)
		values := make([]trs.TRS, l.Samples)
		for i := range values {
			t := (float64(i) + 0.5) / float64(l.Samples)
			values[i] = trs.Position(l.Start.Add(dir.Scale(t)))
		}
		return values
	}

	if l.Samples == 1 {
		return []trs.TRS{trs.Position(vector3.Midpoint(l.Start, l.End))}
	}

	values := make([]trs.TRS, 0, l.Samples)
	values = append(values, trs.Position(l.Start))

	inbetweenSamples := l.Samples - 2
	dir := l.End.Sub(l.Start)
	inc := dir.DivByConstant(float64(inbetweenSamples + 1))
	for i := 0; i < inbetweenSamples; i++ {
		values = append(values, trs.Position(l.Start.Add(inc.Scale(float64(i+1)))))
	}

	values = append(values, trs.Position(l.End))
	return values
}

type LineNode struct {
	Start     nodes.LiftedPort[vector3.Float64] `description:"Position of the first sample. Defaults to the origin."`
	End       nodes.LiftedPort[vector3.Float64] `description:"Position of the last sample. Defaults to the origin."`
	Samples   nodes.LiftedPort[int]             `description:"How many TRS matrices to produce, evenly spaced between Start and End (inclusive unless Exclusive is true). A single sample sits at the midpoint, Exclusive or not. Defaults to 0, clamped to 0 minimum."`
	Exclusive nodes.LiftedPort[bool]            `description:"If true, the line is split into Samples equal segments with one TRS at the center of each, so Start and End are not included. Defaults to false."`
}

func (r LineNode) Description() string {
	return "Produces Samples transforms spaced evenly along a straight line from Start to End, identity rotation."
}

func (r LineNode) Out(out *nodes.Lifted[[]trs.TRS]) {
	nodes.Zip4(
		out,
		r.Start,
		r.End,
		r.Samples,
		r.Exclusive,
		func(start, end vector3.Float64, samples int, exclusive bool) []trs.TRS {
			line := Line{
				Start:     start,
				End:       end,
				Samples:   max(samples, 0),
				Exclusive: exclusive,
			}
			return line.TRS()
		},
	)
}
