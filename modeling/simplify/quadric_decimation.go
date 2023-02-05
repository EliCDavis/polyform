package simplify

import (
	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

func QuadricVector(a mat.Matrix4x4) vector3.Float64 {
	b := mat.Matrix4x4{
		X00: a.X00, X01: a.X01, X02: a.X02, X03: a.X03,
		X10: a.X10, X11: a.X11, X12: a.X12, X13: a.X13,
		X20: a.X20, X21: a.X21, X22: a.X22, X23: a.X23,
		X30: 0, X31: 0, X32: 0, X33: 1,
	}
	return b.Inverse().MulPosition(vector3.Zero[float64]())
}

func QuadricErrorForVector(a mat.Matrix4x4, v vector3.Float64) float64 {
	return (v.X()*a.X00*v.X() + v.Y()*a.X10*v.X() + v.Z()*a.X20*v.X() + a.X30*v.X() +
		v.X()*a.X01*v.Y() + v.Y()*a.X11*v.Y() + v.Z()*a.X21*v.Y() + a.X31*v.Y() +
		v.X()*a.X02*v.Z() + v.Y()*a.X12*v.Z() + v.Z()*a.X22*v.Z() + a.X32*v.Z() +
		v.X()*a.X03 + v.Y()*a.X13 + v.Z()*a.X23 + a.X33)
}

func QuadricDecimation(m modeling.Mesh) modeling.Mesh {
	return m
}
