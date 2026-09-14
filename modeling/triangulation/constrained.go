package triangulation

import (
	"fmt"
	"math"
	"sort"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/predicate"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type edgeKey [2]int

func newEdgeKey(a, b int) edgeKey {
	if a > b {
		a, b = b, a
	}
	return edgeKey{a, b}
}

// winding counts constraint segments along an edge, +1 for each running from
// the lower index to the higher and -1 for the reverse.
type tessellation struct {
	pts         []vector2.Float64
	tris        map[Triangle]struct{}
	adj         map[edgeKey][]Triangle
	incident    map[int][]Triangle
	constrained map[edgeKey]struct{}
	winding     map[edgeKey]int
}

// Which rotation of a triangle gets stored decides which of its neighbours
// opposites reports first, and that leaks all the way out as the order faces
// come back in, so tris has to arrive in a deterministic order.
func newTessellation(pts []vector2.Float64, tris []Triangle) *tessellation {
	t := &tessellation{
		pts:         pts,
		tris:        make(map[Triangle]struct{}, len(tris)),
		adj:         make(map[edgeKey][]Triangle, len(tris)*3),
		incident:    make(map[int][]Triangle, len(pts)),
		constrained: make(map[edgeKey]struct{}),
		winding:     make(map[edgeKey]int),
	}
	for _, tri := range tris {
		t.add(tri)
	}
	return t
}

func (t *tessellation) add(tri Triangle) {
	t.tris[tri] = exists
	for _, e := range tri.Edges() {
		k := newEdgeKey(e[0], e[1])
		t.adj[k] = append(t.adj[k], tri)
	}
	for _, v := range tri {
		t.incident[v] = append(t.incident[v], tri)
	}
}

func without(tris []Triangle, tri Triangle) []Triangle {
	kept := tris[:0]
	for _, other := range tris {
		if other != tri {
			kept = append(kept, other)
		}
	}
	return kept
}

func (t *tessellation) remove(tri Triangle) {
	delete(t.tris, tri)
	for _, e := range tri.Edges() {
		k := newEdgeKey(e[0], e[1])
		if kept := without(t.adj[k], tri); len(kept) == 0 {
			delete(t.adj, k)
		} else {
			t.adj[k] = kept
		}
	}
	for _, v := range tri {
		t.incident[v] = without(t.incident[v], tri)
	}
}

func (t *tessellation) third(tri Triangle, a, b int) int {
	for _, v := range tri {
		if v != a && v != b {
			return v
		}
	}
	return -1
}

func (t *tessellation) across(e edgeKey, from Triangle) (Triangle, bool) {
	for _, tri := range t.adj[e] {
		if tri != from {
			return tri, true
		}
	}
	return Triangle{}, false
}

func (t *tessellation) hasEdge(a, b int) bool {
	_, ok := t.adj[newEdgeKey(a, b)]
	return ok
}

func (t *tessellation) opposites(e edgeKey) (Triangle, Triangle, int, int, bool) {
	tris := t.adj[e]
	if len(tris) != 2 {
		return Triangle{}, Triangle{}, 0, 0, false
	}
	p, q := t.third(tris[0], e[0], e[1]), t.third(tris[1], e[0], e[1])
	if p < 0 || q < 0 {
		return Triangle{}, Triangle{}, 0, 0, false
	}
	return tris[0], tris[1], p, q, true
}

func (t *tessellation) wind(a, b, c int) Triangle {
	return Triangle{a, b, c}.clockwise(t.pts)
}

func (t *tessellation) flip(e edgeKey) (edgeKey, bool) {
	t1, t2, p, q, ok := t.opposites(e)
	if !ok {
		return edgeKey{}, false
	}
	if predicate.Orient2D(t.pts[p], t.pts[q], t.pts[e[0]]) == 0 ||
		predicate.Orient2D(t.pts[p], t.pts[q], t.pts[e[1]]) == 0 {
		return edgeKey{}, false
	}
	if (predicate.Orient2D(t.pts[p], t.pts[q], t.pts[e[0]]) > 0) ==
		(predicate.Orient2D(t.pts[p], t.pts[q], t.pts[e[1]]) > 0) {
		return edgeKey{}, false
	}

	t.remove(t1)
	t.remove(t2)
	t.add(t.wind(p, q, e[0]))
	t.add(t.wind(p, q, e[1]))
	return newEdgeKey(p, q), true
}

// Cocircular points are a genuine tie that leaves more than one correct
// Delaunay triangulation, so which one comes out is decided by the order
// edges are visited. Ranging a map would make that the runtime's choice and
// the same input would not give the same mesh twice.
func (t *tessellation) edges() []edgeKey {
	out := make([]edgeKey, 0, len(t.adj))
	for e := range t.adj {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i][0] != out[j][0] {
			return out[i][0] < out[j][0]
		}
		return out[i][1] < out[j][1]
	})
	return out
}

