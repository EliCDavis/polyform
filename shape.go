package mesh

import (
	"errors"
	"math"
	"math/rand"

	"github.com/EliCDavis/vector"
)

// Shape is a flat (2D) arrangement of points.
type Shape struct {
	points []vector.Vector2
	center vector.Vector2
	origin vector.Vector2
}

// NewShape creates a new shape with a center computed by averaging points positions
func NewShape(bounds []vector.Vector2) (Shape, error) {

	if bounds == nil {
		return Shape{}, errors.New("Can not create a shape without any bounds")
	}

	if len(bounds) < 3 {
		return Shape{}, errors.New("Can not create a shape with less than 3 points")
	}

	center := vector.NewVector2(0, 0)
	for _, point := range bounds {
		center = center.Add(point)
	}

	return Shape{bounds, center.MultByConstant(1.0 / float64(len(bounds))), vector.NewVector2(0, 0)}, nil
}

// NewShapeWithCustomCenter creates a shape with a center you set
func NewShapeWithCustomCenter(bounds []vector.Vector2, center vector.Vector2) Shape {
	return Shape{bounds, center, vector.NewVector2(0, 0)}
}

func (s Shape) GetPoints() []vector.Vector2 {
	return s.points
}

func (s Shape) GetCenter() vector.Vector2 {
	return s.center
}

func (s Shape) GetOrigin() vector.Vector2 {
	return s.origin
}

func (s Shape) GetBounds() (vector.Vector2, vector.Vector2) {

	bottomLeftX := 10000000.
	bottomLeftY := 10000000.

	topRightX := -10000000.
	topRightY := -10000000.

	for _, p := range s.points {
		if p.X() < bottomLeftX {
			bottomLeftX = p.X()
		}
		if p.Y() < bottomLeftY {
			bottomLeftY = p.Y()
		}
		if p.X() > topRightX {
			topRightX = p.X()
		}
		if p.Y() > topRightY {
			topRightY = p.Y()
		}
	}

	return vector.NewVector2(bottomLeftX, bottomLeftY), vector.NewVector2(topRightX, topRightY)
}

func (s Shape) GetBoundingBoxDimensions() (width, height float64) {
	bottomLeft, topRight := s.GetBounds()
	return topRight.X() - bottomLeft.X(), topRight.Y() - bottomLeft.Y()
}

// RandomPointInShape returns a random point inside of the shape
func (s Shape) RandomPointInShape() vector.Vector2 {
	bottomLeftBounds, topRightBounds := s.GetBounds()
	for {
		point := vector.NewVector2(
			bottomLeftBounds.X()+(rand.Float64()*(topRightBounds.X()-bottomLeftBounds.X())),
			bottomLeftBounds.Y()+(rand.Float64()*(topRightBounds.Y()-bottomLeftBounds.Y())),
		)
		if s.IsInside(point) {
			return point
		}
	}
}

// Split figures out which points land on which side of the vertical line and
// builds new shapes from that
func (s Shape) Split(vericalLine float64) ([]Shape, []Shape) {
	return s.shapesOnSide(vericalLine, -1), s.shapesOnSide(vericalLine, 1)
}

func (s Shape) startinPointForSideShape(vericalLine float64, side int) (bool, int) {
	startingPointIndex := 0
	lowestPointHeight := 1000000.0
	lastSide := 0
	crossed := false
	if s.points[len(s.points)-1].X() < vericalLine {
		lastSide = -1
	} else {
		lastSide = 1
	}

	for i := 0; i < len(s.points); i++ {
		n := i * -1 * side
		if n < 0 {
			n += len(s.points)
		}
		if lastSide == side*-1 && s.points[n].Y() < lowestPointHeight {
			if (side == -1 && s.points[n].X() < vericalLine) || (side == 1 && s.points[n].X() > vericalLine) {
				lowestPointHeight = s.points[n].Y()
				startingPointIndex = n
			}
		}

		newSide := 0
		if s.points[n].X() <= vericalLine {
			newSide = -1
		} else {
			newSide = 1
		}
		if lastSide != newSide {
			crossed = true
		}

		lastSide = newSide
	}

	if crossed == false {
		if side == lastSide {
			return true, -1
		}
		return false, -1
	}
	return false, startingPointIndex
}

