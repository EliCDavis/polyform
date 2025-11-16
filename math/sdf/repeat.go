package sdf

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func Repeat(field sample.Vec3ToFloat, transforms []trs.TRS) sample.Vec3ToFloat {
	if len(transforms) == 0 {
		return nullField
	}

	invertedTRS := make([]trs.TRS, len(transforms))
	for i, v := range transforms {
		invertedTRS[i] = trs.FromMatrix(v.Matrix().Inverse())
	}

	return func(v vector3.Float64) float64 {
		closestPoint := field(invertedTRS[0].Transform(v))
		for i := 1; i < len(invertedTRS); i++ {
			closestPoint = min(closestPoint, field(invertedTRS[i].Transform(v)))
		}
		return closestPoint
	}
}

type RepeatNode struct {
	Transforms nodes.Output[[]trs.TRS]
	Field      nodes.Output[sample.Vec3ToFloat]
}

func (cn RepeatNode) Result(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if cn.Field == nil {
		return
	}

	out.Set(Repeat(
		nodes.GetOutputValue(out, cn.Field),
		nodes.TryGetOutputValue(out, cn.Transforms, []trs.TRS{trs.Identity()}),
	))
}
