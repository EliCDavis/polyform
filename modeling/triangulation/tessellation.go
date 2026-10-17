package triangulation

import (
	"fmt"
	"sort"

	"github.com/EliCDavis/polyform/math/predicate"
	"github.com/EliCDavis/polyform/modeling"
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

// Triangles live in flat arrays three entries wide, wound counter clockwise:
// edge k runs from vertex k to vertex k+1, with the triangle across it in the
// matching neighbour slot, or -1 on the hull. winding holds one count per
// group for each constraint edge: +1 for every piece running from the lower
// index to the higher, -1 for the reverse.
type tessellation struct {
	pts        []vector2.Float64
	vertices   []int
	neighbours []int
	startsAt   []int
	winding    map[edgeKey][]int
	groups     int
	reshaped   []edgeKey
}

func newTessellation(pts []vector2.Float64, d *delaunay, groups int) *tessellation {
	remap := make([]int, len(d.dead))
	live := 0
	for i, dead := range d.dead {
		remap[i] = -1
		if !dead {
			remap[i] = live
			live++
		}
	}

	t := &tessellation{
		pts:        pts,
		vertices:   make([]int, 0, 3*live),
		neighbours: make([]int, 0, 3*live),
		startsAt:   make([]int, len(pts)),
		winding:    make(map[edgeKey][]int),
		groups:     max(groups, 1),
	}
	for i := range t.startsAt {
		t.startsAt[i] = -1
	}
	for i, dead := range d.dead {
		if dead {
			continue
		}
		for k := 0; k < 3; k++ {
			v := d.vertices[3*i+k]
			t.vertices = append(t.vertices, v)
			n := d.neighbours[3*i+k]
			if n >= 0 {
				n = remap[n]
			}
			t.neighbours = append(t.neighbours, n)
			t.startsAt[v] = remap[i]
		}
	}
	return t
}

func (t *tessellation) count() int { return len(t.vertices) / 3 }

func (t *tessellation) vertex(tri, k int) int { return t.vertices[3*tri+k%3] }

func (t *tessellation) indexOf(tri, v int) int {
	for k := 0; k < 3; k++ {
		if t.vertices[3*tri+k] == v {
			return k
		}
	}
	return -1
}

func (t *tessellation) third(tri, a, b int) int {
	for k := 0; k < 3; k++ {
		if v := t.vertices[3*tri+k]; v != a && v != b {
			return v
		}
	}
	return -1
}

// aroundVertex visits every triangle holding a, with a's index in it, until
// visit returns false.
func (t *tessellation) aroundVertex(a int, visit func(tri, i int) bool) {
	start := t.startsAt[a]
	if start < 0 {
		return
	}
	for tri := start; ; {
		i := t.indexOf(tri, a)
		if !visit(tri, i) {
			return
		}
		next := t.neighbours[3*tri+i]
		if next == start {
			return
		}
		if next < 0 {
			break
		}
		tri = next
	}
	for tri := start; ; {
		prev := t.neighbours[3*tri+(t.indexOf(tri, a)+2)%3]
		if prev < 0 || prev == start {
			return
		}
		tri = prev
		if !visit(tri, t.indexOf(tri, a)) {
			return
		}
	}
}

// slot finds the triangle holding the directed edge a→b and the edge's index
// in it.
func (t *tessellation) slot(a, b int) (tri, k int, ok bool) {
	t.aroundVertex(a, func(candidate, i int) bool {
		if t.vertices[3*candidate+(i+1)%3] == b {
			tri, k, ok = candidate, i, true
			return false
		}
		return true
	})
	return
}

func (t *tessellation) hasEdge(a, b int) bool {
	if _, _, ok := t.slot(a, b); ok {
		return true
	}
	_, _, ok := t.slot(b, a)
	return ok
}

func (t *tessellation) locked(e edgeKey) bool {
	_, ok := t.winding[e]
	return ok
}

// sides returns the triangle with e running a→b, its slot, and the triangle
// and apex across it.
func (t *tessellation) sides(e edgeKey) (tri, k, across, p, q int, ok bool) {
	tri, k, ok = t.slot(e[0], e[1])
	if !ok {
		if tri, k, ok = t.slot(e[1], e[0]); !ok {
			return
		}
	}
	across = t.neighbours[3*tri+k]
	if across < 0 {
		return tri, k, across, 0, 0, false
	}
	a, b := t.vertex(tri, k), t.vertex(tri, k+1)
	return tri, k, across, t.vertex(tri, k+2), t.third(across, a, b), true
}

// setNeighbour points tri's edge leaving from at neighbour.
func (t *tessellation) setNeighbour(tri, from, neighbour int) {
	if tri >= 0 {
		t.neighbours[3*tri+t.indexOf(tri, from)] = neighbour
	}
}

func (t *tessellation) flip(e edgeKey) bool {
	tri, k, u, p, q, ok := t.sides(e)
	if !ok {
		return false
	}
	a, b := t.vertex(tri, k), t.vertex(tri, k+1)
	oa := predicate.Orient2D(t.pts[p], t.pts[q], t.pts[a])
	ob := predicate.Orient2D(t.pts[p], t.pts[q], t.pts[b])
	if oa == 0 || ob == 0 || (oa > 0) == (ob > 0) {
		return false
	}

	j := t.indexOf(u, b)
	outerBP, outerPA := t.neighbours[3*tri+(k+1)%3], t.neighbours[3*tri+(k+2)%3]
	outerAQ, outerQB := t.neighbours[3*u+(j+1)%3], t.neighbours[3*u+(j+2)%3]

	t.vertices[3*tri], t.vertices[3*tri+1], t.vertices[3*tri+2] = a, q, p
	t.neighbours[3*tri], t.neighbours[3*tri+1], t.neighbours[3*tri+2] = outerAQ, u, outerPA
	t.vertices[3*u], t.vertices[3*u+1], t.vertices[3*u+2] = q, b, p
	t.neighbours[3*u], t.neighbours[3*u+1], t.neighbours[3*u+2] = outerQB, outerBP, tri

	t.setNeighbour(outerAQ, q, tri)
	t.setNeighbour(outerBP, p, u)
	t.startsAt[a], t.startsAt[q], t.startsAt[p], t.startsAt[b] = tri, tri, tri, u
	t.reshaped = append(t.reshaped,
		newEdgeKey(p, q), newEdgeKey(a, q), newEdgeKey(q, b), newEdgeKey(b, p), newEdgeKey(p, a))
	return true
}

// Walks triangle to triangle along a-b, so the edges come back in the order
// the segment meets them, which is the order forceEdge wants to flip them.
func (t *tessellation) crossing(a, b int) []edgeKey {
	pa, pb := t.pts[a], t.pts[b]

	tri, u, v := -1, -1, -1
	t.aroundVertex(a, func(candidate, i int) bool {
		r0, r1 := t.vertex(candidate, i+1), t.vertex(candidate, i+2)
		if predicate.SegmentsCross(pa, pb, t.pts[r0], t.pts[r1]) {
			tri, u, v = candidate, r0, r1
			return false
		}
		return true
	})
	if tri < 0 {
		return nil
	}

	out := []edgeKey{newEdgeKey(u, v)}
	for {
		next := t.neighbours[3*tri+t.indexOf(tri, u)]
		if t.vertex(tri, t.indexOf(tri, u)+1) != v {
			next = t.neighbours[3*tri+t.indexOf(tri, v)]
		}
		if next < 0 {
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

func (t *tessellation) lock(a, b, group int) {
	k := newEdgeKey(a, b)
	w := t.winding[k]
	if w == nil {
		w = make([]int, t.groups)
		t.winding[k] = w
	}
	if a < b {
		w[group]++
	} else {
		w[group]--
	}
}

func (t *tessellation) forceEdge(a, b, group int) error {
	if a == b {
		return nil
	}
	if t.hasEdge(a, b) {
		t.lock(a, b, group)
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
			if t.locked(e) {
				return fmt.Errorf("constraint edge %d-%d crosses another constraint edge", a, b)
			}
			if t.flip(e) {
				flippedAny = true
			}
		}

		if !flippedAny {
			return fmt.Errorf("could not force edge %d-%d; every crossing edge sits in a concave quad", a, b)
		}
	}

	t.lock(a, b, group)
	return nil
}

// Lawson's flip queue. The triangulation was Delaunay before constraints
// were forced, and an edge's Delaunay test only reads the two triangles
// beside it, so only edges whose triangles a flip reshaped need checking.
func (t *tessellation) restoreDelaunay() {
	queue := make([]edgeKey, 0, len(t.reshaped))
	queued := make(map[edgeKey]struct{}, len(t.reshaped))
	for _, e := range t.reshaped {
		if _, already := queued[e]; !already {
			queued[e] = exists
			queue = append(queue, e)
		}
	}
	t.reshaped = nil

	budget := t.count()*t.count() + 64
	for len(queue) > 0 && budget > 0 {
		e := queue[0]
		queue = queue[1:]
		delete(queued, e)

		if t.locked(e) {
			continue
		}
		tri, k, _, p, q, ok := t.sides(e)
		if !ok {
			continue
		}
		a, b := t.vertex(tri, k), t.vertex(tri, k+1)
		if predicate.InCircle(t.pts[a], t.pts[b], t.pts[p], t.pts[q]) <= 0 {
			continue
		}
		if !t.flip(e) {
			continue
		}
		budget--

		for _, around := range [4]edgeKey{
			newEdgeKey(a, p), newEdgeKey(b, p),
			newEdgeKey(a, q), newEdgeKey(b, q),
		} {
			if _, already := queued[around]; already {
				continue
			}
			queued[around] = exists
			queue = append(queue, around)
		}
	}
	t.reshaped = nil
}

// Winding numbers flood in from the hull, where they are zero, and only
// change across constraint edges. Triangle tri's count for each group is at
// [tri*groups, (tri+1)*groups).
func (t *tessellation) windings() []int {
	g := t.groups
	windings := make([]int, t.count()*g)
	seen := make([]bool, t.count())
	queue := make([]int, 0, t.count())

	// tri lies to the left of its own edge k, so entering across it adds the
	// pieces running the same way as that edge.
	enter := func(tri, k int, from []int) {
		if seen[tri] {
			return
		}
		seen[tri] = true
		into := windings[tri*g : (tri+1)*g]
		copy(into, from)
		a, b := t.vertex(tri, k), t.vertex(tri, k+1)
		if w, ok := t.winding[newEdgeKey(a, b)]; ok {
			sign := 1
			if a > b {
				sign = -1
			}
			for i := range into {
				into[i] += sign * w[i]
			}
		}
		queue = append(queue, tri)
	}

	outside := make([]int, g)
	for tri := 0; tri < t.count(); tri++ {
		for k := 0; k < 3; k++ {
			if t.neighbours[3*tri+k] < 0 {
				enter(tri, k, outside)
			}
		}
	}
	for n := 0; n < len(queue); n++ {
		tri := queue[n]
		from := windings[tri*g : (tri+1)*g]
		for k := 0; k < 3; k++ {
			if across := t.neighbours[3*tri+k]; across >= 0 {
				enter(across, t.indexOf(across, t.vertex(tri, k+1)), from)
			}
		}
	}
	return windings
}

// ordered applies orient to every triangle kept accepts, then sorts them so
// the same input always comes back in the same order.
func (t *tessellation) ordered(orient func(Triangle) Triangle, kept func(tri int) bool) []Triangle {
	ordered := make([]Triangle, 0, t.count())
	for tri := 0; tri < t.count(); tri++ {
		if kept == nil || kept(tri) {
			ordered = append(ordered, orient(Triangle{t.vertex(tri, 0), t.vertex(tri, 1), t.vertex(tri, 2)}))
		}
	}
	sort.Slice(ordered, func(i, j int) bool {
		for k := 0; k < 3; k++ {
			if ordered[i][k] != ordered[j][k] {
				return ordered[i][k] < ordered[j][k]
			}
		}
		return false
	})
	return ordered
}

func (t *tessellation) mesh(kept func(tri int) bool) modeling.Mesh {
	tris := make([]int, 0, len(t.vertices))
	for _, tri := range t.ordered(func(tri Triangle) Triangle { return tri.clockwise(t.pts) }, kept) {
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
