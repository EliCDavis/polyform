package stl_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/polyform/formats/stl"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestWriteRead(t *testing.T) {
	// ARRANGE ================================================================
	tri := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
			vector3.New(0., 0., 0.),
			vector3.New(1., 0., 0.),
			vector3.New(1., 1., 0.),
		})
	buf := &bytes.Buffer{}

	// ACT ====================================================================
	assert.NoError(t, stl.WriteMesh(buf, tri))
	cubeBack, err := stl.ReadMesh(buf)
	assert.NoError(t, err)

	// ASSERT =================================================================
	assert.Equal(t, tri.PrimitiveCount(), cubeBack.PrimitiveCount())
	assert.Equal(t, tri.AttributeLength(), cubeBack.AttributeLength())
	assert.False(t, cubeBack.HasFloat3Attribute(modeling.NormalAttribute))

	assert.Equal(t, vector3.New(0., 0., 0.), cubeBack.Tri(0).P1Vec3Attr(modeling.PositionAttribute))
	assert.Equal(t, vector3.New(1., 0., 0.), cubeBack.Tri(0).P2Vec3Attr(modeling.PositionAttribute))
	assert.Equal(t, vector3.New(1., 1., 0.), cubeBack.Tri(0).P3Vec3Attr(modeling.PositionAttribute))
}

func TestWriteReadWithNormals(t *testing.T) {
	// ARRANGE ================================================================
	tri := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
			vector3.New(0., 0., 0.),
			vector3.New(1., 0., 0.),
			vector3.New(1., 1., 0.),
		}).
		SetFloat3Attribute(modeling.NormalAttribute, []vector3.Float64{
			vector3.New(0., 0., 1.),
			vector3.New(0., 0., 1.),
			vector3.New(0., 0., 1.),
		})
	buf := &bytes.Buffer{}

	// ACT ====================================================================
	assert.NoError(t, stl.WriteMesh(buf, tri))
	cubeBack, err := stl.ReadMesh(buf)
	assert.NoError(t, err)

	// ASSERT =================================================================
	assert.Equal(t, tri.PrimitiveCount(), cubeBack.PrimitiveCount())
	assert.Equal(t, tri.AttributeLength(), cubeBack.AttributeLength())
	assert.True(t, cubeBack.HasFloat3Attribute(modeling.NormalAttribute))

	assert.Equal(t, vector3.New(0., 0., 0.), cubeBack.Tri(0).P1Vec3Attr(modeling.PositionAttribute))
	assert.Equal(t, vector3.New(1., 0., 0.), cubeBack.Tri(0).P2Vec3Attr(modeling.PositionAttribute))
	assert.Equal(t, vector3.New(1., 1., 0.), cubeBack.Tri(0).P3Vec3Attr(modeling.PositionAttribute))

	assert.Equal(t, vector3.New(0., 0., 1.), cubeBack.Tri(0).P1Vec3Attr(modeling.NormalAttribute))
	assert.Equal(t, vector3.New(0., 0., 1.), cubeBack.Tri(0).P2Vec3Attr(modeling.NormalAttribute))
	assert.Equal(t, vector3.New(0., 0., 1.), cubeBack.Tri(0).P3Vec3Attr(modeling.NormalAttribute))
}
