package triangulation

import (
	"fmt"
	"math"
	"sort"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
)

// No triangulation holds two crossing constraints whole, so the crossing
// becomes a vertex and splitAtVertices breaks both on it.
func splitCrossingConstraints(weld *geometry.PointWelder2D, segments [][2]int) {
	pts := weld.Points()
	lo := make([]vector2.Float64, len(segments))
	hi := make([]vector2.Float64, len(segments))
	order := make([]int, len(segments))
	for i, s := range segments {
		a, b := pts[s[0]], pts[s[1]]
		lo[i] = vector2.New(min(a.X(), b.X()), min(a.Y(), b.Y()))
		hi[i] = vector2.New(max(a.X(), b.X()), max(a.Y(), b.Y()))
		order[i] = i
	}
	sort.Slice(order, func(x, y int) bool { return lo[order[x]].X() < lo[order[y]].X() })

	for n, i := range order {
		for _, j := range order[n+1:] {
			if lo[j].X() > hi[i].X() {
				break
			}
			if lo[j].Y() > hi[i].Y() || hi[j].Y() < lo[i].Y() {
				continue
			}
			a, b := segments[i], segments[j]
			if a[0] == b[0] || a[0] == b[1] || a[1] == b[0] || a[1] == b[1] {
				continue
			}
			p, err := geometry.NewLine2D(pts[a[0]], pts[a[1]]).
				Intersection(geometry.NewLine2D(pts[b[0]], pts[b[1]]))
			if err != nil {
				continue
			}
			weld.Index(p)
		}
	}
}

type pointsByX struct {
	pts   []vector2.Float64
	order []int
}

func newPointsByX(pts []vector2.Float64) pointsByX {
	order := make([]int, len(pts))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(a, b int) bool { return pts[order[a]].X() < pts[order[b]].X() })
	return pointsByX{pts: pts, order: order}
}

func (p pointsByX) between(minX, maxX float64) []int {
	from := sort.Search(len(p.order), func(i int) bool { return p.pts[p.order[i]].X() >= minX })
	to := sort.Search(len(p.order), func(i int) bool { return p.pts[p.order[i]].X() > maxX })
	return p.order[from:to]
}

// No flips produce a segment through a vertex; never below the weld tolerance, or welded crossings don't split.
func splitAtVertices(byX pointsByX, a, b int, weldTolerance float64) []int {
	pts := byX.pts
	pa, pb := pts[a], pts[b]
	ab := pb.Sub(pa)
	lengthSq := ab.LengthSquared()
	if lengthSq == 0 {
		return []int{a, b}
	}
	tolerance := max(weldTolerance, 1e-9*math.Sqrt(lengthSq))

	type hit struct {
		index int
		t     float64
	}
	hits := []hit{}
	for _, i := range byX.between(min(pa.X(), pb.X())-tolerance, max(pa.X(), pb.X())+tolerance) {
		p := pts[i]
		if i == a || i == b {
			continue
		}
		t := p.Sub(pa).Dot(ab) / lengthSq
		if t <= 0 || t >= 1 {
			continue
		}
		if p.Sub(pa.Add(ab.Scale(t))).Length() > tolerance {
			continue
		}
		hits = append(hits, hit{i, t})
	}

	sort.Slice(hits, func(i, j int) bool { return hits[i].t < hits[j].t })

	out := make([]int, 0, len(hits)+2)
	out = append(out, a)
	for _, h := range hits {
		out = append(out, h.index)
	}
	return append(out, b)
}

// Must stay well above csg's weld tolerance, or a solid built from the result reads as open.
func mergeTolerance(sets ...[]vector2.Float64) float64 {
	min := vector2.New(math.Inf(1), math.Inf(1))
	max := vector2.New(math.Inf(-1), math.Inf(-1))
	for _, set := range sets {
		lo, hi := geometry.Shape(set).GetBounds()
		min = vector2.New(math.Min(lo.X(), min.X()), math.Min(lo.Y(), min.Y()))
		max = vector2.New(math.Max(hi.X(), max.X()), math.Max(hi.Y(), max.Y()))
	}
	extent := math.Max(max.X()-min.X(), max.Y()-min.Y())
	if extent <= 0 || math.IsInf(extent, 0) || math.IsNaN(extent) {
		return 0
	}
	return 1e-8 * extent
}

// Input points keep their indices even when two coincide; only the points
// added later are merged into them.
func welderOver(points []vector2.Float64, tolerance float64) *geometry.PointWelder2D {
	weld := geometry.NewPointWelder2D(tolerance)
	for _, p := range points {
		weld.Add(p)
	}
	return weld
}

// ConstrainedDelaunay triangulates points so every constraint edge survives,
// then keeps only the triangles with a nonzero winding number: overlapping
// outlines union, and an outline wound against the one around it is a hole.
func ConstrainedDelaunay(points []vector2.Float64, constraints []Constraint) (modeling.Mesh, error) {
	sets := make([][]vector2.Float64, 0, len(constraints)+1)
	sets = append(sets, points)
	for _, c := range constraints {
		sets = append(sets, c.shape)
	}
	weld := welderOver(points, mergeTolerance(sets...))

	segments := make([][2]int, 0)
	for _, c := range constraints {
		for i := range c.shape {
			a := weld.Index(c.shape[i])
			b := weld.Index(c.shape[(i+1)%len(c.shape)])
			if a != b {
				segments = append(segments, [2]int{a, b})
			}
		}
	}

	tess, err := triangulate(weld, segments, nil, 1)
	if err != nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), err
	}
	if len(segments) == 0 {
		return tess.mesh(nil), nil
	}
	windings := tess.windings()
	return tess.mesh(func(tri int) bool { return windings[tri] != 0 }), nil
}

// ConstrainedDelaunayEdges triangulates the convex hull of points, forcing
// every listed edge to survive. The edges are open segments rather than a
// closed boundary, so nothing is discarded and the result covers the hull.
func ConstrainedDelaunayEdges(points []vector2.Float64, edges [][2]int) (modeling.Mesh, error) {
	segments := make([][2]int, 0, len(edges))
	for _, e := range edges {
		if e[0] < 0 || e[0] >= len(points) || e[1] < 0 || e[1] >= len(points) {
			return modeling.EmptyMesh(modeling.TriangleTopology),
				fmt.Errorf("edge %v refers to a point outside the set of %d", e, len(points))
		}
		if e[0] != e[1] {
			segments = append(segments, e)
		}
	}

	tess, err := triangulate(welderOver(points, mergeTolerance(points)), segments, nil, 1)
	if err != nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), err
	}
	return tess.mesh(nil), nil
}

// groupOf[i] is segments[i]'s winding group; nil puts every segment in group 0.
func triangulate(weld *geometry.PointWelder2D, segments [][2]int, groupOf []int, groups int) (*tessellation, error) {
	splitCrossingConstraints(weld, segments)
	pts := weld.Points()

	if len(pts) < 3 {
		return nil, fmt.Errorf("need at least 3 distinct points, got %d", len(pts))
	}

	tess := newTessellation(pts, bowyerWatson(pts), groups)

	byX := newPointsByX(pts)
	for si, s := range segments {
		group := 0
		if groupOf != nil {
			group = groupOf[si]
		}
		chain := splitAtVertices(byX, s[0], s[1], weld.Tolerance())
		for i := 0; i < len(chain)-1; i++ {
			if err := tess.forceEdge(chain[i], chain[i+1], group); err != nil {
				return nil, err
			}
		}
	}

	tess.restoreDelaunay()
	return tess, nil
}
