package triangulation_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func area(mesh modeling.Mesh) float64 {
	verts := mesh.Float3Attribute(modeling.PositionAttribute)
	idx := mesh.Indices()
	total := 0.
	for i := 0; i < idx.Len(); i += 3 {
		a, b, c := verts.At(idx.At(i)), verts.At(idx.At(i+1)), verts.At(idx.At(i+2))
		total += math.Abs((b.X()-a.X())*(c.Z()-a.Z())-(c.X()-a.X())*(b.Z()-a.Z())) / 2
	}
	return total
}

func flat(v vector3.Float64) vector2.Float64 { return vector2.New(v.X(), v.Z()) }

func cross2(p1, p2, p3, p4 vector2.Float64) bool {
	o := func(a, b, c vector2.Float64) float64 {
		return (b.X()-a.X())*(c.Y()-a.Y()) - (c.X()-a.X())*(b.Y()-a.Y())
	}
	d1, d2 := o(p1, p2, p3), o(p1, p2, p4)
	d3, d4 := o(p3, p4, p1), o(p3, p4, p2)
	return d1 != 0 && d2 != 0 && d3 != 0 && d4 != 0 &&
		(d1 > 0) != (d2 > 0) && (d3 > 0) != (d4 > 0)
}

// Holds whether or not a constraint got split at vertices lying on it,
// which a whole-segment check would not.
func assertNoEdgeCrossesConstraint(t *testing.T, mesh modeling.Mesh, boundary []vector2.Float64) {
	t.Helper()

	verts := mesh.Float3Attribute(modeling.PositionAttribute)
	idx := mesh.Indices()
	for i := 0; i < idx.Len(); i += 3 {
		a, b, c := idx.At(i), idx.At(i+1), idx.At(i+2)
		for _, e := range [][2]int{{a, b}, {b, c}, {c, a}} {
			p1, p2 := flat(verts.At(e[0])), flat(verts.At(e[1]))
			for j := range boundary {
				p3, p4 := boundary[j], boundary[(j+1)%len(boundary)]
				require.Falsef(t, cross2(p1, p2, p3, p4),
					"edge %v-%v crosses constraint %v-%v", p1, p2, p3, p4)
			}
		}
	}
}

func assertDelaunayExceptConstraints(t *testing.T, mesh modeling.Mesh, boundary []vector2.Float64) {
	t.Helper()

	verts := mesh.Float3Attribute(modeling.PositionAttribute)
	idx := mesh.Indices()
	onBoundary := func(p vector2.Float64) bool {
		for j := range boundary {
			a, b := boundary[j], boundary[(j+1)%len(boundary)]
			ab := b.Sub(a)
			if ab.LengthSquared() == 0 {
				continue
			}
			s := p.Sub(a).Dot(ab) / ab.LengthSquared()
			if s < 0 || s > 1 {
				continue
			}
			if p.Sub(a.Add(ab.Scale(s))).Length() < 1e-9 {
				return true
			}
		}
		return false
	}

	violations := 0
	for i := 0; i < idx.Len(); i += 3 {
		tri := triangulation.Triangle{idx.At(i), idx.At(i + 1), idx.At(i + 2)}
		pts := make([]vector2.Float64, verts.Len())
		for k := 0; k < verts.Len(); k++ {
			pts[k] = flat(verts.At(k))
		}
		for k := 0; k < verts.Len(); k++ {
			if k == tri[0] || k == tri[1] || k == tri[2] {
				continue
			}
			if !tri.InsideCircumcircle(pts[k], pts) {
				continue
			}
			if onBoundary(pts[k]) {
				continue
			}
			violations++
		}
	}
	assert.Zero(t, violations, "interior points found inside a triangle's circumcircle")
}

func lShape() []vector2.Float64 {
	return []vector2.Float64{
		vector2.New(0., 0.), vector2.New(2., 0.), vector2.New(2., 1.),
		vector2.New(1., 1.), vector2.New(1., 2.), vector2.New(0., 2.),
	}
}

func grid(n int) []vector2.Float64 {
	out := []vector2.Float64{}
	for x := 0; x <= n; x++ {
		for y := 0; y <= n; y++ {
			out = append(out, vector2.New(float64(x), float64(y)))
		}
	}
	return out
}

func TestConstrainedDelaunayKeepsOnlyTheConstrainedRegion(t *testing.T) {
	boundary := lShape()
	pts := append(append([]vector2.Float64{}, boundary...),
		vector2.New(0.5, 0.5), vector2.New(1.5, 0.5), vector2.New(0.5, 1.5))

	mesh, err := triangulation.ConstrainedDelaunay(pts,
		[]triangulation.Constraint{triangulation.NewConstraint(boundary)})

	require.NoError(t, err)
	assert.InDelta(t, 3, area(mesh), 1e-9, "the notch should be excluded")
	assertNoEdgeCrossesConstraint(t, mesh, boundary)
}

