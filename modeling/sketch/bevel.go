package sketch

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

type Corners struct {
	Distance float64
	Round    bool
	Segments int
	Closed   bool
}

// A side gives at most half its length to each corner, so neighbouring cuts never cross.
func (c Corners) Apply(shape []vector2.Float64) []vector2.Float64 {
	points := withoutRepeats(shape, c.Closed)
	n := len(points)
	if n < 3 || c.Distance <= 0 {
		return points
	}

	corner := func(i int) bool { return c.Closed || (i > 0 && i < n-1) }
	share := func(from, to int) float64 {
		length := points[to].Distance(points[from])
		if corner(from) && corner(to) {
			return length / 2
		}
		return length
	}

	cut := make([]vector2.Float64, 0, n*(max(c.Segments, 1)+1))
	for i, p := range points {
		if !corner(i) {
			cut = append(cut, p)
			continue
		}
		prev, next := (i+n-1)%n, (i+1)%n
		in := p.Sub(points[prev]).Normalized()
		out := points[next].Sub(p).Normalized()
		turn := math.Atan2(in.X()*out.Y()-in.Y()*out.X(), in.Dot(out))
		if math.Abs(turn) < 1e-9 || math.Abs(turn) > math.Pi-1e-9 {
			cut = append(cut, p)
			continue
		}

		d := math.Min(c.Distance, math.Min(share(prev, i), share(i, next)))
		from, to := p.Sub(in.Scale(d)), p.Add(out.Scale(d))
		if !c.Round || c.Segments <= 1 {
			cut = append(cut, from, to)
			continue
		}

		radius := d / math.Tan(math.Abs(turn)/2)
		side := math.Copysign(1, turn)
		center := from.Add(vector2.New(-in.Y(), in.X()).Scale(radius * side))
		start := math.Atan2(from.Y()-center.Y(), from.X()-center.X())
		for j := 0; j <= c.Segments; j++ {
			angle := start + turn*float64(j)/float64(c.Segments)
			cut = append(cut, center.Add(vector2.New(math.Cos(angle), math.Sin(angle)).Scale(radius)))
		}
	}
	return withoutRepeats(cut, c.Closed)
}

const repeatTolerance = 1e-9

func withoutRepeats(shape []vector2.Float64, closed bool) []vector2.Float64 {
	kept := make([]vector2.Float64, 0, len(shape))
	for _, p := range shape {
		if len(kept) == 0 || kept[len(kept)-1].Distance(p) > repeatTolerance {
			kept = append(kept, p)
		}
	}
	if closed && len(kept) > 1 && kept[0].Distance(kept[len(kept)-1]) <= repeatTolerance {
		kept = kept[:len(kept)-1]
	}
	return kept
}

func BevelArea(area Area, corners Corners) (Area, error) {
	corners.Closed = true
	var outlines, holes []geometry.Shape
	for _, polygon := range area.Polygons {
		outlines = append(outlines, corners.Apply(polygon.Outline))
		for _, hole := range polygon.Holes {
			holes = append(holes, corners.Apply(hole))
		}
	}
	return Fill(outlines, holes)
}

type BevelPathNode struct {
	Outline  nodes.LiftedPort[[]vector2.Float64] `description:"The path or outline whose corners to cut."`
	Distance nodes.LiftedPort[float64]           `description:"How far back along each side a cut starts, at most half the side. On a right angle a round corner's radius equals it."`
	Round    nodes.Output[bool]                  `description:"Rounds each corner with an arc instead of a straight cut."`
	Segments nodes.Output[int]                   `description:"Straight pieces per round corner. Defaults to 8."`
	Closed   nodes.Output[bool]                  `description:"Treats the points as a closed outline, so every corner is cut. Off, the two ends stay put."`
}

func (BevelPathNode) Description() string {
	return "Cuts every corner of a 2D path or outline, straight (a chamfer) or round (a fillet)."
}

func (n BevelPathNode) Out(out *nodes.Lifted[[]vector2.Float64]) {
	round := nodes.TryGetOutputValue(out, n.Round, false)
	segments := max(nodes.TryGetOutputValue(out, n.Segments, 8), 1)
	closed := nodes.TryGetOutputValue(out, n.Closed, false)
	nodes.Zip2(out, n.Outline, n.Distance, func(shape []vector2.Float64, distance float64) []vector2.Float64 {
		return Corners{Distance: distance, Round: round, Segments: segments, Closed: closed}.Apply(shape)
	})
}

type BevelNode struct {
	Area     nodes.LiftedPort[Area]    `description:"The area whose corners to cut, outlines and holes alike."`
	Distance nodes.LiftedPort[float64] `description:"How far back along each side a cut starts, at most half the side."`
	Round    nodes.Output[bool]        `description:"Rounds each corner with an arc instead of a straight cut."`
	Segments nodes.Output[int]         `description:"Straight pieces per round corner. Defaults to 8."`
}

func (BevelNode) Description() string {
	return "Cuts every corner of an area, straight (a chamfer) or round (a fillet), on its outlines and its holes."
}

func (BevelNode) Keywords() []string {
	return []string{"fillet", "chamfer", "round corners"}
}

func (n BevelNode) Out(out *nodes.Lifted[Area]) {
	round := nodes.TryGetOutputValue(out, n.Round, false)
	segments := max(nodes.TryGetOutputValue(out, n.Segments, 8), 1)
	nodes.Zip2(out, n.Area, n.Distance, func(area Area, distance float64) Area {
		cut, err := BevelArea(area, Corners{Distance: distance, Round: round, Segments: segments})
		out.CaptureError(err)
		return cut
	})
}
