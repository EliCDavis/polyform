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
	refutil.RegisterType[nodes.Struct[FilterPositionNode]](factory)
	refutil.RegisterType[nodes.Struct[FilterScaleNode]](factory)
	generator.RegisterTypes(factory)
}

// ============================================================================

type NewNode = nodes.Struct[NewNodeData]

type NewNodeData struct {
	Position nodes.Output[vector3.Float64]
	Rotation nodes.Output[quaternion.Quaternion]
	Scale    nodes.Output[vector3.Float64]
}

func (tnd NewNodeData) Out(out *nodes.StructOutput[TRS]) {
	out.Set(New(
		nodes.TryGetOutputValue(out, tnd.Position, vector3.Zero[float64]()),
		nodes.TryGetOutputValue(out, tnd.Rotation, quaternion.Identity()),
		nodes.TryGetOutputValue(out, tnd.Scale, vector3.One[float64]()),
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
	Array              nodes.Output[[]TRS]
}

func (tnd RandomizeArrayNodeData) Out(out *nodes.StructOutput[[]TRS]) {
	input := nodes.TryGetOutputValue(out, tnd.Array, nil)
	if len(input) == 0 {
		return
	}

	minT := nodes.TryGetOutputValue(out, tnd.TranslationMinimum, vector3.Zero[float64]())
	maxT := nodes.TryGetOutputValue(out, tnd.TranslationMaximum, vector3.Zero[float64]())
	rangeT := maxT.Sub(minT)

	minS := nodes.TryGetOutputValue(out, tnd.ScaleMinimum, vector3.One[float64]())
	maxS := nodes.TryGetOutputValue(out, tnd.ScaleMaximum, vector3.One[float64]())
	rangeS := maxS.Sub(minS)

	minR := nodes.TryGetOutputValue(out, tnd.RotationMinimum, vector3.Zero[float64]())
	maxR := nodes.TryGetOutputValue(out, tnd.RotationMaximum, vector3.Zero[float64]())
	rangeR := maxR.Sub(minR)

	arr := make([]TRS, len(input))
	for i := range input {
		sample := New(
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
			minS.Add(vector3.New(
				rangeS.X()*rand.Float64(),
				rangeS.Y()*rand.Float64(),
				rangeS.Z()*rand.Float64(),
			)),
		)
		arr[i] = input[i].Multiply(sample)
	}

	out.Set(arr)
}

// ============================================================================

type SelectNode = nodes.Struct[SelectNodeData]

type SelectNodeData struct {
	TRS nodes.Output[TRS]
}

func (tnd SelectNodeData) Position(out *nodes.StructOutput[vector3.Float64]) {
	out.Set(nodes.TryGetOutputValue(out, tnd.TRS, Identity()).Position())
}

func (tnd SelectNodeData) Scale(out *nodes.StructOutput[vector3.Float64]) {
	out.Set(nodes.TryGetOutputValue(out, tnd.TRS, Identity()).Scale())
}

func (tnd SelectNodeData) Rotation(out *nodes.StructOutput[quaternion.Quaternion]) {
	out.Set(nodes.TryGetOutputValue(out, tnd.TRS, Identity()).Rotation())
}

// ============================================================================

type SelectArrayNode = nodes.Struct[SelectArrayNodeData]

type SelectArrayNodeData struct {
	TRS nodes.Output[[]TRS]
}

func (tnd SelectArrayNodeData) Position(out *nodes.StructOutput[[]vector3.Float64]) {
	trss := nodes.TryGetOutputValue(out, tnd.TRS, nil)
	arr := make([]vector3.Float64, len(trss))

	for i, trs := range trss {
		arr[i] = trs.Position()
	}

	out.Set(arr)
}

func (tnd SelectArrayNodeData) Scale(out *nodes.StructOutput[[]vector3.Float64]) {
	trss := nodes.TryGetOutputValue(out, tnd.TRS, nil)
	arr := make([]vector3.Float64, len(trss))

	for i, trs := range trss {
		arr[i] = trs.Scale()
	}

	out.Set(arr)
}

func (tnd SelectArrayNodeData) Rotation(out *nodes.StructOutput[[]quaternion.Quaternion]) {
	trss := nodes.TryGetOutputValue(out, tnd.TRS, nil)
	arr := make([]quaternion.Quaternion, len(trss))

	for i, trs := range trss {
		arr[i] = trs.Rotation()
	}

	out.Set(arr)
}

// ============================================================================

type MultiplyNode = nodes.Struct[MultiplyNodeData]

type MultiplyNodeData struct {
	A nodes.Output[TRS]
	B nodes.Output[TRS]
}

func (tnd MultiplyNodeData) Out(out *nodes.StructOutput[TRS]) {
	a := nodes.TryGetOutputValue(out, tnd.A, Identity())
	b := nodes.TryGetOutputValue(out, tnd.B, Identity())
	out.Set(a.Multiply(b))
}

// ============================================================================

type MultiplyArrayNode = nodes.Struct[MultiplyArrayNodeData]

type MultiplyArrayNodeData struct {
	A nodes.Output[[]TRS]
	B nodes.Output[[]TRS]
}

func (tnd MultiplyArrayNodeData) Out(out *nodes.StructOutput[[]TRS]) {
	aVal := nodes.TryGetOutputValue(out, tnd.A, nil)
	bVal := nodes.TryGetOutputValue(out, tnd.B, nil)

	arr := make([]TRS, max(len(aVal), len(bVal)))

	identity := Identity()
	for i := range arr {
		a := identity
		b := identity

		if i < len(aVal) {
			a = aVal[i]
		}

		if i < len(bVal) {
			b = bVal[i]
		}

		arr[i] = a.Multiply(b)
	}

	out.Set(arr)
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

func (n MultiplyToArrayNodeData) Out(out *nodes.StructOutput[[]TRS]) {
	if n.Array == nil {
		return
	}

	in := nodes.GetOutputValue(out, n.Array)
	if n.Left == nil && n.Right == nil {
		out.Set(in)
		return
	}

	arr := make([]TRS, len(in))
	if n.Left == nil && n.Right != nil {
		right := nodes.GetOutputValue(out, n.Right)
		for i, v := range in {
			arr[i] = v.Multiply(right)
		}
	} else if n.Left != nil && n.Right == nil {
		left := nodes.GetOutputValue(out, n.Left)
		for i, v := range in {
			arr[i] = left.Multiply(v)
		}
	} else {
		right := nodes.GetOutputValue(out, n.Right)
		left := nodes.GetOutputValue(out, n.Left)
		for i, v := range in {
			arr[i] = left.Multiply(v.Multiply(right))
		}
	}

	out.Set(arr)
}

// ============================================================================

type TransformArrayNode = nodes.Struct[TransformArrayNodeData]

type TransformArrayNodeData struct {
	Transform nodes.Output[TRS]
	Array     nodes.Output[[]TRS]
}

func (tnd TransformArrayNodeData) Out(out *nodes.StructOutput[[]TRS]) {
	if tnd.Transform == nil {
		out.Set(nodes.TryGetOutputValue(out, tnd.Array, nil))
		return
	}

	v := nodes.GetOutputValue(out, tnd.Transform)
	inArr := nodes.TryGetOutputValue(out, tnd.Array, nil)
	outArr := make([]TRS, len(inArr))
	for i, e := range inArr {
		outArr[i] = v.Multiply(e)
	}

	out.Set(outArr)
}

// ============================================================================

type RotateDirectionNode = nodes.Struct[RotateDirectionNodeData]

type RotateDirectionNodeData struct {
	TRS       nodes.Output[TRS]
	Direction nodes.Output[vector3.Float64]
}

func (tnd RotateDirectionNodeData) Out(out *nodes.StructOutput[vector3.Float64]) {
	if tnd.TRS == nil || tnd.Direction == nil {
		out.Set(nodes.TryGetOutputValue(out, tnd.Direction, vector3.Zero[float64]()))
		return
	}

	out.Set(nodes.GetOutputValue(out, tnd.TRS).RotateDirection(nodes.GetOutputValue(out, tnd.Direction)))
}

// ============================================================================

type RotateDirectionsNode = nodes.Struct[RotateDirectionNodeData]

type RotateDirectionsNodeData struct {
	TRS       nodes.Output[[]TRS]
	Direction nodes.Output[[]vector3.Float64]
}

func (tnd RotateDirectionsNodeData) Out(out *nodes.StructOutput[[]vector3.Float64]) {
	trss := nodes.TryGetOutputValue(out, tnd.TRS, nil)
	directions := nodes.TryGetOutputValue(out, tnd.Direction, nil)
	arr := make([]vector3.Float64, max(len(trss), len(directions)))

	for i := 0; i < len(arr); i++ {
		val := vector3.Zero[float64]()

		if i < len(trss) && i < len(directions) {
			val = trss[i].RotateDirection(directions[i])
		}

		arr[i] = val
	}

	out.Set(arr)
}

// ============================================================================

type ArrayNode = nodes.Struct[ArrayNodeData]

type ArrayNodeData struct {
	Position nodes.Output[[]vector3.Float64]
	Scale    nodes.Output[[]vector3.Float64]
	Rotation nodes.Output[[]quaternion.Quaternion]
}

func (tnd ArrayNodeData) Out(out *nodes.StructOutput[[]TRS]) {
	positions := nodes.TryGetOutputValue(out, tnd.Position, nil)
	rotations := nodes.TryGetOutputValue(out, tnd.Rotation, nil)
	scales := nodes.TryGetOutputValue(out, tnd.Scale, nil)

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

	out.Set(transforms)
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

func filter(
	recorder nodes.ExecutionRecorder,
	Input nodes.Output[[]TRS],
	MinX nodes.Output[float64],
	MinY nodes.Output[float64],
	MinZ nodes.Output[float64],
	MaxX nodes.Output[float64],
	MaxY nodes.Output[float64],
	MaxZ nodes.Output[float64],
	position bool,
) ([]TRS, []TRS) {
	if Input == nil {
		return nil, nil
	}

	inputs := []nodes.Output[float64]{
		MinX, MinY, MinZ,
		MaxX, MaxY, MaxZ,
	}
	allNil := true
	for _, v := range inputs {
		if v != nil {
			allNil = false
			break
		}
	}

	arr := nodes.GetOutputValue(recorder, Input)
	if allNil {
		return arr, nil
	}

	min := vector3.New(
		nodes.TryGetOutputValue(recorder, MinX, -math.MaxFloat64),
		nodes.TryGetOutputValue(recorder, MinY, -math.MaxFloat64),
		nodes.TryGetOutputValue(recorder, MinZ, -math.MaxFloat64),
	)
	max := vector3.New(
		nodes.TryGetOutputValue(recorder, MaxX, math.MaxFloat64),
		nodes.TryGetOutputValue(recorder, MaxY, math.MaxFloat64),
		nodes.TryGetOutputValue(recorder, MaxZ, math.MaxFloat64),
	)

	kept := make([]TRS, 0)
	removed := make([]TRS, 0)

	if position {
		for _, v := range arr {
			if filterV3(v.position, min, max) {
				kept = append(kept, v)
			} else {
				removed = append(removed, v)
			}
		}
	} else {
		for _, v := range arr {
			if filterV3(v.scale, min, max) {
				kept = append(kept, v)
			} else {
				removed = append(removed, v)
			}
		}
	}

	return kept, removed
}

type FilterPositionNode struct {
	Input nodes.Output[[]TRS]
	MinX  nodes.Output[float64]
	MinY  nodes.Output[float64]
	MinZ  nodes.Output[float64]
	MaxX  nodes.Output[float64]
	MaxY  nodes.Output[float64]
	MaxZ  nodes.Output[float64]
}

func (tnd FilterPositionNode) Filter(out *nodes.StructOutput[[]TRS]) ([]TRS, []TRS) {
	return filter(
		out,
		tnd.Input,
		tnd.MinX, tnd.MinY, tnd.MinZ,
		tnd.MaxX, tnd.MaxY, tnd.MaxZ,
		true,
	)
}

func (tnd FilterPositionNode) Kept(out *nodes.StructOutput[[]TRS]) {
	kept, _ := tnd.Filter(out)
	out.Set(kept)
}

func (tnd FilterPositionNode) Removed(out *nodes.StructOutput[[]TRS]) {
	_, removed := tnd.Filter(out)
	out.Set(removed)
}

type FilterScaleNode struct {
	Input nodes.Output[[]TRS]
	MinX  nodes.Output[float64]
	MinY  nodes.Output[float64]
	MinZ  nodes.Output[float64]
	MaxX  nodes.Output[float64]
	MaxY  nodes.Output[float64]
	MaxZ  nodes.Output[float64]
}

func (tnd FilterScaleNode) Filter(out *nodes.StructOutput[[]TRS]) ([]TRS, []TRS) {
	return filter(
		out,
		tnd.Input,
		tnd.MinX, tnd.MinY, tnd.MinZ,
		tnd.MaxX, tnd.MaxY, tnd.MaxZ,
		false,
	)
}

func (tnd FilterScaleNode) Kept(out *nodes.StructOutput[[]TRS]) {
	kept, _ := tnd.Filter(out)
	out.Set(kept)
}

func (tnd FilterScaleNode) Removed(out *nodes.StructOutput[[]TRS]) {
	_, removed := tnd.Filter(out)
	out.Set(removed)
}
