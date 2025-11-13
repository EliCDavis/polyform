package sdf

import (
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
	Position nodes.Output[vector3.Float64]
	Field    nodes.Output[sample.Vec3ToFloat]
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

func Transform(field sample.Vec3ToFloat, transformation trs.TRS) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		return field(transformation.Transform(v))
	}
}

type TransformNode struct {
	Transform nodes.Output[trs.TRS]
	Field     nodes.Output[sample.Vec3ToFloat]
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
