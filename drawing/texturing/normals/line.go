package normals

import (
	"image/color"
	"image/draw"
	"math"

	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Line struct {
	Start           vector2.Float64
	End             vector2.Float64
	Width           float64
	NormalDirection Direction
}

func (l Line) normalLine(src draw.Image, start, end vector2.Float64) {
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

		p := dir.Scale(i)

		circleY := math.Sqrt(rr-(i*i)) * circleYMultiplier

		pixNormal := vector3.New(p.X(), circleY, p.Y()).
			Normalized().
			MultByVector(vector3.Fill(circleYMultiplier)).
			Scale(127).
			Clamp(-128, 127).
			RoundToInt()

		c := color.RGBA{
			R: byte(128 + pixNormal.X()),
			G: byte(128 - pixNormal.Z()),
			B: byte(128 + pixNormal.Y()),
			A: 255,
		}

		pix := start.Add(dir.Scale(i)).FloorToInt()
		src.Set(pix.X(), pix.Y(), c)
	}
}

// Dir is end point - starting point
func (l Line) Dir() vector2.Float64 {
	return l.End.Sub(l.Start)
}

func (l Line) Round(src draw.Image) {
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
