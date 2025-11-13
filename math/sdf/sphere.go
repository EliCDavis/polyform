package sdf

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func Sphere(position vector3.Float64, radius float64) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		return v.Distance(position) - radius
	}
}

type SphereNode struct {
	Position nodes.Output[vector3.Float64]
	Radius   nodes.Output[float64]
}

func (cn SphereNode) Field(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	out.Set(Sphere(
		nodes.TryGetOutputValue(out, cn.Position, vector3.Zero[float64]()),
		nodes.TryGetOutputValue(out, cn.Radius, .5),
	))
}
