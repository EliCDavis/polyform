package animation

import (
	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/vector/vector3"
)

func NewJoint(worldPosition, relativePosition, up, forward vector3.Float64, children ...Joint) Joint {
	return Joint{
		worldPosition:    worldPosition,
		relativePosition: relativePosition,
		up:               up,
		forward:          forward,
		children:         children,
	}
}

type Joint struct {
	worldPosition, relativePosition vector3.Float64
	up, forward                     vector3.Float64
	children                        []Joint
}

func (j Joint) RelativePosition() vector3.Float64 {
	return j.relativePosition
}

func (j Joint) WorldPosition() vector3.Float64 {
	return j.worldPosition
}

func (j Joint) Children() []Joint {
	return j.children
}

func (j Joint) InverseBindMatrix() mat.Matrix4x4 {
	return mat.MatFromDirs(j.up, j.forward, j.worldPosition).Inverse()
}

func (j Joint) RelativeMatrix() mat.Matrix4x4 {
	return mat.MatFromDirs(j.up, j.forward, j.relativePosition)
}
