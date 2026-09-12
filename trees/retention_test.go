package trees

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func sphereShell(count int, radius float64) []Element {
	elements := make([]Element, count)
	step := math.Pi * (3 - math.Sqrt(5))
	extent := 3.5 * radius / math.Sqrt(float64(count))
	size := vector3.New(extent, extent, extent)

	for i := range elements {
		y := 1 - 2*float64(i)/float64(count-1)
		ring := math.Sqrt(math.Max(0, 1-y*y))
		angle := step * float64(i)
		center := vector3.New(math.Cos(angle)*ring, y, math.Sin(angle)*ring).Scale(radius)
		elements[i] = BoundingBoxElement(geometry.NewAABB(center, size))
	}
	return elements
}

func sphereShellWithSpanners(count int, radius float64, spanners int) []Element {
	elements := sphereShell(count-spanners, radius)
	for i := 0; i < spanners; i++ {
		size := vector3.New(radius*2, radius/50, radius*2)
		center := vector3.New(0, radius*(float64(i)/float64(spanners)-0.5), 0)
		elements = append(elements, BoundingBoxElement(geometry.NewAABB(center, size)))
	}
	return elements
}

func testRays(count int, radius float64) []geometry.Ray {
	random := rand.New(rand.NewSource(42))
	rays := make([]geometry.Ray, count)
	for i := range rays {
		origin := vector3.New(random.NormFloat64(), random.NormFloat64(), random.NormFloat64()).
			Normalized().
			Scale(radius * 3)
		aim := vector3.New(random.NormFloat64(), random.NormFloat64(), random.NormFloat64()).
			Scale(radius * 0.5)
		rays[i] = geometry.NewRay(origin, aim.Sub(origin).Normalized())
	}
	return rays
}

func copyOf(indices []int) []int {
	out := make([]int, len(indices))
	copy(out, indices)
	return out
}

func TestRetentionAnswersTheSameQueries(t *testing.T) {
	elements := sphereShellWithSpanners(4000, 1, 6)
	rays := testRays(64, 1)

	baseline := NewOctree(elements)
	baselineHits := make([][]int, len(rays))
	for i, ray := range rays {
		baselineHits[i] = copyOf(baseline.ElementsIntersectingRay(ray, 0, 10))
	}

	probe := vector3.New(0.3, 0.4, 0.5)
	baselineNear := baseline.ElementsWithinRange(probe, 0.2)
	_, baselinePoint := baseline.ClosestPoint(probe)

	for _, tolerance := range []float64{0, 0.1, 0.25, 0.5, 0.75, 0.9} {
		t.Run(fmt.Sprintf("tolerance_%v", tolerance), func(t *testing.T) {
			tree := NewOctree(elements, Tolerance(tolerance))

			for i, ray := range rays {
				assert.ElementsMatch(t, baselineHits[i],
					copyOf(tree.ElementsIntersectingRay(ray, 0, 10)), "ray %d", i)
			}

			assert.ElementsMatch(t, baselineNear, tree.ElementsWithinRange(probe, 0.2))

			_, point := tree.ClosestPoint(probe)
			assert.InDelta(t, baselinePoint.Distance(probe), point.Distance(probe), 1e-12)
		})
	}
}

func TestRetentionHoldsEveryElementExactlyOnce(t *testing.T) {
	elements := sphereShellWithSpanners(2000, 1, 4)

	for _, tolerance := range []float64{noRetention, 0, 0.25, 0.5, 0.9, 1} {
		t.Run(fmt.Sprintf("tolerance_%v", tolerance), func(t *testing.T) {
			seen := make(map[int]int)
			var walk func(*OctTree)
			walk = func(node *OctTree) {
				if node == nil {
					return
				}
				for _, element := range node.elements {
					seen[element.originalIndex]++
				}
				for _, child := range node.children {
					walk(child)
				}
			}
			walk(NewOctree(elements, Tolerance(tolerance)))

			assert.Len(t, seen, len(elements))
			for index, count := range seen {
				assert.Equalf(t, 1, count, "element %d", index)
			}
		})
	}
}

func TestRetentionOfOneKeepsEverythingAtTheRoot(t *testing.T) {
	elements := sphereShell(500, 1)
	tree := NewOctree(elements, Tolerance(1))

	assert.Empty(t, tree.children)
	assert.Len(t, tree.elements, len(elements))
}

func TestRetentionKeepsOnlyTheSpanners(t *testing.T) {
	const spanners = 6
	tree := NewOctree(sphereShellWithSpanners(20000, 1, spanners), Tolerance(0.2))

	retained := 0
	var walk func(*OctTree)
	walk = func(node *OctTree) {
		if len(node.children) > 0 {
			retained += len(node.elements)
		}
		for _, child := range node.children {
			walk(child)
		}
	}
	walk(tree)

	assert.Equal(t, spanners, retained)
}
