package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestRemoveNullFaces3DTransformer_KeepsFacesWithArea(t *testing.T) {
	// ARRANGE ================================================================
	someAttribute := "something"

	mesh := modeling.
		NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(
			someAttribute,
			[]vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 1., 0.),
			},
		)

	removeNullFacesAttribute := meshops.RemoveNullFaces3DTransformer{
		Attribute: someAttribute,
	}

	// ACT ====================================================================
	transformedMesh := mesh.Transform(removeNullFacesAttribute)

	// ASSERT ================================================================-
	arr := transformedMesh.Float3Attribute(someAttribute)
	if assert.Equal(t, 3, arr.Len()) {
		assert.Equal(t, vector3.New(0., 0., 0.), arr.At(0))
		assert.Equal(t, vector3.New(0., 1., 0.), arr.At(1))
		assert.Equal(t, vector3.New(1., 1., 0.), arr.At(2))
	}
}

func TestRemoveNullFaces3DTransformer_RemovesFacesWithNoArea(t *testing.T) {
	// ARRANGE ================================================================
	someAttribute := "something"

	mesh := modeling.
		NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(
			someAttribute,
			[]vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 0., 0.),
				vector3.New(0., 0., 0.),
			},
		)

	removeNullFacesAttribute := meshops.RemoveNullFaces3DTransformer{
		Attribute: someAttribute,
	}

	// ACT ====================================================================
	transformedMesh := mesh.Transform(removeNullFacesAttribute)

	// ASSERT ================================================================-
	assert.False(t, transformedMesh.HasFloat3Attribute(someAttribute))
	assert.Equal(t, 0, transformedMesh.Indices().Len())
}
