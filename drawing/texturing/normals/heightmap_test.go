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
