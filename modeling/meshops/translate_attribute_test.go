package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestTranslateAttribute3D(t *testing.T) {
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

	translateOp := meshops.TranslateAttribute3DTransformer{
		Amount: vector3.New(4., 5., 6.),
	}

	// ACT ====================================================================
	transformedMesh := mesh.Transform(translateOp)

	// ASSERT ================================================================-
	arr := transformedMesh.Float3Attribute(modeling.PositionAttribute)
	assert.Equal(t, 3, arr.Len())
	assert.Equal(t, vector3.New(4., 6., 8.), arr.At(0))
	assert.Equal(t, vector3.New(7., 9., 11.), arr.At(1))
	assert.Equal(t, vector3.New(10., 12., 14.), arr.At(2))
}
