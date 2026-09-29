package sdf_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestSmoothIntersectRoundsTheCreaseAndMatchesHardIntersectAwayFromIt(t *testing.T) {
	a := sdf.Sphere(vector3.New(-0.5, 0., 0.), 1)
	b := sdf.Sphere(vector3.New(0.5, 0., 0.), 1)
	hard := sdf.Intersect(a, b)
	smooth := sdf.SmoothIntersect(0.2, a, b)

	for _, far := range []vector3.Float64{vector3.New(-0.9, 0., 0.), vector3.New(0.3, 0., 0.)} {
		assert.InDelta(t, hard(far), smooth(far), 1e-9, far)
	}

	crease := vector3.New(0., 0.8660254, 0.)
	assert.InDelta(t, 0, hard(crease), 1e-6)
	assert.Greater(t, smooth(crease), 0.0)
	assert.InDelta(t, 0.05, smooth(crease), 1e-6, "h=1 adds radius/4")
}

func TestSmoothIntersectManyFieldsUsesTheTwoLargest(t *testing.T) {
	a := sdf.Sphere(vector3.Zero[float64](), 1)
	b := sdf.Sphere(vector3.Zero[float64](), 1.05)
	c := sdf.Sphere(vector3.Zero[float64](), 5)
	two := sdf.SmoothIntersect(0.2, a, b)
	three := sdf.SmoothIntersect(0.2, a, b, c)

	p := vector3.New(0., 1., 0.)
	assert.InDelta(t, two(p), three(p), 1e-9, "a huge third field never wins the max, so it cannot change the blend")
}

func TestSmoothIntersectZeroRadiusIsHard(t *testing.T) {
	a := sdf.Sphere(vector3.New(-0.5, 0., 0.), 1)
	b := sdf.Sphere(vector3.New(0.5, 0., 0.), 1)
	p := vector3.New(0., 0.8660254, 0.)
	assert.InDelta(t, sdf.Intersect(a, b)(p), sdf.SmoothIntersect(0, a, b)(p), 1e-12)
}
