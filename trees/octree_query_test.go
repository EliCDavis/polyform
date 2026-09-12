package trees

import (
	"math"
	"math/rand"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func probes(count int, spread float64) []vector3.Float64 {
	random := rand.New(rand.NewSource(7))
	out := make([]vector3.Float64, count)
	for i := range out {
		out[i] = vector3.New(random.NormFloat64(), random.NormFloat64(), random.NormFloat64()).Scale(spread)
	}
	return out
}

func TestOctreeQueriesMatchABruteForceScan(t *testing.T) {
	elements := sphereShellWithSpanners(3000, 1, 5)
	rays := testRays(50, 1)
	points := probes(50, 1)

	for _, tolerance := range []float64{noRetention, 0.2} {
		tree := NewOctree(elements, Tolerance(tolerance))

		for i, ray := range rays {
			expected := make([]int, 0)
			for j, e := range elements {
				if e.BoundingBox().IntersectsRayInRange(ray, 0, 10) {
					expected = append(expected, j)
				}
			}
			assert.ElementsMatch(t, expected, copyOf(tree.ElementsIntersectingRay(ray, 0, 10)), "ray %d", i)

			traversed := make([]int, 0)
			tree.TraverseIntersectingRay(ray, 0, 10, func(j int, min, max *float64) {
				traversed = append(traversed, j)
			})
			assert.ElementsMatch(t, expected, traversed, "traverse ray %d", i)
		}

		for i, p := range points {
			containing := make([]int, 0)
			within := make([]int, 0)
			closestDistance := math.Inf(1)
			for j, e := range elements {
				box := e.BoundingBox()
				if box.Contains(p) {
					containing = append(containing, j)
				}
				if box.ClosestPoint(p).Distance(p) <= 0.3 {
					within = append(within, j)
				}
				closestDistance = math.Min(closestDistance, e.ClosestPoint(p).Distance(p))
			}
			assert.ElementsMatch(t, containing, tree.ElementsContainingPoint(p), "contains %d", i)
			assert.ElementsMatch(t, within, tree.ElementsWithinRange(p, 0.3), "within %d", i)

			index, point := tree.ClosestPoint(p)
			require.NotEqual(t, -1, index)
			assert.InDelta(t, closestDistance, point.Distance(p), 1e-12, "closest %d", i)
			assert.InDelta(t, closestDistance, elements[index].ClosestPoint(p).Distance(p), 1e-12, "closest %d index", i)
		}
	}
}

func TestOctreeTraverseNarrowsTheRangeItHandsDown(t *testing.T) {
	elements := sphereShell(2000, 1)
	tree := NewOctree(elements)
	ray := geometry.NewRay(vector3.New(0., 3., 0.), vector3.New(0., -1., 0.))

	everything := 0
	tree.TraverseIntersectingRay(ray, 0, 10, func(i int, min, max *float64) {
		everything++
	})
	require.Greater(t, everything, 1, "sanity: the ray passes through both poles")

	visited := 0
	tree.TraverseIntersectingRay(ray, 0, 10, func(i int, min, max *float64) {
		visited++
		*max = math.Min(*max, 2.5)
	})

	assert.Less(t, visited, everything)
}
