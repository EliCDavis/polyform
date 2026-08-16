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
	Field nodes.Output[sample.Vec3ToFloat] `description:"The field to mirror. Each output port below reflects it across a different axis/plane through the origin; nothing is set on any output until this is connected."`
}

func (n MirrorNode) Description() string {
	return "Reflects an SDF field across one or more axes through the origin, one output port per axis combination."
}

func (n MirrorNode) X(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field != nil {
		out.Set(MirrorX(nodes.GetOutputValue(out, n.Field)))
	}
}

func (n MirrorNode) XDescription() string {
	return "Mirrors across the YZ plane (reflects the X axis, i.e. flips left/right)."
}

func (n MirrorNode) Y(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field != nil {
		out.Set(MirrorY(nodes.GetOutputValue(out, n.Field)))
	}
}

func (n MirrorNode) YDescription() string {
	return "Mirrors across the XZ plane (reflects the Y axis, i.e. flips up/down)."
}

func (n MirrorNode) Z(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field != nil {
		out.Set(MirrorZ(nodes.GetOutputValue(out, n.Field)))
	}
}

func (n MirrorNode) ZDescription() string {
	return "Mirrors across the XY plane (reflects the Z axis, i.e. flips front/back)."
}

func (n MirrorNode) XY(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field != nil {
		out.Set(MirrorXY(nodes.GetOutputValue(out, n.Field)))
	}
}

func (n MirrorNode) XYDescription() string {
	return "Mirrors across both the X and Y axes at once (four-way symmetry in the XY plane)."
}

func (n MirrorNode) XZ(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field != nil {
		out.Set(MirrorXZ(nodes.GetOutputValue(out, n.Field)))
	}
}

func (n MirrorNode) XZDescription() string {
	return "Mirrors across both the X and Z axes at once (four-way symmetry in the XZ plane)."
}

func (n MirrorNode) YZ(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field != nil {
		out.Set(MirrorYZ(nodes.GetOutputValue(out, n.Field)))
	}
}

func (n MirrorNode) YZDescription() string {
	return "Mirrors across both the Y and Z axes at once (four-way symmetry in the YZ plane)."
}

func (n MirrorNode) XYZ(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field != nil {
		out.Set(MirrorXYZ(nodes.GetOutputValue(out, n.Field)))
	}
}

func (n MirrorNode) XYZDescription() string {
	return "Mirrors across all three axes at once (eight-way symmetry, one copy in every octant)."
}
