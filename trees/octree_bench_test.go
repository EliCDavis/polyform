package trees

import (
	"fmt"
	"testing"
)

var benchSizes = []int{10_000, 100_000, 1_000_000}

func BenchmarkOctreeConstruction(b *testing.B) {
	for _, size := range benchSizes {
		elements := sphereShell(size, 1)
		for _, tolerance := range []float64{noRetention, 0.2} {
			b.Run(fmt.Sprintf("n=%d/tolerance=%v", size, tolerance), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					NewOctree(elements, Tolerance(tolerance))
				}
			})
		}
	}
}

func benchWorkloads() map[string][]Element {
	return map[string][]Element{
		"shell":      sphereShell(100_000, 1),
		"shell+span": sphereShellWithSpanners(100_000, 1, 6),
	}
}

func BenchmarkOctreeElementsIntersectingRay(b *testing.B) {
	rays := testRays(1024, 1)
	for name, elements := range benchWorkloads() {
		for _, tolerance := range []float64{noRetention, 0.2} {
			tree := NewOctree(elements, Tolerance(tolerance))
			b.Run(fmt.Sprintf("%s/tolerance=%v", name, tolerance), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					tree.ElementsIntersectingRay(rays[i%len(rays)], 0, 10)
				}
			})
		}
	}
}

func BenchmarkOctreeTraverseIntersectingRay(b *testing.B) {
	rays := testRays(1024, 1)
	for name, elements := range benchWorkloads() {
		tree := NewOctree(elements, Tolerance(0.2))
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				tree.TraverseIntersectingRay(rays[i%len(rays)], 0, 10, func(j int, min, max *float64) {
					*max = 2.5
				})
			}
		})
	}
}

func BenchmarkOctreeClosestPoint(b *testing.B) {
	points := probes(1024, 1)
	tree := NewOctree(sphereShell(100_000, 1), Tolerance(0.2))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tree.ClosestPoint(points[i%len(points)])
	}
}

func BenchmarkOctreeElementsWithinRange(b *testing.B) {
	points := probes(1024, 1)
	tree := NewOctree(sphereShell(100_000, 1), Tolerance(0.2))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tree.ElementsWithinRange(points[i%len(points)], 0.05)
	}
}

func BenchmarkOctreeElementsContainingPoint(b *testing.B) {
	points := probes(1024, 1)
	for i := range points {
		points[i] = points[i].Normalized()
	}
	tree := NewOctree(sphereShell(100_000, 1), Tolerance(0.2))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tree.ElementsContainingPoint(points[i%len(points)])
	}
}
