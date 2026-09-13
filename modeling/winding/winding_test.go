package winding_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/winding"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func reversed(m modeling.Mesh) modeling.Mesh {
	indices := m.Indices()
	positions := m.Float3Attribute(modeling.PositionAttribute)

	verts := make([]vector3.Float64, positions.Len())
	for i := range verts {
		verts[i] = positions.At(i)
	}

	tris := make([]int, 0, indices.Len())
	for i := 0; i+2 < indices.Len(); i += 3 {
		tris = append(tris, indices.At(i), indices.At(i+2), indices.At(i+1))
	}

	return modeling.NewTriangleMesh(tris).
		SetFloat3Data(map[string][]vector3.Float64{modeling.PositionAttribute: verts})
}

func sampler(t *testing.T, m modeling.Mesh) winding.Sampler {
	t.Helper()
	s, err := winding.FromMesh(m)
	require.NoError(t, err)
	return s
}

func TestClosedSolidsWrapOnceInside(t *testing.T) {
	for name, solid := range map[string]modeling.Mesh{
		"cube":        primitives.Cube{Width: 2, Height: 2, Depth: 2}.UnweldedQuads(),
		"welded cube": primitives.Cube{Width: 2, Height: 2, Depth: 2, Dimensions: 4}.Welded(),
		"sphere":      primitives.UVSphere(1, 16, 24),
		"quad sphere": primitives.QuadSphere(1, primitives.Cube{Width: 1, Height: 1, Depth: 1, Dimensions: 6}, false, false),
		"cylinder":    primitives.Cylinder{Sides: 16, Height: 2, Radius: 1}.ToMesh(),
	} {
		t.Run(name, func(t *testing.T) {
			s := sampler(t, solid)
			assert.InDelta(t, 1, s.Number(vector3.Zero[float64]()), 1e-9)
			assert.InDelta(t, 0, s.Number(vector3.New(50., 50., 50.)), 1e-9)
		})
	}
}

func exactCube() modeling.Mesh {
	corners := []vector3.Float64{
		vector3.New(-1., -1., -1.), vector3.New(1., -1., -1.),
		vector3.New(1., 1., -1.), vector3.New(-1., 1., -1.),
		vector3.New(-1., -1., 1.), vector3.New(1., -1., 1.),
		vector3.New(1., 1., 1.), vector3.New(-1., 1., 1.),
	}

	return modeling.NewTriangleMesh([]int{
		0, 3, 2, 0, 2, 1,
		4, 5, 6, 4, 6, 7,
		0, 1, 5, 0, 5, 4,
		3, 7, 6, 3, 6, 2,
		0, 4, 7, 0, 7, 3,
		1, 2, 6, 1, 6, 5,
	}).SetFloat3Data(map[string][]vector3.Float64{
		modeling.PositionAttribute: corners,
	})
}

func TestTheNumberHoldsRightUpToTheSurface(t *testing.T) {
	s := sampler(t, exactCube())

	for _, gap := range []float64{1e-1, 1e-3, 1e-6, 1e-9} {
		for _, on := range []struct {
			inside, outside vector3.Float64
		}{
			{vector3.New(1-gap, 0.5, -0.2), vector3.New(1+gap, 0.5, -0.2)},
			{vector3.New(-1+gap, 0.2, 0.7), vector3.New(-1-gap, 0.2, 0.7)},
			{vector3.New(0.3, -1+gap, -0.6), vector3.New(0.3, -1-gap, -0.6)},
		} {
			assert.InDeltaf(t, 1, s.Number(on.inside), 1e-6, "%v, gap %g", on.inside, gap)
			assert.InDeltaf(t, 0, s.Number(on.outside), 1e-6, "%v, gap %g", on.outside, gap)
		}
	}
}

func TestReversedSolidsWrapBackwards(t *testing.T) {
	cube := primitives.Cube{Width: 2, Height: 2, Depth: 2}.UnweldedQuads()

	s := sampler(t, reversed(cube))
	assert.InDelta(t, -1, s.Number(vector3.Zero[float64]()), 1e-9)
	assert.True(t, s.Inside(vector3.Zero[float64]()))
}

func TestCavitiesCancelTheShellAroundThem(t *testing.T) {
	outer := primitives.Cube{Width: 2, Height: 2, Depth: 2}.UnweldedQuads()
	inner := reversed(primitives.Cube{Width: 1, Height: 1, Depth: 1}.UnweldedQuads())

	s := sampler(t, outer.Append(inner))

	assert.InDelta(t, 1, s.Number(vector3.New(0.75, 0., 0.)), 1e-9)
	assert.InDelta(t, 0, s.Number(vector3.Zero[float64]()), 1e-9)
	assert.InDelta(t, 0, s.Number(vector3.New(50., 0., 0.)), 1e-9)

	assert.True(t, s.Inside(vector3.New(0.75, 0., 0.)))
	assert.False(t, s.Inside(vector3.Zero[float64]()))
}

func TestOpenMeshesStillAnswer(t *testing.T) {
	lidless := primitives.Cylinder{Sides: 16, Height: 2, Radius: 1, NoTop: true}.ToMesh()
	s := sampler(t, lidless)

	assert.True(t, s.Inside(vector3.Zero[float64]()))
	assert.False(t, s.Inside(vector3.New(5., 0., 0.)))
	assert.InDelta(t, 0, s.Number(vector3.New(50., 0., 0.)), 1e-3)
}

func TestInsideAgreesWithTheShapeItCameFrom(t *testing.T) {
	radius := 1.
	s := sampler(t, primitives.UVSphere(radius, 32, 48))

	random := rand.New(rand.NewSource(7))
	for trial := 0; trial < 400; trial++ {
		p := vector3.New(
			random.Float64()*4-2,
			random.Float64()*4-2,
			random.Float64()*4-2,
		)

		distance := p.Length()
		if math.Abs(distance-radius) < 0.1 {
			continue
		}
		assert.Equalf(t, distance < radius, s.Inside(p),
			"point %v at radius %.4f", p, distance)
	}
}

func TestRejectsWhatItCannotSample(t *testing.T) {
	_, err := winding.FromMesh(modeling.EmptyMesh(modeling.PointTopology))
	require.Error(t, err)

	_, err = winding.FromMesh(modeling.EmptyMesh(modeling.TriangleTopology))
	require.Error(t, err)
}
