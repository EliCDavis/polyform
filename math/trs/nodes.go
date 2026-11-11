package trs

import (
	"time"

	"github.com/EliCDavis/polyform/generator/registry"
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
	refutil.RegisterType[nodes.Struct[MultiplyNode]](factory)
	refutil.RegisterType[nodes.Struct[MultiplyToArrayNode]](factory)
	refutil.RegisterType[nodes.Struct[SelectNode]](factory)
	registry.Register(factory)
}

// ============================================================================

type NewNode struct {
	Position nodes.LiftedPort[vector3.Float64]
	Rotation nodes.LiftedPort[quaternion.Quaternion]
	Scale    nodes.LiftedPort[vector3.Float64] `default:"1,1,1"`
}

func (tnd NewNode) Description() string {
	return "Builds a transform from a position, rotation and scale."
}

func (tnd NewNode) Out(out *nodes.Lifted[TRS]) {
	nodes.Zip3(
		out,
		tnd.Position,
		tnd.Rotation,
		tnd.Scale,
		New,
	)
}

// ============================================================================

type SelectNode struct {
	TRS nodes.LiftedPort[TRS]
}

func (tnd SelectNode) Description() string {
	return "Splits a transform into its position, scale and rotation."
}

func (tnd SelectNode) Position(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip1(out, tnd.TRS, TRS.Position)
}

func (tnd SelectNode) Scale(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip1(out, tnd.TRS, TRS.Scale)
}

func (tnd SelectNode) Rotation(out *nodes.Lifted[quaternion.Quaternion]) {
	nodes.Zip1(out, tnd.TRS, TRS.Rotation)
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
			tnd.A,
			tnd.B,
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
			n.Left,
			n.Middle,
			n.Right,
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
	Pivot    nodes.LiftedPort[vector3.Float64]
	Rotation nodes.LiftedPort[quaternion.Quaternion]
	Scale    nodes.LiftedPort[vector3.Float64] `description:"Applied about the pivot as well" default:"1,1,1"`
}

func (PivotNode) Description() string {
	return "A transform that rotates and scales about a point instead of the origin: the position works out to Pivot - R·S·Pivot."
}

func (PivotNode) Keywords() []string {
	return []string{"rotate about", "rotate around"}
}

func (n PivotNode) Out(out *nodes.Lifted[TRS]) {
	nodes.Zip3(
		out,
		n.Pivot,
		n.Rotation,
		n.Scale,
		func(pivot vector3.Float64, rotation quaternion.Quaternion, scale vector3.Float64) TRS {
			moved := rotation.Rotate(pivot.MultByVector(scale))
			return New(pivot.Sub(moved), rotation, scale)
		},
	)
}

// ============================================================================

type LookAtNode struct {
	Position nodes.LiftedPort[vector3.Float64]
	Target   nodes.LiftedPort[vector3.Float64]
	Scale    nodes.LiftedPort[vector3.Float64] `default:"1,1,1"`
}

func (LookAtNode) Description() string {
	return "A transform placed at Position and turned so its local +Z points at Target, with +Y kept up."
}

func (LookAtNode) Keywords() []string {
	return []string{"aim", "orient"}
}

func (n LookAtNode) Out(out *nodes.Lifted[TRS]) {
	nodes.Zip3(out, n.Position, n.Target, n.Scale, func(position, target, scale vector3.Float64) TRS {
		if target.Sub(position).LengthSquared() == 0 {
			return New(position, quaternion.Identity(), scale)
		}
		return New(position, quaternion.Identity(), scale).LookAt(target)
	})
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
		n.TRS,
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
		tnd.TRS,
		tnd.Direction,
		TRS.RotateDirection,
	)
}