// Walks triangle to triangle along a-b, so the edges come back in the order
// the segment meets them, which is the order forceEdge wants to flip them.
func (t *tessellation) crossing(a, b int) []edgeKey {
	pa, pb := t.pts[a], t.pts[b]

	var tri Triangle
	u, v := -1, -1
	for _, candidate := range t.incident[a] {
		var rest [2]int
		n := 0
		for _, vertex := range candidate {
			if vertex != a {
				rest[n] = vertex
				n++
			}
		}
		if predicate.SegmentsCross(pa, pb, t.pts[rest[0]], t.pts[rest[1]]) {
			tri, u, v = candidate, rest[0], rest[1]
			break
		}
	}
	if u < 0 {
		return nil
	}

	out := []edgeKey{newEdgeKey(u, v)}
	for {
		next, ok := t.across(newEdgeKey(u, v), tri)
		if !ok {
			return out
		}
		w := t.third(next, u, v)
		if w == b {
			return out
		}
		switch {
		case predicate.SegmentsCross(pa, pb, t.pts[u], t.pts[w]):
			v = w
		case predicate.SegmentsCross(pa, pb, t.pts[w], t.pts[v]):
			u = w
		default:
			return out
		}
		tri = next
		out = append(out, newEdgeKey(u, v))
	}
}

func (t *tessellation) lock(a, b int) {
	k := newEdgeKey(a, b)
	t.constrained[k] = exists
	if a < b {
		t.winding[k]++
	} else {
		t.winding[k]--
	}
}

func (t *tessellation) forceEdge(a, b int) error {
	if a == b {
		return nil
	}
	if t.hasEdge(a, b) {
		t.lock(a, b)
		return nil
	}

	// Recomputed each round, not hoisted: flipping one edge reshapes its
	// neighbours, so an upfront list empties with the edge still missing.
	budget := len(t.pts)*len(t.pts) + 64

	for !t.hasEdge(a, b) {
		budget--
		if budget < 0 {
			return fmt.Errorf("could not force edge %d-%d after exhausting the flip budget", a, b)
		}

		blocking := t.crossing(a, b)
		if len(blocking) == 0 {
			return fmt.Errorf("could not force edge %d-%d; nothing crosses it yet it is absent", a, b)
		}

		flippedAny := false
		for _, e := range blocking {
			if _, locked := t.constrained[e]; locked {
				return fmt.Errorf("constraint edge %d-%d crosses another constraint edge", a, b)
			}
			if _, ok := t.flip(e); ok {
				flippedAny = true
			}
		}

		if !flippedAny {
			return fmt.Errorf("could not force edge %d-%d; every crossing edge sits in a concave quad", a, b)
		}
	}

	t.lock(a, b)
	return nil
}

// Lawson's flip queue: every edge is checked once, and a flip only puts the
// four edges around it back on the queue.
func (t *tessellation) restoreDelaunay() {
	queue := t.edges()
	queued := make(map[edgeKey]struct{}, len(queue))
	for _, e := range queue {
		queued[e] = exists
	}

	budget := len(t.tris)*len(t.tris) + 64
	for len(queue) > 0 && budget > 0 {
		e := queue[0]
		queue = queue[1:]
		delete(queued, e)

		if _, locked := t.constrained[e]; locked {
			continue
		}
		_, _, p, q, ok := t.opposites(e)
		if !ok {
			continue
		}
		if !t.wind(e[0], e[1], p).InsideCircumcircle(t.pts[q], t.pts) {
			continue
		}
		if _, ok := t.flip(e); !ok {
			continue
		}
		budget--

		for _, around := range [4]edgeKey{
			newEdgeKey(e[0], p), newEdgeKey(e[1], p),
			newEdgeKey(e[0], q), newEdgeKey(e[1], q),
		} {
			if _, already := queued[around]; already {
				continue
			}
			queued[around] = exists
			queue = append(queue, around)
		}
	}
}

