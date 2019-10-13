package mesh

import (
	"errors"

	"github.com/EliCDavis/vector"
)

// Line represents a line segment
type Line struct {
	p1 vector.Vector2
	p2 vector.Vector2
}

// ErrNoIntersection is thrown when Intersection() contains no intersection
var ErrNoIntersection = errors.New("No Intersection")

// NewLine create a new
func NewLine(p1, p2 vector.Vector2) Line {
	return Line{p1, p2}
}

// GetStartPoint returns the starting point of the line segment
func (l Line) GetStartPoint() vector.Vector2 {
	return l.p1
}

// GetEndPoint returns the end point of the line segment
func (l Line) GetEndPoint() vector.Vector2 {
	return l.p2
}

// Intersection finds where two lines intersect
// https://stackoverflow.com/questions/563198/how-do-you-detect-where-two-line-segments-intersect
func (l Line) Intersection(other Line) (vector.Vector2, error) {

	s1_x := l.p2.X() - l.p1.X()
	s1_y := l.p2.Y() - l.p1.Y()

	s2_x := other.p2.X() - other.p1.X()
	s2_y := other.p2.Y() - other.p1.Y()

	s := (-s1_y*(l.p1.X()-other.p1.X()) + s1_x*(l.p1.Y()-other.p1.Y())) / (-s2_x*s1_y + s1_x*s2_y)
	t := (s2_x*(l.p1.Y()-other.p1.Y()) - s2_y*(l.p1.X()-other.p1.X())) / (-s2_x*s1_y + s1_x*s2_y)

	if s >= 0 && s <= 1 && t >= 0 && t <= 1 {
		return vector.NewVector2(l.p1.X()+(t*s1_x), l.p1.Y()+(t*s1_y)), nil
	}

	return vector.Vector2{}, ErrNoIntersection
}

// DoIntersect determines whether two lines intersect eachother
func (l Line) Intersects(other Line) bool {
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
