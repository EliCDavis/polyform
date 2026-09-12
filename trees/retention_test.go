package trees

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"time"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

// Stands in for a triangulated sphere: package trees cannot import modeling,
// and the tree only ever looks at bounds anyway.
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

// A shell with a few elements spanning the whole thing, which is the shape the
// old leftOver branch existed for.
func sphereShellWithSpanners(count int, radius float64, spanners int) []Element {
	elements := sphereShell(count-spanners, radius)
	for i := 0; i < spanners; i++ {
		thickness := radius / 50
		size := vector3.New(radius*2, thickness, radius*2)
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
	depth := OctreeDepthFromCount(len(elements))
	rays := testRays(64, 1)

	baseline := NewOctreeWithDepth(elements, depth)
	baselineHits := make([][]int, len(rays))
	for i, ray := range rays {
		baselineHits[i] = copyOf(baseline.ElementsIntersectingRay(ray, 0, 10))
	}

	probe := vector3.New(0.3, 0.4, 0.5)
	baselineNear := baseline.ElementsWithinRange(probe, 0.2)
	_, baselinePoint := baseline.ClosestPoint(probe)

	for _, tolerance := range []float64{0, 0.1, 0.25, 0.5, 0.75, 0.9} {
		t.Run(fmt.Sprintf("tolerance_%v", tolerance), func(t *testing.T) {
			tree := NewOctreeWithRetention(elements, depth, tolerance)

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
	depth := OctreeDepthFromCount(len(elements))

	for _, tolerance := range []float64{NoRetention, 0, 0.25, 0.5, 0.9, 1} {
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
			walk(NewOctreeWithRetention(elements, depth, tolerance))

			assert.Len(t, seen, len(elements))
			for index, count := range seen {
				assert.Equalf(t, 1, count, "element %d", index)
			}
		})
	}
}

func TestRetentionOfOneKeepsEverythingAtTheRoot(t *testing.T) {
	elements := sphereShell(500, 1)
	tree := NewOctreeWithRetention(elements, OctreeDepthFromCount(len(elements)), 1)

	assert.Empty(t, tree.children)
	assert.Len(t, tree.elements, len(elements))
}

type treeStats struct {
	Nodes, Leaves, Inner, MaxLeaf, Depth int
}

func statsOf(node *OctTree, depth int) treeStats {
	var stats treeStats
	if node == nil {
		return stats
	}

	stats.Nodes = 1
	stats.Depth = depth
	if len(node.children) == 0 {
		stats.Leaves = 1
		stats.MaxLeaf = len(node.elements)
	} else {
		stats.Inner = len(node.elements)
	}

	for _, child := range node.children {
		below := statsOf(child, depth+1)
		stats.Nodes += below.Nodes
		stats.Leaves += below.Leaves
		stats.Inner += below.Inner
		if below.MaxLeaf > stats.MaxLeaf {
			stats.MaxLeaf = below.MaxLeaf
		}
		if below.Depth > stats.Depth {
			stats.Depth = below.Depth
		}
	}
	return stats
}

// ElementsIntersectingRay returns only the elements the ray actually hits,
// which is a property of the geometry and identical whatever the tree looks
// like. What the tree decides is how many bounds get tested to find them.
func rayTests(node *OctTree, ray geometry.Ray, min, max float64) int {
	if !node.bounds.IntersectsRayInRange(ray, min, max) {
		return 1
	}

	tests := 1 + len(node.elements)
	for _, child := range node.children {
		tests += rayTests(child, ray, min, max)
	}
	return tests
}

func workloadNamed(name string) []Element {
	switch name {
	case "even100k":
		return sphereShell(100_000, 1)
	case "even1m":
		return sphereShell(1_000_000, 1)
	case "spanners100k":
		return sphereShellWithSpanners(100_000, 1, 6)
	}
	return nil
}

// One variant per process, driven from outside. Timing several trees in one
// run makes whichever went first absorb the page faults and the GC the others
// then get credit for avoiding.
//
// RETENTION=1 WORKLOAD=even100k TOLERANCE=0.25 go test ./trees/ -count=1 -run TestRetentionSweep -v
func TestRetentionSweep(t *testing.T) {
	if os.Getenv("RETENTION") == "" {
		t.Skip("set RETENTION=1 to run the sweep")
	}

	name := os.Getenv("WORKLOAD")
	elements := workloadNamed(name)
	if elements == nil {
		t.Fatalf("set WORKLOAD to even100k, even1m or spanners100k")
	}

	label := os.Getenv("TOLERANCE")
	tolerance := NoRetention
	if label != "off" {
		parsed, err := strconv.ParseFloat(label, 64)
		if err != nil {
			t.Fatalf("set TOLERANCE to a number or to off: %v", err)
		}
		tolerance = parsed
	}

	depth := OctreeDepthFromCount(len(elements))
	rays := testRays(2000, 1)
	probes := make([]vector3.Float64, 500)
	random := rand.New(rand.NewSource(7))
	for i := range probes {
		probes[i] = vector3.New(random.NormFloat64(), random.NormFloat64(), random.NormFloat64())
	}

	build := time.Duration(math.MaxInt64)
	var tree *OctTree
	for i := 0; i < 5; i++ {
		start := time.Now()
		tree = NewOctreeWithRetention(elements, depth, tolerance)
		if taken := time.Since(start); taken < build {
			build = taken
		}
	}

	stats := statsOf(tree, 1)

	tests := 0
	for _, ray := range rays {
		tests += rayTests(tree, ray, 0, 10)
	}

	rayTime := time.Duration(math.MaxInt64)
	for i := 0; i < 5; i++ {
		start := time.Now()
		for _, ray := range rays {
			tree.ElementsIntersectingRay(ray, 0, 10)
		}
		if taken := time.Since(start) / time.Duration(len(rays)); taken < rayTime {
			rayTime = taken
		}
	}

	closestTime := time.Duration(math.MaxInt64)
	for i := 0; i < 5; i++ {
		start := time.Now()
		for _, probe := range probes {
			tree.ClosestPoint(probe)
		}
		if taken := time.Since(start) / time.Duration(len(probes)); taken < closestTime {
			closestTime = taken
		}
	}

	fmt.Printf("SWEEP %s %-4s %6.1f %8d %8d %8d %8d %8.0f %10.0f %10.0f\n",
		name, label, float64(build.Microseconds())/1000,
		stats.Nodes, stats.Leaves, stats.MaxLeaf, stats.Inner,
		float64(tests)/float64(len(rays)),
		float64(rayTime.Nanoseconds()), float64(closestTime.Nanoseconds()))
}
