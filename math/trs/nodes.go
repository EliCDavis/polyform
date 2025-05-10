package trs

import (
	"math"
	"math/rand/v2"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[ArrayNode](factory)
	refutil.RegisterType[NewNode](factory)
	refutil.RegisterType[RotateDirectionNode](factory)
	refutil.RegisterType[RotateDirectionsNode](factory)
	refutil.RegisterType[RandomizeArrayNode](factory)
	refutil.RegisterType[TransformArrayNode](factory)
	refutil.RegisterType[MultiplyNode](factory)
	refutil.RegisterType[MultiplyToArrayNode](factory)
	refutil.RegisterType[MultiplyArrayNode](factory)
	refutil.RegisterType[SelectNode](factory)
	refutil.RegisterType[SelectArrayNode](factory)
	refutil.RegisterType[nodes.Struct[FilterNode]](factory)
	generator.RegisterTypes(factory)
}

// ============================================================================

type NewNode = nodes.Struct[NewNodeData]

type NewNodeData struct {
	Position nodes.Output[vector3.Float64]
	Rotation nodes.Output[quaternion.Quaternion]
	Scale    nodes.Output[vector3.Float64]
}

func (tnd NewNodeData) Out() nodes.StructOutput[TRS] {
	return nodes.NewStructOutput(New(
		nodes.TryGetOutputValue(tnd.Position, vector3.Zero[float64]()),
		nodes.TryGetOutputValue(tnd.Rotation, quaternion.Identity()),
		nodes.TryGetOutputValue(tnd.Scale, vector3.One[float64]()),
	))
}

// ============================================================================

type RandomizeArrayNode = nodes.Struct[RandomizeArrayNodeData]

type RandomizeArrayNodeData struct {
	TranslationMinimum nodes.Output[vector3.Float64]
	TranslationMaximum nodes.Output[vector3.Float64]
	ScaleMinimum       nodes.Output[vector3.Float64]
	ScaleMaximum       nodes.Output[vector3.Float64]
	RotationMinimum    nodes.Output[vector3.Float64]
	RotationMaximum    nodes.Output[vector3.Float64]
	Samples            nodes.Output[int]
}

func (tnd RandomizeArrayNodeData) Out() nodes.StructOutput[[]TRS] {
	minT := nodes.TryGetOutputValue(tnd.TranslationMinimum, vector3.Zero[float64]())
	maxT := nodes.TryGetOutputValue(tnd.TranslationMaximum, vector3.Zero[float64]())
	rangeT := maxT.Sub(minT)

	minS := nodes.TryGetOutputValue(tnd.ScaleMinimum, vector3.One[float64]())
	maxS := nodes.TryGetOutputValue(tnd.ScaleMaximum, vector3.One[float64]())
	rangeS := maxS.Sub(minS)

	minR := nodes.TryGetOutputValue(tnd.RotationMinimum, vector3.Zero[float64]())
	maxR := nodes.TryGetOutputValue(tnd.RotationMaximum, vector3.Zero[float64]())
	rangeR := maxR.Sub(minR)

	samples := make([]TRS, max(nodes.TryGetOutputValue(tnd.Samples, 0), 0))
	for i := range samples {
		samples[i] = New(
			minT.Add(vector3.New(
				rangeT.X()*rand.Float64(),
				rangeT.Y()*rand.Float64(),
				rangeT.Z()*rand.Float64(),
			)),
			quaternion.FromEulerAngle(minR.Add(vector3.New(
				rangeR.X()*rand.Float64(),
				rangeR.Y()*rand.Float64(),
				rangeR.Z()*rand.Float64(),
			))),
			// quaternion.Identity(),
			minS.Add(vector3.New(
				rangeS.X()*rand.Float64(),
				rangeS.Y()*rand.Float64(),
				rangeS.Z()*rand.Float64(),
			)),
		)
	}

	return nodes.NewStructOutput(samples)
}

// ============================================================================

type SelectNode = nodes.Struct[SelectNodeData]

type SelectNodeData struct {
	TRS nodes.Output[TRS]
}

func (tnd SelectNodeData) Position() nodes.StructOutput[vector3.Float64] {
	return nodes.NewStructOutput(nodes.TryGetOutputValue(tnd.TRS, Identity()).Position())
}

func (tnd SelectNodeData) Scale() nodes.StructOutput[vector3.Float64] {
	return nodes.NewStructOutput(nodes.TryGetOutputValue(tnd.TRS, Identity()).Scale())
}

func (tnd SelectNodeData) Rotation() nodes.StructOutput[quaternion.Quaternion] {
	return nodes.NewStructOutput(nodes.TryGetOutputValue(tnd.TRS, Identity()).Rotation())
}

// ============================================================================

