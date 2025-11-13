package sdf

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

// https://iquilezles.org/articles/distfunctions/
func Plane(position, normal vector3.Float64, height float64) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		return v.Sub(position).Dot(normal) + height
	}
}

type PlaneNode struct {
	Position nodes.Output[vector3.Float64]
	Normal   nodes.Output[vector3.Float64]
	Height   nodes.Output[float64]
}

func (cn PlaneNode) Field(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	out.Set(Plane(
		nodes.TryGetOutputValue(out, cn.Position, vector3.Zero[float64]()),
		nodes.TryGetOutputValue(out, cn.Normal, vector3.Up[float64]()),
		nodes.TryGetOutputValue(out, cn.Height, .0),
	))
}
