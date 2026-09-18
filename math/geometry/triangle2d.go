package geometry

import (
	"github.com/EliCDavis/polyform/math/predicate"
	"github.com/EliCDavis/vector/vector2"
)

type Triangle2D [3]vector2.Float64

// Barycentric are the weights of the corners that add up to point. They
// carry outside the triangle, where one goes negative.
func (t Triangle2D) Barycentric(point vector2.Float64) [3]float64 {
	twiceArea := func(a, b, c vector2.Float64) float64 {
		return (b.X()-a.X())*(c.Y()-a.Y()) - (c.X()-a.X())*(b.Y()-a.Y())
	}
	whole := twiceArea(t[0], t[1], t[2])
	return [3]float64{
		twiceArea(point, t[1], t[2]) / whole,
		twiceArea(t[0], point, t[2]) / whole,
		twiceArea(t[0], t[1], point) / whole,
	}
}

// Contains allows point to sit up to tolerance outside any edge, whichever
// way the triangle winds.
func (t Triangle2D) Contains(point vector2.Float64, tolerance float64) bool {
	positive, negative := false, false
	for i := 0; i < 3; i++ {
		start, end := t[i], t[(i+1)%3]
		// Orient2D carries the edge's length, so the band is a distance.
		side := predicate.Orient2D(start, end, point)
		band := tolerance * start.Distance(end)
		if side > band {
			positive = true
		}
		if side < -band {
			negative = true
		}
	}
	return !(positive && negative)
}
