package mcp

import (
	"math"
	"testing"

	"github.com/fogleman/fauxgl"
	"github.com/stretchr/testify/assert"
)

func TestGridColumnsFillEveryCellWhenACountUpToFiveDivides(t *testing.T) {
	for n, want := range map[int]int{1: 1, 2: 2, 3: 3, 4: 2, 5: 5, 6: 3, 7: 3, 8: 4, 9: 3, 10: 5, 12: 4} {
		assert.Equal(t, want, gridColumns(n, 0), "n=%d", n)
	}
	assert.Equal(t, 2, gridColumns(6, 2))
	assert.Equal(t, 3, gridColumns(3, 9), "never more columns than views")
}

func TestFitDistanceHoldsTheNearFaceInsideTheFrame(t *testing.T) {
	unit := fauxgl.Box{Min: fauxgl.V(-0.5, -0.5, -0.5), Max: fauxgl.V(0.5, 0.5, 0.5)}
	front := fitDistance(unit, fauxgl.V(0, 0, 1), fauxgl.V(0, 1, 0), 1, defaultFov)
	assert.InDelta(t, 0.5+fitMargin*0.5/math.Tan(14*math.Pi/180), front, 1e-9)

	tanHalf := math.Tan(14 * math.Pi / 180)
	post := fauxgl.Box{Min: fauxgl.V(-0.25, -5, -0.25), Max: fauxgl.V(0.25, 5, 0.25)}
	side := fitDistance(post, fauxgl.V(0, 0, 1), fauxgl.V(0, 1, 0), 1, defaultFov)
	above := fitDistance(post, fauxgl.V(0, 1, 0), fauxgl.V(0, 0, -1), 1, defaultFov)
	assert.InDelta(t, 0.25+fitMargin*5/tanHalf, side, 1e-9, "side-on, the height sets the distance")
	assert.InDelta(t, 5+fitMargin*0.25/tanHalf, above, 1e-9, "from above, the footprint does and the top stays past the eye")

	wide := fauxgl.Box{Min: fauxgl.V(-5, -0.5, -0.5), Max: fauxgl.V(5, 0.5, 0.5)}
	square := fitDistance(wide, fauxgl.V(0, 0, 1), fauxgl.V(0, 1, 0), 1, defaultFov)
	widescreen := fitDistance(wide, fauxgl.V(0, 0, 1), fauxgl.V(0, 1, 0), 2, defaultFov)
	assert.InDelta(t, square, 2*(widescreen-0.5)+0.5, 1e-9, "a wider frame halves the distance the width demands")
}

func TestFitRadiusRoundTripsThroughDistance(t *testing.T) {
	assert.InDelta(t, 0.3, fitRadiusOf(distanceForRadius(0.3, defaultFov), defaultFov), 1e-12)
	assert.InDelta(t, 0.3*4.4, distanceForRadius(0.3, defaultFov), 0.01, "the long-standing 4.4x rule still holds at the default fov")
}

func TestLineOfSightTellsEnteringFromLeaving(t *testing.T) {
	// A unit cube's +Z face, wound counter-clockwise seen from +Z.
	face := []*fauxgl.Triangle{
		fauxgl.NewTriangleForPoints(fauxgl.V(-1, -1, 1), fauxgl.V(1, -1, 1), fauxgl.V(1, 1, 1)),
		fauxgl.NewTriangleForPoints(fauxgl.V(-1, -1, 1), fauxgl.V(1, 1, 1), fauxgl.V(-1, 1, 1)),
	}
	leaving := lineOfSight(face, fauxgl.V(0.2, -0.1, 0), fauxgl.V(0.2, -0.1, 5))
	assert.Len(t, leaving, 1)
	assert.False(t, leaving[0].entering, "from inside the cube the ray leaves through its face")
	assert.InDelta(t, 1, leaving[0].t, 1e-9)

	entering := lineOfSight(face, fauxgl.V(0.2, -0.1, 5), fauxgl.V(0.2, -0.1, 0))
	assert.Len(t, entering, 1)
	assert.True(t, entering[0].entering)

	assert.Empty(t, lineOfSight(face, fauxgl.V(0.2, -0.1, 0), fauxgl.V(0.2, -0.1, 0.5)), "a segment that stops short does not reach it")
	assert.Empty(t, lineOfSight(face, fauxgl.V(3, 0, 0), fauxgl.V(3, 0, 5)), "a ray beside the face misses")
}

func TestDefaultViewDirectionRoundTripsThroughAngles(t *testing.T) {
	az, el := directionAngles(defaultViewDirection)
	back := azimuthElevationDirection(az, el)
	assert.InDelta(t, 0, back.Sub(defaultViewDirection).Length(), 1e-12)
}