// Change in winding number on entering tri across e.
func (t *tessellation) step(e edgeKey, tri Triangle) int {
	w := t.winding[e]
	if w == 0 {
		return 0
	}
	third := t.third(tri, e[0], e[1])
	if predicate.Orient2D(t.pts[e[0]], t.pts[e[1]], t.pts[third]) > 0 {
		return w
	}
	return -w
}

// Winding numbers flood in from the hull, where they are zero, and only
// change across constraint edges. Nonzero is kept, so overlapping outlines
// union and an outline wound against the one around it is a hole.
func (t *tessellation) discardOutside() {
	winding := make(map[Triangle]int, len(t.tris))
	queue := make([]Triangle, 0, len(t.tris))
	visit := func(tri Triangle, w int) {
		if _, seen := winding[tri]; seen {
			return
		}
		winding[tri] = w
		queue = append(queue, tri)
	}

	for e, tris := range t.adj {
		if len(tris) == 1 {
			visit(tris[0], t.step(e, tris[0]))
		}
	}

	for len(queue) > 0 {
		tri := queue[0]
		queue = queue[1:]

		for _, edge := range tri.Edges() {
			k := newEdgeKey(edge[0], edge[1])
			if neighbor, ok := t.across(k, tri); ok {
				visit(neighbor, winding[tri]+t.step(k, neighbor))
			}
		}
	}

	for tri, w := range winding {
		if w == 0 {
			t.remove(tri)
		}
	}
}

func (t *tessellation) mesh() modeling.Mesh {
	ordered := make([]Triangle, 0, len(t.tris))
	for tri := range t.tris {
		ordered = append(ordered, tri.clockwise(t.pts))
	}
	sort.Slice(ordered, func(i, j int) bool {
		for k := 0; k < 3; k++ {
			if ordered[i][k] != ordered[j][k] {
				return ordered[i][k] < ordered[j][k]
			}
		}
		return false
	})

	tris := make([]int, 0, len(ordered)*3)
	for _, tri := range ordered {
		tris = append(tris, tri[0], tri[1], tri[2])
	}

	verts := make([]vector3.Float64, len(t.pts))
	uvs := make([]vector2.Float64, len(t.pts))
	for i, p := range t.pts {
		verts[i] = vector3.New(p.X(), 0, p.Y())
		uvs[i] = vector2.Zero[float64]()
	}

	return modeling.NewTriangleMesh(tris).
		SetFloat3Attribute(modeling.PositionAttribute, verts).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs)
}

func segmentIntersection(p1, p2, p3, p4 vector2.Float64) (vector2.Float64, bool) {
	if !predicate.SegmentsCross(p1, p2, p3, p4) {
		return vector2.Zero[float64](), false
	}
	d1 := predicate.Orient2D(p3, p4, p1)
	d2 := predicate.Orient2D(p3, p4, p2)
	if d1 == d2 {
		return vector2.Zero[float64](), false
	}
	return p1.Add(p2.Sub(p1).Scale(d1 / (d1 - d2))), true
}

// No triangulation holds two crossing constraints whole, so the crossing
// becomes a vertex and splitAtVertices breaks both on it.
func splitCrossingConstraints(pts *[]vector2.Float64, segments [][2]int, tolerance float64) {
	for i := 0; i < len(segments); i++ {
		for j := i + 1; j < len(segments); j++ {
			a, b := segments[i], segments[j]
			if a[0] == b[0] || a[0] == b[1] || a[1] == b[0] || a[1] == b[1] {
				continue
			}
			p, ok := segmentIntersection(
				(*pts)[a[0]], (*pts)[a[1]], (*pts)[b[0]], (*pts)[b[1]])
			if !ok {
				continue
			}
			indexOfPoint(pts, p, tolerance)
		}
	}
}

