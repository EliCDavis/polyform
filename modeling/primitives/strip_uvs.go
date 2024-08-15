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

type StripUVsNode = nodes.StructNode[StripUVs, StripUVsNodeData]

type StripUVsNodeData struct {
	Width nodes.NodeOutput[float64]
	Start nodes.NodeOutput[vector2.Float64]
	End   nodes.NodeOutput[vector2.Float64]
}

func (sund StripUVsNodeData) Process() (StripUVs, error) {
	return StripUVs{
		Start: nodes.TryGetOutputValue(sund.Start, vector2.New(0, 0.5)),
		End:   nodes.TryGetOutputValue(sund.End, vector2.New(1, 0.5)),
		Width: nodes.TryGetOutputValue(sund.Width, 1.),
	}, nil
}
