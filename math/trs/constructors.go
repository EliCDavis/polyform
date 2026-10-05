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
// Errors on matrixes with shear. An axis m collapses keeps a scale of zero,
// and the rotation comes from the axes that are left.
func FromMatrix(m mat.Matrix4x4) (TRS, error) {
	// https://github.com/CedricGuillemet/ImGuizmo/blob/cf287a3fd48d503ad19a8f8ca81a1a15b63bccf1/ImGuizmo.cpp#L2359
	// https://github.com/UltravioletFramework/ultraviolet/blob/main/Source/Ultraviolet/Mathematics/Matrix.cs#L2251

	columns := [3]vector3.Float64{
		vector3.New(m.X00, m.X10, m.X20),
		vector3.New(m.X01, m.X11, m.X21),
		vector3.New(m.X02, m.X12, m.X22),
	}
	scale := vector3.New(columns[0].Length(), columns[1].Length(), columns[2].Length())

	if shear := Shear(m); shear > ShearTolerance {
		return TRS{}, fmt.Errorf("matrix contains shear (%.4f off perpendicular) and is not a TRS; a non-uniform scale composed with a rotation produces this", shear)
	}

	axes := rotationAxes(columns)
	rotM := mat.Matrix4x4{
		X00: axes[0].X(), X01: axes[1].X(), X02: axes[2].X(),
		X10: axes[0].Y(), X11: axes[1].Y(), X12: axes[2].Y(),
		X20: axes[0].Z(), X21: axes[1].Z(), X22: axes[2].Z(),
		X33: 1,
	}

	return TRS{
		position: vector3.New(m.X03, m.X13, m.X23),
		scale:    scale,
		rotation: quaternion.FromMatrix(rotM),
	}, nil
}

// rotationAxes normalizes each column with any length and rebuilds the
// collapsed ones square to them, so the axes always form a rotation.
func rotationAxes(columns [3]vector3.Float64) [3]vector3.Float64 {
	var axes [3]vector3.Float64
	var kept []int
	for i, column := range columns {
		if length := column.Length(); length > 0 {
			axes[i] = column.Scale(1 / length)
			kept = append(kept, i)
		}
	}

	switch len(kept) {
	case 0:
		return [3]vector3.Float64{vector3.Right[float64](), vector3.Up[float64](), vector3.Forward[float64]()}
	case 1:
		i := kept[0]
		axes[(i+1)%3] = axes[i].Perpendicular().Normalized()
		axes[(i+2)%3] = axes[i].Cross(axes[(i+1)%3])
	case 2:
		missing := 3 - kept[0] - kept[1]
		axes[missing] = axes[(missing+1)%3].Cross(axes[(missing+2)%3])
	}
	return axes
}
