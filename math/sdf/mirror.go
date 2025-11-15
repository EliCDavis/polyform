package sdf

import (
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func MirrorX(f sample.Vec3ToFloat) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		return f(vector3.New(math.Abs(v.X()), v.Y(), v.Z()))
	}
}

func MirrorY(f sample.Vec3ToFloat) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		return f(vector3.New(v.X(), math.Abs(v.Y()), v.Z()))
	}
}

func MirrorZ(f sample.Vec3ToFloat) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		return f(vector3.New(v.X(), v.Y(), math.Abs(v.Z())))
	}
}

func MirrorXY(f sample.Vec3ToFloat) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		return f(vector3.New(math.Abs(v.X()), math.Abs(v.Y()), v.Z()))
	}
}

func MirrorYZ(f sample.Vec3ToFloat) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		return f(vector3.New(v.X(), math.Abs(v.Y()), math.Abs(v.Z())))
	}
}
func MirrorXZ(f sample.Vec3ToFloat) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		return f(vector3.New(math.Abs(v.X()), v.Y(), math.Abs(v.Z())))
	}
}

func MirrorXYZ(f sample.Vec3ToFloat) sample.Vec3ToFloat {
	return func(v vector3.Float64) float64 {
		return f(v.Abs())
	}
}

type MirrorNode struct {
	Field nodes.Output[sample.Vec3ToFloat]
}

func (n MirrorNode) X(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field != nil {
		out.Set(MirrorX(nodes.GetOutputValue(out, n.Field)))
	}
}

func (n MirrorNode) Y(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field != nil {
		out.Set(MirrorY(nodes.GetOutputValue(out, n.Field)))
	}
}

func (n MirrorNode) Z(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field != nil {
		out.Set(MirrorZ(nodes.GetOutputValue(out, n.Field)))
	}
}

func (n MirrorNode) XY(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field != nil {
		out.Set(MirrorXY(nodes.GetOutputValue(out, n.Field)))
	}
}

func (n MirrorNode) XZ(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field != nil {
		out.Set(MirrorXZ(nodes.GetOutputValue(out, n.Field)))
	}
}

func (n MirrorNode) YZ(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field != nil {
		out.Set(MirrorYZ(nodes.GetOutputValue(out, n.Field)))
	}
}

func (n MirrorNode) XYZ(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field != nil {
		out.Set(MirrorXYZ(nodes.GetOutputValue(out, n.Field)))
	}
}
