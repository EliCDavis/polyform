package sketch_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling/sketch"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOffsetGrowsWithRoundCorners(t *testing.T) {
	grown, err := sketch.Offset(squareArea(t, 10, vector2.New(0., 0.)), 1, 32)
	require.NoError(t, err)

	require.Len(t, grown.Polygons, 1)
	assert.InDelta(t, 100+4*10+math.Pi, grown.Size(), 0.01)
	for _, p := range grown.Polygons[0].Outline {
		outside := vector2.New(math.Max(math.Abs(p.X())-5, 0), math.Max(math.Abs(p.Y())-5, 0))
		assert.InDelta(t, 1, outside.Length(), 1e-6, "every boundary point is the distance from the square, at %v", p)
	}
	requireWellFormed(t, grown)
}

func TestOffsetShrinksWithSharpCorners(t *testing.T) {
	shrunk, err := sketch.Offset(squareArea(t, 10, vector2.New(0., 0.)), -1, 16)
	require.NoError(t, err)

	require.Len(t, shrunk.Polygons, 1)
	assert.InDelta(t, 64, shrunk.Size(), 1e-9)
}

func TestOffsetMergesAreasItGrowsTogether(t *testing.T) {
	apart, err := sketch.Union(squareArea(t, 2, vector2.New(-2., 0.)), squareArea(t, 2, vector2.New(2., 0.)))
	require.NoError(t, err)
	require.Len(t, apart.Polygons, 2)

	merged, err := sketch.Offset(apart, 1.5, 16)
	require.NoError(t, err)

	assert.Len(t, merged.Polygons, 1)
}

func TestOffsetClosesAHoleNarrowerThanTwiceTheDistance(t *testing.T) {
	ring, err := sketch.Stroke(squarePath(10), 2, 16, true)
	require.NoError(t, err)
	require.Len(t, ring.Polygons[0].Holes, 1)

	grown, err := sketch.Offset(ring, 5, 16)
	require.NoError(t, err)

	require.Len(t, grown.Polygons, 1)
	assert.Empty(t, grown.Polygons[0].Holes)
}

func TestOffsetShrinksAThinAreaToNothing(t *testing.T) {
	gone, err := sketch.Offset(squareArea(t, 2, vector2.New(0., 0.)), -1.5, 16)
	require.NoError(t, err)

	assert.Empty(t, gone.Polygons)
}

func TestOffsetNodeGivesOneAreaPerDistance(t *testing.T) {
	areas := nodes.GetNodeOutputPort[[]sketch.Area](&nodes.Struct[sketch.OffsetNode]{Data: sketch.OffsetNode{
		Area:     nodes.ConstOutput[sketch.Area]{Val: squareArea(t, 10, vector2.New(0., 0.))},
		Distance: nodes.ConstOutput[[]float64]{Val: []float64{-1, 1}},
	}}, "Out").Value()

	require.Len(t, areas, 2)
	assert.Less(t, areas[0].Size(), 100.)
	assert.Greater(t, areas[1].Size(), 100.)
}

func TestBoundaryNodeSplitsOutlinesFromHoles(t *testing.T) {
	ring, err := sketch.Stroke(squarePath(10), 2, 16, true)
	require.NoError(t, err)
	node := &nodes.Struct[sketch.BoundaryNode]{Data: sketch.BoundaryNode{Area: nodes.ConstOutput[sketch.Area]{Val: ring}}}

	outlines := nodes.GetNodeOutputPort[[][]vector2.Float64](node, "Outlines").Value()
	holes := nodes.GetNodeOutputPort[[][]vector2.Float64](node, "Holes").Value()

	require.Len(t, outlines, 1)
	require.Len(t, holes, 1)
	assert.Positive(t, geometry.Shape(outlines[0]).SignedArea())
	assert.Negative(t, geometry.Shape(holes[0]).SignedArea())
}

func offsetReference(t *testing.T, area sketch.Area, distance float64, segments int) sketch.Area {
	t.Helper()
	var band []sketch.Area
	for _, loop := range area.Loops() {
		stroke, err := sketch.Stroke(loop, 2*math.Abs(distance), segments, true)
		require.NoError(t, err)
		band = append(band, stroke)
	}
	if distance > 0 {
		grown, err := sketch.Union(append([]sketch.Area{area}, band...)...)
		require.NoError(t, err)
		return grown
	}
	shrunk, err := sketch.Subtract(area, band...)
	require.NoError(t, err)
	return shrunk
}

func offsetTestShapes(t *testing.T) map[string]sketch.Area {
	ring, err := sketch.Stroke(squarePath(10), 2, 16, true)
	require.NoError(t, err)
	star := make([]vector2.Float64, 0, 14)
	for i := range 14 {
		r := 5.
		if i%2 == 1 {
			r = 2
		}
		angle := float64(i) * math.Pi / 7
		star = append(star, vector2.New(math.Cos(angle), math.Sin(angle)).Scale(r))
	}
	starArea, err := sketch.Fill([]geometry.Shape{star}, nil)
	require.NoError(t, err)
	ell, err := sketch.Fill([]geometry.Shape{{
		vector2.New(0., 0.), vector2.New(6., 0.), vector2.New(6., 2.), vector2.New(2., 2.), vector2.New(2., 6.), vector2.New(0., 6.),
	}}, nil)
	require.NoError(t, err)
	comb, err := sketch.Fill([]geometry.Shape{{
		vector2.New(0., 0.), vector2.New(9., 0.), vector2.New(9., 5.), vector2.New(8., 5.), vector2.New(8., 1.), vector2.New(5., 1.),
		vector2.New(5., 5.), vector2.New(4., 5.), vector2.New(4., 1.), vector2.New(1., 1.), vector2.New(1., 5.), vector2.New(0., 5.),
	}}, nil)
	require.NoError(t, err)
	pair, err := sketch.Union(squareArea(t, 2, vector2.New(-2., 0.)), squareArea(t, 2, vector2.New(2., 0.)))
	require.NoError(t, err)
	trace, err := sketch.Stroke([]vector2.Float64{vector2.New(0., 0.), vector2.New(6., 0.), vector2.New(6., 4.), vector2.New(1., 4.)}, 1, 8, false)
	require.NoError(t, err)
	return map[string]sketch.Area{
		"square": squareArea(t, 10, vector2.New(0., 0.)), "ring": ring, "star": starArea,
		"ell": ell, "comb": comb, "pair": pair, "trace": trace,
	}
}

func TestOffsetMatchesTheMinkowskiReference(t *testing.T) {
	for name, area := range offsetTestShapes(t) {
		for _, distance := range []float64{-1.5, -0.6, -0.2, 0.2, 0.6, 1.5, 4} {
			got, err := sketch.Offset(area, distance, 64)
			require.NoError(t, err, "%s by %g", name, distance)
			want := offsetReference(t, area, distance, 64)

			assert.InDelta(t, want.Size(), got.Size(), 1e-3*math.Max(want.Size(), 1), "%s by %g", name, distance)
			assert.Equal(t, len(want.Polygons), len(got.Polygons), "%s by %g pieces", name, distance)
			requireWellFormed(t, got)
		}
	}
}
