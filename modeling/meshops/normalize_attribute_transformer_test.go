package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeAttribute3DTransformer(t *testing.T) {
	// ARRANGE ================================================================
	someAttribute := "something"

	mesh := modeling.
		NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(
			someAttribute,
			[]vector3.Float64{
				vector3.New(0., 1., 0.),
				vector3.New(0., 0., 2.),
				vector3.New(4., 0., 0.),
			},
		)

	normalizeAttribute := meshops.NormalizeAttribute3DTransformer{
		Attribute: someAttribute,
	}

	// ACT ====================================================================
	transformedMesh := mesh.Transform(normalizeAttribute)

	// ASSERT ================================================================-
	arr := transformedMesh.Float3Attribute(someAttribute)
	assert.Equal(t, 3, arr.Len())
	assert.Equal(t, vector3.New(0., .25, 0.), arr.At(0))
	assert.Equal(t, vector3.New(0., 0., 0.5), arr.At(1))
	assert.Equal(t, vector3.New(1., 0., 0.), arr.At(2))
}

func TestNormalizeAttribute2DTransformer(t *testing.T) {
	// ARRANGE ================================================================
	someAttribute := "something"

	mesh := modeling.
		NewTriangleMesh([]int{0, 1, 2}).
		SetFloat2Attribute(
			someAttribute,
			[]vector2.Float64{
				vector2.New(1., 0.),
				vector2.New(0., -2.),
				vector2.New(0., 4.),
			},
		)

	normalizeAttribute := meshops.NormalizeAttribute2DTransformer{
		Attribute: someAttribute,
	}

	// ACT ====================================================================
	transformedMesh := mesh.Transform(normalizeAttribute)

	// ASSERT ================================================================-
	arr := transformedMesh.Float2Attribute(someAttribute)
	assert.Equal(t, 3, arr.Len())
	assert.Equal(t, vector2.New(.25, 0.), arr.At(0))
	assert.Equal(t, vector2.New(0., -0.5), arr.At(1))
	assert.Equal(t, vector2.New(0., 1.), arr.At(2))
}
