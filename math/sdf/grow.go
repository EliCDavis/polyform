package sdf

import (
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func Grow(field sample.Vec3ToFloat, distance float64) sample.Vec3ToFloat {
	if distance == 0 {
		return field
	}

	return func(p vector3.Float64) float64 {
		return field(p) - distance
	}
}

func Shell(field sample.Vec3ToFloat, thickness float64) sample.Vec3ToFloat {
	half := thickness / 2
	return func(p vector3.Float64) float64 {
		return math.Abs(field(p)) - half
	}
}

type GrowNode struct {
	Field    nodes.Output[sample.Vec3ToFloat] `description:"The field to grow."`
	Distance nodes.Output[float64]            `description:"World units to move the surface along its normal. Positive grows the shape, negative shrinks it. Defaults to 0."`
}

func (n GrowNode) Description() string {
	return "Moves a field's surface outward or inward by a fixed distance, keeping its shape."
}

func (n GrowNode) Keywords() []string {
	return []string{"offset", "inflate"}
}

func (n GrowNode) Grow(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field == nil {
		return
	}
	out.Set(Grow(nodes.GetOutputValue(out, n.Field), nodes.TryGetOutputValue(out, n.Distance, 0)))
}

type ShellNode struct {
	Field     nodes.Output[sample.Vec3ToFloat] `description:"The field whose surface becomes a hollow shell."`
	Thickness nodes.Output[float64]            `description:"Total wall thickness in world units, centered on the original surface. Defaults to 0.01."`
}

func (n ShellNode) Description() string {
	return "Turns a solid into a thin hollow wall centered on its surface. Cut it with a subtraction or intersection to get open edges."
}

func (n ShellNode) Shell(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field == nil {
		return
	}
	out.Set(Shell(nodes.GetOutputValue(out, n.Field), nodes.TryGetOutputValue(out, n.Thickness, 0.01)))
}
