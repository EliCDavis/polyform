package sdf

import (
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// https://iquilezles.org/articles/distfunctions/
func RoundedCylinder(pos vector3.Float64, radius, roundingRadius, bodyHeight float64) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		p := v.Sub(pos)
		d := vector2.New(p.XZ().Length()-radius+roundingRadius, math.Abs(p.Y())-bodyHeight)
		return math.Min(math.Max(d.X(), d.Y()), 0.0) + vector2.New(math.Max(d.X(), 0.0), math.Max(d.Y(), 0.0)).Length() - roundingRadius
	}
}

type RoundedCylinderNode struct {
	Position       nodes.Output[vector3.Float64] `description:"Center of the cylinder. Defaults to the origin."`
	Radius         nodes.Output[float64]         `description:"Radius of the cylinder's body. Defaults to 0.5."`
	RoundingRadius nodes.Output[float64]         `description:"Radius of the fillet rounding the top/bottom rim edges. Defaults to 0.25."`
	BodyHeight     nodes.Output[float64]         `description:"Half the cylinder's total height (distance from Position to each flat cap along Y). Defaults to 1."`
}

func (cn RoundedCylinderNode) Description() string {
	return "A cylinder standing along the Y axis with rounded top/bottom rim edges."
}

func (cn RoundedCylinderNode) Field(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	out.Set(RoundedCylinder(
		nodes.TryGetOutputValue(out, cn.Position, vector3.Zero[float64]()),
		nodes.TryGetOutputValue(out, cn.Radius, .5),
		nodes.TryGetOutputValue(out, cn.RoundingRadius, .25),
		nodes.TryGetOutputValue(out, cn.BodyHeight, 1.),
	))
}
