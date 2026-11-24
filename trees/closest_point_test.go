package trees

import (
	"fmt"
	"testing"

	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

type countingElement struct {
	Element
	tests *int
}

func (c countingElement) ClosestPoint(p vector3.Float64) vector3.Float64 {
	*c.tests++
	return c.Element.ClosestPoint(p)
}

func TestOctreeClosestPointTestsFewElements(t *testing.T) {
	tests := 0
	shell := sphereShellWithSpanners(20_000, 1, 5)
	elements := make([]Element, len(shell))
	for i, e := range shell {
		elements[i] = countingElement{Element: e, tests: &tests}
	}
	points := probes(500, 1)

	for _, tolerance := range []float64{noRetention, 0.2} {
		t.Run(fmt.Sprintf("tolerance=%v", tolerance), func(t *testing.T) {
			tree := NewOctree(elements, Tolerance(tolerance))
			tests = 0
			for _, p := range points {
				tree.ClosestPoint(p)
			}
			assert.Less(t, tests/len(points), 20)
		})
	}
}

func TestOctreeClosestPointDoesNotAllocate(t *testing.T) {
	tree := NewOctree(sphereShell(20_000, 1))
	points := probes(64, 1)
	i := 0
	allocs := testing.AllocsPerRun(200, func() {
		tree.ClosestPoint(points[i%len(points)])
		i++
	})
	assert.Zero(t, allocs)
}
