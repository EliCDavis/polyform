package primitives

import "github.com/EliCDavis/vector/vector2"

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
