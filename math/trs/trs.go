package trs

import (
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/vector/vector3"
)

type TRS struct {
	position vector3.Float64
	scale    vector3.Float64
	rotation quaternion.Quaternion
}

// The position of the TRS
func (trs TRS) Position() vector3.Float64 {
	return trs.position
}

// The scale of the TRS
func (trs TRS) Scale() vector3.Float64 {
	return trs.scale
}

// The rotation of the TRS
func (trs TRS) Rotation() quaternion.Quaternion {
	return trs.rotation
}

// Transform a point by the TRS
func (trs TRS) Transform(in vector3.Float64) vector3.Float64 {
	return trs.rotation.Rotate(trs.scale.MultByVector(in)).Add(trs.position)
}

// Create a new TRS with the position translated by "in"
func (trs TRS) Translate(in vector3.Float64) TRS {
	return TRS{
		position: trs.position.Add(in),
		scale:    trs.scale,
		rotation: trs.rotation,
	}
}

func (trs TRS) Multiply(other TRS) TRS {
	return TRS{
		position: trs.position.Add(other.position),
		rotation: trs.rotation.Multiply(other.rotation),
		scale:    trs.scale.MultByVector(other.scale),
	}
}

// Transform an array of points by the TRS
func (trs TRS) TransformArray(in []vector3.Float64) []vector3.Float64 {
	out := make([]vector3.Float64, len(in))
	for i, v := range in {
		out[i] = trs.Transform(v)
	}
	return out
}

// Transform an array of points by the TRS and store those changes in the
// array passed in
func (trs TRS) TransformInPlace(in []vector3.Float64) {
	for i, v := range in {
		in[i] = trs.Transform(v)
	}
}