type SelectArrayNode = nodes.Struct[SelectArrayNodeData]

type SelectArrayNodeData struct {
	TRS nodes.Output[[]TRS]
}

func (tnd SelectArrayNodeData) Position() nodes.StructOutput[[]vector3.Float64] {
	trss := nodes.TryGetOutputValue(tnd.TRS, nil)
	out := make([]vector3.Float64, len(trss))

	for i, trs := range trss {
		out[i] = trs.Position()
	}

	return nodes.NewStructOutput(out)
}

func (tnd SelectArrayNodeData) Scale() nodes.StructOutput[[]vector3.Float64] {
	trss := nodes.TryGetOutputValue(tnd.TRS, nil)
	out := make([]vector3.Float64, len(trss))

	for i, trs := range trss {
		out[i] = trs.Scale()
	}

	return nodes.NewStructOutput(out)
}

func (tnd SelectArrayNodeData) Rotation() nodes.StructOutput[[]quaternion.Quaternion] {
	trss := nodes.TryGetOutputValue(tnd.TRS, nil)
	out := make([]quaternion.Quaternion, len(trss))

	for i, trs := range trss {
		out[i] = trs.Rotation()
	}

	return nodes.NewStructOutput(out)
}

// ============================================================================

type MultiplyNode = nodes.Struct[MultiplyNodeData]

type MultiplyNodeData struct {
	A nodes.Output[TRS]
	B nodes.Output[TRS]
}

func (tnd MultiplyNodeData) Out() nodes.StructOutput[TRS] {
	a := nodes.TryGetOutputValue(tnd.A, Identity())
	b := nodes.TryGetOutputValue(tnd.B, Identity())
	return nodes.NewStructOutput(a.Multiply(b))
}

// ============================================================================

type MultiplyArrayNode = nodes.Struct[MultiplyArrayNodeData]

type MultiplyArrayNodeData struct {
	A nodes.Output[[]TRS]
	B nodes.Output[[]TRS]
}

func (tnd MultiplyArrayNodeData) Out() nodes.StructOutput[[]TRS] {
	aVal := nodes.TryGetOutputValue(tnd.A, nil)
	bVal := nodes.TryGetOutputValue(tnd.B, nil)

	out := make([]TRS, max(len(aVal), len(bVal)))

	identity := Identity()
	for i := range out {
		a := identity
		b := identity

		if i < len(aVal) {
			a = aVal[i]
		}

		if i < len(bVal) {
			b = bVal[i]
		}

		out[i] = a.Multiply(b)
	}

	return nodes.NewStructOutput(out)
}

// ============================================================================

type MultiplyToArrayNode = nodes.Struct[MultiplyToArrayNodeData]

type MultiplyToArrayNodeData struct {
	Left  nodes.Output[TRS]
	Array nodes.Output[[]TRS]
	Right nodes.Output[TRS]
}

func (n MultiplyToArrayNodeData) Description() string {
	return "Multiplies each element by the left and right values provided. If left or right is not defined, they are considered the identity matrix. Each value in the resulting array is computed by `left * arr[i] * right`"
}

func (n MultiplyToArrayNodeData) Out() nodes.StructOutput[[]TRS] {
	if n.Array == nil {
		return nodes.NewStructOutput[[]TRS](nil)
	}

	arr := n.Array.Value()

	if n.Left == nil && n.Right == nil {
		return nodes.NewStructOutput(arr)
	}

	out := make([]TRS, len(arr))
	if n.Left == nil && n.Right != nil {
		right := n.Right.Value()
		for i, v := range arr {
			out[i] = v.Multiply(right)
		}
	} else if n.Left != nil && n.Right == nil {
		left := n.Left.Value()
		for i, v := range arr {
			out[i] = left.Multiply(v)
		}
	} else {
		right := n.Right.Value()
		left := n.Left.Value()
		for i, v := range arr {
			out[i] = left.Multiply(v.Multiply(right))
		}
	}

	return nodes.NewStructOutput(out)
}

// ============================================================================

type TransformArrayNode = nodes.Struct[TransformArrayNodeData]

type TransformArrayNodeData struct {
	Transform nodes.Output[TRS]
	Array     nodes.Output[[]TRS]
}

func (tnd TransformArrayNodeData) Out() nodes.StructOutput[[]TRS] {
	if tnd.Transform == nil {
		return nodes.NewStructOutput(nodes.TryGetOutputValue(tnd.Array, nil))
	}

	v := tnd.Transform.Value()
	inArr := nodes.TryGetOutputValue(tnd.Array, nil)

	out := make([]TRS, len(inArr))
	for i, e := range inArr {
		out[i] = v.Multiply(e)
	}

	return nodes.NewStructOutput(out)
}

// ============================================================================

