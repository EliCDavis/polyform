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
	X     nodes.NodeOutput[[]float64]
	Y     nodes.NodeOutput[[]float64]
	Z     nodes.NodeOutput[[]float64]
	Theta nodes.NodeOutput[[]float64]
}

func (snd FromThetaArrayNodeData) Process() ([]Quaternion, error) {
	xArr := nodes.TryGetOutputValue(snd.X, nil)
	yArr := nodes.TryGetOutputValue(snd.Y, nil)
	zArr := nodes.TryGetOutputValue(snd.Z, nil)
	thetaArr := nodes.TryGetOutputValue(snd.Theta, nil)

	out := make([]Quaternion, max(len(xArr), len(yArr), len(zArr), len(thetaArr)))
	for i := 0; i < len(out); i++ {
		x := 0.
		y := 0.
		z := 0.
		theta := 0.

		if i < len(xArr) {
			x = xArr[i]
		}

		if i < len(yArr) {
			y = yArr[i]
		}

		if i < len(zArr) {
			z = zArr[i]
		}

		if i < len(thetaArr) {
			theta = thetaArr[i]
		}

		out[i] = FromTheta(theta, vector3.New(x, y, z))
	}

	return out, nil
}
