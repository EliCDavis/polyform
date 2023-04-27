package animation

import (
	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/vector/vector3"
)

func NewJoint(name string, weight float64, worldPosition, up, forward vector3.Float64, children ...Joint) Joint {
	return Joint{
		name:          name,
		weight:        weight,
		worldPosition: worldPosition,
		up:            up,
		forward:       forward,
		children:      children,
	}
}

type Joint struct {
	name          string
	weight        float64
	worldPosition vector3.Float64
	up, forward   vector3.Float64
	children      []Joint
}

func (j Joint) WorldPosition() vector3.Float64 {
	return j.worldPosition
}

func (j Joint) Children() []Joint {
	return j.children
}

func (j Joint) Matrix() mat.Matrix4x4 {
	return mat.MatFromDirs(j.up, j.forward, j.worldPosition)
}

func (j Joint) InverseBindMatrix() mat.Matrix4x4 {
	return j.Matrix().Inverse()
}

type Axis int

const (
	XAxis Axis = iota
	YAxis
	ZAxis
)

func MirrorJoint(joint Joint, name string, axis Axis) Joint {
	newWorld := joint.worldPosition

	switch axis {
	case XAxis:
		newWorld = newWorld.MultByVector(vector3.New(-1., 1., 1.))

	case YAxis:
		newWorld = newWorld.MultByVector(vector3.New(1., -1., 1.))

	case ZAxis:
		newWorld = newWorld.MultByVector(vector3.New(1., 1., -1.))
	}

	newChildren := make([]Joint, len(joint.children))
	for i, j := range joint.children {
		newChildren[i] = MirrorJoint(j, j.name, axis)
	}

	return NewJoint(
		name,
		joint.weight,
		newWorld,
		joint.up,
		joint.forward,
		newChildren...,
	)
}
