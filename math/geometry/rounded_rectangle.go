package geometry

import (
	"math"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

func RoundedRectangle(size vector2.Float64, radius float64, center vector2.Float64, segments int) Shape {
	half := vector2.New(math.Abs(size.X())/2, math.Abs(size.Y())/2)
	radius = math.Min(math.Max(radius, 0), math.Min(half.X(), half.Y()))
	inner := half.Sub(vector2.New(radius, radius))

	corners := []struct {
		at         vector2.Float64
		straightIn float64
	}{
		{vector2.New(inner.X(), -inner.Y()), inner.X()},
		{vector2.New(inner.X(), inner.Y()), inner.Y()},
		{vector2.New(-inner.X(), inner.Y()), inner.X()},
		{vector2.New(-inner.X(), -inner.Y()), inner.Y()},
	}
	if radius == 0 {
		segments = 0
	}

	outline := make(Shape, 0, 4*(segments+1))
	for k, corner := range corners {
		for j := 0; j <= segments; j++ {
			if j == 0 && corner.straightIn == 0 && segments > 0 {
				continue
			}
			angle := (float64(k) - 1 + float64(j)/math.Max(float64(segments), 1)) * math.Pi / 2
			outline = append(outline, center.Add(corner.at).Add(vector2.New(math.Cos(angle), math.Sin(angle)).Scale(radius)))
		}
	}
	return outline
}

type RoundedRectangleNode struct {
	Size     nodes.LiftedPort[vector2.Float64] `description:"Full width and height. Defaults to 1 by 1."`
	Radius   nodes.LiftedPort[float64]         `description:"Corner radius, at most half the shorter side. Defaults to 0."`
	Center   nodes.LiftedPort[vector2.Float64] `description:"Defaults to the origin."`
	Segments nodes.Output[int]                 `description:"Straight segments per corner. Defaults to 8."`
}

func (RoundedRectangleNode) Description() string {
	return "A closed 2D rounded-rectangle outline. An array of sizes, radii or centers gives one outline each."
}

func (RoundedRectangleNode) Keywords() []string {
	return []string{"rectangle", "rounded", "fillet", "outline"}
}

func (n RoundedRectangleNode) Out(out *nodes.Lifted[[]vector2.Float64]) {
	segments := max(nodes.TryGetOutputValue(out, n.Segments, 8), 1)
	nodes.Zip3(
		out,
		nodes.LiftedOr(n.Size, vector2.One[float64]()),
		nodes.LiftedOr(n.Radius, 0.),
		nodes.LiftedOr(n.Center, vector2.Zero[float64]()),
		func(size vector2.Float64, radius float64, center vector2.Float64) []vector2.Float64 {
			return RoundedRectangle(size, radius, center, segments)
		},
	)
}
