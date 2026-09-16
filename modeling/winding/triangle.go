package winding

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
)

type TriangleSampler struct {
	tris [][3]vector3.Float64
}

func (s TriangleSampler) Number(p vector3.Float64) float64 {
	total := 0.
	for _, tri := range s.tris {
		total += geometry.SolidAngle(
			tri[0].Sub(p),
			tri[1].Sub(p),
			tri[2].Sub(p),
		)
	}
	return total / (4 * math.Pi)
}

// 0.5 because a point exactly on the surface sees half a wrap
func (s TriangleSampler) Inside(p vector3.Float64) bool {
	return math.Abs(s.Number(p)) > 0.5
}
