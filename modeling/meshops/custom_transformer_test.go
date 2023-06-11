package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestCustomTransformer(t *testing.T) {
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

	customOp := meshops.CustomTransformer{
		Func: func(m modeling.Mesh) (results modeling.Mesh, err error) {
			return m.ModifyFloat3Attribute(modeling.PositionAttribute, func(i int, v vector3.Float64) vector3.Float64 {
				return vector3.New(-10., -10., -10.).Scale(float64(i + 1))
			}), nil
		},
	}

	// ACT ====================================================================
	transformedMesh := mesh.Transform(customOp)

	// ASSERT ================================================================-
	arr := transformedMesh.Float3Attribute(modeling.PositionAttribute)
	assert.Equal(t, 3, arr.Len())
	assert.Equal(t, vector3.New(-10., -10., -10.), arr.At(0))
	assert.Equal(t, vector3.New(-20., -20., -20.), arr.At(1))
	assert.Equal(t, vector3.New(-30., -30., -30.), arr.At(2))
}
