package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestScaleAttribute3D(t *testing.T) {
	// ARRANGE ================================================================
	mesh := modeling.
		NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(
			modeling.PositionAttribute,
			[]vector3.Float64{
				vector3.New(0., 1., 2.),
				vector3.New(3., 4., 5.),
				vector3.New(6., 7., 8.),
			},
		)

	translateOp := meshops.ScaleAttribute3DTransformer{
		Amount: vector3.New(2., 3., 4.),
	}

	// ACT ====================================================================
	transformedMesh := mesh.Transform(translateOp)

	// ASSERT ================================================================-
	arr := transformedMesh.Float3Attribute(modeling.PositionAttribute)
	assert.Equal(t, 3, arr.Len())
	assert.Equal(t, vector3.New(0., 3., 8.), arr.At(0))
	assert.Equal(t, vector3.New(6., 12., 20.), arr.At(1))
	assert.Equal(t, vector3.New(12., 21., 32.), arr.At(2))
}
