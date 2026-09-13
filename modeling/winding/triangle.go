package winding

import (
	"math"

	"github.com/EliCDavis/vector/vector3"
)

// Van Oosterom and Strackee, "The Solid Angle of a Plane Triangle", 1983.
// Signed by the triangle's winding.
func solidAngle(a, b, c vector3.Float64) float64 {
	lengthA, lengthB, lengthC := a.Length(), b.Length(), c.Length()

	return 2 * math.Atan2(
		a.Dot(b.Cross(c)),
		lengthA*lengthB*lengthC+
			a.Dot(b)*lengthC+
			a.Dot(c)*lengthB+
			b.Dot(c)*lengthA,
	)
}

type TriangleSampler struct {
	tris [][3]vector3.Float64
}

func (s TriangleSampler) Number(p vector3.Float64) float64 {
	total := 0.
	for _, tri := range s.tris {
		total += solidAngle(
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
