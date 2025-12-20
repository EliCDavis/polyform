package sdf

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func Torus(position vector3.Float64, minorRadius, majorRadius float64) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		q := vector2.New(v.XZ().Length()-minorRadius, v.Y())
		return q.Length() - majorRadius
	}
}

type TorusNode struct {
	Position    nodes.Output[vector3.Float64]
	MinorRadius nodes.Output[float64]
	MajorRadius nodes.Output[float64]
}

func (cn TorusNode) Field(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	out.Set(Torus(
		nodes.TryGetOutputValue(out, cn.Position, vector3.Zero[float64]()),
		nodes.TryGetOutputValue(out, cn.MinorRadius, 0.1),
		nodes.TryGetOutputValue(out, cn.MajorRadius, 1),
	))
}
