package geometry_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestSolidAngle_OctantIsAnEighthOfTheSphere(t *testing.T) {
	x, y, z := vector3.Right[float64](), vector3.Up[float64](), vector3.Forward[float64]()

	assert.InDelta(t, math.Pi/2, geometry.SolidAngle(x, y, z), 1e-12)
	assert.InDelta(t, -math.Pi/2, geometry.SolidAngle(x, z, y), 1e-12)
}

func TestSolidAngle_ShrinksWithDistance(t *testing.T) {
	near := geometry.SolidAngle(
		vector3.New(1., 0., 0.), vector3.New(0., 1., 0.), vector3.New(0., 0., 1.))
	far := geometry.SolidAngle(
		vector3.New(100., 0., 0.), vector3.New(0., 100., 0.), vector3.New(0., 0., 100.))
	assert.InDelta(t, near, far, 1e-12, "scaling about the origin leaves the angle alone")

	shifted := geometry.SolidAngle(
		vector3.New(101., 100., 100.), vector3.New(100., 101., 100.), vector3.New(100., 100., 101.))
	assert.Less(t, math.Abs(shifted), 1e-4)
}
