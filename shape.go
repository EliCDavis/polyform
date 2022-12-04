package mesh

import (
	"math"
	"math/rand"

	"github.com/EliCDavis/vector"
)

// Shape is a flat (2D) arrangement of points.
type Shape []vector.Vector2

func (s Shape) GetBounds() (vector.Vector2, vector.Vector2) {
	bottomLeftX := math.Inf(1)
	bottomLeftY := math.Inf(1)

	topRightX := math.Inf(-1)
	topRightY := math.Inf(-1)

	for _, p := range s {
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
	if s[len(s)-1].X() < vericalLine {
		lastSide = -1
	} else {
		lastSide = 1
	}

	for i := 0; i < len(s); i++ {
		n := i * -1 * side
		if n < 0 {
			n += len(s)
		}
		if lastSide == side*-1 && s[n].Y() < lowestPointHeight {
			if (side == -1 && s[n].X() < vericalLine) || (side == 1 && s[n].X() > vericalLine) {
				lowestPointHeight = s[n].Y()
				startingPointIndex = n
			}
		}

		newSide := 0
		if s[n].X() <= vericalLine {
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
			return []Shape{s}
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
	if pointBefore >= len(s) {
		pointBefore -= len(s)
	} else if pointBefore < 0 {
		pointBefore += len(s)
	}

	verticalLine := NewLine2D(vector.NewVector2(vericalLineX, -1000000), vector.NewVector2(vericalLineX, 1000000))
	curLine := NewLine2D(s[startingPointIndex], s[pointBefore])
	intersection, err := verticalLine.Intersection(curLine)
	if err == ErrNoIntersection {
		panic("Intersection is nil!")
	}

	regions := []region{{-100000, 100000, make([]vector.Vector2, 1), false}}
	regions[0].points[0] = intersection
	regions[0].lowestPoint = intersection.Y()

	currentRegion := 0

	// -1 for left, +1 for right, 0 for unset
	lastPointsSide := side

	for i := 0; i < len(s); i++ {
		n := (i * -1 * side) + startingPointIndex
		if n >= len(s) {
			n -= len(s)
		} else if n < 0 {
			n += len(s)
		}
		var currentSide int

		if s[n].X() <= vericalLineX {
			currentSide = -1
		} else {
			currentSide = 1
		}

		// Change the region we're working with.
		if currentSide != lastPointsSide {

			pointBefore = n + side
			if pointBefore >= len(s) {
				pointBefore -= len(s)
			} else if pointBefore < 0 {
				pointBefore += len(s)
			}

			intersection, err := NewLine2D(s[n], s[pointBefore]).Intersection(verticalLine)
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
					if regions[regionIndex].lowestPoint <= s[n].Y() &&
						regions[regionIndex].highestPoint >= s[n].Y() {
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
				regions[currentRegion].highestPoint = s[n].Y()
			}
			regions[currentRegion].points = append(regions[currentRegion].points, s[n])
		}

		lastPointsSide = currentSide
	}

	resultingShapes := make([]Shape, 0)

	for r := range regions {
		if len(regions[r].points) > 2 {
			resultingShapes = append(resultingShapes, Shape(regions[r].points))
		}
	}

	return resultingShapes
}

// IsInside returns true if the point p lies inside the polygon[] with n vertices
func (s Shape) IsInside(p vector.Vector2) bool {
	// There must be at least 3 vertices in polygon[]
	if len(s) < 3 {
		return false
	}

	// Create a point for line segment from p to infinite
	extreme := vector.NewVector2(math.MaxFloat64, p.Y())

	// Count intersections of the above line with sides of polygon
	count := 0
	i := 0
	for {
		next := (i + 1) % len(s)

		// Check if the line segment from 'p' to 'extreme' intersects
		// with the line segment from 'polygon[i]' to 'polygon[next]'
		if NewLine2D(s[i], s[next]).Intersects(NewLine2D(p, extreme)) {
			// If the point 'p' is colinear with line segment 'i-next',
			// then check if it lies on segment. If it lies, return true,
			// otherwise false
			if calculateOrientation(s[i], p, s[next]) == Colinear {
				return onSegment(s[i], p, s[next])
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
	newShapePonts := make([]vector.Vector2, len(s))
	for i, point := range s {
		newShapePonts[i] = point.Add(amount)
	}
	return Shape(newShapePonts)
}

// Rotate will rotate all points in the shape around the pivot by the passed in amount
func (s Shape) Rotate(amount float64, pivot vector.Vector2) Shape {
	newPoints := make([]vector.Vector2, s.Len())

	for p, point := range s {

		// https://play.golang.org/p/qWUotd3Lb56
		directionWithMag := point.Sub(pivot)

		magnitude := directionWithMag.Length()

		newRot := math.Atan2(directionWithMag.Y(), directionWithMag.X()) + amount

		newPoints[p] = vector.NewVector2(
			math.Cos(newRot)*magnitude,
			math.Sin(newRot)*magnitude,
		).Add(pivot)

	}

	return Shape(newPoints)
}

// Scale shifts all points towards or away from the origin
func (s Shape) Scale(amount float64, origin vector.Vector2) Shape {
	newShapePonts := make([]vector.Vector2, len(s))

	for i, point := range s {
		newShapePonts[i] = origin.Add(point.Sub(origin).Normalized().MultByConstant(amount * origin.Distance(point)))
	}
	return newShapePonts
}

// Len returns the number of points in the polygon
func (s Shape) Len() int {
	return len(s)
}

// Swap switches two points indeces so the polygon is ordered a different way
func (s Shape) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less determines which point is more oriented more clockwise from the center than the other
func (s Shape) Less(i, j int) bool {
	center := vector.Vector2Zero()

	a := s[i]
	b := s[j]

	if a.X()-center.X() >= 0 && b.X()-center.X() < 0 {
		return true
	}

	if a.X()-center.X() < 0 && b.X()-center.X() >= 0 {
		return false
	}

	if a.X()-center.X() == 0 && b.X()-center.X() == 0 {
		if a.Y()-center.Y() >= 0 || b.Y()-center.Y() >= 0 {
			return a.Y() > b.Y()
		}
		return b.Y() > a.Y()
	}

	// compute the cross product of vectors (center -> a) x (center -> b)
	det := (a.X()-center.X())*(b.Y()-center.Y()) - (b.X()-center.X())*(a.Y()-center.Y())
	if det < 0 {
		return true
	}
	if det > 0 {
		return false
	}

	// points a and b are on the same line from the center
	// check which point is closer to the center
	d1 := (a.X()-center.X())*(a.X()-center.X()) + (a.Y()-center.Y())*(a.Y()-center.Y())
	d2 := (b.X()-center.X())*(b.X()-center.X()) + (b.Y()-center.Y())*(b.Y()-center.Y())
	return d1 > d2
}
