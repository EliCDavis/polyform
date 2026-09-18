package csg

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type splitPair struct {
	a, b      half
	tolerance float64
}

// Everything combine builds before it classifies.
func splitBoth(t *testing.T, a, b modeling.Mesh) splitPair {
	t.Helper()

	first, err := NewSolid(a)
	require.NoError(t, err)
	second, err := NewSolid(b)
	require.NoError(t, err)

	tolerance := max(first.tolerance, second.tolerance)

	cutsA, cornersA, touchingA := cutsAgainst(first.faces, second.faces, tolerance)
	cutsB, cornersB, touchingB := cutsAgainst(second.faces, first.faces, tolerance)
	corners := append(cornersA, cornersB...)

	halfA, err := splitAll(first.faces, first.cornerIDs, cutsA, corners, touchingA, tolerance, newIDSpace(first.weld, tolerance))
	require.NoError(t, err)
	halfB, err := splitAll(second.faces, second.cornerIDs, cutsB, corners, touchingB, tolerance, newIDSpace(second.weld, tolerance))
	require.NoError(t, err)

	return splitPair{a: halfA, b: halfB, tolerance: tolerance}
}

// Section 7 read on its own: every face gets its own ray. Grouping has to
// agree with this, which is the whole reason it is allowed to skip rays.
func classifyEachFace(faces []face, against *target) []classification {
	answers := make([]classification, len(faces))
	for i, f := range faces {
		answers[i] = against.classify(f)
	}
	return answers
}

func testPairs() map[string][2]modeling.Mesh {
	cube := func(center vector3.Float64, side float64) modeling.Mesh {
		return primitives.Cube{Height: side, Width: side, Depth: side}.
			UnweldedQuads().
			Translate(center)
	}

	return map[string][2]modeling.Mesh{
		"overlapping cubes": {
			cube(vector3.Zero[float64](), 2),
			cube(vector3.New(0.8, 0.5, 0.3), 2),
		},
		"cubes meeting face to face": {
			cube(vector3.Zero[float64](), 2),
			cube(vector3.New(2., 0., 0.), 2),
		},
		"cubes sharing part of a face": {
			cube(vector3.Zero[float64](), 2),
			cube(vector3.New(2., 0.7, 0.4), 2),
		},
		"sphere through sphere": {
			primitives.UVSphere(1, 16, 24),
			primitives.UVSphere(1, 16, 24).Translate(vector3.New(0.7, 0.3, 0.2)),
		},
		"sphere through cube": {
			cube(vector3.Zero[float64](), 2),
			primitives.UVSphere(1.1, 20, 28).Translate(vector3.New(0.4, 0.4, 0.0)),
		},
		"cylinder through sphere": {
			primitives.UVSphere(1.28, 20, 32),
			primitives.Cylinder{Sides: 28, Height: 4, Radius: 0.62}.ToMesh(),
		},
		"corner touch": {
			cube(vector3.Zero[float64](), 2),
			cube(vector3.New(1.9, 1.9, 1.9), 2),
		},
		"deeply nested": {
			primitives.UVSphere(2, 18, 26),
			primitives.UVSphere(0.4, 14, 20),
		},
	}
}

func TestGroupingAgreesWithPerFaceClassification(t *testing.T) {
	for name, pair := range testPairs() {
		t.Run(name, func(t *testing.T) {
			p := splitBoth(t, pair[0], pair[1])

			groupedA, groupedB := classifyBoth(p.a, p.b, p.tolerance)

			targetA := newTarget(p.a.faces, p.tolerance, len(p.b.faces))
			targetB := newTarget(p.b.faces, p.tolerance, len(p.a.faces))

			require.Greater(t, len(p.a.faces), 0, "sanity: this pair should split into faces")
			assert.Equal(t, classifyEachFace(p.a.faces, targetB), groupedA, "first solid")
			assert.Equal(t, classifyEachFace(p.b.faces, targetA), groupedB, "second solid")
		})
	}
}

func TestGroupingSpendsFarFewerRaysThanFaces(t *testing.T) {
	p := splitBoth(t,
		primitives.UVSphere(1.28, 24, 36),
		primitives.UVSphere(1.28, 24, 36).Translate(vector3.New(0.6, 0.4, 0.3)),
	)

	patches := patchesOf(p.a.cornerIDs, p.a.onCurve, p.a.touching)
	rays := rayBudget(patches)

	require.Greater(t, len(p.a.faces), 1000, "sanity: this should be a mesh worth grouping")
	assert.Less(t, rays, len(p.a.faces)/50,
		"grouping fired %d rays for %d faces, which is not the saving it exists for",
		rays, len(p.a.faces))
}
