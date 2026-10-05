package geometry_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
)

func TestRoundedRectangleRoundsOnlyAsFarAsItCan(t *testing.T) {
	stadium := geometry.RoundedRectangle(vector2.New(4., 2.), 5, vector2.New(0., 0.), 16)
	assert.InDelta(t, 2*2+math.Pi, stadium.SignedArea(), 0.02, "the radius stops at half the short side")

	circle := geometry.RoundedRectangle(vector2.New(2., 2.), 1, vector2.New(3., 0.), 64)
	assert.InDelta(t, math.Pi, circle.SignedArea(), 0.01)
	for i, p := range circle {
		assert.InDeltaf(t, 1, p.Sub(vector2.New(3., 0.)).Length(), 1e-9, "point %d", i)
		next := circle[(i+1)%len(circle)]
		assert.Greaterf(t, next.Sub(p).Length(), 1e-9, "no point repeats where two arcs meet, at %d", i)
	}
}

func TestRoundedRectangleWithoutRadiusIsItsFourCorners(t *testing.T) {
	rectangle := geometry.RoundedRectangle(vector2.New(4., 2.), 0, vector2.New(1., 1.), 8)

	assert.Len(t, rectangle, 4)
	assert.InDelta(t, 8, rectangle.SignedArea(), 1e-9, "counter-clockwise")
}
