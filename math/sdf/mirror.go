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

// mirrorUnion reflects f across every axis flagged true and unions every
// resulting copy together - evaluating f at each sign combination of the
// mirrored axes and taking the minimum - rather than folding a query onto
// one canonical side before evaluating f once. The fold (MirrorX etc.
// above) silently discards whatever f actually defines on the non-
// canonical side; this preserves it, at the cost of evaluating f once per
// mirrored copy (2 calls for one axis, 4 for two, 8 for three) instead of
// once.
func mirrorUnion(f sample.Vec3ToFloat, mirrorX, mirrorY, mirrorZ bool) sample.Vec3ToFloat {
	xSigns := []float64{1}
	if mirrorX {
		xSigns = append(xSigns, -1)
	}
	ySigns := []float64{1}
	if mirrorY {
		ySigns = append(ySigns, -1)
	}
	zSigns := []float64{1}
	if mirrorZ {
		zSigns = append(zSigns, -1)
	}

	return func(v vector3.Float64) float64 {
		result := math.Inf(1)
		for _, sx := range xSigns {
			for _, sy := range ySigns {
				for _, sz := range zSigns {
					result = math.Min(result, f(vector3.New(v.X()*sx, v.Y()*sy, v.Z()*sz)))
				}
			}
		}
		return result
	}
}

type MirrorNode struct {
	Field nodes.Output[sample.Vec3ToFloat] `description:"The field to mirror. Each output port below reflects it across a different axis/plane through the origin; nothing is set on any output until this is connected."`
	Union nodes.Output[bool]               `description:"When true (the default), every reflected copy is unioned with the original by evaluating the field on every mirrored side and combining the results, so real content already present on more than one side of the mirror plane is preserved instead of being silently overwritten. When false: cheaper but less safe - a query is folded onto one canonical side before the field is evaluated once, which discards whatever the field actually defines on the other side. Only set this false when the field being mirrored is known to have content on one side of every mirrored axis only, e.g. a single limb built entirely on the positive side."`
}

func (n MirrorNode) Description() string {
	return "Reflects an SDF field across one or more axes through the origin, one output port per axis combination."
}

func (n MirrorNode) union(out nodes.ExecutionRecorder) bool {
	return nodes.TryGetOutputValue(out, n.Union, true)
}

func (n MirrorNode) X(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field == nil {
		return
	}
	field := nodes.GetOutputValue(out, n.Field)
	if n.union(out) {
		out.Set(mirrorUnion(field, true, false, false))
	} else {
		out.Set(MirrorX(field))
	}
}

func (n MirrorNode) XDescription() string {
	return "Mirrors across the YZ plane (reflects the X axis, i.e. flips left/right)."
}

func (n MirrorNode) Y(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field == nil {
		return
	}
	field := nodes.GetOutputValue(out, n.Field)
	if n.union(out) {
		out.Set(mirrorUnion(field, false, true, false))
	} else {
		out.Set(MirrorY(field))
	}
}

func (n MirrorNode) YDescription() string {
	return "Mirrors across the XZ plane (reflects the Y axis, i.e. flips up/down)."
}

func (n MirrorNode) Z(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field == nil {
		return
	}
	field := nodes.GetOutputValue(out, n.Field)
	if n.union(out) {
		out.Set(mirrorUnion(field, false, false, true))
	} else {
		out.Set(MirrorZ(field))
	}
}

func (n MirrorNode) ZDescription() string {
	return "Mirrors across the XY plane (reflects the Z axis, i.e. flips front/back)."
}

func (n MirrorNode) XY(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field == nil {
		return
	}
	field := nodes.GetOutputValue(out, n.Field)
	if n.union(out) {
		out.Set(mirrorUnion(field, true, true, false))
	} else {
		out.Set(MirrorXY(field))
	}
}

func (n MirrorNode) XYDescription() string {
	return "Mirrors across both the X and Y axes at once (four-way symmetry in the XY plane)."
}

func (n MirrorNode) XZ(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field == nil {
		return
	}
	field := nodes.GetOutputValue(out, n.Field)
	if n.union(out) {
		out.Set(mirrorUnion(field, true, false, true))
	} else {
		out.Set(MirrorXZ(field))
	}
}

func (n MirrorNode) XZDescription() string {
	return "Mirrors across both the X and Z axes at once (four-way symmetry in the XZ plane)."
}

func (n MirrorNode) YZ(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field == nil {
		return
	}
	field := nodes.GetOutputValue(out, n.Field)
	if n.union(out) {
		out.Set(mirrorUnion(field, false, true, true))
	} else {
		out.Set(MirrorYZ(field))
	}
}

func (n MirrorNode) YZDescription() string {
	return "Mirrors across both the Y and Z axes at once (four-way symmetry in the YZ plane)."
}

func (n MirrorNode) XYZ(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Field == nil {
		return
	}
	field := nodes.GetOutputValue(out, n.Field)
	if n.union(out) {
		out.Set(mirrorUnion(field, true, true, true))
	} else {
		out.Set(MirrorXYZ(field))
	}
}

func (n MirrorNode) XYZDescription() string {
	return "Mirrors across all three axes at once (eight-way symmetry, one copy in every octant)."
}
