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
	Position nodes.Output[vector3.Float64] `description:"Center of the sphere."`
	Radius   nodes.Output[float64]         `description:"Radius of the sphere." default:"0.5"`
}

func (cn SphereNode) Description() string {
	return "Sphere field of a given radius, centered on Position."
}

func (cn SphereNode) Field(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	out.Set(Sphere(
		nodes.GetOutputValue(out, cn.Position),
		nodes.GetOutputValue(out, cn.Radius),
	))
}
