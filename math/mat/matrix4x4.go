package mat

import "github.com/EliCDavis/vector/vector3"

type Matrix4x4 struct {
	X00, X01, X02, X03 float64
	X10, X11, X12, X13 float64
	X20, X21, X22, X23 float64
	X30, X31, X32, X33 float64
}

func MatFromDirs(up, forward, offset vector3.Float64) Matrix4x4 {
	left := up.Cross(forward).Normalized()
	newFwd := left.Cross(up).Normalized()
	return Matrix4x4{
		left.X(), up.X(), newFwd.X(), offset.X(),
		left.Y(), up.Y(), newFwd.Y(), offset.Y(),
		left.Z(), up.Z(), newFwd.Z(), offset.Z(),
		0, 0, 0, 1,
	}
}

// FromColArray creates a Matrix4x4 from a 1D array of 16 elements, interpreting them as columns.
func FromColArray(matrix [16]float64) Matrix4x4 {
	return Matrix4x4{
		X00: matrix[0], X01: matrix[4], X02: matrix[8], X03: matrix[12], // Row 0
		X10: matrix[1], X11: matrix[5], X12: matrix[9], X13: matrix[13], // Row 1
		X20: matrix[2], X21: matrix[6], X22: matrix[10], X23: matrix[14], // Row 2
		X30: matrix[3], X31: matrix[7], X32: matrix[11], X33: matrix[15], // Row 3
	}
}