func TestConstrainedDelaunayAddsBoundaryPointsItWasNotGiven(t *testing.T) {
	boundary := lShape()
	interior := []vector2.Float64{
		vector2.New(0.5, 0.5), vector2.New(1.5, 0.5), vector2.New(0.5, 1.5),
	}

	mesh, err := triangulation.ConstrainedDelaunay(interior,
		[]triangulation.Constraint{triangulation.NewConstraint(boundary)})

	require.NoError(t, err)
	assert.InDelta(t, 3, area(mesh), 1e-9)
	assertNoEdgeCrossesConstraint(t, mesh, boundary)
}

func TestConstrainedDelaunayHandlesConstraintsThroughVertices(t *testing.T) {
	square := []vector2.Float64{
		vector2.New(0., 0.), vector2.New(4., 0.), vector2.New(4., 4.), vector2.New(0., 4.),
	}

	mesh, err := triangulation.ConstrainedDelaunay(grid(4),
		[]triangulation.Constraint{triangulation.NewConstraint(square)})

	require.NoError(t, err)
	assert.InDelta(t, 16, area(mesh), 1e-9)
	assertNoEdgeCrossesConstraint(t, mesh, square)
}

func TestConstrainedDelaunayCutsAcrossAGrid(t *testing.T) {
	shape := []vector2.Float64{
		vector2.New(0., 0.), vector2.New(4., 1.), vector2.New(4., 4.), vector2.New(0., 4.),
	}

	mesh, err := triangulation.ConstrainedDelaunay(grid(4),
		[]triangulation.Constraint{triangulation.NewConstraint(shape)})

	require.NoError(t, err)
	assert.InDelta(t, 14, area(mesh), 1e-9)
	assertNoEdgeCrossesConstraint(t, mesh, shape)
}

func TestConstrainedDelaunayStaysDelaunayAwayFromConstraints(t *testing.T) {
	square := []vector2.Float64{
		vector2.New(0., 0.), vector2.New(4., 0.), vector2.New(4., 4.), vector2.New(0., 4.),
	}
	pts := append(grid(4),
		vector2.New(1.3, 2.7), vector2.New(2.6, 1.4), vector2.New(3.1, 3.3))

	mesh, err := triangulation.ConstrainedDelaunay(pts,
		[]triangulation.Constraint{triangulation.NewConstraint(square)})

	require.NoError(t, err)
	assertDelaunayExceptConstraints(t, mesh, square)
}

func TestConstrainedDelaunayEveryInputPointSurvives(t *testing.T) {
	boundary := lShape()
	pts := append(append([]vector2.Float64{}, boundary...),
		vector2.New(0.5, 0.5), vector2.New(1.5, 0.5), vector2.New(0.5, 1.5))

	mesh, err := triangulation.ConstrainedDelaunay(pts,
		[]triangulation.Constraint{triangulation.NewConstraint(boundary)})
	require.NoError(t, err)

	verts := mesh.Float3Attribute(modeling.PositionAttribute)
	for _, p := range pts {
		found := false
		for i := 0; i < verts.Len(); i++ {
			if flat(verts.At(i)).Sub(p).Length() < 1e-9 {
				found = true
				break
			}
		}
		assert.Truef(t, found, "input point %v is missing from the result", p)
	}
}

