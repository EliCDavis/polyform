package geometry_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var unitCorner = geometry.Triangle{
	vector3.New(0., 0., 0.), vector3.New(1., 0., 0.), vector3.New(0., 1., 0.),
}

func TestTriangleBasics(t *testing.T) {
	assert.Equal(t, vector3.New(0., 0., 1.), unitCorner.Normal())
	assert.InDelta(t, 0.5, unitCorner.Area(), 1e-12)
	assert.InDelta(t, 1./3., unitCorner.Centroid().X(), 1e-12)
	assert.InDelta(t, 0.745, unitCorner.Reach(), 1e-3)
	assert.Equal(t, vector3.Zero[float64](), geometry.Triangle{unitCorner[0], unitCorner[0], unitCorner[1]}.Normal())
}

func TestTriangleDegenerate(t *testing.T) {
	sliver := geometry.Triangle{
		vector3.New(0., 0., 0.), vector3.New(1., 0., 0.), vector3.New(0.5, 1e-12, 0.),
	}
	assert.True(t, sliver.Degenerate(1e-9))
	assert.False(t, sliver.Degenerate(1e-15))
	assert.False(t, unitCorner.Degenerate(1e-9))
}

func TestTriangleContains(t *testing.T) {
	assert.True(t, unitCorner.Contains(vector3.New(0.2, 0.2, 0.), 0))
	assert.False(t, unitCorner.Contains(vector3.New(0.2, 0.2, 1e-6), 1e-9), "off the plane")
	assert.True(t, unitCorner.Contains(vector3.New(0.2, 0.2, 1e-10), 1e-9))
	assert.False(t, unitCorner.Contains(vector3.New(0.6, 0.6, 0.), 0), "past the hypotenuse")
	assert.True(t, unitCorner.Contains(vector3.New(-1e-10, 0.5, 0.), 1e-9), "a hair outside an edge")
}

func TestDegenerateTriangleContainsNothing(t *testing.T) {
	flat := geometry.Triangle{vector3.New(0., 0., 0.), vector3.New(1., 0., 0.), vector3.New(2., 0., 0.)}
	above := vector3.New(0.5, 5., 0.)

	assert.False(t, flat.ProjectsInside(above))
	assert.False(t, flat.Contains(above, 1e-9))
	assert.False(t, flat.Contains(vector3.New(0.5, 0., 0.), 1e-9), "not even a point on its line")
	assert.InDelta(t, 5, flat.ClosestPoint(above).Distance(above), 1e-12, "falls back to the edges")
}

func TestTriangleClosestPoint(t *testing.T) {
	assert.InDelta(t, 0, unitCorner.ClosestPoint(vector3.New(0.2, 0.2, 5.)).Distance(vector3.New(0.2, 0.2, 0.)), 1e-12)
	assert.InDelta(t, 0, unitCorner.ClosestPoint(vector3.New(1., 1., 0.)).Distance(vector3.New(0.5, 0.5, 0.)), 1e-12)
	assert.InDelta(t, 0, unitCorner.ClosestPoint(vector3.New(3., -1., 2.)).Distance(vector3.New(1., 0., 0.)), 1e-12)

	onEdge, distance := unitCorner.ClosestPointOnEdges(vector3.New(0.5, -1., 0.))
	assert.Equal(t, vector3.New(0.5, 0., 0.), onEdge)
	assert.InDelta(t, 1, distance, 1e-12)
}

func TestTriangleIntersect(t *testing.T) {
	floor := geometry.Triangle{
		vector3.New(-2., 0., -2.), vector3.New(2., 0., -2.), vector3.New(0., 0., 2.),
	}
	wall := geometry.Triangle{
		vector3.New(-1., -1., 0.), vector3.New(1., -1., 0.), vector3.New(0., 1., 0.),
	}

	cut, ok := floor.Intersect(wall, 1e-9)
	require.True(t, ok)
	assert.InDelta(t, 0, cut.GetStartPoint().Y(), 1e-12)
	assert.InDelta(t, 0, cut.GetEndPoint().Y(), 1e-12)
	assert.InDelta(t, 0, cut.GetStartPoint().Z(), 1e-12)
	assert.InDelta(t, 1, cut.Length(), 1e-12, "the wall crosses y=0 between x=-0.5 and x=0.5")

	back, ok := wall.Intersect(floor, 1e-9)
	require.True(t, ok)
	assert.InDelta(t, cut.Length(), back.Length(), 1e-12)

	apart := wall.Translate(vector3.New(0., 5., 0.))
	_, ok = floor.Intersect(apart, 1e-9)
	assert.False(t, ok)

	_, ok = floor.Intersect(floor.Translate(vector3.New(0., 1e-12, 0.)), 1e-9)
	assert.False(t, ok, "coplanar within tolerance")

	resting := geometry.Triangle{
		vector3.New(-1., 0., 0.), vector3.New(1., 0., 0.), vector3.New(0., 1., 0.),
	}
	shared, ok := floor.Intersect(resting, 1e-9)
	require.True(t, ok, "an edge lying on the other triangle is shared")
	assert.InDelta(t, 2, shared.Length(), 1e-12)

	touching := geometry.Triangle{
		vector3.New(0., 0., 0.), vector3.New(1., 1., 0.), vector3.New(-1., 1., 0.),
	}
	_, ok = floor.Intersect(touching, 1e-9)
	assert.False(t, ok, "a single corner on the plane is no segment")
}

func TestTriangle2D(t *testing.T) {
	tri := geometry.Triangle2D{vector2.New(0., 0.), vector2.New(2., 0.), vector2.New(0., 2.)}

	weights := tri.Barycentric(vector2.New(0.5, 0.5))
	assert.InDelta(t, 0.5, weights[0], 1e-12)
	assert.InDelta(t, 0.25, weights[1], 1e-12)
	assert.InDelta(t, 0.25, weights[2], 1e-12)

	outside := tri.Barycentric(vector2.New(3., 0.))
	assert.Negative(t, outside[0])
	assert.InDelta(t, 1, outside[0]+outside[1]+outside[2], 1e-12)

	assert.True(t, tri.Contains(vector2.New(0.5, 0.5), 0))
	assert.False(t, tri.Contains(vector2.New(2., 2.), 0))
	assert.True(t, tri.Contains(vector2.New(-1e-10, 1.), 1e-9))
	reversed := geometry.Triangle2D{tri[0], tri[2], tri[1]}
	assert.True(t, reversed.Contains(vector2.New(0.5, 0.5), 0))
}

func TestPlaneLiftInvertsDropAxis(t *testing.T) {
	plane := geometry.NewPlane(vector3.New(1., 2., 3.), vector3.New(1., 1., 1.).Normalized())
	point := plane.ClosestPoint(vector3.New(4., -2., 0.5))

	for axis := 0; axis < 3; axis++ {
		lifted := plane.Lift(geometry.DropAxis(point, axis), axis)
		assert.InDeltaf(t, 0, lifted.Distance(point), 1e-12, "axis %d", axis)
	}
	assert.Equal(t, 2, geometry.DominantAxis(vector3.New(0.1, -0.2, 0.9)))
	assert.LessOrEqual(t, 1/math.Sqrt(3), math.Abs(vector3.New(1., 1., 1.).Normalized().Z()))
}
