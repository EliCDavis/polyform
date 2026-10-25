package geometry

import (
	"cmp"
	"slices"

	"github.com/EliCDavis/polyform/math/predicate"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

// ConvexHull returns the hull counter-clockwise, leaving out points that sit
// on a hull edge. Fewer than three distinct points come back sorted by X then Y.
func ConvexHull(points []vector2.Float64) []vector2.Float64 {
	sorted := slices.Clone(points)
	slices.SortFunc(sorted, func(a, b vector2.Float64) int {
		if c := cmp.Compare(a.X(), b.X()); c != 0 {
			return c
		}
		return cmp.Compare(a.Y(), b.Y())
	})
	sorted = slices.Compact(sorted)
	if len(sorted) < 3 {
		return sorted
	}

	hull := make([]vector2.Float64, 0, 2*len(sorted))
	for _, p := range sorted {
		for len(hull) >= 2 && predicate.Orient2D(hull[len(hull)-2], hull[len(hull)-1], p) <= 0 {
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, p)
	}

	lower := len(hull) + 1
	for i := len(sorted) - 2; i >= 0; i-- {
		p := sorted[i]
		for len(hull) >= lower && predicate.Orient2D(hull[len(hull)-2], hull[len(hull)-1], p) <= 0 {
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, p)
	}
	return hull[:len(hull)-1]
}

type ConvexHullNode struct {
	Points nodes.LiftedPort[[]vector2.Float64] `description:"Points to wrap."`
}

func (n ConvexHullNode) Description() string {
	return "Smallest convex outline containing every point, wound counter-clockwise."
}

func (n ConvexHullNode) Out(out *nodes.Lifted[[]vector2.Float64]) {
	nodes.Zip1(out, n.Points, ConvexHull)
}