func (s Shape) shapesOnSide(vericalLineX float64, side int) []Shape {

	onOurSide, startingPointIndex := s.startinPointForSideShape(vericalLineX, side)

	if startingPointIndex == -1 {
		if onOurSide {
			newShape, _ := NewShape(s.points)
			return []Shape{newShape}
		}
		return []Shape{}
	}

	type region struct {
		highestPoint float64
		lowestPoint  float64
		points       []vector.Vector2
		started      bool
	}

	pointBefore := startingPointIndex + side
	if pointBefore >= len(s.points) {
		pointBefore -= len(s.points)
	} else if pointBefore < 0 {
		pointBefore += len(s.points)
	}

	verticalLine := NewLine2D(vector.NewVector2(vericalLineX, -1000000), vector.NewVector2(vericalLineX, 1000000))
	curLine := NewLine2D(s.points[startingPointIndex], s.points[pointBefore])
	intersection, err := verticalLine.Intersection(curLine)
	if err == ErrNoIntersection {
		panic("Intersection is nil!")
	}

	regions := []region{region{-100000, 100000, make([]vector.Vector2, 1), false}}
	regions[0].points[0] = intersection
	regions[0].lowestPoint = intersection.Y()

	currentRegion := 0

	// -1 for left, +1 for right, 0 for unset
	lastPointsSide := side

	for i := 0; i < len(s.points); i++ {
		n := (i * -1 * side) + startingPointIndex
		if n >= len(s.points) {
			n -= len(s.points)
		} else if n < 0 {
			n += len(s.points)
		}
		var currentSide int

		if s.points[n].X() <= vericalLineX {
			currentSide = -1
		} else {
			currentSide = 1
		}

		// Change the region we're working with.
		if currentSide != lastPointsSide {

			pointBefore = n + side
			if pointBefore >= len(s.points) {
				pointBefore -= len(s.points)
			} else if pointBefore < 0 {
				pointBefore += len(s.points)
			}

			intersection, err := NewLine2D(s.points[n], s.points[pointBefore]).Intersection(verticalLine)
			if err == ErrNoIntersection {
				panic("Intersection is nil!")
			}

			if currentRegion != -1 {
				if regions[currentRegion].started == false {
					regions[currentRegion].highestPoint = intersection.Y()
				}
				regions[currentRegion].started = true
			}

			if currentSide == side {
				foundRegion := false

				// Find region we're in.
				for regionIndex := range regions {

					if regions[regionIndex].lowestPoint <= s.points[n].Y() &&
						regions[regionIndex].highestPoint >= s.points[n].Y() {
						currentRegion = regionIndex
						foundRegion = true
						break
					}
				}

				// If can't find one, create one.
				if foundRegion == false {
					regions = append(regions, region{-100000, 100000, make([]vector.Vector2, 0), false})
					currentRegion = len(regions) - 1
					regions[currentRegion].lowestPoint = intersection.Y()
				}

				regions[currentRegion].points = append(regions[currentRegion].points, intersection)

			} else {
				regions[currentRegion].points = append(regions[currentRegion].points, intersection)
				currentRegion = -1
			}

		}

		if currentRegion != -1 {
			if regions[currentRegion].started == false {
				regions[currentRegion].highestPoint = s.points[n].Y()
			}
			regions[currentRegion].points = append(regions[currentRegion].points, s.points[n])
		}

		lastPointsSide = currentSide
	}

	resultingShapes := make([]Shape, 0)

	for r := range regions {
		newShape, err := NewShape(regions[r].points)
		if err == nil {
			resultingShapes = append(resultingShapes, newShape)
		}
	}

	return resultingShapes
}

