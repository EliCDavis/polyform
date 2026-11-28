package scalar_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/scalar"
	"github.com/stretchr/testify/assert"
)

func TestClamp(t *testing.T) {
	assert.Equal(t, 0.5, scalar.Clamp(0.5, 0, 1))
	assert.Equal(t, 0., scalar.Clamp(-.0001, 0, 1))
	assert.Equal(t, 1., scalar.Clamp(1.0001, 0, 1))
	assert.Equal(t, -1, scalar.Clamp(-4, -1, 2))
	assert.Equal(t, 2, scalar.Clamp(9, -1, 2))
	assert.True(t, math.IsNaN(scalar.Clamp(math.NaN(), 0, 1)))
}

func TestFract(t *testing.T) {
	assert.Equal(t, 0.25, scalar.Fract(3.25))
	assert.Equal(t, 0.75, scalar.Fract(-3.25))
	assert.Equal(t, 0., scalar.Fract(2))
}

func TestModuloTakesTheDivisorsSign(t *testing.T) {
	assert.Equal(t, 1, scalar.Modulo(7, 3))
	assert.Equal(t, 2, scalar.Modulo(-7, 3))
	assert.Equal(t, -2, scalar.Modulo(7, -3))
	assert.Equal(t, -1, scalar.Modulo(-7, -3))
	assert.Equal(t, 0, scalar.Modulo(-6, 3))
	assert.Equal(t, 1.5, scalar.Modulo(-0.5, 2.))
	assert.Equal(t, -1.5, scalar.Modulo(0.5, -2.))
}

func TestModuloByZeroIsZero(t *testing.T) {
	assert.Equal(t, 0, scalar.Modulo(7, 0))
	assert.Equal(t, 0., scalar.Modulo(7., 0))
}

func TestSmoothStep(t *testing.T) {
	assert.Equal(t, 0., scalar.SmoothStep(-1, 0, 1))
	assert.Equal(t, 1., scalar.SmoothStep(2, 0, 1))
	assert.Equal(t, 0.5, scalar.SmoothStep(0.5, 0, 1))
	assert.Equal(t, 0.15625, scalar.SmoothStep(0.25, 0, 1))
}

func TestSmoothStepDescendsWhenFromIsAboveTo(t *testing.T) {
	assert.Equal(t, 0.15625, scalar.SmoothStep(0.75, 1, 0))
	assert.Equal(t, 1., scalar.SmoothStep(-1, 1, 0))
}

func TestSmoothStepWithEqualEdgesIsAHardStep(t *testing.T) {
	assert.Equal(t, 0., scalar.SmoothStep(0.4, 0.5, 0.5))
	assert.Equal(t, 1., scalar.SmoothStep(0.5, 0.5, 0.5))
}
