package triangulation_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func square(size float64, center vector2.Float64) geometry.Shape {
	return geometry.RoundedRectangle(vector2.New(size, size), 0, center, 1)
}

func requireCounterClockwiseArea(t *testing.T, region triangulation.Region) float64 {
	t.Helper()
	total := 0.
	for i, tri := range region.Triangles {
		area := geometry.Shape{region.Points[tri[0]], region.Points[tri[1]], region.Points[tri[2]]}.SignedArea()
		require.Positivef(t, area, "triangle %d is not counter-clockwise", i)
		total += area
	}
	return total
}

func TestFillDifferenceCutsOverlappingHolesOutOnce(t *testing.T) {
	region, err := triangulation.FillDifference(
		[]geometry.Shape{square(10, vector2.New(0., 0.))},
		[]geometry.Shape{square(4, vector2.New(-1., 0.)), square(4, vector2.New(1., 0.))},
	)
	require.NoError(t, err)

	assert.InDelta(t, 100.-6*4, requireCounterClockwiseArea(t, region), 1e-9, "the holes' overlap is cut, not filled back in")
}

func TestFillDifferenceMergesOverlappingOutlines(t *testing.T) {
	region, err := triangulation.FillDifference(
		[]geometry.Shape{square(2, vector2.New(0., 0.)), square(2, vector2.New(1., 0.))},
		nil,
	)
	require.NoError(t, err)

	assert.InDelta(t, 6., requireCounterClockwiseArea(t, region), 1e-9)
}

func TestFillDifferenceCutsANotchWhereAHoleCrossesTheEdge(t *testing.T) {
	region, err := triangulation.FillDifference(
		[]geometry.Shape{square(4, vector2.New(0., 0.))},
		[]geometry.Shape{square(2, vector2.New(2., 0.)), square(1, vector2.New(10., 0.))},
	)
	require.NoError(t, err)

	assert.InDelta(t, 16.-2, requireCounterClockwiseArea(t, region), 1e-9, "only the part over the outline is cut; a hole outside it does nothing")
}

func TestFillDifferenceNamesOutlinesThatEncloseNothing(t *testing.T) {
	_, err := triangulation.FillDifference(
		[]geometry.Shape{{vector2.New(0., 0.), vector2.New(1., 0.), vector2.New(2., 0.)}},
		nil,
	)
	assert.ErrorContains(t, err, "no outline encloses any area")

	empty, err := triangulation.FillDifference(nil, []geometry.Shape{square(1, vector2.New(0., 0.))})
	require.NoError(t, err)
	assert.Empty(t, empty.Triangles, "nothing to fill is not an error")
}

func TestFillCutsAContourWoundAgainstTheOneAroundIt(t *testing.T) {
	outer := square(4, vector2.New(0., 0.))
	inner := square(2, vector2.New(0., 0.))
	reversed := make(geometry.Shape, len(inner))
	for i, p := range inner {
		reversed[len(inner)-1-i] = p
	}

	region, err := triangulation.Fill(outer, reversed)
	require.NoError(t, err)

	assert.InDelta(t, 12., requireCounterClockwiseArea(t, region), 1e-9)
}

func TestFillKeepsOnlyThePointsItsTrianglesUse(t *testing.T) {
	region, err := triangulation.Fill(square(2, vector2.New(0., 0.)), square(2, vector2.New(5., 0.)))
	require.NoError(t, err)

	used := map[int]bool{}
	for _, tri := range region.Triangles {
		for _, v := range tri {
			used[v] = true
		}
	}
	assert.Len(t, used, len(region.Points))
	assert.InDelta(t, 8., requireCounterClockwiseArea(t, region), 1e-9)
}
