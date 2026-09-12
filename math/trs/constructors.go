package trs

import (
	"fmt"
	"math"

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

// ShearTolerance is how far from perpendicular a matrix's axes may drift
// before FromMatrix refuses it.
const ShearTolerance = 1e-9

// Shear reports how far m is from being a rotation and an axis-aligned
// scale, as the largest cosine between its axes. Zero means the axes are
// perpendicular and the matrix is a TRS.
func Shear(m mat.Matrix4x4) float64 {
	axes := [3]vector3.Float64{
		vector3.New(m.X00, m.X10, m.X20),
		vector3.New(m.X01, m.X11, m.X21),
		vector3.New(m.X02, m.X12, m.X22),
	}

	worst := 0.
	for i := 0; i < 3; i++ {
		for j := i + 1; j < 3; j++ {
			lengths := axes[i].Length() * axes[j].Length()
			if lengths == 0 {
				continue
			}
			worst = math.Max(worst, math.Abs(axes[i].Dot(axes[j]))/lengths)
		}
	}
	return worst
}

// FromMatrix decomposes m into a TRS.
//
// Errors on matrixes with shear
func FromMatrix(m mat.Matrix4x4) (TRS, error) {
	// https://github.com/CedricGuillemet/ImGuizmo/blob/cf287a3fd48d503ad19a8f8ca81a1a15b63bccf1/ImGuizmo.cpp#L2359
	// https://github.com/UltravioletFramework/ultraviolet/blob/main/Source/Ultraviolet/Mathematics/Matrix.cs#L2251

	scale := vector3.New(
		vector3.New(m.X00, m.X10, m.X20).Length(),
		vector3.New(m.X01, m.X11, m.X21).Length(),
		vector3.New(m.X02, m.X12, m.X22).Length(),
	)

	if scale.X() == 0 || scale.Y() == 0 || scale.Z() == 0 {
		return TRS{}, fmt.Errorf("matrix collapses an axis (scale %v), leaving no rotation to extract", scale)
	}

	if shear := Shear(m); shear > ShearTolerance {
		return TRS{}, fmt.Errorf("matrix contains shear (%.4f off perpendicular) and is not a TRS; a non-uniform scale composed with a rotation produces this", shear)
	}

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

	return TRS{
		position: vector3.New(m.X03, m.X13, m.X23),
		scale:    scale,
		rotation: quaternion.FromMatrix(rotM),
	}, nil
}
