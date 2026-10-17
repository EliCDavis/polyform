package sketch

import (
	"math"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

func Transform(shape []vector2.Float64, translation vector2.Float64, degrees float64, scale vector2.Float64) []vector2.Float64 {
	sin, cos := math.Sincos(degrees * math.Pi / 180)
	moved := make([]vector2.Float64, len(shape))
	for i, p := range shape {
		scaled := vector2.New(p.X()*scale.X(), p.Y()*scale.Y())
		moved[i] = vector2.New(
			scaled.X()*cos-scaled.Y()*sin,
			scaled.X()*sin+scaled.Y()*cos,
		).Add(translation)
	}
	return moved
}

type TransformNode struct {
	Outline     nodes.LiftedPort[[]vector2.Float64] `description:"The outline or path to move."`
	Translation nodes.LiftedPort[vector2.Float64]   `description:"Applied last. Defaults to no move."`
	Rotation    nodes.LiftedPort[float64]           `description:"Degrees counter-clockwise about the origin. Defaults to 0."`
	Scale       nodes.LiftedPort[vector2.Float64]   `description:"Applied first, about the origin. Defaults to 1 by 1."`
}

func (TransformNode) Description() string {
	return "Scales, rotates about the origin, then moves a 2D outline or path. Lists give one each."
}

func (TransformNode) Keywords() []string {
	return []string{"rotate", "translate", "move", "scale", "place"}
}

func (n TransformNode) Out(out *nodes.Lifted[[]vector2.Float64]) {
	nodes.Zip4(
		out,
		n.Outline,
		nodes.LiftedOr(n.Translation, vector2.Zero[float64]()),
		n.Rotation,
		nodes.LiftedOr(n.Scale, vector2.One[float64]()),
		Transform,
	)
}
