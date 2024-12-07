package trs

import (
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/vector/vector3"
)

// Create a new TRS
func New(position vector3.Float64, rotation quaternion.Quaternion, scale vector3.Float64) TRS {
	return TRS{
		position: position,
		rotation: rotation,
		scale:    scale,
	}
}

// Create a new TRS with a specified position, with a scale of (1, 1, 1) and a
// identity rotation
func Position(position vector3.Float64) TRS {
	return TRS{
		position: position,
		rotation: quaternion.Identity(),
		scale:    vector3.One[float64](),
	}
}

// Create a new TRS with a specified scale, with a position of (0, 0, 0) and a
// identity rotation
func Scale(scale vector3.Float64) TRS {
	return TRS{
		scale:    scale,
		rotation: quaternion.Identity(),
	}
}

// Create a new TRS with a specified rotation, a position of (0, 0, 0) and a
// scale of (1, 1, 1)
func Rotation(rotation quaternion.Quaternion) TRS {
	return TRS{
		scale:    vector3.One[float64](),
		rotation: rotation,
	}
}
