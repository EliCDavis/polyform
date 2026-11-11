package normals_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/drawing/texturing/normals"
	"github.com/stretchr/testify/assert"
)

func TestFlatHeightMapIsUnoccluded(t *testing.T) {
	flat := texturing.Empty[float64](8, 8)
	flat.Fill(0.5)

	occlusion := normals.OcclusionFromHeightmap(flat, 3, 2)

	occlusion.Scan(func(x, y int, v float64) {
		assert.Equal(t, 1., v)
	})
}

func TestPitIsDarkerThanTheRidgeAroundIt(t *testing.T) {
	heights := texturing.Empty[float64](16, 16)
	heights.Fill(1)
	heights.Set(8, 8, 0)

	occlusion := normals.OcclusionFromHeightmap(heights, 2, 2)

	assert.Less(t, occlusion.Get(8, 8), 0.1)
	assert.Equal(t, 1., occlusion.Get(2, 2))
}

func TestOcclusionWrapsAcrossTheEdge(t *testing.T) {
	heights := texturing.Empty[float64](8, 8)
	heights.Set(7, 0, 1)

	occlusion := normals.OcclusionFromHeightmap(heights, 1, 2)

	assert.Less(t, occlusion.Get(0, 0), 1.)
}
