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

func FromMatrix(m mat.Matrix4x4) TRS {
	// https://github.com/CedricGuillemet/ImGuizmo/blob/cf287a3fd48d503ad19a8f8ca81a1a15b63bccf1/ImGuizmo.cpp#L2359
	// https://github.com/UltravioletFramework/ultraviolet/blob/main/Source/Ultraviolet/Mathematics/Matrix.cs#L2251

	// TODO: Do we panic when any component of scale is 0?

	// xNeg := (m.X00 * m.X10 * m.X20) < 0
	// xSign := 1.
	// if xNeg {
	// 	xSign = -1
	// }

	// yNeg := (m.X01 * m.X11 * m.X21) < 0
	// ySign := 1.
	// if yNeg {
	// 	ySign = -1
	// }

	// zNeg := (m.X02 * m.X12 * m.X22) < 0
	// zSign := 1.
	// if zNeg {
	// 	zSign = -1
	// }

	scale := vector3.New(
		vector3.New(m.X00, m.X10, m.X20).Length(),
		vector3.New(m.X01, m.X11, m.X21).Length(),
		vector3.New(m.X02, m.X12, m.X22).Length(),
	)

	rotM := mat.Matrix4x4{}
	rotM.X00 = m.X00 / scale.X()
	rotM.X10 = m.X10 / scale.X()
	rotM.X20 = m.X20 / scale.X()

	rotM.X01 = m.X01 / scale.Y()
	rotM.X11 = m.X11 / scale.Y()
	rotM.X21 = m.X21 / scale.Y()

	rotM.X02 = m.X02 / scale.Z()
	rotM.X12 = m.X12 / scale.Z()
	rotM.X22 = m.X22 / scale.Z()

	rotM.X33 = 1

	// scale = vector3.New(
	// 	vector3.New(m.X00, m.X10, m.X20).Length(),
	// 	vector3.New(m.X01, m.X11, m.X21).Length(),
	// 	vector3.New(m.X02, m.X12, m.X22).Length(),
	// )

	return TRS{
		position: vector3.New(m.X03, m.X13, m.X23),
		scale:    scale,
		rotation: quaternion.FromMatrix(rotM),
	}
}
