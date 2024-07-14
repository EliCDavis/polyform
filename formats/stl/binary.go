package stl

import (
	"github.com/EliCDavis/vector/vector3"
)

type Header [80]byte

type Vec struct {
	X float32
	Y float32
	Z float32
}

func (v Vec) Zero() bool {
	return v.X == 0 && v.Y == 0 && v.Z == 0
}

func (v Vec) Float64() vector3.Float64 {
	return vector3.New(v.X, v.Y, v.Z).ToFloat64()
}

type Triangle struct {
	Normal    Vec
	Vertex1   Vec
	Vertex2   Vec
	Vertex3   Vec
	Attribute uint16
}

type Binary struct {
	Header    Header
	Triangles []Triangle
}
