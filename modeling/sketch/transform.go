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
	Translation nodes.LiftedPort[vector2.Float64]   `description:"Applied last."`
	Rotation    nodes.LiftedPort[float64]           `description:"Degrees counter-clockwise about the origin."`
	Scale       nodes.LiftedPort[vector2.Float64]   `description:"Applied first, about the origin." default:"1,1"`
}

func (TransformNode) Description() string {
	return "Scales, rotates about the origin, then moves a 2D outline or path."
}

func (n TransformNode) Out(out *nodes.Lifted[[]vector2.Float64]) {
	nodes.Zip4(
		out,
		n.Outline,
		n.Translation,
		n.Rotation,
		n.Scale,
		Transform,
	)
}
