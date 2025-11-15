package normals

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Line struct {
	Start           vector2.Float64
	End             vector2.Float64
	Width           float64
	NormalDirection Direction
}

func (l Line) normalLine(src NormalMap, start, end vector2.Float64) {
	line := end.Sub(start)
	len := line.Length()
	dir := line.Normalized()

	// a = center.X()
	// b := 0
	// c := 0
	// r := len

	circleYMultiplier := 1.
	if l.NormalDirection == Subtractive {
		circleYMultiplier = -1
	}

	// Equation of a Sphere
	// (x - a)^2 + (y - b)^2 + (z - c)^2 = r^2
	// (y - b)^2 = r^2 - (x - a)^2 - (z - c)^2
	// y - b = math.sqrt(r^2 - (x - a)^2 - (z - c)^2)
	// y = math.sqrt(r^2 - (x - a)^2 - (z - c)^2) + b
	// y = math.sqrt(r^2 - (x - a)^2)
	rr := len * len
	for i := len; i >= 0; i -= 0.5 {
		pix := start.Add(dir.Scale(i)).FloorToInt()
		if pix.MinComponent() < 0 || pix.Y() >= src.Height() || pix.X() >= src.Width() {
			continue
		}

		p := dir.Scale(i)

		circleY := math.Sqrt(rr-(i*i)) * circleYMultiplier

		pixNormal := vector3.New(p.X(), circleY, p.Y()).
			Normalized().
			MultByVector(vector3.Fill(circleYMultiplier)).
			XZY().
			FlipY().
			Clamp(-1, 1)

		src.Set(pix.X(), pix.Y(), pixNormal)
	}
}

// Dir is end point - starting point
func (l Line) Dir() vector2.Float64 {
	return l.End.Sub(l.Start)
}

func (l Line) Round(src NormalMap) {
	dir := l.Dir()
	n := dir.Normalized()
	perp := n.Perpendicular().Scale(l.Width / 2)
	length := dir.Length()
	start := l.Start
	for i := 0.; i < length; i += 0.5 {
		newStart := start.Add(n.Scale(float64(i)))
		l.normalLine(src, newStart, newStart.Add(perp))
		l.normalLine(src, newStart, newStart.Sub(perp))
	}
}

type DrawLinesNode struct {
	Thicknesses nodes.Output[[]float64]
	Lines       nodes.Output[[]geometry.Line2D]
	Texture     nodes.Output[NormalMap] `description:"texture to draw on"`
}

func (n DrawLinesNode) NormalMap(out *nodes.StructOutput[NormalMap]) {
	if n.Texture == nil {
		return
	}
	img := nodes.GetOutputValue(out, n.Texture).Copy()
	dim := vector2.New(img.Width(), img.Height()).ToFloat64()

	thicknesses := nodes.TryGetOutputValue(out, n.Thicknesses, nil)
	lines := nodes.TryGetOutputValue(out, n.Lines, nil)

	for i := range len(lines) {
		radius := 0.5
		if i < len(thicknesses) {
			radius = thicknesses[i]
		}

		line := lines[i]

		s := Line{
			Start:           line.GetStartPoint().MultByVector(dim),
			End:             line.GetEndPoint().MultByVector(dim),
			Width:           radius * float64(dim.MinComponent()),
			NormalDirection: Additive,
		}
		s.Round(img)
	}

	out.Set(img)
}

type DrawLineNode struct {
	Thicknesses nodes.Output[float64]
	Start       nodes.Output[vector2.Float64]
	End         nodes.Output[vector2.Float64]
	Subtract    nodes.Output[bool]
	Texture     nodes.Output[NormalMap] `description:"texture to draw on"`
}

func (n DrawLineNode) NormalMap(out *nodes.StructOutput[NormalMap]) {
	if n.Texture == nil {
		return
	}
	img := nodes.GetOutputValue(out, n.Texture).Copy()
	dim := vector2.New(img.Width(), img.Height()).ToFloat64()

	start := nodes.TryGetOutputValue(out, n.Start, vector2.New(0., 0.))
	end := nodes.TryGetOutputValue(out, n.End, vector2.New(1., 1.))

	radius := nodes.TryGetOutputValue(out, n.Thicknesses, 0.5)

	dir := Additive
	if nodes.TryGetOutputValue(out, n.Subtract, false) {
		dir = Subtractive
	}

	s := Line{
		Start:           start.MultByVector(dim),
		End:             end.MultByVector(dim),
		Width:           radius * float64(dim.MinComponent()),
		NormalDirection: dir,
	}
	s.Round(img)

	out.Set(img)
}
