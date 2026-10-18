package triangulation

import (
	"errors"
	"math"
	"slices"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector2"
)

type Region struct {
	Points    []vector2.Float64
	Triangles []Triangle
}

// Fill triangulates the area the contours enclose by nonzero winding:
// overlapping contours union, and a contour wound against the one around it
// is a hole.
func Fill(contours ...geometry.Shape) (Region, error) {
	return FillWinding([][]geometry.Shape{contours}, func(w []int) bool { return w[0] != 0 })
}

// FillDifference triangulates the area inside any of keep and outside every
// one of cut, however they overlap. Winding plays no part.
func FillDifference(keep, cut []geometry.Shape) (Region, error) {
	if len(keep) == 0 {
		return Region{}, nil
	}
	if !slices.ContainsFunc(keep, func(s geometry.Shape) bool { return s.SignedArea() != 0 }) {
		return Region{}, errors.New("no outline encloses any area")
	}

	return FillWinding(
		[][]geometry.Shape{CounterClockwise(keep), CounterClockwise(cut)},
		func(w []int) bool { return w[0] != 0 && w[1] == 0 },
	)
}

// FillWinding keeps the triangles keep accepts, given each group's winding number there.
func FillWinding(groups [][]geometry.Shape, keep func(winding []int) bool) (Region, error) {
	weld, segments, groupOf := contourSegments(groups)
	if len(segments) == 0 {
		return Region{}, nil
	}
	tess, err := triangulate(weld, segments, groupOf, len(groups))
	if err != nil {
		return Region{}, err
	}

	windings := tess.windings()
	g := tess.groups
	return tess.region(func(tri int) bool {
		return keep(windings[tri*g : (tri+1)*g])
	}), nil
}

func CounterClockwise(shapes []geometry.Shape) []geometry.Shape {
	out := make([]geometry.Shape, len(shapes))
	for i, s := range shapes {
		out[i] = s
		if s.SignedArea() < 0 {
			out[i] = slices.Clone(s)
			slices.Reverse(out[i])
		}
	}
	return out
}

// Outer loops run counter-clockwise, holes clockwise; points on straight runs are dropped.
func (r Region) Boundaries() []geometry.Shape {
	uses := map[[2]int]int{}
	for _, tri := range r.Triangles {
		for k := range tri {
			a, b := tri[k], tri[(k+1)%3]
			uses[[2]int{min(a, b), max(a, b)}]++
		}
	}

	next := map[int][]int{}
	for _, tri := range r.Triangles {
		for k := range tri {
			a, b := tri[k], tri[(k+1)%3]
			if uses[[2]int{min(a, b), max(a, b)}] == 1 {
				next[a] = append(next[a], b)
			}
		}
	}

	starts := make([]int, 0, len(next))
	for v := range next {
		starts = append(starts, v)
	}
	slices.Sort(starts)

	var loops []geometry.Shape
	for _, start := range starts {
		for len(next[start]) > 0 {
			var loop []int
			for v := start; ; {
				loop = append(loop, v)
				to := next[v][0]
				next[v] = next[v][1:]
				v = to
				if v == start {
					break
				}
				if len(next[v]) == 0 {
					loop = nil
					break
				}
			}
			if len(loop) >= 3 {
				loops = append(loops, r.withoutStraightRuns(loop))
			}
		}
	}
	return loops
}

func (r Region) withoutStraightRuns(loop []int) geometry.Shape {
	tolerance := mergeTolerance(r.Points)
	kept := make(geometry.Shape, 0, len(loop))
	for i, v := range loop {
		prev, at, next := r.Points[loop[(i+len(loop)-1)%len(loop)]], r.Points[v], r.Points[loop[(i+1)%len(loop)]]
		span := next.Sub(prev)
		if length := span.Length(); length > 0 {
			offLine := math.Abs(span.X()*at.Sub(prev).Y()-span.Y()*at.Sub(prev).X()) / length
			if offLine <= tolerance && at.Sub(prev).Dot(span) > 0 && next.Sub(at).Dot(span) > 0 {
				continue
			}
		}
		kept = append(kept, at)
	}
	return kept
}

func contourSegments(groups [][]geometry.Shape) (*geometry.PointWelder2D, [][2]int, []int) {
	var sets [][]vector2.Float64
	for _, contours := range groups {
		for _, contour := range contours {
			sets = append(sets, contour)
		}
	}
	weld := geometry.NewPointWelder2D(mergeTolerance(sets...))

	var segments [][2]int
	var groupOf []int
	for g, contours := range groups {
		for _, contour := range contours {
			for i := range contour {
				a, b := weld.Index(contour[i]), weld.Index(contour[(i+1)%len(contour)])
				if a != b {
					segments = append(segments, [2]int{a, b})
					groupOf = append(groupOf, g)
				}
			}
		}
	}
	return weld, segments, groupOf
}

func (t *tessellation) region(kept func(tri int) bool) Region {
	var triangles []Triangle
	used := make([]bool, len(t.pts))
	for _, tri := range t.ordered(func(tri Triangle) Triangle {
		clockwise := tri.clockwise(t.pts)
		return Triangle{clockwise[0], clockwise[2], clockwise[1]}
	}, kept) {
		triangles = append(triangles, tri)
		for _, v := range tri {
			used[v] = true
		}
	}

	remap := make([]int, len(t.pts))
	var points []vector2.Float64
	for i, p := range t.pts {
		if used[i] {
			remap[i] = len(points)
			points = append(points, p)
		}
	}
	for i, tri := range triangles {
		triangles[i] = Triangle{remap[tri[0]], remap[tri[1]], remap[tri[2]]}
	}
	return Region{Points: points, Triangles: triangles}
}
