package primitives

import "github.com/EliCDavis/vector/vector2"

type CircleUVs struct {
	Center vector2.Float64
	Radius float64
}

type EuclideanUVSpace interface {
	AtXY(p vector2.Float64) vector2.Float64
	AtXYs(xys []vector2.Float64) []vector2.Float64
}

type PolarUVSpace interface {
	AtRt(radius, theta float64) vector2.Float64
	AtRts(rts []vector2.Float64) []vector2.Float64
}
