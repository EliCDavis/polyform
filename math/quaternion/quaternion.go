package quaternion

import (
	"math"

	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

// Quaternion is a 4 component imaginary number thingy for rotating 3D meshes.
type Quaternion struct {
	v vector3.Float64
	w float64
}

// NewQuaternion creates a quaternion
func New(v vector3.Float64, w float64) Quaternion {
	return Quaternion{v, w}
}

// QuaternionZero returns a quaternion with 0 for all it's components
func Zero() Quaternion {
	return Quaternion{vector3.Zero[float64](), 0}
}

// Rotate takes a given vector and rotates it with by this quaternion.
//
// Resources Used:
//
//	https://gamedev.stackexchange.com/questions/28395
func (q Quaternion) Rotate(v vector3.Float64) vector3.Float64 {
	return q.v.Scale(q.v.Dot(v) * 2.0).
		Add(v.Scale(math.Pow(q.w, 2.0) - q.v.Dot(q.v))).
		Add(q.v.Cross(v).Scale(2.0 * q.w))
}

func (q Quaternion) RotateArray(arr []vector3.Float64) []vector3.Float64 {
	results := make([]vector3.Float64, len(arr))
	for i, v := range arr {
		results[i] = q.Rotate(v)
	}
	return results
}

func (q Quaternion) Normalize() Quaternion {
	vq := vector4.New(q.v.X(), q.v.Y(), q.v.Z(), q.w).Normalized()

	return Quaternion{
		v: vector3.New(vq.X(), vq.Y(), vq.Z()),
		w: vq.W(),
	}
}

// https://github.com/toji/gl-matrix/blob/f0583ef53e94bc7e78b78c8a24f09ed5e2f7a20c/src/gl-matrix/quat.js#L179
func (q Quaternion) Multiply(other Quaternion) Quaternion {
	ax := q.v.X()
	ay := q.v.Y()
	az := q.v.Z()
	aw := q.w
	bx := other.v.X()
	by := other.v.Y()
	bz := other.v.Z()
	bw := other.w

	return Quaternion{
		v: vector3.New(
			ax*bw+aw*bx+ay*bz-az*by,
			ay*bw+aw*by+az*bx-ax*bz,
			az*bw+aw*bz+ax*by-ay*bx,
		),
		w: aw*bw - ax*bx - ay*by - az*bz,
	}
}

// https://github.com/toji/gl-matrix/blob/f0583ef53e94bc7e78b78c8a24f09ed5e2f7a20c/src/gl-matrix/quat.js#L54
func RotationTo(from, to vector3.Float64) Quaternion {
	dot := from.Dot(to)

	if dot < -0.999999 {
		cross := vector3.Right[float64]().Cross(from).Normalized()

		if cross.Length() < 0.000001 {
			cross = vector3.Up[float64]().Cross(from).Normalized()
		}

		return FromTheta(math.Pi, cross.Normalized())
	} else if dot > 0.999999 {
		return New(vector3.Zero[float64](), 1)
	}

	cross := from.Cross(to)
	return New(cross, 1+dot).Normalize()
}

// UnitQuaternionFromTheta takes a vector and angle and builds a unit
// quaternion in the form (cos(theta/2.0), sin(theta/2.0))
//
// Resources Used:
//
//	https://www.youtube.com/watch?v=mHVwd8gYLnI
//	https://en.wikipedia.org/wiki/Quaternions_and_spatial_rotation
func FromTheta(theta float64, v vector3.Float64) Quaternion {
	return Quaternion{
		w: math.Cos(theta / 2.0),
		v: v.Normalized().Scale(math.Sin(theta / 2.0)),
	}
}
