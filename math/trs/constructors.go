package trs

import (
	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/vector/vector3"
)

func Identity() TRS {
	return TRS{
		position: vector3.Zero[float64](),
		rotation: quaternion.Identity(),
		scale:    vector3.One[float64](),
	}
}

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

// https://github.com/CedricGuillemet/ImGuizmo/blob/cf287a3fd48d503ad19a8f8ca81a1a15b63bccf1/ImGuizmo.cpp#L2359
func FromMatrix(m mat.Matrix4x4) TRS {
	return TRS{
		position: vector3.New(m.X03, m.X13, m.X23),
		scale: vector3.New(
			vector3.New(m.X00, m.X10, m.X20).Length(),
			vector3.New(m.X01, m.X11, m.X21).Length(),
			vector3.New(m.X02, m.X12, m.X22).Length(),
		),
		rotation: quaternion.FromMatrix(m),
	}
}
