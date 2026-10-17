package sketch

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

// Offset grows the area by distance, or shrinks it when distance is
// negative, keeping it round where it grows around a corner.
func Offset(area Area, distance float64, segments int) (Area, error) {
	if distance == 0 || len(area.Polygons) == 0 {
		return area, nil
	}
	segments = max(segments, 1)

	var curves []geometry.Shape
	for _, loop := range area.Loops() {
		if curve := offsetLoop(loop, distance, segments); len(curve) >= 3 {
			curves = append(curves, curve)
		}
	}

	// Only positive winding is kept: it drops the loops concave corners
	// leave and every piece a shrink turned inside out.
	region, err := triangulation.FillWinding([][]geometry.Shape{curves}, func(w []int) bool { return w[0] > 0 })
	if err != nil {
		return Area{}, err
	}
	return areaOf(region), nil
}

// offsetLoop moves every edge of the loop distance to its right, which is
// away from the area for outlines and holes alike. Where moved edges pull
// apart an arc joins them; where they overlap the path runs back through
// the corner.
func offsetLoop(loop geometry.Shape, distance float64, segments int) geometry.Shape {
	pts := withoutRepeats(loop, true)
	n := len(pts)
	if n < 3 {
		return nil
	}
	right := func(v vector2.Float64) vector2.Float64 { return vector2.New(v.Y(), -v.X()) }
	radius := math.Abs(distance)
	step := math.Pi / float64(segments)

	curve := make(geometry.Shape, 0, 3*n)
	for i := range n {
		prev, at, next := pts[(i+n-1)%n], pts[i], pts[(i+1)%n]
		in, out := at.Sub(prev).Normalized(), next.Sub(at).Normalized()
		normalIn, normalOut := right(in).Scale(distance), right(out).Scale(distance)
		turn := in.X()*out.Y() - in.Y()*out.X()

		switch {
		case in.Dot(out) > 1-1e-12:
			curve = append(curve, at.Add(normalIn))
		case turn*distance > 0 || in.Dot(out) < -1+1e-12:
			from := math.Atan2(normalIn.Y(), normalIn.X())
			sweep := math.Atan2(normalOut.Y(), normalOut.X()) - from
			for distance > 0 && sweep <= 0 {
				sweep += 2 * math.Pi
			}
			for distance < 0 && sweep >= 0 {
				sweep -= 2 * math.Pi
			}
			steps := max(1, int(math.Ceil(math.Abs(sweep)/step-1e-9)))
			for k := 0; k <= steps; k++ {
				angle := from + sweep*float64(k)/float64(steps)
				curve = append(curve, at.Add(vector2.New(math.Cos(angle), math.Sin(angle)).Scale(radius)))
			}
		default:
			curve = append(curve, at.Add(normalIn), at, at.Add(normalOut))
		}
	}
	return curve
}

type OffsetNode struct {
	Area     nodes.LiftedPort[Area]    `description:"The area to grow or shrink."`
	Distance nodes.LiftedPort[float64] `description:"How far to move every edge outward; negative moves it inward. Corners it grows around come out round."`
	Segments nodes.Output[int]         `description:"Straight segments per half circle at grown corners. Defaults to 8."`
}

func (OffsetNode) Description() string {
	return "Grows or shrinks an area by a distance: a margin around it, or an inset within it. Shrinking can close holes or split the area apart; growing can merge pieces."
}

func (OffsetNode) Keywords() []string {
	return []string{"buffer", "inflate"}
}

func (n OffsetNode) Out(out *nodes.Lifted[Area]) {
	segments := max(nodes.TryGetOutputValue(out, n.Segments, 8), 1)
	nodes.Zip2(out, n.Area, n.Distance, func(area Area, distance float64) Area {
		offset, err := Offset(area, distance, segments)
		out.CaptureError(err)
		return offset
	})
}

type BoundaryNode struct {
	Area nodes.Output[Area] `description:"The area whose loops to return."`
}

func (BoundaryNode) Description() string {
	return "The loops around an area: Outlines run counter-clockwise, Holes clockwise. Each output is a list, ready for a path, a stroke or a sweep."
}

func (n BoundaryNode) loops(out *nodes.StructOutput[[][]vector2.Float64], holes bool) {
	if n.Area == nil {
		return
	}
	var loops [][]vector2.Float64
	for _, polygon := range nodes.GetOutputValue(out, n.Area).Polygons {
		if holes {
			loops = append(loops, polygon.Holes...)
		} else {
			loops = append(loops, polygon.Outline)
		}
	}
	out.Set(loops)
}

func (n BoundaryNode) Outlines(out *nodes.StructOutput[[][]vector2.Float64]) {
	n.loops(out, false)
}

func (n BoundaryNode) Holes(out *nodes.StructOutput[[][]vector2.Float64]) {
	n.loops(out, true)
}
