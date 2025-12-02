package sdf

import (
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func CutSphere(position vector3.Float64, radius, cutDistance float64) sample.Vec3ToFloat {
	w := math.Sqrt((radius * radius) - (cutDistance * cutDistance))
	return func(v vector3.Float64) float64 {
		q := vector2.New(v.XZ().Length(), v.Y())
		s := math.Max((cutDistance-radius)*q.X()*q.X()+w*w*(cutDistance+radius-2.0*q.Y()), cutDistance*q.X()-w*q.Y())

		if s < 0.0 {
			return q.Length() - radius
		}

		if q.X() < w {
			return q.Y() - q.Y()
		}

		return q.Sub(vector2.New(w, cutDistance)).Length()
	}
}

type CutSphereNode struct {
	Position    nodes.Output[vector3.Float64]
	Radius      nodes.Output[float64]
	CutDistance nodes.Output[float64]
}

func (cn CutSphereNode) Field(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	out.Set(CutSphere(
		nodes.TryGetOutputValue(out, cn.Position, vector3.Zero[float64]()),
		nodes.TryGetOutputValue(out, cn.Radius, .5),
		nodes.TryGetOutputValue(out, cn.CutDistance, .25),
	))
}
