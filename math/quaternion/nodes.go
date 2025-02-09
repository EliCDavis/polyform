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
	refutil.RegisterType[FromEulerAnglesNode](factory)

	generator.RegisterTypes(factory)
}

type NewNode = nodes.Struct[Quaternion, NewNodeData]

type NewNodeData struct {
	X nodes.NodeOutput[float64]
	Y nodes.NodeOutput[float64]
	Z nodes.NodeOutput[float64]
	W nodes.NodeOutput[float64]
}

func (cn NewNodeData) Process() (Quaternion, error) {
	return New(
		vector3.New(
			nodes.TryGetOutputValue(cn.X, 0.),
			nodes.TryGetOutputValue(cn.Y, 0.),
			nodes.TryGetOutputValue(cn.Z, 0.),
		),
		nodes.TryGetOutputValue(cn.W, 0),
	), nil
}

// From Theta =================================================================

type FromThetaNode = nodes.Struct[Quaternion, FromThetaNodeData]

type FromThetaNodeData struct {
	Theta     nodes.NodeOutput[float64]
	Direction nodes.NodeOutput[vector3.Float64]
}

func (cn FromThetaNodeData) Process() (Quaternion, error) {
	return FromTheta(
		nodes.TryGetOutputValue(cn.Theta, 0),
		nodes.TryGetOutputValue(cn.Direction, vector3.Zero[float64]()),
	), nil
}

// ============================================================================

type FromThetaArrayNode = nodes.Struct[[]Quaternion, FromThetaArrayNodeData]

type FromThetaArrayNodeData struct {
	Direction nodes.NodeOutput[[]vector3.Float64]
	Theta     nodes.NodeOutput[[]float64]
}

func (snd FromThetaArrayNodeData) Process() ([]Quaternion, error) {
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

	return out, nil
}

// From Euler Angles ==========================================================
type FromEulerAnglesNode = nodes.Struct[Quaternion, FromEulerAnglesNodeData]

type FromEulerAnglesNodeData struct {
	Angles nodes.NodeOutput[vector3.Float64]
}

func (cn FromEulerAnglesNodeData) Process() (Quaternion, error) {
	return FromEulerAngles(nodes.TryGetOutputValue(cn.Angles, vector3.Zero[float64]())), nil
}
