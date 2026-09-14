package triangulation_test

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func turn(a, b, c vector2.Float64) float64 {
	return (b.X()-a.X())*(c.Y()-a.Y()) - (c.X()-a.X())*(b.Y()-a.Y())
}

func convexHullArea(pts []vector2.Float64) float64 {
	p := append([]vector2.Float64{}, pts...)
	sort.Slice(p, func(i, j int) bool {
		if p[i].X() != p[j].X() {
			return p[i].X() < p[j].X()
		}
		return p[i].Y() < p[j].Y()
	})

	build := func(in []vector2.Float64) []vector2.Float64 {
		out := []vector2.Float64{}
		for _, q := range in {
			for len(out) >= 2 && turn(out[len(out)-2], out[len(out)-1], q) <= 0 {
				out = out[:len(out)-1]
			}
			out = append(out, q)
		}
		return out
	}

	lower := build(p)
	reversed := append([]vector2.Float64{}, p...)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	upper := build(reversed)

	hull := append(lower[:len(lower)-1], upper[:len(upper)-1]...)
	total := 0.
	for i := range hull {
		j := (i + 1) % len(hull)
		total += hull[i].X()*hull[j].Y() - hull[j].X()*hull[i].Y()
	}
	return math.Abs(total) / 2
}

// A Delaunay triangulation tiles the convex hull of its points exactly. A
// super triangle too small to contain them leaves slivers missing against
// that hull instead, which reads as a hole a constraint can slip through.
func TestBowyerWatsonCoversTheConvexHull(t *testing.T) {
	for _, seed := range []int64{7, 11, 42, 1234} {
		rng := rand.New(rand.NewSource(seed))
		for trial := 0; trial < 200; trial++ {
			n := 5 + rng.Intn(25)
			pts := make([]vector2.Float64, n)
			for i := range pts {
				pts[i] = vector2.New(rng.Float64()*4-2, rng.Float64()*4-2)
			}

			want := convexHullArea(pts)
			got := area(triangulation.BowyerWatson(pts))
			require.InDeltaf(t, want, got, 1e-9*want+1e-12,
				"seed %d trial %d: %d points", seed, trial, n)
		}
	}
}

// Points are normalized before triangulating, so the result depends on the
// shape of the set and not on where it sits or how big it is. A set at 1e-9
// and the same set at 1e11 must come out identically.
func TestBowyerWatsonIsScaleInvariant(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	base := make([]vector2.Float64, 20)
	for i := range base {
		base[i] = vector2.New(rng.Float64(), rng.Float64())
	}

	reference := area(triangulation.BowyerWatson(base)) / convexHullArea(base)

	for _, scale := range []float64{1e-9, 1e-3, 1e3, 1e6, 1e11} {
		scaled := make([]vector2.Float64, len(base))
		for i, p := range base {
			scaled[i] = p.Scale(scale).Add(vector2.New(scale, scale))
		}

		covered := area(triangulation.BowyerWatson(scaled)) / convexHullArea(scaled)
		assert.InDeltaf(t, reference, covered, 1e-9,
			"scale %g covered a different fraction of its hull", scale)
	}
}

func TestBowyerWatsonCoversTheHullOfThinPointSets(t *testing.T) {
	for _, tc := range []struct {
		name string
		w, h float64
	}{
		{"wide and short", 100, 0.05},
		{"tall and narrow", 0.05, 100},
		{"extremely wide", 2000, 1},
		{"extremely tall", 1, 2000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(3))
			for trial := 0; trial < 100; trial++ {
				pts := make([]vector2.Float64, 12)
				for i := range pts {
					pts[i] = vector2.New(rng.Float64()*tc.w, rng.Float64()*tc.h)
				}

				want := convexHullArea(pts)
				require.InDeltaf(t, want, area(triangulation.BowyerWatson(pts)), 1e-9*want,
					"trial %d", trial)
			}
		})
	}
}

// Points are squashed to a unit square before triangulating, which is not a
// Delaunay preserving transform. Only the flip pass afterwards makes the
// result Delaunay again, and a sliver is where it would show first.
func TestBowyerWatsonIsDelaunayOnThinPointSets(t *testing.T) {
	for _, aspect := range []float64{1, 100, 2000} {
		rng := rand.New(rand.NewSource(8))
		for trial := 0; trial < 100; trial++ {
			pts := make([]vector2.Float64, 15)
			for i := range pts {
				pts[i] = vector2.New(rng.Float64()*aspect, rng.Float64())
			}

			idx := triangulation.BowyerWatson(pts).Indices()
			for i := 0; i < idx.Len(); i += 3 {
				tri := triangulation.Triangle{idx.At(i), idx.At(i + 1), idx.At(i + 2)}
				for j := range pts {
					if j == tri[0] || j == tri[1] || j == tri[2] {
						continue
					}
					require.Falsef(t, tri.InsideCircumcircle(pts[j], pts),
						"aspect %g trial %d: point %d sits inside the circumcircle of %v",
						aspect, trial, j, tri)
				}
			}
		}
	}
}

