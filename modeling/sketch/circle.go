package sketch

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

// Circle is a regular polygon with a corner on +X, counter-clockwise.
func Circle(radius float64, sides int, center vector2.Float64) geometry.Shape {
	sides = max(sides, 3)
	shape := make(geometry.Shape, sides)
	for i := range shape {
		angle := 2 * math.Pi * float64(i) / float64(sides)
		shape[i] = center.Add(vector2.New(math.Cos(angle), math.Sin(angle)).Scale(math.Abs(radius)))
	}
	return shape
}

type CircleNode struct {
	Radius nodes.LiftedPort[float64] `default:"0.5"`
	Sides  nodes.Output[int]         `default:"32"`
	Center nodes.LiftedPort[vector2.Float64]
}

func (CircleNode) Description() string {
	return "A 2D circle or regular polygon."
}

func (CircleNode) Keywords() []string {
	return []string{"polygon"}
}

func (n CircleNode) Outline(out *nodes.Lifted[[]vector2.Float64]) {
	sides := nodes.GetOutputValue(out, n.Sides)
	nodes.Zip2(out, n.Radius, n.Center, func(radius float64, center vector2.Float64) []vector2.Float64 {
		return Circle(radius, sides, center)
	})
}

func (n CircleNode) Area(out *nodes.Lifted[Area]) {
	sides := nodes.GetOutputValue(out, n.Sides)
	nodes.Zip2(out, n.Radius, n.Center, func(radius float64, center vector2.Float64) Area {
		return enclosed(out, Circle(radius, sides, center))
	})
}
