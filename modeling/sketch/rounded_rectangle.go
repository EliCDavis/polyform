package sketch

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

func RoundedRectangle(size vector2.Float64, radius float64, center vector2.Float64, segments int) geometry.Shape {
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

	shape := make(geometry.Shape, 0, 4*(segments+1))
	for k, corner := range corners {
		for j := 0; j <= segments; j++ {
			if j == 0 && corner.straightIn == 0 && segments > 0 {
				continue
			}
			angle := (float64(k) - 1 + float64(j)/math.Max(float64(segments), 1)) * math.Pi / 2
			shape = append(shape, center.Add(corner.at).Add(vector2.New(math.Cos(angle), math.Sin(angle)).Scale(radius)))
		}
	}
	return shape
}

type RoundedRectangleNode struct {
	Size     nodes.LiftedPort[vector2.Float64] `description:"Full width and height." default:"1,1"`
	Radius   nodes.LiftedPort[float64]         `description:"Corner radius, at most half the shorter side."`
	Center   nodes.LiftedPort[vector2.Float64] `description:"Middle of the rectangle."`
	Segments nodes.Output[int]                 `description:"Straight segments per corner." default:"8"`
}

func (RoundedRectangleNode) Description() string {
	return "A closed 2D rounded-rectangle outline."
}

func (RoundedRectangleNode) Keywords() []string {
	return []string{"fillet"}
}

func (n RoundedRectangleNode) Out(out *nodes.Lifted[[]vector2.Float64]) {
	segments := max(nodes.GetOutputValue(out, n.Segments), 1)
	nodes.Zip3(
		out,
		n.Size,
		n.Radius,
		n.Center,
		func(size vector2.Float64, radius float64, center vector2.Float64) []vector2.Float64 {
			return RoundedRectangle(size, radius, center, segments)
		},
	)
}
