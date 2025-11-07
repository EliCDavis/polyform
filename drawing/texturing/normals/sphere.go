package normals

import (
	"math"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Sphere struct {
	Center    vector2.Float64
	Radius    float64
	Direction Direction
}

func (s Sphere) Draw(src NormalMap) {

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

	for x := max(bottom.X(), 0); x < min(top.X(), float64(src.Width())); x++ {
		xA := x - a
		xAxA := xA * xA
		for y := max(bottom.Y(), 0); y < min(top.Y(), float64(src.Height())); y++ {
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
				Scale(circleYMultiplier).
				XZY().
				FlipY().
				Clamp(-1, 1)

			src.Set(int(x), int(y), pixNormal)
		}
	}
}

type DrawSpheresNode struct {
	Radii     nodes.Output[[]float64]
	Positions nodes.Output[[]vector2.Float64]
	Subtract  nodes.Output[bool]
	Texture   nodes.Output[NormalMap] `description:"texture to draw on"`
}

func (n DrawSpheresNode) NormalMap(out *nodes.StructOutput[NormalMap]) {
	if n.Texture == nil {
		return
	}
	img := nodes.GetOutputValue(out, n.Texture).Copy()
	dim := vector2.New(img.Width(), img.Height())

	radii := nodes.TryGetOutputValue(out, n.Radii, nil)
	positions := nodes.TryGetOutputValue(out, n.Positions, nil)

	dir := Additive
	if nodes.TryGetOutputValue(out, n.Subtract, false) {
		dir = Subtractive
	}

	size := max(len(radii), len(positions))
	for i := range size {
		radius := 0.5
		if i < len(radii) {
			radius = radii[i]
		}

		position := vector2.New(0.5, 0.5)
		if i < len(positions) {
			position = positions[i]
		}

		s := Sphere{
			Center:    position.MultByVector(dim.ToFloat64()),
			Radius:    float64(dim.MinComponent()) * radius,
			Direction: dir,
		}
		s.Draw(img)
	}

	out.Set(img)
}
