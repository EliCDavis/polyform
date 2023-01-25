package geometry

import (
	"errors"
	"math"

	"github.com/EliCDavis/vector/vector2"
)

// Line2D represents a line segment
type Line2D struct {
	p1 vector2.Float64
	p2 vector2.Float64
}

// ErrNoIntersection is thrown when Intersection() contains no intersection
var ErrNoIntersection = errors.New("no intersection")

// NewLine2D create a new line
func NewLine2D(p1, p2 vector2.Float64) Line2D {
	return Line2D{p1, p2}
}

// GetStartPoint returns the starting point of the line segment
func (l Line2D) GetStartPoint() vector2.Float64 {
	return l.p1
}

// GetEndPoint returns the end point of the line segment
func (l Line2D) GetEndPoint() vector2.Float64 {
	return l.p2
}

// Dir is end point - starting point
func (l Line2D) Dir() vector2.Float64 {
	return l.p2.Sub(l.p1)
}

// ScaleOutwards multiplies the current length of the line by extending it out
// further in the two different directions it's heading
func (l Line2D) ScaleOutwards(amount float64) Line2D {
	dirAndMag := l.p2.Sub(l.p1).DivByConstant(2.0)
	center := dirAndMag.Add(l.p1)
	return NewLine2D(
		center.Add(dirAndMag.Scale(amount)),
		center.Add(dirAndMag.Scale(-amount)),
	)
}

func (l Line2D) ClosestPointOnLine(p vector2.Float64) vector2.Float64 {
	l2 := math.Pow(l.p1.Distance(l.p2), 2)
	if l2 == 0.0 {
		return l.p1
	}

	// Consider the line extending the segment, parameterized as v + t (w - v).
	// We find projection of point p onto the line.
	// It falls where t = [(p-v) . (w-v)] / |w-v|^2
	// We clamp t from [0,1] to handle points outside the segment vw.
	t := math.Max(0, math.Min(1, p.Sub(l.p1).Dot(l.p2.Sub(l.p1))/l2))
	projection := l.p1.Add(l.p2.Sub(l.p1).Scale(t)) // Projection falls on the segment
	return projection
}

// Intersection finds where two lines intersect
// https://stackoverflow.com/questions/563198/how-do-you-detect-where-two-line-segments-intersect
func (l Line2D) Intersection(other Line2D) (vector2.Float64, error) {
	s1_x := l.p2.X() - l.p1.X()
	s1_y := l.p2.Y() - l.p1.Y()

	s2_x := other.p2.X() - other.p1.X()
	s2_y := other.p2.Y() - other.p1.Y()

	div := (-s2_x*s1_y + s1_x*s2_y)
	s := (-s1_y*(l.p1.X()-other.p1.X()) + s1_x*(l.p1.Y()-other.p1.Y())) / div
	t := (s2_x*(l.p1.Y()-other.p1.Y()) - s2_y*(l.p1.X()-other.p1.X())) / div

	if s >= 0 && s <= 1 && t >= 0 && t <= 1 {
		return vector2.New(l.p1.X()+(t*s1_x), l.p1.Y()+(t*s1_y)), nil
	}

	return vector2.Float64{}, ErrNoIntersection
}

// Intersects determines whether two lines intersect eachother
func (l Line2D) Intersects(other Line2D) bool {
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
