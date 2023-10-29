package normals

import (
	"image"
	"image/color"
	"math"

	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Sphere struct {
	Center    vector2.Float64
	Radius    float64
	Direction Direction
}

func (s Sphere) Draw(src *image.RGBA) {

	center3 := vector3.New(s.Center.X(), 0, s.Center.Y())

	corner := vector2.Fill(s.Radius)
	bottom := s.Center.Sub(corner)
	top := s.Center.Add(corner)

	circleYMultiplier := 1.
	if s.Direction == Subtractive {
		circleYMultiplier = -1
	}

	// a = center.X()
	// b := 0
	// c := center.Y()

	// Equation of a Sphere
	// (x - a)^2 + (y - b)^2 + (z - c)^2 = r^2
	// (y - b)^2 = r^2 - (x - a)^2 - (z - c)^2
	// y - b = math.sqrt(r^2 - (x - a)^2 - (z - c)^2)
	// y = math.sqrt(r^2 - (x - a)^2 - (z - c)^2) + b
	// y = math.sqrt(r^2 - (x - a)^2 - (z - c)^2)
	rr := s.Radius * s.Radius
	a := s.Center.X()
	c := s.Center.Y()

	for x := bottom.X(); x < top.X(); x++ {
		xA := x - a
		xAxA := xA * xA
		for y := bottom.Y(); y < top.Y(); y++ {
			pix := vector2.New(x, y)
			dist := pix.Distance(s.Center)
			if dist > s.Radius {
				continue
			}

			zC := y - c
			circleY := math.Sqrt(rr-xAxA-(zC*zC)) * circleYMultiplier
			pixNormal := vector3.New(x, circleY, y).
				Sub(center3).
				Normalized().
				MultByVector(vector3.Fill(circleYMultiplier)).
				Scale(127).
				Clamp(-127, 127)
			c := color.RGBA{
				R: byte(128 + pixNormal.X()),
				G: byte(128 - pixNormal.Z()),
				B: byte(128 + pixNormal.Y()),
				A: 255,
			}
			src.Set(int(x), int(y), c)
		}
	}
}
