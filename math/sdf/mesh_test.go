package sdf_test

import (
	"math/rand"
	"testing"

	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertMatchesBox(t *testing.T, mesh modeling.Mesh) {
	field, err := sdf.Mesh(mesh)
	require.NoError(t, err)
	box := sdf.Box(vector3.Zero[float64](), vector3.One[float64]())

	rng := rand.New(rand.NewSource(4))
	for i := 0; i < 3000; i++ {
		p := vector3.New(rng.Float64()*3-1.5, rng.Float64()*3-1.5, rng.Float64()*3-1.5)
		assert.InDelta(t, box(p), field(p), 1e-9, "at %v", p)
	}

	for _, p := range []vector3.Float64{
		vector3.New(0.7, 0.7, 0.7), vector3.New(-0.6, 0.6, 0.), vector3.New(0.45, 0.45, 0.45),
		vector3.New(0.5, 0.5, 0.9), vector3.New(0.6, -0.5, 0.5),
	} {
		assert.InDelta(t, box(p), field(p), 1e-9, "at %v", p)
	}
}

func TestMeshWeldedCubeMatchesBox(t *testing.T) {
	assertMatchesBox(t, primitives.UnitCube())
}

func TestMeshUnweldedCubeMatchesBox(t *testing.T) {
	cube := primitives.Cube{Width: 1, Height: 1, Depth: 1}.UnweldedQuads()
	assertMatchesBox(t, cube)
}

func TestMeshSphereSignMatchesInside(t *testing.T) {
	sphere := primitives.UVSphere(1, 24, 24)
	field, err := sdf.Mesh(sphere)
	require.NoError(t, err)

	rng := rand.New(rand.NewSource(5))
	for i := 0; i < 2000; i++ {
		p := vector3.New(rng.Float64()*3-1.5, rng.Float64()*3-1.5, rng.Float64()*3-1.5)
		length := p.Length()
		if length > 0.95 && length < 1.05 {
			continue
		}
		assert.Equal(t, length < 1, field(p) < 0, "at %v", p)
	}
}

func TestMeshOpenSurfaceSignFollowsWinding(t *testing.T) {
	quad := modeling.NewTriangleMesh([]int{0, 1, 2, 0, 2, 3}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
			vector3.New(-1., 0., -1.), vector3.New(-1., 0., 1.), vector3.New(1., 0., 1.), vector3.New(1., 0., -1.),
		})
	field, err := sdf.Mesh(quad)
	require.NoError(t, err)

	assert.InDelta(t, 0.5, field(vector3.New(0., 0.5, 0.)), 1e-9)
	assert.InDelta(t, -0.5, field(vector3.New(0., -0.5, 0.)), 1e-9)
}

func TestMeshEmptyIsNothing(t *testing.T) {
	field, err := sdf.Mesh(modeling.EmptyMesh(modeling.TriangleTopology))
	require.NoError(t, err)
	assert.Greater(t, field(vector3.Zero[float64]()), 1e9)
}

func TestMeshRejectsPointClouds(t *testing.T) {
	_, err := sdf.Mesh(modeling.EmptyPointcloud())
	assert.Error(t, err)
}
