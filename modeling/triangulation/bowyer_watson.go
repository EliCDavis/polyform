package triangulation

import (
	"math"
	"sort"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/predicate"
	"github.com/EliCDavis/polyform/math/sfc"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Edge [2]int

func (e Edge) Length(points []vector2.Float64) float64 {
	return points[e[0]].Distance(points[e[1]])
}

func (e Edge) Connected(other Edge) bool {
	return e[0] == other[0] || e[0] == other[1] || e[1] == other[0] || e[1] == other[1]
}

type Triangle [3]int

func (t Triangle) Edges() []Edge {
	return []Edge{{t[0], t[1]}, {t[1], t[2]}, {t[2], t[0]}}
}

func (t Triangle) CounterClockwise(points []vector2.Float64) bool {
	return ccw(points[t[0]], points[t[1]], points[t[2]])
}

// Points map to 3D as (x, 0, y), where clockwise in 2D is what puts the
// face normal on +Y. Every triangle reaching a mesh goes through here, so
// flips and hole filling can not drift into disagreeing about winding.
func (t Triangle) clockwise(points []vector2.Float64) Triangle {
	if ccw(points[t[0]], points[t[1]], points[t[2]]) {
		return Triangle{t[0], t[2], t[1]}
	}
	return t
}

func (t Triangle) InsideCircumcircle(p vector2.Float64, points []vector2.Float64) bool {
	a, b, c := points[t[0]], points[t[1]], points[t[2]]
	if predicate.Orient2D(a, b, c) < 0 {
		a, c = c, a
	}
	return predicate.InCircle(a, b, c, p) > 0
}

func ccw(a, b, c vector2.Float64) bool {
	return predicate.Orient2D(a, b, c) > 0
}

// SuperTriangle is an equilateral triangle containing every point in the
// set, built around the bounding circle.
//
// Its vertices are culled at the end, so any circumcircle test they win
// costs a real triangle at the hull. A hull sliver of relative width w has a
// circumradius near 1/(8w) times the extent, so the margin has to beat that
// for the thinnest sliver double precision can represent.
func SuperTriangle(points []vector2.Float64) []vector2.Float64 {
	min, max := geometry.Shape(points).GetBounds()

	center := min.Add(max).Scale(0.5)
	radius := max.Sub(min).Length() * superTriangleMargin
	if radius == 0 {
		radius = superTriangleMargin
	}

	return []vector2.Float64{
		vector2.New(center.X()-radius*math.Sqrt(3), center.Y()-radius),
		vector2.New(center.X(), center.Y()+2*radius),
		vector2.New(center.X()+radius*math.Sqrt(3), center.Y()-radius),
	}
}

const superTriangleMargin = 1e20

var exists = struct{}{}

// Insertion order decides how far each point-location walk has to travel.
// Points sorted along a space filling curve land next to the last one
// inserted, so the walk is a few steps rather than a crossing of the set.
func hilbertOrder(points []vector2.Float64) []int {
	min, max := geometry.Shape(points).GetBounds()
	span := math.Max(max.X()-min.X(), max.Y()-min.Y())
	if span == 0 {
		span = 1
	}

	keys := sfc.Hilbert2D{
		Min:        min,
		Max:        min.Add(vector2.Fill(span)),
		Resolution: 16,
	}.EncodeArray(points)

	order := make([]int, len(points))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool {
		return keys[order[i]] < keys[order[j]]
	})
	return order
}

// Triangles live in flat arrays three entries wide. Vertices wind counter
// clockwise, and edge k runs from vertex k to vertex k+1 with the triangle
// across it recorded in the matching neighbour slot.
type delaunay struct {
	points     []vector2.Float64
	vertices   []int
	neighbours []int
	dead       []bool
	free       []int
	last       int

	stamp    []int
	startsAt []int
}

func newDelaunay(points []vector2.Float64) *delaunay {
	super := SuperTriangle(points)
	all := make([]vector2.Float64, 0, len(points)+3)
	all = append(all, points...)
	all = append(all, super...)

	d := &delaunay{
		points:     all,
		vertices:   make([]int, 0, len(points)*6),
		neighbours: make([]int, 0, len(points)*6),
		dead:       make([]bool, 0, len(points)*2),
		startsAt:   make([]int, len(all)),
	}
	d.newTriangle(len(all)-3, len(all)-1, len(all)-2)
	return d
}

func (d *delaunay) newTriangle(a, b, c int) int {
	if n := len(d.free); n > 0 {
		t := d.free[n-1]
		d.free = d.free[:n-1]
		d.vertices[3*t], d.vertices[3*t+1], d.vertices[3*t+2] = a, b, c
		d.neighbours[3*t], d.neighbours[3*t+1], d.neighbours[3*t+2] = -1, -1, -1
		d.dead[t] = false
		return t
	}
	d.vertices = append(d.vertices, a, b, c)
	d.neighbours = append(d.neighbours, -1, -1, -1)
	d.dead = append(d.dead, false)
	d.stamp = append(d.stamp, 0)
	return len(d.dead) - 1
}

func (d *delaunay) edgeIndex(t, from, to int) int {
	for k := 0; k < 3; k++ {
		if d.vertices[3*t+k] == from && d.vertices[3*t+(k+1)%3] == to {
			return k
		}
	}
	return -1
}

