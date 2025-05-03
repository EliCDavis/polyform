package trs

import (
	"fmt"
	"math"
	"strconv"

	"github.com/EliCDavis/polyform/math/mat"
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

// https://github.com/UltravioletFramework/ultraviolet/issues/92
func (trs TRS) Matrix() mat.Matrix4x4 {
	p := trs.position
	s := trs.scale

	rotX2 := trs.rotation.Dir().X() * 2.
	rotY2 := trs.rotation.Dir().Y() * 2.
	rotZ2 := trs.rotation.Dir().Z() * 2.

	xx := trs.rotation.Dir().X() * rotX2
	xy := trs.rotation.Dir().X() * rotY2
	xz := trs.rotation.Dir().X() * rotZ2
	yy := trs.rotation.Dir().Y() * rotY2
	yz := trs.rotation.Dir().Y() * rotZ2
	zz := trs.rotation.Dir().Z() * rotZ2
	wx := trs.rotation.W() * rotX2
	wy := trs.rotation.W() * rotY2
	wz := trs.rotation.W() * rotZ2

	return mat.Matrix4x4{
		(1 - (yy + zz)) * s.X(), (xy - wz) * s.Y(), (xz + wy) * s.Z(), p.X(),
		(xy + wz) * s.X(), (1 - (xx + zz)) * s.Y(), (yz - wx) * s.Z(), p.Y(),
		(xz - wy) * s.X(), (yz + wx) * s.Y(), (1 - (xx + yy)) * s.Z(), p.Z(),
		0, 0, 0, 1,
	}
}

func (trs TRS) RotateDirection(in vector3.Float64) vector3.Float64 {
	return trs.rotation.Rotate(in)
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

func (trs TRS) SetScale(in vector3.Float64) TRS {
	return TRS{
		position: trs.position,
		scale:    in,
		rotation: trs.rotation,
	}
}

func (trs TRS) SetRotation(in quaternion.Quaternion) TRS {
	return TRS{
		position: trs.position,
		scale:    trs.scale,
		rotation: in,
	}
}

func (trs TRS) SetTranslation(in vector3.Float64) TRS {
	return TRS{
		position: in,
		scale:    trs.scale,
		rotation: trs.rotation,
	}
}

func (trs TRS) Multiply(other TRS) TRS {
	return FromMatrix(trs.Matrix().Multiply(other.Matrix()))
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

func inDelta(a, b, d float64, component string) error {
	diff := math.Abs(a - b)
	if diff <= d {
		return nil
	}
	return fmt.Errorf(
		"expected %s %s to be within delta (%s) of %s",
		component,
		strconv.FormatFloat(a, 'f', -1, 64),
		strconv.FormatFloat(d, 'f', -1, 64),
		strconv.FormatFloat(b, 'f', -1, 64),
	)
}

// Checks if each of the components of this TRS is within delta of TRS passed
// in.
// If they aren't, an error is returned describing which component is out of
// range
func (trs TRS) WithinDelta(in TRS, delta float64) error {
	cases := []struct {
		component string
		a         float64
		b         float64
	}{
		{component: "position.x", a: trs.position.X(), b: in.position.X()},
		{component: "position.y", a: trs.position.Y(), b: in.position.Y()},
		{component: "position.z", a: trs.position.Z(), b: in.position.Z()},

		{component: "rotation.x", a: trs.rotation.Dir().X(), b: in.rotation.Dir().X()},
		{component: "rotation.y", a: trs.rotation.Dir().Y(), b: in.rotation.Dir().Y()},
		{component: "rotation.z", a: trs.rotation.Dir().Z(), b: in.rotation.Dir().Z()},
		{component: "rotation.w", a: trs.rotation.W(), b: in.rotation.W()},

		{component: "scale.x", a: trs.scale.X(), b: in.scale.X()},
		{component: "scale.y", a: trs.scale.Y(), b: in.scale.Y()},
		{component: "scale.z", a: trs.scale.Z(), b: in.scale.Z()},
	}

	for _, c := range cases {
		if err := inDelta(c.a, c.b, delta, c.component); err != nil {
			return err
		}
	}

	return nil
}
