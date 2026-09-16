package predicate_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/predicate"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrient2DAnswersTheObviousCases(t *testing.T) {
	a := vector2.New(0., 0.)
	b := vector2.New(1., 0.)

	assert.Positive(t, predicate.Orient2D(a, b, vector2.New(0., 1.)))
	assert.Negative(t, predicate.Orient2D(a, b, vector2.New(0., -1.)))
	assert.Zero(t, predicate.Orient2D(a, b, vector2.New(2., 0.)))
}

func TestOrient2DDoesNotMistakeCancellationForCollinear(t *testing.T) {
	a := vector2.New(60.466028797961954, 94.05090880450125)
	b := vector2.New(126.922034119811, 137.82232752319925)
	c := vector2.New(342.6631464502089, 279.92076568416394)

	naive := (b.X()-a.X())*(c.Y()-a.Y()) - (c.X()-a.X())*(b.Y()-a.Y())
	require.Zero(t, naive, "sanity: plain arithmetic really does cancel here")

	assert.Positive(t, predicate.Orient2D(a, b, c),
		"these three points do turn, and the answer has to say so")
}

func TestOrient2DStaysExactAcrossAWideExponentSpread(t *testing.T) {
	tiny, huge := math.Ldexp(1, -100), math.Ldexp(1, 100)
	a := vector2.New(tiny, tiny)
	b := vector2.New(huge, huge)
	c := vector2.New(1., 1.)

	assert.Zero(t, predicate.Orient2D(a, b, c))
	assert.Positive(t, predicate.Orient2D(a, b, vector2.New(1., math.Nextafter(1., 2.))))
	assert.Negative(t, predicate.Orient2D(a, b, vector2.New(math.Nextafter(1., 2.), 1.)))
}

func TestInCircleStaysExactAcrossAWideExponentSpread(t *testing.T) {
	huge := math.Ldexp(1, 100)
	corner1 := vector2.New(huge, 0.)
	corner2 := vector2.New(0., huge)
	corner3 := vector2.New(-huge, 0.)
	bottom := vector2.New(0., -huge)

	assert.Zero(t, predicate.InCircle(corner1, corner2, corner3, bottom))
	assert.Positive(t, predicate.InCircle(corner1, corner2, corner3, vector2.New(0., -math.Nextafter(huge, 0))))
	assert.Negative(t, predicate.InCircle(corner1, corner2, corner3, vector2.New(0., -math.Nextafter(huge, math.Inf(1)))))
	assert.Positive(t, predicate.InCircle(corner1, corner2, corner3, vector2.New(math.Ldexp(1, -100), 0.)))
}

func TestOrient3DAnswersTheObviousCases(t *testing.T) {
	a := vector3.New(0., 0., 0.)
	b := vector3.New(1., 0., 0.)
	c := vector3.New(0., 1., 0.)

	above := predicate.Orient3D(a, b, c, vector3.New(0., 0., 1.))
	below := predicate.Orient3D(a, b, c, vector3.New(0., 0., -1.))

	assert.NotZero(t, above)
	assert.NotEqual(t, above > 0, below > 0, "the two sides should disagree")
	assert.Zero(t, predicate.Orient3D(a, b, c, vector3.New(2., 3., 0.)),
		"a point in the plane should be exactly zero, not merely small")
}

func TestSegmentsCrossSeparatesCrossingFromTouching(t *testing.T) {
	cross := predicate.SegmentsCross(
		vector2.New(0., 0.), vector2.New(2., 2.),
		vector2.New(0., 2.), vector2.New(2., 0.))
	assert.True(t, cross, "an X crosses")

	apart := predicate.SegmentsCross(
		vector2.New(0., 0.), vector2.New(1., 0.),
		vector2.New(0., 1.), vector2.New(1., 1.))
	assert.False(t, apart, "parallel segments do not")

	touching := predicate.SegmentsCross(
		vector2.New(0., 0.), vector2.New(2., 0.),
		vector2.New(1., 0.), vector2.New(1., 2.))
	assert.False(t, touching,
		"a T only touches, and counting it would reject outlines that are fine")

	shared := predicate.SegmentsCross(
		vector2.New(0., 0.), vector2.New(1., 1.),
		vector2.New(1., 1.), vector2.New(2., 0.))
	assert.False(t, shared, "segments meeting end to end do not cross")
}
