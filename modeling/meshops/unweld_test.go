package meshops_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
	"github.com/stretchr/testify/assert"
)

func TestUnweld(t *testing.T) {
	// ARRANGE ================================================================
	inMesh := modeling.NewTriangleMesh([]int{0, 1, 2, 0, 1, 3}).
		SetFloat1Attribute("attr-1", []float64{
			1, 2, 3, 4,
		}).
		SetFloat2Attribute("attr-2", []vector2.Float64{
			vector2.New(0., 1.), vector2.New(0., 2.), vector2.New(0., 3.), vector2.New(0., 4.),
		}).
		SetFloat3Attribute("attr-3", []vector3.Float64{
			vector3.New(0., 1., 0.), vector3.New(0., 2., 0.), vector3.New(0., 3., 0.), vector3.New(0., 4., 0.),
		}).
		SetFloat4Attribute("attr-4", []vector4.Float64{
			vector4.New(0., 1., 0., 0.), vector4.New(0., 2., 0., 0.), vector4.New(0., 3., 0., 0.), vector4.New(0., 4., 0., 0.),
		})

	op := meshops.UnweldTransformer{}

	// ACT ====================================================================
	unweldedMesh := inMesh.Transform(op)

	// ASSERT =================================================================
	attr1 := unweldedMesh.Float1Attribute("attr-1")
	attr2 := unweldedMesh.Float2Attribute("attr-2")
	attr3 := unweldedMesh.Float3Attribute("attr-3")
	attr4 := unweldedMesh.Float4Attribute("attr-4")

	assert.Equal(t, 6, attr1.Len())
	assert.Equal(t, 1., attr1.At(0))
	assert.Equal(t, 2., attr1.At(1))
	assert.Equal(t, 3., attr1.At(2))
	assert.Equal(t, 1., attr1.At(3))
	assert.Equal(t, 2., attr1.At(4))
	assert.Equal(t, 4., attr1.At(5))

	assert.Equal(t, 6, attr2.Len())
	assert.Equal(t, vector2.New(0., 1.), attr2.At(0))
	assert.Equal(t, vector2.New(0., 2.), attr2.At(1))
	assert.Equal(t, vector2.New(0., 3.), attr2.At(2))
	assert.Equal(t, vector2.New(0., 1.), attr2.At(3))
	assert.Equal(t, vector2.New(0., 2.), attr2.At(4))
	assert.Equal(t, vector2.New(0., 4.), attr2.At(5))

	assert.Equal(t, 6, attr3.Len())
	assert.Equal(t, vector3.New(0., 1., 0.), attr3.At(0))
	assert.Equal(t, vector3.New(0., 2., 0.), attr3.At(1))
	assert.Equal(t, vector3.New(0., 3., 0.), attr3.At(2))
	assert.Equal(t, vector3.New(0., 1., 0.), attr3.At(3))
	assert.Equal(t, vector3.New(0., 2., 0.), attr3.At(4))
	assert.Equal(t, vector3.New(0., 4., 0.), attr3.At(5))

	assert.Equal(t, 6, attr4.Len())
	assert.Equal(t, vector4.New(0., 1., 0., 0.), attr4.At(0))
	assert.Equal(t, vector4.New(0., 2., 0., 0.), attr4.At(1))
	assert.Equal(t, vector4.New(0., 3., 0., 0.), attr4.At(2))
	assert.Equal(t, vector4.New(0., 1., 0., 0.), attr4.At(3))
	assert.Equal(t, vector4.New(0., 2., 0., 0.), attr4.At(4))
	assert.Equal(t, vector4.New(0., 4., 0., 0.), attr4.At(5))
}
