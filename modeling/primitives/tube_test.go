package primitives_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTubeWindingMatchesNormals(t *testing.T) {
	mesh := primitives.Tube{Sides: 16, Height: 2, InnerRadius: 0.4, OuterRadius: 1}.ToMesh()

	require.True(t, mesh.HasFloat3Attribute(modeling.NormalAttribute))
	positions := mesh.Float3Attribute(modeling.PositionAttribute)
	normals := mesh.Float3Attribute(modeling.NormalAttribute)

	indices := mesh.Indices()
	for i := 0; i < indices.Len(); i += 3 {
		a, b, c := indices.At(i), indices.At(i+1), indices.At(i+2)
		pa, pb, pc := positions.At(a), positions.At(b), positions.At(c)

		face := pb.Sub(pa).Cross(pc.Sub(pa))
		require.Greater(t, face.Length(), 1e-12, "degenerate triangle at index %d", i)
		face = face.Normalized()

		for _, v := range []int{a, b, c} {
			assert.Greaterf(t, face.Dot(normals.At(v)), 0.0,
				"triangle %d faces %v but vertex %d's normal is %v", i/3, face, v, normals.At(v))
		}
	}
}

func TestTubeIsHollow(t *testing.T) {
	mesh := primitives.Tube{Sides: 12, Height: 1, InnerRadius: 0.4, OuterRadius: 1}.ToMesh()
	positions := mesh.Float3Attribute(modeling.PositionAttribute)

	minRadius, maxRadius := 1e9, 0.
	for i := 0; i < positions.Len(); i++ {
		p := positions.At(i)
		r := vector3.New(p.X(), 0., p.Z()).Length()
		minRadius = min(minRadius, r)
		maxRadius = max(maxRadius, r)
	}

	assert.InDelta(t, 0.4, minRadius, 1e-9)
	assert.InDelta(t, 1.0, maxRadius, 1e-9)
}

func TestTubeSwappedRadiiStillBuilds(t *testing.T) {
	swapped := primitives.Tube{Sides: 12, Height: 1, InnerRadius: 1, OuterRadius: 0.4}.ToMesh()
	normal := primitives.Tube{Sides: 12, Height: 1, InnerRadius: 0.4, OuterRadius: 1}.ToMesh()
	assert.Equal(t, normal.PrimitiveCount(), swapped.PrimitiveCount())
}
