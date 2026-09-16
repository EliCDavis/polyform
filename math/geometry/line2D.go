package geometry

import (
	"errors"
	"math"

	"github.com/EliCDavis/polyform/math/predicate"
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

func (l Line2D) ExtendEnd(amount float64) Line2D {
	dirAndMag := l.p2.Sub(l.p1).Normalized().Scale(amount)
	return NewLine2D(l.p1, l.p2.Add(dirAndMag))
}

func (l Line2D) ExtendStart(amount float64) Line2D {
	dirAndMag := l.p1.Sub(l.p2).Normalized().Scale(amount)
	return NewLine2D(l.p1.Add(dirAndMag), l.p2)
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

// Intersection is where the segments meet. Touching, at an endpoint or along
// the other's interior, counts. Which side each end falls on is decided
// exactly; only the point itself is rounded.
func (l Line2D) Intersection(other Line2D) (vector2.Float64, error) {
	side1 := predicate.Orient2D(other.p1, other.p2, l.p1)
	side2 := predicate.Orient2D(other.p1, other.p2, l.p2)
	side3 := predicate.Orient2D(l.p1, l.p2, other.p1)
	side4 := predicate.Orient2D(l.p1, l.p2, other.p2)

	switch {
	case side1 == 0 && onSegment(other.p1, l.p1, other.p2):
		return l.p1, nil
	case side2 == 0 && onSegment(other.p1, l.p2, other.p2):
		return l.p2, nil
	case side3 == 0 && onSegment(l.p1, other.p1, l.p2):
		return other.p1, nil
	case side4 == 0 && onSegment(l.p1, other.p2, l.p2):
		return other.p2, nil
	case (side1 > 0) == (side2 > 0) || (side3 > 0) == (side4 > 0):
		return vector2.Float64{}, ErrNoIntersection
	}

	return l.p1.Add(l.p2.Sub(l.p1).Scale(side1 / (side1 - side2))), nil
}

func (l Line2D) Intersects(other Line2D) bool {
	_, err := l.Intersection(other)
	return err == nil
}
