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

func (trs TRS) Matrix() mat.Matrix4x4 {
	// https://github.com/UltravioletFramework/ultraviolet/issues/92
	p := trs.position
	s := trs.scale

	rotX := trs.rotation.Dir().X()
	rotY := trs.rotation.Dir().Y()
	rotZ := trs.rotation.Dir().Z()
	rotW := trs.rotation.W()

	xx := rotX * rotX
	yy := rotY * rotY
	zz := rotZ * rotZ
	xy := rotX * rotY
	zw := rotZ * rotW
	zx := rotZ * rotX
	yw := rotY * rotW
	yz := rotY * rotZ
	xw := rotX * rotW

	var result mat.Matrix4x4
	result.X00 = s.X() * (1 - (2 * (yy + zz)))
	result.X10 = s.X() * (2 * (xy + zw))
	result.X20 = s.X() * (2 * (zx - yw))
	result.X30 = 0
	result.X01 = s.Y() * (2 * (xy - zw))
	result.X11 = s.Y() * (1 - (2 * (zz + xx)))
	result.X21 = s.Y() * (2 * (yz + xw))
	result.X31 = 0
	result.X02 = s.Z() * (2 * (zx + yw))
	result.X12 = s.Z() * (2 * (yz - xw))
	result.X22 = s.Z() * (1 - (2 * (yy + xx)))
	result.X32 = 0
	result.X03 = p.X()
	result.X13 = p.Y()
	result.X23 = p.Z()
	result.X33 = 1
	return result
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

func (trs TRS) Inverse() TRS {
	return FromMatrix(trs.Matrix().Inverse())
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

// Look at returns a new TRS with it's rotation modified to be looking at the
// position provided
func (trs TRS) LookAt(positionToLookAt vector3.Float64) TRS {
	forward := positionToLookAt.Sub(trs.position).Normalized()
	up := vector3.Up[float64]()
	right := forward.Cross(up).Normalized()
	up = forward.Cross(right).Normalized()

	trs.rotation = quaternion.FromMatrix(mat.Matrix4x4{
		right.X(), up.X(), forward.X(), 0,
		right.Y(), up.Y(), forward.Y(), 0,
		right.Z(), up.Z(), forward.Z(), 0,
		0, 0, 0, 1,
	})

	return trs
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

func Positions(transforms []TRS) []vector3.Float64 {
	positions := make([]vector3.Float64, len(transforms))
	for i, v := range transforms {
		positions[i] = v.position
	}
	return positions
}

func Scales(transforms []TRS) []vector3.Float64 {
	scales := make([]vector3.Float64, len(transforms))
	for i, v := range transforms {
		scales[i] = v.scale
	}
	return scales
}

func Rotations(transforms []TRS) []quaternion.Quaternion {
	rotations := make([]quaternion.Quaternion, len(transforms))
	for i, v := range transforms {
		rotations[i] = v.rotation
	}
	return rotations
}
