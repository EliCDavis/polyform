package sdf

import (
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/math/quaternion"
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

func Rotate(field sample.Vec3ToFloat, rotation quaternion.Quaternion) sample.Vec3ToFloat {
	q := rotation.Normalize()
	inverse := quaternion.New(q.Dir().Scale(-1), q.W())
	return func(v vector3.Float64) float64 {
		return field(inverse.Rotate(v))
	}
}

type RotateNode struct {
	Rotation nodes.Output[quaternion.Quaternion] `description:"Rotation to apply about the origin. Defaults to no rotation."`
	Field    nodes.Output[sample.Vec3ToFloat]    `description:"The field to rotate."`
}

func (cn RotateNode) Description() string {
	return "Rotates an SDF field about the origin."
}

func (cn RotateNode) Result(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if cn.Field == nil {
		return
	}

	out.Set(Rotate(
		nodes.GetOutputValue(out, cn.Field),
		nodes.TryGetOutputValue(out, cn.Rotation, quaternion.Identity()),
	))
}

// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

func Scale(field sample.Vec3ToFloat, scale vector3.Float64) (sample.Vec3ToFloat, error) {
	factor := math.Min(math.Abs(scale.X()), math.Min(math.Abs(scale.Y()), math.Abs(scale.Z())))
	if factor == 0 {
		return nil, fmt.Errorf("scale %v collapses an axis, leaving no volume behind", scale)
	}

	inverse := vector3.New(1/scale.X(), 1/scale.Y(), 1/scale.Z())
	return func(v vector3.Float64) float64 {
		return field(v.MultByVector(inverse)) * factor
	}, nil
}

type ScaleNode struct {
	Scale nodes.Output[vector3.Float64]    `description:"Per-axis scale about the origin. Equal components keep distances exact; unequal ones report the shortest distance, which stays safe to march and blend. A zero component is an error. Defaults to no scaling."`
	Field nodes.Output[sample.Vec3ToFloat] `description:"The field to scale."`
}

func (cn ScaleNode) Description() string {
	return "Scales an SDF field about the origin, per axis."
}

func (cn ScaleNode) Result(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if cn.Field == nil {
		return
	}

	scaled, err := Scale(
		nodes.GetOutputValue(out, cn.Field),
		nodes.TryGetOutputValue(out, cn.Scale, vector3.One[float64]()),
	)
	if err != nil {
		out.CaptureError(err)
		return
	}
	out.Set(scaled)
}

// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

func Transform(field sample.Vec3ToFloat, transformation trs.TRS) sample.Vec3ToFloat {
	inverse := transformation.Inverse()
	scale := transformation.Scale()
	factor := math.Min(math.Abs(scale.X()), math.Min(math.Abs(scale.Y()), math.Abs(scale.Z())))
	return func(v vector3.Float64) float64 {
		return field(inverse.MulPosition(v)) * factor
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
