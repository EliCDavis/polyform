package predicate_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/predicate"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestPredicatesAnswerRatherThanPanicOnNaN(t *testing.T) {
	nan := math.NaN()
	inf := math.Inf(1)

	for name, run := range map[string]func() float64{
		"Orient2D NaN": func() float64 {
			return predicate.Orient2D(
				vector2.New(0., 0.), vector2.New(1., nan), vector2.New(2., 2.))
		},
		"Orient2D Inf": func() float64 {
			return predicate.Orient2D(
				vector2.New(0., 0.), vector2.New(inf, 1.), vector2.New(2., 2.))
		},
		"Orient3D NaN": func() float64 {
			return predicate.Orient3D(
				vector3.New(0., 0., 0.), vector3.New(1., 0., 0.),
				vector3.New(0., 1., 0.), vector3.New(nan, 0., 1.))
		},
		"InCircle NaN": func() float64 {
			return predicate.InCircle(
				vector2.New(0., 0.), vector2.New(1., 0.),
				vector2.New(0., 1.), vector2.New(nan, nan))
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				assert.True(t, math.IsNaN(run()), "an unanswerable test should come back NaN")
			})
		})
	}
}

func TestSegmentsCrossSaysNoWhenItCannotTell(t *testing.T) {
	nan := math.NaN()
	assert.NotPanics(t, func() {
		assert.False(t, predicate.SegmentsCross(
			vector2.New(0., 0.), vector2.New(2., 2.),
			vector2.New(0., nan), vector2.New(2., 0.)))
	})
}
