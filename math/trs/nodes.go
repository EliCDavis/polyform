package trs

import (
	"fmt"
	"math"
	"time"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/chance"
	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[NewNode]](factory)
	refutil.RegisterType[nodes.Struct[RotateDirectionNode]](factory)
	refutil.RegisterType[nodes.Struct[TransformPointNode]](factory)
	refutil.RegisterType[nodes.Struct[LookAtNode]](factory)
	refutil.RegisterType[nodes.Struct[PivotNode]](factory)
	refutil.RegisterType[nodes.Struct[RandomizeArrayNode]](factory)
	refutil.RegisterType[nodes.Struct[MultiplyNode]](factory)
	refutil.RegisterType[nodes.Struct[MultiplyToArrayNode]](factory)
	refutil.RegisterType[nodes.Struct[SelectNode]](factory)
	refutil.RegisterType[nodes.Struct[FilterPositionNode]](factory)
	refutil.RegisterType[nodes.Struct[FilterScaleNode]](factory)
	generator.RegisterTypes(factory)
}

// ============================================================================

type NewNode struct {
	Position nodes.LiftedPort[vector3.Float64]
	Rotation nodes.LiftedPort[quaternion.Quaternion]
	Scale    nodes.LiftedPort[vector3.Float64]
}

func (tnd NewNode) Description() string {
	return "Builds a transform from a position, rotation and scale."
}

func (tnd NewNode) Out(out *nodes.Lifted[TRS]) {
	nodes.Zip3(
		out,
		nodes.LiftedOr(tnd.Position, vector3.Zero[float64]()),
		nodes.LiftedOr(tnd.Rotation, quaternion.Identity()),
		nodes.LiftedOr(tnd.Scale, vector3.One[float64]()),
		New,
	)
}

// ============================================================================

type RandomizeArrayNode struct {
	TranslationMinimum nodes.Output[vector3.Float64]
	TranslationMaximum nodes.Output[vector3.Float64]
	ScaleMinimum       nodes.Output[vector3.Float64]
	ScaleMaximum       nodes.Output[vector3.Float64]
	RotationMinimum    nodes.Output[vector3.Float64]
	RotationMaximum    nodes.Output[vector3.Float64]
	Array              nodes.Output[[]TRS]
	Seed               nodes.Output[int] `description:"The same seed always produces the same offsets. Defaults to 0."`
}

func (tnd RandomizeArrayNode) Description() string {
	return "Randomly offsets each transform's position, scale and rotation within the given ranges."
}

func (tnd RandomizeArrayNode) Out(out *nodes.StructOutput[[]TRS]) {
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

	rnd := chance.FromSeed(nodes.TryGetOutputValue(out, tnd.Seed, 0))

	arr := make([]TRS, len(input))
	for i := range input {
		sample := New(
			minT.Add(vector3.New(
				rangeT.X()*rnd.Float64(),
				rangeT.Y()*rnd.Float64(),
				rangeT.Z()*rnd.Float64(),
			)),
			quaternion.FromEulerAngle(minR.Add(vector3.New(
				rangeR.X()*rnd.Float64(),
				rangeR.Y()*rnd.Float64(),
				rangeR.Z()*rnd.Float64(),
			))),
			minS.Add(vector3.New(
				rangeS.X()*rnd.Float64(),
				rangeS.Y()*rnd.Float64(),
				rangeS.Z()*rnd.Float64(),
			)),
		)
		composed, err := FromMatrix(input[i].Multiply(sample))
		if err != nil {
			out.CaptureError(fmt.Errorf("entry %d: %w", i, err))
			return
		}
		arr[i] = composed
	}

	out.Set(arr)
}

// ============================================================================

type SelectNode struct {
	TRS nodes.LiftedPort[TRS]
}

func (tnd SelectNode) Description() string {
	return "Splits a transform into its position, scale and rotation."
}

func (tnd SelectNode) Position(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip1(out, nodes.LiftedOr(tnd.TRS, Identity()), TRS.Position)
}

func (tnd SelectNode) Scale(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip1(out, nodes.LiftedOr(tnd.TRS, Identity()), TRS.Scale)
}

func (tnd SelectNode) Rotation(out *nodes.Lifted[quaternion.Quaternion]) {
	nodes.Zip1(out, nodes.LiftedOr(tnd.TRS, Identity()), TRS.Rotation)
}

// ============================================================================

type MultiplyNode struct {
	A nodes.LiftedPort[TRS]
	B nodes.LiftedPort[TRS]
}

func (MultiplyNode) Description() string {
	return "Composes two transforms into one, with A as the parent of B: a point goes through B first, then A."
}

func (MultiplyNode) Keywords() []string {
	return []string{"compose", "parent"}
}

func (tnd MultiplyNode) Out(out *nodes.Lifted[TRS]) {
	compose(out, func(recorder nodes.ExecutionRecorder) {
		nodes.Zip2(
			out,
			nodes.LiftedOr(tnd.A, Identity()),
			nodes.LiftedOr(tnd.B, Identity()),
			func(a, b TRS) TRS {
				return decompose(recorder, a.Multiply(b))
			},
		)
	})
}

