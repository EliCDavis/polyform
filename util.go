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

// Intersect determines whether two lines intersect eachother
func doIntersect(l, other Line) bool {
	// Find the four orientations needed for general and
	// special cases
	o1 := calculateOrientation(l.p1, l.p2, other.p1)
	o2 := calculateOrientation(l.p1, l.p2, other.p2)
	o3 := calculateOrientation(other.p1, other.p2, l.p1)
	o4 := calculateOrientation(other.p1, other.p2, l.p2)

	// General case
	if o1 != o2 && o3 != o4 {
		return true
	}

	// Special Cases
	// l.p1, l.p2 and other.p1 are colinear and other.p1 lies on segment l.p1l.p2
	if o1 == Colinear && onSegment(l.p1, other.p1, l.p2) {
		return true
	}

	// l.p1, l.p2 and other.p2 are colinear and other.p2 lies on segment l.p1l.p2
	if o2 == Colinear && onSegment(l.p1, other.p2, l.p2) {
		return true
	}

	// p2, other.p2 and l.p1 are colinear and l.p1 lies on segment p2other.p2
	if o3 == Colinear && onSegment(other.p1, l.p1, other.p2) {
		return true
	}

	// p2, other.p2 and l.p2 are colinear and l.p2 lies on segment p2other.p2
	if o4 == 0 && onSegment(other.p1, l.p2, other.p2) {
		return true
	}

	return false // Doesn't fall in any of the above cases
}

// Given three colinear points p, q, r, the function checks if
// point q lies on line segment 'pr'
func onSegment(p, q, r vector.Vector2) bool {
	return q.X() <= math.Max(p.X(), r.X()) && q.X() >= math.Min(p.X(), r.X()) && q.Y() <= math.Max(p.Y(), r.Y()) && q.Y() >= math.Min(p.Y(), r.Y())
}
