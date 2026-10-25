package triangulation

import (
	"errors"
	"math"
	"slices"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/predicate"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

const (
	defaultVoronoiDomainSegments = 64
	defaultVoronoiDomainScale    = 0.75
)

// Voronoi returns, for every site, the part of domain closer to that site
// than to any other. Domain must be a convex loop; fewer than 3 points falls
// back to a circle around the sites.
func Voronoi(sites, domain []vector2.Float64) ([][]vector2.Float64, error) {
	if len(sites) == 0 {
		return nil, nil
	}

	if len(domain) < 3 {
		domain = defaultVoronoiDomain(sites)
	}
	domain, err := convexCounterClockwise(domain)
	if err != nil {
		return nil, err
	}

	neighbours := voronoiNeighbours(sites)
	weld := geometry.NewPointWelder2D(mergeTolerance(domain))
	cells := make([][]vector2.Float64, len(sites))
	for i, site := range sites {
		cell := slices.Clone(domain)
		for _, j := range neighbours[i] {
			cell = clipToBisector(cell, site, sites[j])
		}
		cells[i] = weldLoop(weld, cell)
	}
	return cells, nil
}

func weldLoop(weld *geometry.PointWelder2D, loop []vector2.Float64) []vector2.Float64 {
	welded := make([]vector2.Float64, 0, len(loop))
	for _, p := range loop {
		i := weld.Index(p)
		p = weld.Points()[i]
		if len(welded) == 0 || welded[len(welded)-1] != p {
			welded = append(welded, p)
		}
	}
	if len(welded) > 1 && welded[0] == welded[len(welded)-1] {
		welded = welded[:len(welded)-1]
	}
	if len(welded) < 3 {
		return nil
	}
	return welded
}

func defaultVoronoiDomain(sites []vector2.Float64) []vector2.Float64 {
	min, max := geometry.Shape(sites).GetBounds()
	center := min.Add(max).Scale(0.5)
	radius := max.Sub(min).Length() * defaultVoronoiDomainScale
	if radius == 0 {
		radius = 1
	}

	circle := make([]vector2.Float64, defaultVoronoiDomainSegments)
	for i := range circle {
		angle := 2 * math.Pi * float64(i) / defaultVoronoiDomainSegments
		circle[i] = center.Add(vector2.New(math.Cos(angle), math.Sin(angle)).Scale(radius))
	}
	return circle
}

func convexCounterClockwise(loop []vector2.Float64) ([]vector2.Float64, error) {
	area := geometry.Shape(loop).SignedArea()
	if area == 0 {
		return nil, errors.New("voronoi domain has no area")
	}

	out := slices.Clone(loop)
	if area < 0 {
		slices.Reverse(out)
	}

	for i := range out {
		if predicate.Orient2D(out[i], out[(i+1)%len(out)], out[(i+2)%len(out)]) < 0 {
			return nil, errors.New("voronoi domain must be convex")
		}
	}
	return out, nil
}

func voronoiNeighbours(sites []vector2.Float64) [][]int {
	neighbours := make([][]int, len(sites))
	if len(sites) >= 3 {
		for _, tri := range bowyerWatson(sites).triangles() {
			for _, e := range tri.Edges() {
				neighbours[e[0]] = append(neighbours[e[0]], e[1])
				neighbours[e[1]] = append(neighbours[e[1]], e[0])
			}
		}
	}

	for i := range neighbours {
		if len(neighbours[i]) == 0 {
			for j := range sites {
				if j != i {
					neighbours[i] = append(neighbours[i], j)
				}
			}
			continue
		}
		slices.Sort(neighbours[i])
		neighbours[i] = slices.Compact(neighbours[i])
	}
	return neighbours
}

func clipToBisector(cell []vector2.Float64, keep, other vector2.Float64) []vector2.Float64 {
	normal := other.Sub(keep)
	mid := keep.Add(other).Scale(0.5)

	out := make([]vector2.Float64, 0, len(cell)+1)
	for i, cur := range cell {
		prev := cell[(i+len(cell)-1)%len(cell)]
		dPrev := prev.Sub(mid).Dot(normal)
		dCur := cur.Sub(mid).Dot(normal)
		if (dPrev < 0 && dCur > 0) || (dPrev > 0 && dCur < 0) {
			out = append(out, prev.Add(cur.Sub(prev).Scale(dPrev/(dPrev-dCur))))
		}
		if dCur <= 0 {
			out = append(out, cur)
		}
	}
	return out
}

type VoronoiNode struct {
	Points nodes.Output[[]vector2.Float64]
	Domain nodes.Output[[]vector2.Float64] `description:"Convex outline the cells fill. Fewer than 3 points uses a circle around the points."`
}

func (node VoronoiNode) Description() string {
	return "Splits a domain into one cell per point, each holding the area closest to its point."
}

func (node VoronoiNode) Out(out *nodes.StructOutput[[][]vector2.Float64]) {
	points := nodes.TryGetOutputValue(out, node.Points, nil)
	domain := nodes.TryGetOutputValue(out, node.Domain, nil)

	cells, err := Voronoi(points, domain)
	if err != nil {
		out.CaptureError(err)
		return
	}
	out.Set(cells)
}