// IsInside returns true if the point p lies inside the polygon[] with n vertices
func (s Shape) IsInside(p vector.Vector2) bool {

	// There must be at least 3 vertices in polygon[]
	if len(s.points) < 3 {
		return false
	}

	// Create a point for line segment from p to infinite
	extreme := vector.NewVector2(100000, p.Y())

	// Count intersections of the above line with sides of polygon
	count := 0
	i := 0
	for {
		next := (i + 1) % len(s.points)

		// Check if the line segment from 'p' to 'extreme' intersects
		// with the line segment from 'polygon[i]' to 'polygon[next]'
		if NewLine2D(s.points[i], s.points[next]).Intersects(NewLine2D(p, extreme)) {
			// If the point 'p' is colinear with line segment 'i-next',
			// then check if it lies on segment. If it lies, return true,
			// otherwise false
			if calculateOrientation(s.points[i], p, s.points[next]) == Colinear {
				return onSegment(s.points[i], p, s.points[next])
			}

			count++
		}
		i = next
		if i == 0 {
			break
		}
	}

	// log.Print(count)
	// Return true if count is odd, false otherwise
	return count%2 == 1
}

// Translate Moves all points over by the specified amount
func (s Shape) Translate(amount vector.Vector2) Shape {
	newShapePonts := make([]vector.Vector2, len(s.points))
	for i, point := range s.points {
		newShapePonts[i] = point.Add(amount)
	}
	newshape, _ := NewShape(newShapePonts)
	return newshape
}

// Rotate will rotate all points in the shape around the pivot by the passed in amount
func (s Shape) Rotate(amount float64, pivot vector.Vector2) Shape {
	newPoints := make([]vector.Vector2, s.Len())

	for p, point := range s.points {

		// https://play.golang.org/p/qWUotd3Lb56
		directionWithMag := point.Sub(pivot)

		magnitude := directionWithMag.Length()

		newRot := math.Atan2(directionWithMag.Y(), directionWithMag.X()) + amount

		newPoints[p] = vector.NewVector2(
			math.Cos(newRot)*magnitude,
			math.Sin(newRot)*magnitude,
		).Add(pivot)

	}

	newShape, _ := NewShape(newPoints)
	return newShape
}

// Scale shifts all points towards or away from the origin
func (s Shape) Scale(amount float64) Shape {
	newShapePonts := make([]vector.Vector2, len(s.points))

	for i, point := range s.points {
		newShapePonts[i] = s.origin.Add(point.Sub(s.origin).Normalized().MultByConstant(amount * s.origin.Distance(point)))
	}
	newshape, _ := NewShape(newShapePonts)
	return newshape
}

// Len returns the number of points in the polygon
func (s Shape) Len() int {
	return len(s.points)
}

// Swap switches two points indeces so the polygon is ordered a different way
func (s *Shape) Swap(i, j int) {
	s.points[i], s.points[j] = s.points[j], s.points[i]
}

// PointClosestToCenter returns the index of the closest point to the center of the polygon
func (s Shape) PointClosestToCenter() int {
	bestDistance := math.Inf(0)
	curPoint := -1

	for i, point := range s.points {
		dist := s.center.Distance(point)
		if dist < bestDistance {
			bestDistance = dist
			curPoint = i
		}
	}

	return curPoint
}

// Less determines which point is more orriented more clockwise from the center than the other
func (s Shape) Less(i, j int) bool {
	a := s.points[i]
	b := s.points[j]

	if a.X()-s.center.X() >= 0 && b.X()-s.center.X() < 0 {
		return true
	}

	if a.X()-s.center.X() < 0 && b.X()-s.center.X() >= 0 {
		return false
	}

	if a.X()-s.center.X() == 0 && b.X()-s.center.X() == 0 {
		if a.Y()-s.center.Y() >= 0 || b.Y()-s.center.Y() >= 0 {
			return a.Y() > b.Y()
		}
		return b.Y() > a.Y()
	}

	// compute the cross product of vectors (center -> a) x (center -> b)
	det := (a.X()-s.center.X())*(b.Y()-s.center.Y()) - (b.X()-s.center.X())*(a.Y()-s.center.Y())
	if det < 0 {
		return true
	}
	if det > 0 {
		return false
	}

	// points a and b are on the same line from the center
	// check which point is closer to the center
	d1 := (a.X()-s.center.X())*(a.X()-s.center.X()) + (a.Y()-s.center.Y())*(a.Y()-s.center.Y())
	d2 := (b.X()-s.center.X())*(b.X()-s.center.X()) + (b.Y()-s.center.Y())*(b.Y()-s.center.Y())
	return d1 > d2
}
