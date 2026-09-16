package csg

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Normal (1,1,1)/sqrt3, so dropping the dominant axis shrinks the in-plane
// direction (2,-1,-1) to 58% of its length.
func slantedFace(t *testing.T) face {
	t.Helper()
	f, ok := newFace(vector3.New(0., 0., 0.), vector3.New(1., -1., 0.), vector3.New(1., 0., -1.))
	require.True(t, ok)
	return f
}

func cornerIDs(weld *welder, f face) [3]int {
	return [3]int{weld.Index(f.verts[0]), weld.Index(f.verts[1]), weld.Index(f.verts[2])}
}

func TestACutEndpointNearACornerIsThatCornerOnBothSides(t *testing.T) {
	const tolerance = 1e-6
	f := slantedFace(t)
	inward := vector3.New(2., -1., -1.).Normalized()
	across := f.verts[1].Add(f.verts[2]).Scale(0.5)

	for _, gap := range []float64{0.9 * tolerance, 1.5 * tolerance} {
		weld := newWelder(tolerance)
		ids := cornerIDs(weld, f)
		onCurve := map[int]bool{}

		near := f.verts[0].Add(inward.Scale(gap))
		pieces, pieceIDs, err := splitFace(f, ids, []segment{{near, across}}, nil, tolerance, weld, onCurve)
		require.NoErrorf(t, err, "gap %g", gap)
		require.Greaterf(t, len(pieces), 1, "gap %g", gap)

		for _, corners := range pieceIDs {
			for _, id := range corners {
				if id != ids[0] && id != ids[1] && id != ids[2] {
					assert.Truef(t, onCurve[id], "gap %g: piece corner %d is a cut point the curve does not know", gap, id)
				}
			}
		}

		touching := make([]bool, len(pieces))
		assert.Lenf(t, patchesOf(pieceIDs, onCurve, touching), 2, "gap %g: the cut should separate the face", gap)
	}
}

// A right-angled sliver 1.5 tolerances wide, laid along the direction that
// dropping the dominant axis shrinks most, so it is under a tolerance wide in
// projection but not in space.
func TestANeedleFaceIsStillSplit(t *testing.T) {
	const tolerance = 1e-6
	across := vector3.New(2., -1., -1.).Normalized()
	a := vector3.New(0., 0., 0.)
	b := a.Add(across.Scale(1.5 * tolerance))
	c := vector3.New(0., 1., -1.)
	f, ok := newFace(a, b, c)
	require.True(t, ok)

	weld := newWelder(tolerance)
	ids := cornerIDs(weld, f)
	require.NotEqual(t, ids[0], ids[1], "sanity: the two close corners are distinct to the welder")

	// Across the needle near its wide end, where the pieces on both sides
	// are still taller than the tolerance.
	onCurve := map[int]bool{}
	cut := segment{c.Add(a.Sub(c).Scale(0.95)), c.Add(b.Sub(c).Scale(0.95))}
	pieces, pieceIDs, err := splitFace(f, ids, []segment{cut}, nil, tolerance, weld, onCurve)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(pieces), 2)

	for _, corners := range pieceIDs {
		for _, id := range corners {
			if id != ids[0] && id != ids[1] && id != ids[2] {
				assert.Truef(t, onCurve[id], "piece corner %d is a cut point the curve does not know", id)
			}
		}
	}
}

func TestAFaceTouchedAtItsBarycenterByAnApexIsOutside(t *testing.T) {
	box, err := facesOf(primitives.Cube{Width: 2, Height: 2, Depth: 2}.UnweldedQuads())
	require.NoError(t, err)

	var top face
	for _, f := range box {
		if f.normal.Y() > 0.9 {
			top = f
			break
		}
	}
	apex := top.barycenter()

	base := [4]vector3.Float64{
		apex.Add(vector3.New(-1., 2., -1.)),
		apex.Add(vector3.New(1., 2., -1.)),
		apex.Add(vector3.New(1., 2., 1.)),
		apex.Add(vector3.New(-1., 2., 1.)),
	}
	pyramid := modeling.NewTriangleMesh([]int{
		0, 1, 2,
		0, 2, 3,
		0, 3, 4,
		0, 4, 1,
		1, 3, 2,
		1, 4, 3,
	}).SetFloat3Data(map[string][]vector3.Float64{
		modeling.PositionAttribute: {apex, base[0], base[1], base[2], base[3]},
	})
	require.NoError(t, CheckClosed(pyramid), "sanity: the pyramid is wound consistently")

	faces, err := facesOf(pyramid)
	require.NoError(t, err)
	against := newSolid(faces, toleranceFor(box, faces), 0)

	assert.Equal(t, outside, against.classify(top))
}
