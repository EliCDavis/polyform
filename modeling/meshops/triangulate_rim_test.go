package meshops

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/predicate"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lShapedRim() []vector3.Float64 {
	corners := [][2]float64{
		{0, 0}, {1, 0}, {2, 0}, {3, 0},
		{3, 1}, {2, 1}, {1, 1},
		{1, 2}, {1, 3},
		{0, 3}, {0, 2}, {0, 1},
	}
	rim := make([]vector3.Float64, len(corners))
	for i, c := range corners {
		rim[i] = vector3.New(c[0], 0, c[1])
	}
	return rim
}

func outlineOf(rim []vector3.Float64) []vector2.Float64 {
	return geometry.NewPlaneFromPolygon(rim).Project(rim)
}

func area2D(flat [][2]float64) float64 {
	twice := 0.
	for i, p := range flat {
		q := flat[(i+1)%len(flat)]
		twice += p[0]*q[1] - q[0]*p[1]
	}
	return math.Abs(twice) / 2
}

func TestTriangulateOutlineCoversItExactlyOnce(t *testing.T) {
	flat := outlineOf(lShapedRim())

	patch, ok := triangulateOutline(flat)
	require.True(t, ok, "a flat L should be a usable outline")
	require.Len(t, patch, len(flat)-2, "n points take n-2 triangles")

	outline := make([][2]float64, len(flat))
	for i, p := range flat {
		outline[i] = [2]float64{p.X(), p.Y()}
	}

	covered := 0.
	for _, tri := range patch {
		covered += area2D([][2]float64{
			{flat[tri[0]].X(), flat[tri[0]].Y()},
			{flat[tri[1]].X(), flat[tri[1]].Y()},
			{flat[tri[2]].X(), flat[tri[2]].Y()},
		})
	}

	assert.InDelta(t, area2D(outline), covered, 1e-9,
		"the triangles should tile the hole, not double back over it")
}

func TestTriangulateOutlineWindsEveryTriangleTheSameWay(t *testing.T) {
	flat := outlineOf(lShapedRim())

	patch, ok := triangulateOutline(flat)
	require.True(t, ok)

	var reference float64
	for i, tri := range patch {
		turn := predicate.Orient2D(flat[tri[0]], flat[tri[1]], flat[tri[2]])
		require.NotZero(t, turn, "triangle %d collapsed to a line", i)
		if reference == 0 {
			reference = turn
			continue
		}
		assert.Equal(t, reference > 0, turn > 0, "triangle %d faces the other way", i)
	}

	assert.NotEqual(t, geometry.Shape(flat).SignedArea() > 0, reference > 0,
		"the fill has to run against the rim for the two to agree on which side is out")
}

func TestTriangulateOutlineRefusesOneThatCrossesItself(t *testing.T) {
	rim := []vector3.Float64{
		vector3.New(0., 0., 0.),
		vector3.New(2., 0., 1.),
		vector3.New(2., 0., 0.),
		vector3.New(0., 0., 3.),
	}

	_, ok := triangulateOutline(outlineOf(rim))
	assert.False(t, ok, "should decline rather than produce a crossed patch")
}

func TestClosesOutlineAcceptsOnlyAPatchBorderedByEveryRimEdge(t *testing.T) {
	assert.True(t, closesOutline([][3]int{{0, 2, 1}, {0, 3, 2}}, 4),
		"two triangles wound against a square rim")
	assert.False(t, closesOutline([][3]int{{0, 1, 2}, {0, 2, 3}}, 4),
		"the same two triangles wound with the rim")
	assert.False(t, closesOutline([][3]int{{0, 2, 1}, {0, 4, 2}}, 5),
		"a patch that skips a merged rim point")
	assert.False(t, closesOutline(nil, 3),
		"an empty patch")
}
