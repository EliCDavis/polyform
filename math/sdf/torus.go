package sdf

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func Torus(position vector3.Float64, majorRadius, minorRadius float64) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		q := vector2.New(v.XZ().Length()-majorRadius, v.Y())
		return q.Length() - minorRadius
	}
}

type TorusNode struct {
	Position   nodes.Output[vector3.Float64] `description:"Center of the torus. Defaults to the origin."`
	RingRadius nodes.Output[float64]         `description:"The distance from Position to the center of the tube, in the XZ plane. Defaults to 1."`
	TubeRadius nodes.Output[float64]         `description:"The thickness of the ring. Defaults to 0.1."`
}

func (cn TorusNode) Description() string {
	return "A torus lying flat in the XZ plane."
}

func (cn TorusNode) Field(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	out.Set(Torus(
		nodes.TryGetOutputValue(out, cn.Position, vector3.Zero[float64]()),
		nodes.TryGetOutputValue(out, cn.RingRadius, 1),
		nodes.TryGetOutputValue(out, cn.TubeRadius, 0.1),
	))
}