func TestSuperTriangleContainsEveryPoint(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	for _, scale := range []vector2.Float64{
		vector2.New(1., 1.), vector2.New(1000., 0.001),
		vector2.New(0.001, 1000.), vector2.New(1e7, 1e7),
	} {
		pts := make([]vector2.Float64, 30)
		for i := range pts {
			pts[i] = vector2.New(rng.Float64()*scale.X(), rng.Float64()*scale.Y())
		}
		super := triangulation.SuperTriangle(pts)
		require.Len(t, super, 3)

		for _, p := range pts {
			inside := true
			for i := range super {
				a, b := super[i], super[(i+1)%len(super)]
				if turn(a, b, p) < 0 {
					inside = false
				}
			}
			// Winding of the returned triangle is not guaranteed, so accept
			// consistently-negative too.
			if !inside {
				allNegative := true
				for i := range super {
					a, b := super[i], super[(i+1)%len(super)]
					if turn(a, b, p) > 0 {
						allNegative = false
					}
				}
				inside = allNegative
			}
			assert.Truef(t, inside, "point %v falls outside the super triangle %v", p, super)
		}
	}
}

func TestBowyerWatsonUsesEveryPoint(t *testing.T) {
	rng := rand.New(rand.NewSource(5))
	pts := make([]vector2.Float64, 40)
	for i := range pts {
		pts[i] = vector2.New(rng.Float64()*10, rng.Float64()*10)
	}

	mesh := triangulation.BowyerWatson(pts)
	idx := mesh.Indices()
	used := map[int]bool{}
	for i := 0; i < idx.Len(); i++ {
		used[idx.At(i)] = true
	}

	verts := mesh.Float3Attribute(modeling.PositionAttribute)
	require.Equal(t, len(pts), verts.Len())
	assert.Len(t, used, len(pts), "every input point should belong to a triangle")
}

func windings(t *testing.T, m modeling.Mesh) (up, down int) {
	t.Helper()
	idx := m.Indices()
	pos := m.Float3Attribute(modeling.PositionAttribute)
	for i := 0; i < idx.Len(); i += 3 {
		a, b, c := pos.At(idx.At(i)), pos.At(idx.At(i+1)), pos.At(idx.At(i+2))
		if b.Sub(a).Cross(c.Sub(a)).Y() < 0 {
			down++
		} else {
			up++
		}
	}
	return
}

func TestTriangulationWindsEveryTriangleTheSameWay(t *testing.T) {
	pts := []vector2.Float64{
		vector2.New(0., 0.),
		vector2.New(1.8964066195723626, -1.040857292644666),
		vector2.New(3.979983360114563, -0.08663011346505423),
		vector2.New(3.792813239144725, 2.0473053825681147),
		vector2.New(2.3274080354040336, 4.181240878601283),
	}

	t.Run("bowyer watson", func(t *testing.T) {
		up, down := windings(t, triangulation.BowyerWatson(pts))
		require.Positive(t, up)
		assert.Zero(t, down, "%d of %d triangles face the other way", down, up+down)
	})

	t.Run("constrained", func(t *testing.T) {
		mesh, err := triangulation.ConstrainedDelaunay(pts,
			[]triangulation.Constraint{triangulation.NewConstraint(pts)})
		require.NoError(t, err)
		up, down := windings(t, mesh)
		require.Positive(t, up)
		assert.Zero(t, down, "%d of %d triangles face the other way", down, up+down)
	})
}

func TestTriangulationWindingSurvivesRandomInput(t *testing.T) {
	rng := rand.New(rand.NewSource(4))

	for trial := 0; trial < 500; trial++ {
		n := 4 + rng.Intn(12)
		poly := make([]vector2.Float64, n)
		for i := range poly {
			angle := 2 * math.Pi * float64(i) / float64(n)
			radius := 0.5 + rng.Float64()*1.5
			poly[i] = vector2.New(math.Cos(angle)*radius, math.Sin(angle)*radius)
		}
		pts := append([]vector2.Float64{}, poly...)
		for i := 0; i < rng.Intn(20); i++ {
			pts = append(pts, vector2.New(rng.Float64()*4-2, rng.Float64()*4-2))
		}

		_, down := windings(t, triangulation.BowyerWatson(pts))
		require.Zerof(t, down, "trial %d: BowyerWatson", trial)

		mesh, err := triangulation.ConstrainedDelaunay(pts,
			[]triangulation.Constraint{triangulation.NewConstraint(poly)})
		require.NoErrorf(t, err, "trial %d", trial)
		_, down = windings(t, mesh)
		require.Zerof(t, down, "trial %d: ConstrainedDelaunay", trial)
	}
}

