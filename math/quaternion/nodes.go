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
	refutil.RegisterType[nodes.Struct[FromThetaArrayNode]](factory)
	refutil.RegisterType[nodes.Struct[FromEulerAngleNode]](factory)
	refutil.RegisterType[nodes.Struct[FromEulerAnglesNode]](factory)

	generator.RegisterTypes(factory)
}

type NewNode struct {
	X nodes.Output[float64]
	Y nodes.Output[float64]
	Z nodes.Output[float64]
	W nodes.Output[float64]
}

func (cn NewNode) Out(out *nodes.StructOutput[Quaternion]) {
	out.Set(New(
		vector3.New(
			nodes.TryGetOutputValue(out, cn.X, 0.),
			nodes.TryGetOutputValue(out, cn.Y, 0.),
			nodes.TryGetOutputValue(out, cn.Z, 0.),
		),
		nodes.TryGetOutputValue(out, cn.W, 0),
	))
}

// From Theta =================================================================

type FromThetaNode struct {
	Theta     nodes.Output[float64]
	Direction nodes.Output[vector3.Float64]
}

func (cn FromThetaNode) Out(out *nodes.StructOutput[Quaternion]) {
	out.Set(FromTheta(
		nodes.TryGetOutputValue(out, cn.Theta, 0),
		nodes.TryGetOutputValue(out, cn.Direction, vector3.Zero[float64]()),
	))
}

// ============================================================================

type FromThetaArrayNode struct {
	Direction nodes.Output[[]vector3.Float64]
	Theta     nodes.Output[[]float64]
}

func (snd FromThetaArrayNode) Out(out *nodes.StructOutput[[]Quaternion]) {
	directions := nodes.TryGetOutputValue(out, snd.Direction, nil)
	thetaArr := nodes.TryGetOutputValue(out, snd.Theta, nil)

	arr := make([]Quaternion, max(len(directions), len(directions)))
	for i := range arr {
		direction := vector3.Zero[float64]()
		theta := 0.

		if i < len(directions) {
			direction = directions[i]
		}

		if i < len(thetaArr) {
			theta = thetaArr[i]
		}

		arr[i] = FromTheta(theta, direction)
	}
	out.Set(arr)
}

// From Euler Angles ==========================================================
type FromEulerAngleNode struct {
	Angle nodes.Output[vector3.Float64]
}

func (cn FromEulerAngleNode) Out(out *nodes.StructOutput[Quaternion]) {
	out.Set(FromEulerAngle(nodes.TryGetOutputValue(out, cn.Angle, vector3.Zero[float64]())))
}

type FromEulerAnglesNode struct {
	Angles nodes.Output[[]vector3.Float64]
}

func (cn FromEulerAnglesNode) Out(out *nodes.StructOutput[[]Quaternion]) {
	angles := nodes.TryGetOutputValue(out, cn.Angles, nil)

	arr := make([]Quaternion, len(angles))
	for i, v := range angles {
		arr[i] = FromEulerAngle(v)
	}
	out.Set(arr)
}