func TestConstrainedDelaunayRejectsTooFewPoints(t *testing.T) {
	_, err := triangulation.ConstrainedDelaunay(
		[]vector2.Float64{vector2.New(0., 0.), vector2.New(1., 0.)}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least 3")
}

func TestConstrainedDelaunayProducesNoDegenerateTriangles(t *testing.T) {
	square := []vector2.Float64{
		vector2.New(0., 0.), vector2.New(4., 0.), vector2.New(4., 4.), vector2.New(0., 4.),
	}

	mesh, err := triangulation.ConstrainedDelaunay(grid(4),
		[]triangulation.Constraint{triangulation.NewConstraint(square)})
	require.NoError(t, err)

	verts := mesh.Float3Attribute(modeling.PositionAttribute)
	idx := mesh.Indices()
	require.Greater(t, idx.Len(), 0)
	for i := 0; i < idx.Len(); i += 3 {
		a, b, c := verts.At(idx.At(i)), verts.At(idx.At(i+1)), verts.At(idx.At(i+2))
		cross := (b.X()-a.X())*(c.Z()-a.Z()) - (c.X()-a.X())*(b.Z()-a.Z())
		assert.Greaterf(t, math.Abs(cross), 1e-12, "degenerate triangle at %d", i/3)
	}
}

func TestConstrainedDelaunaySplitsCrossingConstraints(t *testing.T) {
	a := []vector2.Float64{
		vector2.New(0., 0.), vector2.New(2., 0.), vector2.New(2., 2.), vector2.New(0., 2.),
	}
	b := []vector2.Float64{
		vector2.New(1., 1.), vector2.New(3., 1.), vector2.New(3., 3.), vector2.New(1., 3.),
	}

	mesh, err := triangulation.ConstrainedDelaunay(
		append(append([]vector2.Float64{}, a...), b...),
		[]triangulation.Constraint{triangulation.NewConstraint(a), triangulation.NewConstraint(b)})

	require.NoError(t, err)
	assert.InDelta(t, 7, area(mesh), 1e-9, "two overlapping 2x2 squares cover 7")
	assertNoEdgeCrossesConstraint(t, mesh, a)
	assertNoEdgeCrossesConstraint(t, mesh, b)
}

func TestConstrainedDelaunayHandlesASelfIntersectingOutline(t *testing.T) {
	bowtie := []vector2.Float64{
		vector2.New(0., 0.), vector2.New(2., 2.), vector2.New(2., 0.), vector2.New(0., 2.),
	}

	mesh, err := triangulation.ConstrainedDelaunay(bowtie,
		[]triangulation.Constraint{triangulation.NewConstraint(bowtie)})

	require.NoError(t, err)
	assert.InDelta(t, 2, area(mesh), 1e-9, "the two lobes of the bowtie")
	assertNoEdgeCrossesConstraint(t, mesh, bowtie)
}

func TestConstrainedDelaunayHandlesATJunction(t *testing.T) {
	outer := []vector2.Float64{
		vector2.New(0., 0.), vector2.New(4., 0.), vector2.New(4., 4.), vector2.New(0., 4.),
	}
	inner := []vector2.Float64{
		vector2.New(2., 0.), vector2.New(3., 2.), vector2.New(1., 2.),
	}

	mesh, err := triangulation.ConstrainedDelaunay(
		append(append([]vector2.Float64{}, outer...), inner...),
		[]triangulation.Constraint{
			triangulation.NewConstraint(outer), triangulation.NewConstraint(inner)})

	require.NoError(t, err)
	assert.InDelta(t, 16, area(mesh), 1e-9)
	assertNoEdgeCrossesConstraint(t, mesh, outer)
	assertNoEdgeCrossesConstraint(t, mesh, inner)
}

func TestConstrainedDelaunayHandlesTouchingButNotCrossingShapes(t *testing.T) {
	left := []vector2.Float64{
		vector2.New(0., 0.), vector2.New(2., 0.), vector2.New(2., 2.), vector2.New(0., 2.),
	}
	right := []vector2.Float64{
		vector2.New(2., 0.), vector2.New(4., 0.), vector2.New(4., 2.), vector2.New(2., 2.),
	}

	mesh, err := triangulation.ConstrainedDelaunay(
		append(append([]vector2.Float64{}, left...), right...),
		[]triangulation.Constraint{
			triangulation.NewConstraint(left), triangulation.NewConstraint(right)})

	require.NoError(t, err)
	assert.InDelta(t, 8, area(mesh), 1e-9, "two squares sharing an edge")
}

func TestConstrainedDelaunaySurvivesRandomOutlines(t *testing.T) {
	rng := rand.New(rand.NewSource(1))

	for trial := 0; trial < 400; trial++ {
		n := 4 + rng.Intn(10)
		poly := make([]vector2.Float64, n)
		for i := range poly {
			angle := 2 * math.Pi * float64(i) / float64(n)
			radius := 0.5 + rng.Float64()*1.5
			poly[i] = vector2.New(math.Cos(angle)*radius, math.Sin(angle)*radius)
		}

		pts := append([]vector2.Float64{}, poly...)
		for i := 0; i < rng.Intn(25); i++ {
			pts = append(pts, vector2.New(rng.Float64()*4-2, rng.Float64()*4-2))
		}

		_, err := triangulation.ConstrainedDelaunay(pts,
			[]triangulation.Constraint{triangulation.NewConstraint(poly)})
		require.NoErrorf(t, err, "trial %d with %d boundary points", trial, n)
	}
}

func TestConstrainedDelaunaySurvivesRandomSelfIntersectingOutlines(t *testing.T) {
	rng := rand.New(rand.NewSource(2))

	for trial := 0; trial < 300; trial++ {
		n := 5 + rng.Intn(8)
		poly := make([]vector2.Float64, n)
		for i := range poly {
			poly[i] = vector2.New(rng.Float64()*4-2, rng.Float64()*4-2)
		}

		pts := append([]vector2.Float64{}, poly...)
		for i := 0; i < rng.Intn(20); i++ {
			pts = append(pts, vector2.New(rng.Float64()*4-2, rng.Float64()*4-2))
		}

		_, err := triangulation.ConstrainedDelaunay(pts,
			[]triangulation.Constraint{triangulation.NewConstraint(poly)})
		require.NoErrorf(t, err, "trial %d with %d boundary points", trial, n)
	}
}
