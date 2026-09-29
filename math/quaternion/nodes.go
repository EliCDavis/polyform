package quaternion

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[NewNode]](factory)
	refutil.RegisterType[nodes.Struct[FromThetaNode]](factory)
	refutil.RegisterType[nodes.Struct[RotationToNode]](factory)
	refutil.RegisterType[nodes.Struct[MultiplyNode]](factory)
	refutil.RegisterType[nodes.Struct[FromEulerAngleNode]](factory)

	generator.RegisterTypes(factory)
}

type NewNode struct {
	X nodes.LiftedPort[float64]
	Y nodes.LiftedPort[float64]
	Z nodes.LiftedPort[float64]
	W nodes.LiftedPort[float64]
}

func (cn NewNode) Description() string {
	return "Builds a rotation from its raw X/Y/Z/W components."
}

func (cn NewNode) Out(out *nodes.Lifted[Quaternion]) {
	nodes.ZipAll(
		out,
		[]nodes.LiftedPort[float64]{
			nodes.LiftedOr(cn.X, 0.),
			nodes.LiftedOr(cn.Y, 0.),
			nodes.LiftedOr(cn.Z, 0.),
			nodes.LiftedOr(cn.W, 0.),
		},
		func(v []float64) Quaternion {
			return New(vector3.New(v[0], v[1], v[2]), v[3])
		},
	)
}

// From Theta =================================================================

type FromThetaNode struct {
	Theta     nodes.LiftedPort[float64]
	Direction nodes.LiftedPort[vector3.Float64]
}

func (cn FromThetaNode) Description() string {
	return "Rotation of an angle in radians about an axis."
}

func (cn FromThetaNode) Out(out *nodes.Lifted[Quaternion]) {
	nodes.Zip2(out, cn.Theta, cn.Direction, FromTheta)
}

// ============================================================================

type RotationToNode struct {
	From nodes.LiftedPort[vector3.Float64]
	To   nodes.LiftedPort[vector3.Float64]
}

func (RotationToNode) Description() string {
	return "The shortest rotation that turns the From direction into the To direction."
}

func (RotationToNode) Keywords() []string {
	return []string{"look at", "aim"}
}

func (n RotationToNode) Out(out *nodes.Lifted[Quaternion]) {
	forward := vector3.Forward[float64]()
	nodes.Zip2(
		out,
		nodes.LiftedOr(n.From, forward),
		nodes.LiftedOr(n.To, forward),
		func(from, to vector3.Float64) Quaternion {
			return RotationTo(from.Normalized(), to.Normalized())
		},
	)
}

// ============================================================================

type MultiplyNode struct {
	First nodes.LiftedPort[Quaternion]
	Then  nodes.LiftedPort[Quaternion]
}

func (MultiplyNode) Description() string {
	return "Composes two rotations into one: the result applies First, then Then."
}

func (MultiplyNode) Keywords() []string {
	return []string{"compose", "chain"}
}

func (n MultiplyNode) Out(out *nodes.Lifted[Quaternion]) {
	nodes.Zip2(
		out,
		nodes.LiftedOr(n.First, Identity()),
		nodes.LiftedOr(n.Then, Identity()),
		func(first, then Quaternion) Quaternion {
			return then.Multiply(first)
		},
	)
}

// From Euler Angles ==========================================================

type FromEulerAngleNode struct {
	Angle nodes.LiftedPort[vector3.Float64]
}

func (cn FromEulerAngleNode) Description() string {
	return "Rotation from euler angles in radians."
}

func (cn FromEulerAngleNode) Out(out *nodes.Lifted[Quaternion]) {
	nodes.Zip1(out, cn.Angle, FromEulerAngle)
}
