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

func (cn NewNodeData) Out(out *nodes.StructOutput[Quaternion]) {
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

type FromThetaNode = nodes.Struct[FromThetaNodeData]

type FromThetaNodeData struct {
	Theta     nodes.Output[float64]
	Direction nodes.Output[vector3.Float64]
}

func (cn FromThetaNodeData) Out(out *nodes.StructOutput[Quaternion]) {
	out.Set(FromTheta(
		nodes.TryGetOutputValue(out, cn.Theta, 0),
		nodes.TryGetOutputValue(out, cn.Direction, vector3.Zero[float64]()),
	))
}

// ============================================================================

type FromThetaArrayNode = nodes.Struct[FromThetaArrayNodeData]

type FromThetaArrayNodeData struct {
	Direction nodes.Output[[]vector3.Float64]
	Theta     nodes.Output[[]float64]
}

func (snd FromThetaArrayNodeData) Out(out *nodes.StructOutput[[]Quaternion]) {
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
type FromEulerAngleNode = nodes.Struct[FromEulerAngleNodeData]

type FromEulerAngleNodeData struct {
	Angle nodes.Output[vector3.Float64]
}

func (cn FromEulerAngleNodeData) Out(out *nodes.StructOutput[Quaternion]) {
	out.Set(FromEulerAngle(nodes.TryGetOutputValue(out, cn.Angle, vector3.Zero[float64]())))
}

type FromEulerAnglesNode = nodes.Struct[FromEulerAnglesNodeData]

type FromEulerAnglesNodeData struct {
	Angles nodes.Output[[]vector3.Float64]
}

func (cn FromEulerAnglesNodeData) Out(out *nodes.StructOutput[[]Quaternion]) {
	angles := nodes.TryGetOutputValue(out, cn.Angles, nil)

	arr := make([]Quaternion, len(angles))
	for i, v := range angles {
		arr[i] = FromEulerAngle(v)
	}
	out.Set(arr)
}
