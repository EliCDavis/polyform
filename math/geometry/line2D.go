package geometry

import (
	"errors"
	"math"

	"github.com/EliCDavis/polyform/math/predicate"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
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

func (l Line2D) Length() float64 {
	return l.p2.Sub(l.p1).Length()
}

func (l Line2D) AtTime(time float64) vector2.Float64 {
	return l.p2.Sub(l.p1).Scale(time).Add(l.p1)
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

// The plane layouts match extrude's X, Y and Z outputs, so promoted lines
// land on an area extruded from the same points.
func onPlaneX(p vector2.Float64, offset float64) vector3.Float64 {
	return vector3.New(offset, p.X(), p.Y())
}

func onPlaneY(p vector2.Float64, offset float64) vector3.Float64 {
	return vector3.New(p.X(), offset, p.Y())
}

func onPlaneZ(p vector2.Float64, offset float64) vector3.Float64 {
	return vector3.New(p.X(), p.Y(), offset)
}

func (l Line2D) promote(offset float64, place func(vector2.Float64, float64) vector3.Float64) Line3D {
	return NewLine3D(place(l.p1, offset), place(l.p2, offset))
}

// PlaneX lays the line on the YZ plane at x = offset: its X along Y, its Y along Z.
func (l Line2D) PlaneX(offset float64) Line3D { return l.promote(offset, onPlaneX) }

// PlaneY lays the line on the XZ plane at y = offset: its X along X, its Y along Z.
func (l Line2D) PlaneY(offset float64) Line3D { return l.promote(offset, onPlaneY) }

// PlaneZ lays the line on the XY plane at z = offset.
func (l Line2D) PlaneZ(offset float64) Line3D { return l.promote(offset, onPlaneZ) }

func LineStripsFromPoints2D(points []vector2.Float64, closed bool) []Line2D {
	if len(points) < 2 {
		return nil
	}

	lines := make([]Line2D, 0, len(points))
	for i := 0; i < len(points)-1; i++ {
		lines = append(lines, NewLine2D(points[i], points[i+1]))
	}
	if closed && len(points) > 2 {
		lines = append(lines, NewLine2D(points[len(points)-1], points[0]))
	}

	return lines
}

func LinesFromPoints2D(points []vector2.Float64) []Line2D {
	if len(points) < 2 {
		return nil
	}

	lines := make([]Line2D, 0, len(points)/2)
	for i := 0; i < len(points)-1; i += 2 {
		lines = append(lines, NewLine2D(points[i], points[i+1]))
	}

	return lines
}

// ============================================================================

type LinesFromPoints2DNode struct {
	Points nodes.LiftedPort[[]vector2.Float64]
	Closed nodes.LiftedPort[bool] `description:"Line Strips also joins the last point back to the first."`
}

func (n LinesFromPoints2DNode) Description() string {
	return "Builds 2D line segments connecting a sequence of points."
}

func (n LinesFromPoints2DNode) LineStrips(out *nodes.Lifted[[]Line2D]) {
	nodes.Zip2(out, n.Points, n.Closed, LineStripsFromPoints2D)
}

func (n LinesFromPoints2DNode) Lines(out *nodes.Lifted[[]Line2D]) {
	nodes.Zip1(out, n.Points, LinesFromPoints2D)
}

// ============================================================================

type Lines2DTo3DNode struct {
	Lines  nodes.LiftedPort[[]Line2D] `description:"The 2D lines to place in 3D."`
	Offset nodes.LiftedPort[float64]  `description:"How far along the output's axis the plane sits. Defaults to 0."`
}

func (n Lines2DTo3DNode) Description() string {
	return "Places 2D lines on an axis-aligned plane in 3D, laid out the same way as an extruded area."
}

func (n Lines2DTo3DNode) Keywords() []string {
	return []string{"promote", "lift", "plane"}
}

func (n Lines2DTo3DNode) promote(out *nodes.Lifted[[]Line3D], place func(Line2D, float64) Line3D) {
	nodes.Zip2(out, n.Lines, n.Offset, func(lines []Line2D, offset float64) []Line3D {
		promoted := make([]Line3D, len(lines))
		for i, line := range lines {
			promoted[i] = place(line, offset)
		}
		return promoted
	})
}

func (n Lines2DTo3DNode) X(out *nodes.Lifted[[]Line3D]) {
	n.promote(out, Line2D.PlaneX)
}

func (Lines2DTo3DNode) XDescription() string {
	return "On the YZ plane at x = Offset: each line's X along Y, its Y along Z."
}

func (n Lines2DTo3DNode) Y(out *nodes.Lifted[[]Line3D]) {
	n.promote(out, Line2D.PlaneY)
}

func (Lines2DTo3DNode) YDescription() string {
	return "On the XZ plane at y = Offset: each line's X along X, its Y along Z."
}

func (n Lines2DTo3DNode) Z(out *nodes.Lifted[[]Line3D]) {
	n.promote(out, Line2D.PlaneZ)
}

func (Lines2DTo3DNode) ZDescription() string {
	return "On the XY plane at z = Offset."
}
