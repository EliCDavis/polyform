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
	Position nodes.Output[vector3.Float64] `description:"A point the plane passes through."`
	Normal   nodes.Output[vector3.Float64] `description:"Direction the plane faces. The side it points toward is 'outside'." default:"0,1,0"`
	Height   nodes.Output[float64]         `description:"Offset along Normal, added to the signed distance."`
}

func (cn PlaneNode) Description() string {
	return "An infinite flat plane."
}

func (cn PlaneNode) Field(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	out.Set(Plane(
		nodes.GetOutputValue(out, cn.Position),
		nodes.GetOutputValue(out, cn.Normal).Normalized(),
		nodes.GetOutputValue(out, cn.Height),
	))
}
