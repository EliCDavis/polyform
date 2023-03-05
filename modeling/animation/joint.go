package animation

import (
	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/vector/vector3"
)

func NewJoint(name string, worldPosition, up, forward vector3.Float64, children ...Joint) Joint {
	return Joint{
		name:          name,
		worldPosition: worldPosition,
		up:            up,
		forward:       forward,
		children:      children,
	}
}

type Joint struct {
	name          string
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
