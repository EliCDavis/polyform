package trees

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
)

type axis int

const (
	xAxis axis = iota
	yAxis
	zAxis
	none
)

func nextkdAxis(current axis) axis {
	switch current {
	case xAxis:
		return yAxis
	case yAxis:
		return zAxis
	case zAxis:
		return xAxis
	}
	panic("unimplemented axis")
}

type KDTree struct {
	left       *KDTree
	right      *KDTree
	axis       axis
	splitValue float64
	bounds     geometry.AABB
	elements   []elementReference
}

// func (kdt KDTree) ElementsContainingPoint(v vector3.Float64) []int {
// }

// func (kdt KDTree) ClosestPoint(v vector3.Float64) (int, vector3.Float64) {
// }

// func (kdt KDTree) ElementsIntersectingRay(ray geometry.Ray, min, max float64) []int {
// 	if !kdt.bounds.IntersectsRayInRange(ray, min, max) {
// 		return nil
// 	}
// }

func (kdt KDTree) BoundingBox() geometry.AABB {
	return kdt.bounds
}

func newKDTreeWithDepth(elements []elementReference, maxDepth int, axis axis) *KDTree {
	if len(elements) == 0 {
		return nil
	}

	if len(elements) == 1 {
		return &KDTree{
			bounds:   elements[0].primitive.BoundingBox(),
			axis:     none,
			elements: elements,
		}
	}

	bounds := elements[0].primitive.BoundingBox()

	min, max := 0., 0.
	for _, item := range elements {
		box := item.primitive.BoundingBox()
		bounds.EncapsulateBounds(box)
		switch axis {
		case xAxis:
			min = math.Min(min, box.Min().X())
			max = math.Max(max, box.Max().X())
		case yAxis:
			min = math.Min(min, box.Min().Y())
			max = math.Max(max, box.Max().Y())
		case zAxis:
			min = math.Min(min, box.Min().Z())
			max = math.Max(max, box.Max().Z())
		}
	}

	if maxDepth == 0 {
		return &KDTree{
			bounds:   bounds,
			elements: elements,
			left:     nil,
			right:    nil,
			axis:     none,
		}
	}

	split := min + ((max - min) / 2.)
	left := make([]elementReference, 0)
	right := make([]elementReference, 0)

	for _, item := range elements {
		center := item.primitive.BoundingBox().Center()
		leftSide := true

		switch axis {
		case xAxis:
			leftSide = center.X() < split
		case yAxis:
			leftSide = center.Y() < split
		case zAxis:
			leftSide = center.Z() < split
		}

		if leftSide {
			left = append(left, item)
		} else {
			right = append(right, item)
		}
	}

	return &KDTree{
		left:       newKDTreeWithDepth(left, maxDepth-1, nextkdAxis(axis)),
		right:      newKDTreeWithDepth(right, maxDepth-1, nextkdAxis(axis)),
		axis:       axis,
		splitValue: split,
		bounds:     bounds,
		elements:   nil,
	}
}

func NewKDTreeWithDepth(elements []Element, maxDepth int) *KDTree {
	primitives := make([]elementReference, len(elements))
	for i, ele := range elements {
		primitives[i] = elementReference{
			primitive:     ele,
			originalIndex: i,
		}
	}
	return newKDTreeWithDepth(primitives, maxDepth, xAxis)
}
