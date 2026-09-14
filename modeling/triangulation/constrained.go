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

type tessellation struct {
	pts         []vector2.Float64
	tris        map[Triangle]struct{}
	adj         map[edgeKey][]Triangle
	constrained map[edgeKey]struct{}
}

// Which rotation of a triangle gets stored decides which of its neighbours
// opposites reports first, and that leaks all the way out as the order faces
// come back in, so tris has to arrive in a deterministic order.
func newTessellation(pts []vector2.Float64, tris []Triangle) *tessellation {
	t := &tessellation{
		pts:         pts,
		tris:        make(map[Triangle]struct{}, len(tris)),
		adj:         make(map[edgeKey][]Triangle, len(tris)*3),
		constrained: make(map[edgeKey]struct{}),
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
}

func (t *tessellation) remove(tri Triangle) {
	delete(t.tris, tri)
	for _, e := range tri.Edges() {
		k := newEdgeKey(e[0], e[1])
		kept := t.adj[k][:0]
		for _, other := range t.adj[k] {
			if other != tri {
				kept = append(kept, other)
			}
		}
		if len(kept) == 0 {
			delete(t.adj, k)
		} else {
			t.adj[k] = kept
		}
	}
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
	off := func(tri Triangle) int {
		for _, v := range tri {
			if v != e[0] && v != e[1] {
				return v
			}
		}
		return -1
	}
	p, q := off(tris[0]), off(tris[1])
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

// Ordered by where each one cuts a-b. Index order is just as deterministic
// but leaves forceEdge flipping in an order that can stall; walking the
// crossings along the segment is the order the flips actually want.
func (t *tessellation) crossing(a, b int) []edgeKey {
	pa, pb := t.pts[a], t.pts[b]

	type crossed struct {
		edge edgeKey
		at   float64
	}

	found := []crossed{}
	for _, e := range t.edges() {
		if e[0] == a || e[1] == a || e[0] == b || e[1] == b {
			continue
		}
		if !predicate.SegmentsCross(pa, pb, t.pts[e[0]], t.pts[e[1]]) {
			continue
		}

		from := predicate.Orient2D(t.pts[e[0]], t.pts[e[1]], pa)
		to := predicate.Orient2D(t.pts[e[0]], t.pts[e[1]], pb)
		at := 0.
		if from != to {
			at = from / (from - to)
		}
		found = append(found, crossed{e, at})
	}

	sort.Slice(found, func(i, j int) bool {
		if found[i].at != found[j].at {
			return found[i].at < found[j].at
		}
		if found[i].edge[0] != found[j].edge[0] {
			return found[i].edge[0] < found[j].edge[0]
		}
		return found[i].edge[1] < found[j].edge[1]
	})

	out := make([]edgeKey, len(found))
	for i, c := range found {
		out[i] = c.edge
	}
	return out
}

func (t *tessellation) forceEdge(a, b int) error {
	if a == b {
		return nil
	}
	if t.hasEdge(a, b) {
		t.constrained[newEdgeKey(a, b)] = exists
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

	t.constrained[newEdgeKey(a, b)] = exists
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

// Flood from outside the hull rather than testing centroids: a centroid on
// the line through a boundary edge gets misclassified by ray casting.
func (t *tessellation) discardOutside() {
	outside := make(map[Triangle]struct{})
	queue := make([]Triangle, 0)

	for e, tris := range t.adj {
		if len(tris) != 1 {
			continue
		}
		if _, locked := t.constrained[e]; locked {
			continue
		}
		if _, seen := outside[tris[0]]; !seen {
			outside[tris[0]] = exists
			queue = append(queue, tris[0])
		}
	}

	for len(queue) > 0 {
		tri := queue[0]
		queue = queue[1:]

		for _, edge := range tri.Edges() {
			k := newEdgeKey(edge[0], edge[1])
			if _, locked := t.constrained[k]; locked {
				continue
			}
			for _, neighbor := range t.adj[k] {
				if neighbor == tri {
					continue
				}
				if _, seen := outside[neighbor]; seen {
					continue
				}
				outside[neighbor] = exists
				queue = append(queue, neighbor)
			}
		}
	}

	for tri := range outside {
		t.remove(tri)
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
func splitCrossingConstraints(pts *[]vector2.Float64, segments [][2]int) {
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
			indexOfPoint(pts, p)
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

func indexOfPoint(pts *[]vector2.Float64, p vector2.Float64) int {
	for i, existing := range *pts {
		if existing.Sub(p).Length() < 1e-9 {
			return i
		}
	}
	*pts = append(*pts, p)
	return len(*pts) - 1
}

// ConstrainedDelaunay triangulates points so every constraint edge survives,
// then discards the triangles outside the constraints.
func ConstrainedDelaunay(points []vector2.Float64, constraints []Constraint) (modeling.Mesh, error) {
	pts := make([]vector2.Float64, len(points))
	copy(pts, points)

	segments := make([][2]int, 0)
	for _, c := range constraints {
		for i := range c.shape {
			a := indexOfPoint(&pts, c.shape[i])
			b := indexOfPoint(&pts, c.shape[(i+1)%len(c.shape)])
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
	splitCrossingConstraints(&pts, segments)

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
