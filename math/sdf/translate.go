package sdf

import (
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func Translate(field sample.Vec3ToFloat, translation vector3.Float64) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		return field(v.Sub(translation))
	}
}

type TranslateNode struct {
	Position nodes.Output[vector3.Float64]    `description:"Offset to shift the field by. Defaults to no movement."`
	Field    nodes.Output[sample.Vec3ToFloat] `description:"The field to move."`
}

func (cn TranslateNode) Description() string {
	return "Moves an SDF field by a fixed offset."
}

func (cn TranslateNode) Result(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if cn.Field == nil {
		return
	}

	out.Set(Translate(
		nodes.GetOutputValue(out, cn.Field),
		nodes.TryGetOutputValue(out, cn.Position, vector3.Zero[float64]()),
	))
}

// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

// Transform maps the query point into the field's local space and scales
// the distance that comes back into world units.
//
// Without that last step the field reports local distances: a field scaled
// to a quarter size answers -1 at its center where the truth is -0.25, a
// gradient of 4 where a distance field must never exceed 1. Marching cubes
// interpolates the surface position linearly between voxel corners and a
// smooth union compares two fields' values, so both read a field like that
// as noise - speckled, degenerate geometry rather than a smaller shape.
//
// A non-uniform scale has no exact distance correction, since the true
// distance depends on direction. The smallest scale component is the
// standard conservative choice: it under-estimates rather than
// over-estimates, which keeps the gradient at or below 1 and so keeps the
// field safe to march and to blend.
func Transform(field sample.Vec3ToFloat, transformation trs.TRS) sample.Vec3ToFloat {
	inverse := transformation.Inverse()
	scale := transformation.Scale()
	factor := math.Min(math.Abs(scale.X()), math.Min(math.Abs(scale.Y()), math.Abs(scale.Z())))
	return func(v vector3.Float64) float64 {
		return field(inverse.Transform(v)) * factor
	}
}

type TransformNode struct {
	Transform nodes.Output[trs.TRS]            `description:"The translation/rotation/scale to apply. Defaults to identity (no change)."`
	Field     nodes.Output[sample.Vec3ToFloat] `description:"The field to transform."`
}

func (cn TransformNode) Description() string {
	return "Moves/rotates/scales an SDF field by a TRS transform."
}

func (cn TransformNode) Result(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if cn.Field == nil {
		return
	}

	out.Set(Transform(
		nodes.GetOutputValue(out, cn.Field),
		nodes.TryGetOutputValue(out, cn.Transform, trs.Identity()),
	))
}