func (d *delaunay) inCircumcircle(t int, p vector2.Float64) bool {
	return predicate.InCircle(
		d.points[d.vertices[3*t]],
		d.points[d.vertices[3*t+1]],
		d.points[d.vertices[3*t+2]],
		p,
	) > 0
}

// Walks from the last triangle touched, crossing whichever edge has the
// point on its far side, until no edge does. Delaunay triangulations are the
// case where this walk is known to terminate; the budget covers the rest.
func (d *delaunay) locate(p vector2.Float64) int {
	t, previous := d.last, -1
	for step := 0; step < len(d.dead); step++ {
		next := -1
		for k := 0; k < 3; k++ {
			across := d.neighbours[3*t+k]
			if across < 0 || across == previous {
				continue
			}
			a, b := d.points[d.vertices[3*t+k]], d.points[d.vertices[3*t+(k+1)%3]]
			if predicate.Orient2D(a, b, p) < 0 {
				next = across
				break
			}
		}
		if next < 0 {
			return t
		}
		t, previous = next, t
	}

	for t := range d.dead {
		if !d.dead[t] && d.contains(t, p) {
			return t
		}
	}
	return -1
}

func (d *delaunay) contains(t int, p vector2.Float64) bool {
	for k := 0; k < 3; k++ {
		a, b := d.points[d.vertices[3*t+k]], d.points[d.vertices[3*t+(k+1)%3]]
		if predicate.Orient2D(a, b, p) < 0 {
			return false
		}
	}
	return true
}

type cavityEdge struct {
	from, to, across int
}

func (d *delaunay) insert(index int) {
	p := d.points[index]

	start := d.locate(p)
	if start < 0 || !d.inCircumcircle(start, p) {
		return
	}

	mark := index + 1
	bad := []int{start}
	d.stamp[start] = mark
	for i := 0; i < len(bad); i++ {
		t := bad[i]
		for k := 0; k < 3; k++ {
			n := d.neighbours[3*t+k]
			if n < 0 || d.stamp[n] == mark {
				continue
			}
			if d.inCircumcircle(n, p) {
				d.stamp[n] = mark
				bad = append(bad, n)
			}
		}
	}

	boundary := make([]cavityEdge, 0, len(bad)+2)
	for _, t := range bad {
		for k := 0; k < 3; k++ {
			n := d.neighbours[3*t+k]
			if n >= 0 && d.stamp[n] == mark {
				continue
			}
			boundary = append(boundary, cavityEdge{
				from:   d.vertices[3*t+k],
				to:     d.vertices[3*t+(k+1)%3],
				across: n,
			})
		}
	}

	for _, t := range bad {
		d.dead[t] = true
		d.free = append(d.free, t)
	}

	created := make([]int, 0, len(boundary))
	for _, e := range boundary {
		t := d.newTriangle(e.from, e.to, index)
		d.neighbours[3*t] = e.across
		if e.across >= 0 {
			d.neighbours[3*e.across+d.edgeIndex(e.across, e.to, e.from)] = t
		}
		d.startsAt[e.from] = t
		created = append(created, t)
	}

	for _, t := range created {
		next := d.startsAt[d.vertices[3*t+1]]
		d.neighbours[3*t+1] = next
		d.neighbours[3*next+2] = t
	}

	d.last = created[0]
}

func (d *delaunay) cullSuperTriangle(count int) {
	for t := range d.dead {
		if d.dead[t] {
			continue
		}
		if d.vertices[3*t] < count && d.vertices[3*t+1] < count && d.vertices[3*t+2] < count {
			continue
		}
		d.dead[t] = true
		for k := 0; k < 3; k++ {
			if n := d.neighbours[3*t+k]; n >= 0 {
				d.neighbours[3*n+d.edgeIndex(n, d.vertices[3*t+(k+1)%3], d.vertices[3*t+k])] = -1
			}
		}
	}
}

func (d *delaunay) triangles() []Triangle {
	out := make([]Triangle, 0, len(d.dead))
	for t := range d.dead {
		if d.dead[t] {
			continue
		}
		out = append(out, Triangle{d.vertices[3*t], d.vertices[3*t+1], d.vertices[3*t+2]})
	}
	sort.Slice(out, func(i, j int) bool {
		for k := 0; k < 3; k++ {
			if out[i][k] != out[j][k] {
				return out[i][k] < out[j][k]
			}
		}
		return false
	})
	return out
}

func bowyerWatson(points []vector2.Float64) *delaunay {
	if len(points) < 3 {
		panic("can not tesselate without at least 3 points")
	}

	d := newDelaunay(points)
	for _, i := range hilbertOrder(points) {
		d.insert(i)
	}
	d.cullSuperTriangle(len(points))
	return d
}

// Triangles wind counter clockwise throughout, and the mesh wants clockwise,
// so each one is emitted reversed.
func BowyerWatson(points []vector2.Float64) modeling.Mesh {
	d := bowyerWatson(points)

	tris := make([]int, 0, len(d.dead)*3)
	for t := range d.dead {
		if d.dead[t] {
			continue
		}
		tris = append(tris, d.vertices[3*t], d.vertices[3*t+2], d.vertices[3*t+1])
	}

	verts := make([]vector3.Float64, len(points))
	uvs := make([]vector2.Float64, len(points))
	for i, p := range points {
		verts[i] = vector3.New(p.X(), 0, p.Y())
		uvs[i] = vector2.Zero[float64]()
	}

	return modeling.
		NewTriangleMesh(tris).
		SetFloat3Attribute(modeling.PositionAttribute, verts).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs)
}
