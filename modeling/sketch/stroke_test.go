package sketch_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling/sketch"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func distanceToPath(p vector2.Float64, path []vector2.Float64) float64 {
	best := math.Inf(1)
	for i := 0; i+1 < len(path); i++ {
		best = math.Min(best, geometry.NewLine2D(path[i], path[i+1]).ClosestPointOnLine(p).Distance(p))
	}
	return best
}

func TestStrokeOfAStraightPathIsACapsule(t *testing.T) {
	area, err := sketch.Stroke([]vector2.Float64{vector2.New(0., 0.), vector2.New(10., 0.)}, 2, 32, false)
	require.NoError(t, err)
	require.Len(t, area.Polygons, 1)
	shape := geometry.Shape(area.Polygons[0].Outline)

	assert.InDelta(t, 10*2+math.Pi, shape.SignedArea(), 0.01, "counter-clockwise")
	lo, hi := shape.GetBounds()
	assert.InDelta(t, -1, lo.X(), 1e-9)
	assert.InDelta(t, 11, hi.X(), 1e-9)
	assert.InDelta(t, 1, hi.Y(), 1e-9)
}

func TestStrokeBoundaryStaysHalfAWidthFromTheBentPath(t *testing.T) {
	path := []vector2.Float64{vector2.New(0., 0.), vector2.New(10., 0.), vector2.New(10., 10.), vector2.New(3., 4.)}
	area, err := sketch.Stroke(path, 2, 16, false)
	require.NoError(t, err)
	require.Len(t, area.Polygons, 1)
	shape := geometry.Shape(area.Polygons[0].Outline)

	for i, p := range shape {
		assert.InDeltaf(t, 1, distanceToPath(p, path), 0.01, "point %d at %v", i, p)
	}
	assert.Positive(t, shape.SignedArea())
}

func TestTransformScalesThenRotatesThenMoves(t *testing.T) {
	moved := sketch.Transform([]vector2.Float64{vector2.New(1., 0.)}, vector2.New(5., 5.), 90, vector2.New(2., 1.))

	assert.InDelta(t, 5, moved[0].X(), 1e-9)
	assert.InDelta(t, 7, moved[0].Y(), 1e-9)
}
