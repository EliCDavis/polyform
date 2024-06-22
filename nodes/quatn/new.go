package quatn

import (
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type New = nodes.StructNode[quaternion.Quaternion, NewData]

type NewData struct {
	X nodes.NodeOutput[float64]
	Y nodes.NodeOutput[float64]
	Z nodes.NodeOutput[float64]
	W nodes.NodeOutput[float64]
}

func (cn NewData) Process() (quaternion.Quaternion, error) {

	var x float64
	if cn.X != nil {
		x = cn.X.Value()
	}

	var y float64
	if cn.Y != nil {
		y = cn.Y.Value()
	}

	var z float64
	if cn.Z != nil {
		z = cn.Z.Value()
	}

	var w float64
	if cn.W != nil {
		w = cn.W.Value()
	}

	return quaternion.New(vector3.New(x, y, z), w), nil
}

// From Theta =================================================================

type FromTheta = nodes.StructNode[quaternion.Quaternion, FromThetaData]

type FromThetaData struct {
	Theta     nodes.NodeOutput[float64]
	Direction nodes.NodeOutput[vector3.Float64]
}

func (cn FromThetaData) Process() (quaternion.Quaternion, error) {
	var theta float64
	if cn.Theta != nil {
		theta = cn.Theta.Value()
	}

	var direction vector3.Float64
	if cn.Direction != nil {
		direction = cn.Direction.Value()
	}

	return quaternion.FromTheta(theta, direction), nil
}
