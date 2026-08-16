package sdf

import (
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func Repeat(field sample.Vec3ToFloat, transforms []trs.TRS, radius float64) sample.Vec3ToFloat {
	if len(transforms) == 0 {
		return nullField
	}

	if len(transforms) == 1 {
		inverted := trs.FromMatrix(transforms[0].Matrix().Inverse())
		return func(f vector3.Float64) float64 {
			return field(inverted.Transform(f))
		}
	}

	invertedTRS := make([]trs.TRS, len(transforms))
	for i, v := range transforms {
		invertedTRS[i] = trs.FromMatrix(v.Matrix().Inverse())
	}

	if radius <= 0 {
		return func(v vector3.Float64) float64 {
			closestPoint := field(invertedTRS[0].Transform(v))
			for i := 1; i < len(invertedTRS); i++ {
				closestPoint = min(closestPoint, field(invertedTRS[i].Transform(v)))
			}
			return closestPoint
		}
	}

	return func(v vector3.Float64) float64 {
		min1, min2 := math.Inf(1), math.Inf(1)
		for i := range invertedTRS {
			val := field(invertedTRS[i].Transform(v))
			if val < min1 {
				min1, min2 = val, min1
			} else if val < min2 {
				min2 = val
			}
		}
		return smoothUnionBlend(min1, min2, radius)
	}

}

type RepeatNode struct {
	Transforms nodes.Output[[]trs.TRS]          `description:"One TRS per copy to place. Defaults to a single identity transform (one unmoved copy). An empty array produces an empty field (nothing)."`
	Field      nodes.Output[sample.Vec3ToFloat] `description:"The field to repeat. Nothing is set on the output until this is connected."`
	Radius     nodes.Output[float64]            `description:"Width of the blend region in world units. Zero or less is a regular union. Defaults to 0."`
}

func (cn RepeatNode) Description() string {
	return "Places one copy of an SDF field at each given transform and unions them together."
}

func (cn RepeatNode) Result(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if cn.Field == nil {
		return
	}

	out.Set(Repeat(
		nodes.GetOutputValue(out, cn.Field),
		nodes.TryGetOutputValue(out, cn.Transforms, []trs.TRS{trs.Identity()}),
		nodes.TryGetOutputValue(out, cn.Radius, 0),
	))
}
