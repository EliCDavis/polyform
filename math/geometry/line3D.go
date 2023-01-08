package geometry

import (
	"math"

	"github.com/EliCDavis/vector"
)

// Line3D is a series of ordered points that make up a line segment
// through 3D space.
type Line3D struct {
	p1 vector.Vector3
	p2 vector.Vector3
}

// NewLine3D create a new line
func NewLine3D(p1, p2 vector.Vector3) Line3D {
	return Line3D{p1, p2}
}

// GetStartPoint returns the starting point of the line segment
func (l Line3D) GetStartPoint() vector.Vector3 {
	return l.p1
}

// GetEndPoint returns the end point of the line segment
func (l Line3D) GetEndPoint() vector.Vector3 {
	return l.p2
}

func (l Line3D) SetStartPoint(newStart vector.Vector3) Line3D {
	return NewLine3D(newStart, l.GetEndPoint())
}

func (l Line3D) SetEndPoint(newEnd vector.Vector3) Line3D {
	return NewLine3D(l.GetStartPoint(), newEnd)
}

func (l Line3D) Translate(amt vector.Vector3) Line3D {
	return Line3D{
		l.p1.Add(amt),
		l.p2.Add(amt),
	}
}

// YIntersection Uses the paremetric equations of the line
func (l Line3D) YIntersection(x float64, z float64) float64 {
	v := l.p2.Sub(l.p1)
	t := (x - l.p1.X()) / v.X()

	// This would mean that the X direction is 0 (x never changes) so we'll
	// have to figure out where we are on the line using the z axis.
	if math.IsNaN(t) {
		t = (z - l.p1.Z()) / v.Z()
	}

	// Well then uh... return y slope I guess? Maybe I should throw NaN?
	// Ima throw a NaN
	if math.IsNaN(t) {
		return math.NaN()
	}

	return l.p1.Y() + (v.Y() * t)
}

// ScaleOutwards assumes line segment 3d is only two points
// multiplies the current length of the line by extending it out
// further in the two different directions it's heading
func (l Line3D) ScaleOutwards(amount float64) Line3D {
	dirAndMag := l.p2.Sub(l.p1).DivByConstant(2.0)
	center := dirAndMag.Add(l.p1)
	return Line3D{
		center.Add(dirAndMag.MultByConstant(amount)),
		center.Add(dirAndMag.MultByConstant(-amount)),
	}
}

func (l Line3D) ClosestPointOnLine(p vector.Vector3) vector.Vector3 {
	l2 := math.Pow(l.p1.Distance(l.p2), 2)
	if l2 == 0.0 {
		return l.p1
	}

	// Consider the line extending the segment, parameterized as v + t (w - v).
	// We find projection of point p onto the line.
	// It falls where t = [(p-v) . (w-v)] / |w-v|^2
	// We clamp t from [0,1] to handle points outside the segment vw.
	t := math.Max(0, math.Min(1, p.Sub(l.p1).Dot(l.p2.Sub(l.p1))/l2))
	projection := l.p1.Add(l.p2.Sub(l.p1).MultByConstant(t)) // Projection falls on the segment
	return projection
}
