package geometry

import (
	"math"

	"github.com/EliCDavis/vector/vector3"
)

// SolidAngle of the triangle abc seen from the origin, signed by its winding.
// Van Oosterom and Strackee, "The Solid Angle of a Plane Triangle", 1983.
func SolidAngle(a, b, c vector3.Float64) float64 {
	lengthA, lengthB, lengthC := a.Length(), b.Length(), c.Length()

	return 2 * math.Atan2(
		a.Dot(b.Cross(c)),
		lengthA*lengthB*lengthC+
			a.Dot(b)*lengthC+
			a.Dot(c)*lengthB+
			b.Dot(c)*lengthA,
	)
}
