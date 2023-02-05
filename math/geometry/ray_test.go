package geometry_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestRay(t *testing.T) {
	ray := geometry.NewRay(vector3.Zero[float64](), vector3.Up[float64]())

	assert.Equal(t, vector3.Zero[float64](), ray.Origin())
	assert.Equal(t, vector3.Up[float64](), ray.Direction())
	assert.Equal(t, vector3.Up[float64]().Scale(2), ray.At(2))
}