type RotateDirectionNode = nodes.Struct[RotateDirectionNodeData]

type RotateDirectionNodeData struct {
	TRS       nodes.Output[TRS]
	Direction nodes.Output[vector3.Float64]
}

func (tnd RotateDirectionNodeData) Out() nodes.StructOutput[vector3.Float64] {
	if tnd.TRS == nil || tnd.Direction == nil {
		return nodes.NewStructOutput(nodes.TryGetOutputValue(tnd.Direction, vector3.Zero[float64]()))
	}

	return nodes.NewStructOutput(tnd.TRS.Value().RotateDirection(tnd.Direction.Value()))
}

// ============================================================================

type RotateDirectionsNode = nodes.Struct[RotateDirectionNodeData]

type RotateDirectionsNodeData struct {
	TRS       nodes.Output[[]TRS]
	Direction nodes.Output[[]vector3.Float64]
}

func (tnd RotateDirectionsNodeData) Out() nodes.StructOutput[[]vector3.Float64] {
	trss := nodes.TryGetOutputValue(tnd.TRS, nil)
	directions := nodes.TryGetOutputValue(tnd.Direction, nil)
	out := make([]vector3.Float64, max(len(trss), len(directions)))

	for i := 0; i < len(out); i++ {
		val := vector3.Zero[float64]()

		if i < len(trss) && i < len(directions) {
			val = trss[i].RotateDirection(directions[i])
		}

		out[i] = val
	}

	return nodes.NewStructOutput(out)
}

// ============================================================================

type ArrayNode = nodes.Struct[ArrayNodeData]

type ArrayNodeData struct {
	Position nodes.Output[[]vector3.Float64]
	Scale    nodes.Output[[]vector3.Float64]
	Rotation nodes.Output[[]quaternion.Quaternion]
}

func (tnd ArrayNodeData) Out() nodes.StructOutput[[]TRS] {
	positions := nodes.TryGetOutputValue(tnd.Position, nil)
	rotations := nodes.TryGetOutputValue(tnd.Rotation, nil)
	scales := nodes.TryGetOutputValue(tnd.Scale, nil)

	transforms := make([]TRS, max(len(positions), len(rotations), len(scales)))
	for i := 0; i < len(transforms); i++ {
		p := vector3.Zero[float64]()
		r := quaternion.Identity()
		s := vector3.One[float64]()

		if i < len(positions) {
			p = positions[i]
		}

		if i < len(rotations) {
			r = rotations[i]
		}

		if i < len(scales) {
			s = scales[i]
		}

		transforms[i] = New(p, r, s)
	}

	return nodes.NewStructOutput(transforms)
}

// ============================================================================

func filterV3(v, min, max vector3.Float64) bool {
	if v.X() < min.X() || v.X() > max.X() {
		return false
	}

	if v.Y() < min.Y() || v.Y() > max.Y() {
		return false
	}

	if v.Z() < min.Z() || v.Z() > max.Z() {
		return false
	}

	return true
}

type FilterNode struct {
	Input       nodes.Output[[]TRS]
	MinScale    nodes.Output[vector3.Float64]
	MaxScale    nodes.Output[vector3.Float64]
	MinPosition nodes.Output[vector3.Float64]
	MaxPosition nodes.Output[vector3.Float64]
}

func (tnd FilterNode) Filter() ([]TRS, []TRS) {
	if tnd.Input == nil {
		return nil, nil
	}

	if tnd.MinScale == nil && tnd.MaxScale == nil && tnd.MinPosition == nil && tnd.MaxPosition == nil {
		return tnd.Input.Value(), nil
	}

	arr := tnd.Input.Value()
	minS := nodes.TryGetOutputValue(tnd.MinScale, vector3.Fill(-math.MaxFloat64))
	maxS := nodes.TryGetOutputValue(tnd.MaxScale, vector3.Fill(math.MaxFloat64))
	minP := nodes.TryGetOutputValue(tnd.MinPosition, vector3.Fill(-math.MaxFloat64))
	maxP := nodes.TryGetOutputValue(tnd.MaxPosition, vector3.Fill(math.MaxFloat64))

	kept := make([]TRS, 0)
	removed := make([]TRS, 0)

	for _, v := range arr {
		if filterV3(v.scale, minS, maxS) && filterV3(v.position, minP, maxP) {
			kept = append(kept, v)
		} else {
			removed = append(removed, v)
		}
	}
	return kept, removed
}

func (tnd FilterNode) Kept() nodes.StructOutput[[]TRS] {
	kept, _ := tnd.Filter()
	return nodes.NewStructOutput(kept)
}

func (tnd FilterNode) Removed() nodes.StructOutput[[]TRS] {
	_, removed := tnd.Filter()
	return nodes.NewStructOutput(removed)
}
