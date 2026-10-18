package sketch

import (
	"math"

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

func withoutRepeats(shape []vector2.Float64, closed bool) []vector2.Float64 {
	kept := make([]vector2.Float64, 0, len(shape))
	for _, p := range shape {
		if len(kept) == 0 || kept[len(kept)-1] != p {
			kept = append(kept, p)
		}
	}
	if closed && len(kept) > 1 && kept[0] == kept[len(kept)-1] {
		kept = kept[:len(kept)-1]
	}
	return kept
}

type BevelNode struct {
	Outline  nodes.LiftedPort[[]vector2.Float64] `description:"The path or outline whose corners to cut."`
	Distance nodes.LiftedPort[float64]           `description:"How far back along each side a cut starts, at most half the side. On a right angle a round corner's radius equals it."`
	Round    nodes.Output[bool]                  `description:"Rounds each corner with an arc instead of a straight cut."`
	Segments nodes.Output[int]                   `description:"Straight pieces per round corner. Defaults to 8."`
	Closed   nodes.Output[bool]                  `description:"Treats the points as a closed outline, so every corner is cut. Off, the two ends stay put."`
}

func (BevelNode) Description() string {
	return "Cuts every corner of a 2D path or outline, straight (a chamfer) or round (a fillet)."
}

func (n BevelNode) Out(out *nodes.Lifted[[]vector2.Float64]) {
	round := nodes.TryGetOutputValue(out, n.Round, false)
	segments := max(nodes.TryGetOutputValue(out, n.Segments, 8), 1)
	closed := nodes.TryGetOutputValue(out, n.Closed, false)
	nodes.Zip2(out, n.Outline, n.Distance, func(shape []vector2.Float64, distance float64) []vector2.Float64 {
		return Corners{Distance: distance, Round: round, Segments: segments, Closed: closed}.Apply(shape)
	})
}
