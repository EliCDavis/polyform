package geometry

import (
	"math"

	"github.com/EliCDavis/vector/vector2"
)

// For collinear p, q, r: whether q lies on the segment pr.
func onSegment(p, q, r vector2.Float64) bool {
	return q.X() <= math.Max(p.X(), r.X()) && q.X() >= math.Min(p.X(), r.X()) && q.Y() <= math.Max(p.Y(), r.Y()) && q.Y() >= math.Min(p.Y(), r.Y())
}