func TestBowyerWatsonIsDelaunayAtScale(t *testing.T) {
	rng := rand.New(rand.NewSource(9))
	pts := make([]vector2.Float64, 3000)
	for i := range pts {
		pts[i] = vector2.New(rng.Float64()*100, rng.Float64()*100)
	}

	idx := triangulation.BowyerWatson(pts).Indices()
	require.Equal(t, 3*(2*len(pts)-2-hullSize(pts)), idx.Len(), "a triangulation of the hull has 2n-2-h triangles")

	for i := 0; i < idx.Len(); i += 3 {
		tri := triangulation.Triangle{idx.At(i), idx.At(i + 1), idx.At(i + 2)}
		for _, j := range rng.Perm(len(pts))[:50] {
			if j == tri[0] || j == tri[1] || j == tri[2] {
				continue
			}
			require.Falsef(t, tri.InsideCircumcircle(pts[j], pts), "point %d inside circumcircle of %v", j, tri)
		}
	}
}

func hullSize(pts []vector2.Float64) int {
	p := append([]vector2.Float64{}, pts...)
	sort.Slice(p, func(i, j int) bool {
		if p[i].X() != p[j].X() {
			return p[i].X() < p[j].X()
		}
		return p[i].Y() < p[j].Y()
	})
	build := func(in []vector2.Float64) []vector2.Float64 {
		out := []vector2.Float64{}
		for _, q := range in {
			for len(out) >= 2 && turn(out[len(out)-2], out[len(out)-1], q) <= 0 {
				out = out[:len(out)-1]
			}
			out = append(out, q)
		}
		return out
	}
	lower := build(p)
	reversed := append([]vector2.Float64{}, p...)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return len(lower) - 1 + len(build(reversed)) - 1
}

func TestBowyerWatsonIsDeterministic(t *testing.T) {
	rng := rand.New(rand.NewSource(10))
	pts := make([]vector2.Float64, 500)
	for i := range pts {
		pts[i] = vector2.New(rng.Float64(), rng.Float64())
	}

	first := triangulation.BowyerWatson(pts).Indices()
	second := triangulation.BowyerWatson(pts).Indices()
	require.Equal(t, first.Len(), second.Len())
	for i := 0; i < first.Len(); i++ {
		require.Equal(t, first.At(i), second.At(i), "index %d", i)
	}
}

func TestBowyerWatsonLeavesDuplicatePointsUnreferenced(t *testing.T) {
	pts := []vector2.Float64{
		vector2.New(0., 0.), vector2.New(1., 0.), vector2.New(0., 1.), vector2.New(1., 1.),
		vector2.New(1., 0.), vector2.New(0.5, 0.5), vector2.New(0.5, 0.5),
	}

	var mesh modeling.Mesh
	require.NotPanics(t, func() { mesh = triangulation.BowyerWatson(pts) })

	idx := mesh.Indices()
	used := map[int]bool{}
	for i := 0; i < idx.Len(); i++ {
		used[idx.At(i)] = true
	}
	assert.False(t, used[4], "the second (1,0) is a duplicate")
	assert.False(t, used[6], "the second (0.5,0.5) is a duplicate")
	assert.True(t, used[5])
	assert.Equal(t, 4*3, idx.Len(), "a square with its centre is four triangles")
}

func TestBowyerWatsonHandlesPointsOnExistingEdges(t *testing.T) {
	pts := []vector2.Float64{}
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			pts = append(pts, vector2.New(float64(x), float64(y)))
		}
	}

	mesh := triangulation.BowyerWatson(pts)
	require.Equal(t, 4*4*2*3, mesh.Indices().Len())
	assert.InDelta(t, 16., area(mesh), 1e-9)
	_, down := windings(t, mesh)
	assert.Zero(t, down)
}

func BenchmarkBowyerWatson(b *testing.B) {
	for _, n := range []int{1_000, 10_000, 100_000} {
		rng := rand.New(rand.NewSource(1))
		pts := make([]vector2.Float64, n)
		for i := range pts {
			pts[i] = vector2.New(rng.Float64()*100, rng.Float64()*100)
		}
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				triangulation.BowyerWatson(pts)
			}
		})
	}
}