// A segment running through an existing vertex has nothing properly
// crossing it, so no sequence of flips can produce it whole.
func splitAtVertices(pts []vector2.Float64, a, b int) []int {
	pa, pb := pts[a], pts[b]
	ab := pb.Sub(pa)
	lengthSq := ab.LengthSquared()
	if lengthSq == 0 {
		return []int{a, b}
	}
	tolerance := 1e-9 * math.Sqrt(lengthSq)

	type hit struct {
		index int
		t     float64
	}
	hits := []hit{}
	for i, p := range pts {
		if i == a || i == b {
			continue
		}
		if math.Abs(predicate.Orient2D(pa, pb, p)) > tolerance*math.Sqrt(lengthSq) {
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

func indexOfPoint(pts *[]vector2.Float64, p vector2.Float64, tolerance float64) int {
	for i, existing := range *pts {
		if existing.Sub(p).Length() < tolerance {
			return i
		}
	}
	*pts = append(*pts, p)
	return len(*pts) - 1
}

// Distance under which two points count as the same, relative to the
// extent of everything being triangulated.
func mergeTolerance(sets ...[]vector2.Float64) float64 {
	min := vector2.New(math.Inf(1), math.Inf(1))
	max := vector2.New(math.Inf(-1), math.Inf(-1))
	for _, set := range sets {
		for _, p := range set {
			min = vector2.New(math.Min(p.X(), min.X()), math.Min(p.Y(), min.Y()))
			max = vector2.New(math.Max(p.X(), max.X()), math.Max(p.Y(), max.Y()))
		}
	}
	extent := math.Max(max.X()-min.X(), max.Y()-min.Y())
	if extent <= 0 || math.IsInf(extent, 0) || math.IsNaN(extent) {
		return 0
	}
	return 1e-9 * extent
}

// ConstrainedDelaunay triangulates points so every constraint edge survives,
// then keeps only the triangles with a nonzero winding number: overlapping
// outlines union, and an outline wound against the one around it is a hole.
func ConstrainedDelaunay(points []vector2.Float64, constraints []Constraint) (modeling.Mesh, error) {
	pts := make([]vector2.Float64, len(points))
	copy(pts, points)

	sets := make([][]vector2.Float64, 0, len(constraints)+1)
	sets = append(sets, points)
	for _, c := range constraints {
		sets = append(sets, c.shape)
	}
	tolerance := mergeTolerance(sets...)

	segments := make([][2]int, 0)
	for _, c := range constraints {
		for i := range c.shape {
			a := indexOfPoint(&pts, c.shape[i], tolerance)
			b := indexOfPoint(&pts, c.shape[(i+1)%len(c.shape)], tolerance)
			if a != b {
				segments = append(segments, [2]int{a, b})
			}
		}
	}

	return triangulate(pts, segments, len(segments) > 0)
}

// ConstrainedDelaunayEdges triangulates the convex hull of points, forcing
// every listed edge to survive. The edges are open segments rather than a
// closed boundary, so nothing is discarded and the result covers the hull.
func ConstrainedDelaunayEdges(points []vector2.Float64, edges [][2]int) (modeling.Mesh, error) {
	pts := make([]vector2.Float64, len(points))
	copy(pts, points)

	segments := make([][2]int, 0, len(edges))
	for _, e := range edges {
		if e[0] < 0 || e[0] >= len(pts) || e[1] < 0 || e[1] >= len(pts) {
			return modeling.EmptyMesh(modeling.TriangleTopology),
				fmt.Errorf("edge %v refers to a point outside the set of %d", e, len(pts))
		}
		if e[0] != e[1] {
			segments = append(segments, e)
		}
	}

	return triangulate(pts, segments, false)
}

func triangulate(pts []vector2.Float64, segments [][2]int, discard bool) (modeling.Mesh, error) {
	splitCrossingConstraints(&pts, segments, mergeTolerance(pts))

	if len(pts) < 3 {
		return modeling.EmptyMesh(modeling.TriangleTopology),
			fmt.Errorf("need at least 3 distinct points, got %d", len(pts))
	}

	tess := newTessellation(pts, bowyerWatson(pts).triangles())

	for _, s := range segments {
		chain := splitAtVertices(pts, s[0], s[1])
		for i := 0; i < len(chain)-1; i++ {
			if err := tess.forceEdge(chain[i], chain[i+1]); err != nil {
				return modeling.EmptyMesh(modeling.TriangleTopology), err
			}
		}
	}

	tess.restoreDelaunay()

	if discard {
		tess.discardOutside()
	}

	return tess.mesh(), nil
}
