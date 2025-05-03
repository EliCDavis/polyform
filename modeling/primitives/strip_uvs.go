package primitives

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

type StripUVs struct {
	Start vector2.Float64
	End   vector2.Float64
	Width float64
}

func (suv StripUVs) At(time float64) vector2.Float64 {
	return suv.End.
		Sub(suv.Start).
		Scale(time).
		Add(suv.Start)
}

// This is slow, shouldn't really use it if you're gonna be calling it a bunch
func (suv StripUVs) AtXY(x, y float64) vector2.Float64 {
	return suv.StartLeft().
		Add(suv.Dir().Scale(y)).
		Add(suv.LeftToRight().Scale(x))
}

func (suv StripUVs) Dir() vector2.Float64 {
	return suv.End.Sub(suv.Start)
}

func (suv StripUVs) perpendicular() vector2.Float64 {
	return suv.Dir().Perpendicular().Normalized().Scale(suv.Width / 2)
}

func (suv StripUVs) StartLeft() vector2.Float64 {
	return suv.Start.Sub(suv.perpendicular())
}

func (suv StripUVs) StartRight() vector2.Float64 {
	return suv.Start.Add(suv.perpendicular())
}

func (suv StripUVs) EndLeft() vector2.Float64 {
	return suv.End.Sub(suv.perpendicular())
}

func (suv StripUVs) EndRight() vector2.Float64 {
	return suv.End.Add(suv.perpendicular())
}

func (suv StripUVs) LeftToRight() vector2.Float64 {
	return suv.StartRight().Sub(suv.StartLeft())
}

type StripUVsNode = nodes.Struct[StripUVsNodeData]

type StripUVsNodeData struct {
	Width nodes.Output[float64]
	Start nodes.Output[vector2.Float64]
	End   nodes.Output[vector2.Float64]
}

func (sund StripUVsNodeData) Out() nodes.StructOutput[StripUVs] {
	return nodes.NewStructOutput(StripUVs{
		Start: nodes.TryGetOutputValue(sund.Start, vector2.New(0, 0.5)),
		End:   nodes.TryGetOutputValue(sund.End, vector2.New(1, 0.5)),
		Width: nodes.TryGetOutputValue(sund.Width, 1.),
	})
}