// ============================================================================

type MultiplyToArrayNode struct {
	Left   nodes.LiftedPort[TRS]
	Middle nodes.LiftedPort[TRS]
	Right  nodes.LiftedPort[TRS]
}

func (n MultiplyToArrayNode) Description() string {
	return "Composes three transforms as `Left * Middle * Right`. An unwired end is the identity."
}

func (n MultiplyToArrayNode) Out(out *nodes.Lifted[TRS]) {
	compose(out, func(recorder nodes.ExecutionRecorder) {
		nodes.Zip3(
			out,
			nodes.LiftedOr(n.Left, Identity()),
			nodes.LiftedOr(n.Middle, Identity()),
			nodes.LiftedOr(n.Right, Identity()),
			func(left, middle, right TRS) TRS {
				inner := decompose(recorder, middle.Multiply(right))
				return decompose(recorder, left.Multiply(inner))
			},
		)
	})
}

// A matrix that cannot be split back into a position, rotation and scale is
// reported once for the whole output rather than once per element, which for
// a long array would otherwise be thousands of identical errors.
type composeFailure struct {
	err error
}

func (c *composeFailure) CaptureError(err error) {
	if err != nil && c.err == nil {
		c.err = err
	}
}

func (c *composeFailure) CaptureTiming(string, time.Duration) {}

func compose(out *nodes.Lifted[TRS], run func(nodes.ExecutionRecorder)) {
	failure := &composeFailure{}
	run(failure)
	out.CaptureError(failure.err)
}

func decompose(recorder nodes.ExecutionRecorder, m mat.Matrix4x4) TRS {
	composed, err := FromMatrix(m)
	if err != nil {
		recorder.CaptureError(err)
		return Identity()
	}
	return composed
}

// ============================================================================

type PivotNode struct {
	Pivot    nodes.Output[vector3.Float64]
	Rotation nodes.Output[quaternion.Quaternion]
	Scale    nodes.Output[vector3.Float64] `description:"Defaults to (1,1,1); applied about the pivot as well"`
}

func (PivotNode) Description() string {
	return "A transform that rotates and scales about a point instead of the origin: the position works out to Pivot - R·S·Pivot."
}

func (PivotNode) Keywords() []string {
	return []string{"rotate about", "hinge"}
}

func (n PivotNode) Out(out *nodes.StructOutput[TRS]) {
	pivot := nodes.TryGetOutputValue(out, n.Pivot, vector3.Zero[float64]())
	rotation := nodes.TryGetOutputValue(out, n.Rotation, quaternion.Identity())
	scale := nodes.TryGetOutputValue(out, n.Scale, vector3.One[float64]())
	moved := rotation.Rotate(pivot.MultByVector(scale))
	out.Set(New(pivot.Sub(moved), rotation, scale))
}

// ============================================================================

type LookAtNode struct {
	Position nodes.Output[vector3.Float64]
	Target   nodes.Output[vector3.Float64]
	Scale    nodes.Output[vector3.Float64] `description:"Defaults to (1,1,1)"`
}

func (LookAtNode) Description() string {
	return "A transform placed at Position and turned so its local +Z points at Target, with +Y kept up."
}

func (LookAtNode) Keywords() []string {
	return []string{"aim", "orient"}
}

func (n LookAtNode) Out(out *nodes.StructOutput[TRS]) {
	position := nodes.TryGetOutputValue(out, n.Position, vector3.Zero[float64]())
	target := nodes.TryGetOutputValue(out, n.Target, vector3.Forward[float64]())
	scale := nodes.TryGetOutputValue(out, n.Scale, vector3.One[float64]())
	if target.Sub(position).LengthSquared() == 0 {
		out.Set(New(position, quaternion.Identity(), scale))
		return
	}
	out.Set(New(position, quaternion.Identity(), scale).LookAt(target))
}

// ============================================================================

type TransformPointNode struct {
	TRS   nodes.LiftedPort[TRS]
	Point nodes.LiftedPort[vector3.Float64]
}

func (TransformPointNode) Description() string {
	return "Moves a point through a transform: scaled, rotated, then translated."
}

func (TransformPointNode) Keywords() []string {
	return []string{"local to world"}
}

func (n TransformPointNode) Out(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip2(
		out,
		nodes.LiftedOr(n.TRS, Identity()),
		n.Point,
		TRS.Transform,
	)
}

// ============================================================================

type RotateDirectionNode struct {
	TRS       nodes.LiftedPort[TRS]
	Direction nodes.LiftedPort[vector3.Float64]
}

func (tnd RotateDirectionNode) Description() string {
	return "Rotates a direction by a transform's rotation, ignoring its position and scale."
}

func (tnd RotateDirectionNode) Out(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip2(
		out,
		nodes.LiftedOr(tnd.TRS, Identity()),
		tnd.Direction,
		TRS.RotateDirection,
	)
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

func (tnd FilterPositionNode) Description() string {
	return "Splits transforms into those whose position falls inside the given bounds and those that don't."
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

func (tnd FilterScaleNode) Description() string {
	return "Splits transforms into those whose scale falls inside the given bounds and those that don't."
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
