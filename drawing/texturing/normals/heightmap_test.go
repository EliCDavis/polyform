package normals_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/drawing/texturing/normals"
	"github.com/stretchr/testify/assert"
)

func TestHeightmapNormalsTiltAwayFromHigherGroundWithGreenUp(t *testing.T) {
	risingDown := texturing.Empty[float64](5, 5)
	risingDown.MutateParallel(func(x, y int, _ float64) float64 { return float64(y) })
	risingRight := texturing.Empty[float64](5, 5)
	risingRight.MutateParallel(func(x, y int, _ float64) float64 { return float64(x) })

	down := normals.FromHeightmap(risingDown, 1).Get(2, 2)
	right := normals.FromHeightmap(risingRight, 1).Get(2, 2)

	assert.Greater(t, down.Y(), 0.)
	assert.Less(t, right.X(), 0.)
}

func TestHeightmapNormalsMatchAcrossATilingSeam(t *testing.T) {
	tile := normals.FromHeightmap(texturing.FromArray([]float64{0, 2, 3, 1}, 4, 1), 1)
	shifted := normals.FromHeightmap(texturing.FromArray([]float64{1, 0, 2, 3}, 4, 1), 1)

	assert.InDelta(t, shifted.Get(1, 0).X(), tile.Get(0, 0).X(), 1e-9)
	assert.InDelta(t, shifted.Get(0, 0).X(), tile.Get(3, 0).X(), 1e-9)
}
