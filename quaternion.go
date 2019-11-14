package mesh

import (
	"math"

	"github.com/EliCDavis/vector"
)

// Quaternion is a 4 component imaginary number thingy for rotating 3D meshes.
type Quaternion struct {
	v vector.Vector3
	w float64
}

// NewQuaternion creates a quaternion
func NewQuaternion(v vector.Vector3, w float64) Quaternion {
	return Quaternion{v, w}
}

// QuaternionZero returns a quaternion with 0 for all it's components
func QuaternionZero() Quaternion {
	return Quaternion{vector.Vector3Zero(), 0}
}

// Rotate takes a given vector and rotates it with by this quaternion.
//
// Resources Used:
//     https://gamedev.stackexchange.com/questions/28395
func (q Quaternion) Rotate(v vector.Vector3) vector.Vector3 {
	return q.v.MultByConstant(q.v.Dot(v) * 2.0).
		Add(v.MultByConstant(math.Pow(q.w, 2.0) - q.v.Dot(q.v))).
		Add(q.v.Cross(v).MultByConstant(2.0 * q.w))
}

// UnitQuaternionFromTheta takes a vector and angle and builds a unit
// quaternion in the form (cos(theta/2.0), sin(theta/2.0))
//
// Resources Used:
//     https://www.youtube.com/watch?v=mHVwd8gYLnI
//     https://en.wikipedia.org/wiki/Quaternions_and_spatial_rotation
func UnitQuaternionFromTheta(theta float64, v vector.Vector3) Quaternion {
	return Quaternion{
		w: math.Cos(theta / 2.0),
		v: v.Normalized().MultByConstant(math.Sin(theta / 2.0)),
	}
}
