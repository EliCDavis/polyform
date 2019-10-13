package mesh

import (
	"math"

	"github.com/EliCDavis/vector"
)

type Orientation int

const (
	Colinear Orientation = iota
	Clockwise
	Counterclockwise
)

// To find orientation of ordered triplet (p, q, r).
// The function returns following values
// 0 --> p, q and r are colinear
// 1 --> Clockwise
// 2 --> Counterclockwise
func calculateOrientation(p, q, r vector.Vector2) Orientation {
	// See https://www.geeksforgeeks.org/orientation-3-ordered-points/
	// for details of below formula.
	val := ((q.Y() - p.Y()) * (r.X() - q.X())) - ((q.X() - p.X()) * (r.Y() - q.Y()))

	if val == 0 {
		return Colinear
	}

	if val > 0 {
		return Clockwise
	}

	return Counterclockwise
}

// Given three colinear points p, q, r, the function checks if
// point q lies on line segment 'pr'
func onSegment(p, q, r vector.Vector2) bool {
	return q.X() <= math.Max(p.X(), r.X()) && q.X() >= math.Min(p.X(), r.X()) && q.Y() <= math.Max(p.Y(), r.Y()) && q.Y() >= math.Min(p.Y(), r.Y())
}
