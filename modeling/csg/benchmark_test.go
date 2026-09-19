package csg

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
)

// 12 triangles per cell squared, so dimensions 92 lands near 100k and 289
// near a million.
func benchSphere(dimensions int, offset vector3.Float64) modeling.Mesh {
	return primitives.QuadSphere(0.5, primitives.Cube{
		Width: 1, Height: 1, Depth: 1, Dimensions: dimensions,
	}, false, true).Translate(offset)
}

func megabytes() float64 {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return float64(stats.HeapAlloc) / (1 << 20)
}

func TestBenchmarkStages(t *testing.T) {
	if os.Getenv("BENCH") == "" {
		t.Skip("set BENCH=1")
	}

	sizes := []int{10, 30, 92}
	if os.Getenv("BENCH_BIG") != "" {
		sizes = append(sizes, 160, 289)
	}

	fmt.Printf("%9s %10s %9s %10s %9s %9s %9s %9s %9s %9s %10s %8s\n",
		"tris/in", "read", "weld", "closed", "cuts", "split", "patches", "trees",
		"classify", "emit", "TOTAL", "heapMB")

	for _, dimensions := range sizes {
		a := benchSphere(dimensions, vector3.Zero[float64]())
		b := benchSphere(dimensions, vector3.New(0.31, 0.27, 0.19))

		runtime.GC()
		wall := time.Now()

		mark := time.Now()
		facesA, err := facesOf(a)
		if err != nil {
			t.Fatal(err)
		}
		facesB, _ := facesOf(b)
		read := time.Since(mark)

		mark = time.Now()
		first := &Solid{mesh: a, faces: facesA}
		second := &Solid{mesh: b, faces: facesB}
		first.once.Do(func() {
			first.tolerance = toleranceFor(facesA, nil)
			first.weld = newWelder(first.tolerance)
			first.cornerIDs = weldFaces(first.weld, facesA)
		})
		second.once.Do(func() {
			second.tolerance = toleranceFor(facesB, nil)
			second.weld = newWelder(second.tolerance)
			second.cornerIDs = weldFaces(second.weld, facesB)
		})
		welding := time.Since(mark)

		mark = time.Now()
		if err := validate(a, facesA, first.cornerIDs); err != nil {
			t.Fatal(err)
		}
		if err := validate(b, facesB, second.cornerIDs); err != nil {
			t.Fatal(err)
		}
		checked := time.Since(mark)

		tolerance := max(first.tolerance, second.tolerance)

		mark = time.Now()
		cutsA, cornersA, touchingA := cutsAgainst(facesA, facesB, tolerance)
		cutsB, cornersB, touchingB := cutsAgainst(facesB, facesA, tolerance)
		corners := append(cornersA, cornersB...)
		cutting := time.Since(mark)

		mark = time.Now()
		halfA, err := splitAll(facesA, first.cornerIDs, cutsA, corners, touchingA, tolerance, newIDSpace(first.weld, tolerance))
		if err != nil {
			t.Fatal(err)
		}
		halfB, err := splitAll(facesB, second.cornerIDs, cutsB, corners, touchingB, tolerance, newIDSpace(second.weld, tolerance))
		if err != nil {
			t.Fatal(err)
		}
		splitting := time.Since(mark)

		mark = time.Now()
		patchesA := patchesOf(halfA.cornerIDs, halfA.onCurve, halfA.touching)
		patchesB := patchesOf(halfB.cornerIDs, halfB.onCurve, halfB.touching)
		grouping := time.Since(mark)

		mark = time.Now()
		targetA := newTarget(halfA.faces, tolerance, rayBudget(patchesB))
		targetB := newTarget(halfB.faces, tolerance, rayBudget(patchesA))
		building := time.Since(mark)

		mark = time.Now()
		answersA := classifyPatches(halfA.faces, patchesA, targetB)
		answersB := classifyPatches(halfB.faces, patchesB, targetA)
		classifying := time.Since(mark)

		fromA := kept{source: a}
		fromB := kept{source: b, inverted: true}
		for i, f := range halfA.faces {
			if keepFromA[difference][answersA[i]] {
				fromA.faces = append(fromA.faces, f)
			}
		}
		for i, f := range halfB.faces {
			if keepFromB[difference][answersB[i]] {
				fromB.faces = append(fromB.faces, f.reversed())
			}
		}

		mark = time.Now()
		out, _ := meshFromFaces(fromA, fromB)
		emitting := time.Since(mark)

		total := time.Since(wall)
		round := func(d time.Duration) string {
			if d > time.Second {
				return fmt.Sprintf("%.2fs", d.Seconds())
			}
			return fmt.Sprintf("%dms", d.Milliseconds())
		}

		fmt.Printf("%9d %10s %9s %10s %9s %9s %9s %9s %9s %9s %10s %8.0f   (%d patches, %d rays, %d out)\n",
			len(facesA), round(read), round(welding), round(checked), round(cutting), round(splitting),
			round(grouping), round(building), round(classifying), round(emitting), round(total),
			megabytes(), len(patchesA)+len(patchesB),
			rayBudget(patchesA)+rayBudget(patchesB), out.Indices().Len()/3)
	}
}
