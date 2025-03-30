package quaternion

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[NewNode](factory)
	refutil.RegisterType[FromThetaNode](factory)
	refutil.RegisterType[FromThetaArrayNode](factory)
	refutil.RegisterType[FromEulerAngleNode](factory)
	refutil.RegisterType[FromEulerAnglesNode](factory)

	generator.RegisterTypes(factory)
}

type NewNode = nodes.Struct[NewNodeData]

type NewNodeData struct {
	X nodes.Output[float64]
	Y nodes.Output[float64]
	Z nodes.Output[float64]
	W nodes.Output[float64]
}

func (cn NewNodeData) Out() nodes.StructOutput[Quaternion] {
	return nodes.NewStructOutput(New(
		vector3.New(
			nodes.TryGetOutputValue(cn.X, 0.),
			nodes.TryGetOutputValue(cn.Y, 0.),
			nodes.TryGetOutputValue(cn.Z, 0.),
		),
		nodes.TryGetOutputValue(cn.W, 0),
	))
}

// From Theta =================================================================

type FromThetaNode = nodes.Struct[FromThetaNodeData]

type FromThetaNodeData struct {
	Theta     nodes.Output[float64]
	Direction nodes.Output[vector3.Float64]
}

func (cn FromThetaNodeData) Out() nodes.StructOutput[Quaternion] {
	return nodes.NewStructOutput(FromTheta(
		nodes.TryGetOutputValue(cn.Theta, 0),
		nodes.TryGetOutputValue(cn.Direction, vector3.Zero[float64]()),
	))
}

// ============================================================================

type FromThetaArrayNode = nodes.Struct[FromThetaArrayNodeData]

type FromThetaArrayNodeData struct {
	Direction nodes.Output[[]vector3.Float64]
	Theta     nodes.Output[[]float64]
}

func (snd FromThetaArrayNodeData) Out() nodes.StructOutput[[]Quaternion] {
	directions := nodes.TryGetOutputValue(snd.Direction, nil)
	thetaArr := nodes.TryGetOutputValue(snd.Theta, nil)

	out := make([]Quaternion, max(len(directions), len(directions)))
	for i := 0; i < len(out); i++ {
		direction := vector3.Zero[float64]()
		theta := 0.

		if i < len(directions) {
			direction = directions[i]
		}

		if i < len(thetaArr) {
			theta = thetaArr[i]
		}

		out[i] = FromTheta(theta, direction)
	}

	return nodes.NewStructOutput(out)
}

// From Euler Angles ==========================================================
type FromEulerAngleNode = nodes.Struct[FromEulerAngleNodeData]

type FromEulerAngleNodeData struct {
	Angle nodes.Output[vector3.Float64]
}

func (cn FromEulerAngleNodeData) Out() nodes.StructOutput[Quaternion] {
	return nodes.NewStructOutput(FromEulerAngle(nodes.TryGetOutputValue(cn.Angle, vector3.Zero[float64]())))
}

type FromEulerAnglesNode = nodes.Struct[FromEulerAnglesNodeData]

type FromEulerAnglesNodeData struct {
	Angles nodes.Output[[]vector3.Float64]
}

func (cn FromEulerAnglesNodeData) Out() nodes.StructOutput[[]Quaternion] {
	angles := nodes.TryGetOutputValue(cn.Angles, nil)

	out := make([]Quaternion, len(angles))
	for i, v := range angles {
		out[i] = FromEulerAngle(v)
	}

	return nodes.NewStructOutput(out)
}
