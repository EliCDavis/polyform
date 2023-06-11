package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestCenterAttribute3D(t *testing.T) {
	// ARRANGE ================================================================
	mesh := modeling.
		NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(
			modeling.PositionAttribute,
			[]vector3.Float64{
				vector3.New(1., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(0., 0., 1.),
			},
		)

	centerOp := meshops.CenterAttribute3DTransformer{
		Attribute: modeling.PositionAttribute,
	}

	// ACT ====================================================================
	transformedMesh := mesh.Transform(centerOp)

	// ASSERT ================================================================-
	arr := transformedMesh.Float3Attribute(modeling.PositionAttribute)
	assert.Equal(t, 3, arr.Len())
	assert.Equal(t, vector3.New(0.5, -0.5, -0.5), arr.At(0))
	assert.Equal(t, vector3.New(-0.5, 0.5, -0.5), arr.At(1))
	assert.Equal(t, vector3.New(-0.5, -0.5, 0.5), arr.At(2))
}