func Identity() Matrix4x4 {
	return Matrix4x4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func (a Matrix4x4) Add(b Matrix4x4) Matrix4x4 {
	return Matrix4x4{
		a.X00 + b.X00, a.X10 + b.X10, a.X20 + b.X20, a.X30 + b.X30,
		a.X01 + b.X01, a.X11 + b.X11, a.X21 + b.X21, a.X31 + b.X31,
		a.X02 + b.X02, a.X12 + b.X12, a.X22 + b.X22, a.X32 + b.X32,
		a.X03 + b.X03, a.X13 + b.X13, a.X23 + b.X23, a.X33 + b.X33,
	}
}

func (a Matrix4x4) MulPosition(b vector3.Float64) vector3.Float64 {
	X := a.X00*b.X() + a.X01*b.Y() + a.X02*b.Z() + a.X03
	y := a.X10*b.X() + a.X11*b.Y() + a.X12*b.Z() + a.X13
	z := a.X20*b.X() + a.X21*b.Y() + a.X22*b.Z() + a.X23
	return vector3.New(X, y, z)
}

func (a Matrix4x4) Determinant() float64 {
	return a.X00*a.X11*a.X22*a.X33 - a.X00*a.X11*a.X23*a.X32 +
		a.X00*a.X12*a.X23*a.X31 - a.X00*a.X12*a.X21*a.X33 +
		a.X00*a.X13*a.X21*a.X32 - a.X00*a.X13*a.X22*a.X31 -
		a.X01*a.X12*a.X23*a.X30 + a.X01*a.X12*a.X20*a.X33 -
		a.X01*a.X13*a.X20*a.X32 + a.X01*a.X13*a.X22*a.X30 -
		a.X01*a.X10*a.X22*a.X33 + a.X01*a.X10*a.X23*a.X32 +
		a.X02*a.X13*a.X20*a.X31 - a.X02*a.X13*a.X21*a.X30 +
		a.X02*a.X10*a.X21*a.X33 - a.X02*a.X10*a.X23*a.X31 +
		a.X02*a.X11*a.X23*a.X30 - a.X02*a.X11*a.X20*a.X33 -
		a.X03*a.X10*a.X21*a.X32 + a.X03*a.X10*a.X22*a.X31 -
		a.X03*a.X11*a.X22*a.X30 + a.X03*a.X11*a.X20*a.X32 -
		a.X03*a.X12*a.X20*a.X31 + a.X03*a.X12*a.X21*a.X30
}

func (a Matrix4x4) Inverse() Matrix4x4 {
	m := Matrix4x4{}
	r := 1 / a.Determinant()
	m.X00 = (a.X12*a.X23*a.X31 - a.X13*a.X22*a.X31 + a.X13*a.X21*a.X32 - a.X11*a.X23*a.X32 - a.X12*a.X21*a.X33 + a.X11*a.X22*a.X33) * r
	m.X01 = (a.X03*a.X22*a.X31 - a.X02*a.X23*a.X31 - a.X03*a.X21*a.X32 + a.X01*a.X23*a.X32 + a.X02*a.X21*a.X33 - a.X01*a.X22*a.X33) * r
	m.X02 = (a.X02*a.X13*a.X31 - a.X03*a.X12*a.X31 + a.X03*a.X11*a.X32 - a.X01*a.X13*a.X32 - a.X02*a.X11*a.X33 + a.X01*a.X12*a.X33) * r
	m.X03 = (a.X03*a.X12*a.X21 - a.X02*a.X13*a.X21 - a.X03*a.X11*a.X22 + a.X01*a.X13*a.X22 + a.X02*a.X11*a.X23 - a.X01*a.X12*a.X23) * r
	m.X10 = (a.X13*a.X22*a.X30 - a.X12*a.X23*a.X30 - a.X13*a.X20*a.X32 + a.X10*a.X23*a.X32 + a.X12*a.X20*a.X33 - a.X10*a.X22*a.X33) * r
	m.X11 = (a.X02*a.X23*a.X30 - a.X03*a.X22*a.X30 + a.X03*a.X20*a.X32 - a.X00*a.X23*a.X32 - a.X02*a.X20*a.X33 + a.X00*a.X22*a.X33) * r
	m.X12 = (a.X03*a.X12*a.X30 - a.X02*a.X13*a.X30 - a.X03*a.X10*a.X32 + a.X00*a.X13*a.X32 + a.X02*a.X10*a.X33 - a.X00*a.X12*a.X33) * r
	m.X13 = (a.X02*a.X13*a.X20 - a.X03*a.X12*a.X20 + a.X03*a.X10*a.X22 - a.X00*a.X13*a.X22 - a.X02*a.X10*a.X23 + a.X00*a.X12*a.X23) * r
	m.X20 = (a.X11*a.X23*a.X30 - a.X13*a.X21*a.X30 + a.X13*a.X20*a.X31 - a.X10*a.X23*a.X31 - a.X11*a.X20*a.X33 + a.X10*a.X21*a.X33) * r
	m.X21 = (a.X03*a.X21*a.X30 - a.X01*a.X23*a.X30 - a.X03*a.X20*a.X31 + a.X00*a.X23*a.X31 + a.X01*a.X20*a.X33 - a.X00*a.X21*a.X33) * r
	m.X22 = (a.X01*a.X13*a.X30 - a.X03*a.X11*a.X30 + a.X03*a.X10*a.X31 - a.X00*a.X13*a.X31 - a.X01*a.X10*a.X33 + a.X00*a.X11*a.X33) * r
	m.X23 = (a.X03*a.X11*a.X20 - a.X01*a.X13*a.X20 - a.X03*a.X10*a.X21 + a.X00*a.X13*a.X21 + a.X01*a.X10*a.X23 - a.X00*a.X11*a.X23) * r
	m.X30 = (a.X12*a.X21*a.X30 - a.X11*a.X22*a.X30 - a.X12*a.X20*a.X31 + a.X10*a.X22*a.X31 + a.X11*a.X20*a.X32 - a.X10*a.X21*a.X32) * r
	m.X31 = (a.X01*a.X22*a.X30 - a.X02*a.X21*a.X30 + a.X02*a.X20*a.X31 - a.X00*a.X22*a.X31 - a.X01*a.X20*a.X32 + a.X00*a.X21*a.X32) * r
	m.X32 = (a.X02*a.X11*a.X30 - a.X01*a.X12*a.X30 - a.X02*a.X10*a.X31 + a.X00*a.X12*a.X31 + a.X01*a.X10*a.X32 - a.X00*a.X11*a.X32) * r
	m.X33 = (a.X01*a.X12*a.X20 - a.X02*a.X11*a.X20 + a.X02*a.X10*a.X21 - a.X00*a.X12*a.X21 - a.X01*a.X10*a.X22 + a.X00*a.X11*a.X22) * r
	return m
}

func (a Matrix4x4) Multiply(b Matrix4x4) Matrix4x4 {
	return Matrix4x4{
		X00: (a.X00 * b.X00) + (a.X01 * b.X10) + (a.X02 * b.X20) + (a.X03 * b.X30),
		X01: (a.X00 * b.X01) + (a.X01 * b.X11) + (a.X02 * b.X21) + (a.X03 * b.X31),
		X02: (a.X00 * b.X02) + (a.X01 * b.X12) + (a.X02 * b.X22) + (a.X03 * b.X32),
		X03: (a.X00 * b.X03) + (a.X01 * b.X13) + (a.X02 * b.X23) + (a.X03 * b.X33),

		X10: (a.X10 * b.X00) + (a.X11 * b.X10) + (a.X12 * b.X20) + (a.X13 * b.X30),
		X11: (a.X10 * b.X01) + (a.X11 * b.X11) + (a.X12 * b.X21) + (a.X13 * b.X31),
		X12: (a.X10 * b.X02) + (a.X11 * b.X12) + (a.X12 * b.X22) + (a.X13 * b.X32),
		X13: (a.X10 * b.X03) + (a.X11 * b.X13) + (a.X12 * b.X23) + (a.X13 * b.X33),

		X20: (a.X20 * b.X00) + (a.X21 * b.X10) + (a.X22 * b.X20) + (a.X23 * b.X30),
		X21: (a.X20 * b.X01) + (a.X21 * b.X11) + (a.X22 * b.X21) + (a.X23 * b.X31),
		X22: (a.X20 * b.X02) + (a.X21 * b.X12) + (a.X22 * b.X22) + (a.X23 * b.X32),
		X23: (a.X20 * b.X03) + (a.X21 * b.X13) + (a.X22 * b.X23) + (a.X23 * b.X33),

		X30: (a.X30 * b.X00) + (a.X31 * b.X10) + (a.X32 * b.X20) + (a.X33 * b.X30),
		X31: (a.X30 * b.X01) + (a.X31 * b.X11) + (a.X32 * b.X21) + (a.X33 * b.X31),
		X32: (a.X30 * b.X02) + (a.X31 * b.X12) + (a.X32 * b.X22) + (a.X33 * b.X32),
		X33: (a.X30 * b.X03) + (a.X31 * b.X13) + (a.X32 * b.X23) + (a.X33 * b.X33),
	}
}
