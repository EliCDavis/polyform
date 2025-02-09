package quaternion

import (
	"math"

	"github.com/EliCDavis/polyform/math/mat"
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

// Zero returns a quaternion with 0 for all it's components
func Zero() Quaternion {
	return Quaternion{vector3.Zero[float64](), 0}
}

func Identity() Quaternion {
	return Quaternion{vector3.Zero[float64](), 1}
}

func (q Quaternion) Vector4() vector4.Float64 {
	return vector4.New(q.v.X(), q.v.Y(), q.v.Z(), q.w)
}

func (q Quaternion) Dir() vector3.Float64 {
	return q.v
}

func (q Quaternion) W() float64 {
	return q.w
}

func (q Quaternion) ToArr() [4]float64 {
	return [4]float64{q.v.X(), q.v.Y(), q.v.Z(), q.w}
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

// https://en.wikipedia.org/wiki/Conversion_between_quaternions_and_Euler_angles
func (q Quaternion) ToEulerAngles() vector3.Float64 {
	v := q.v

	// roll (x-axis rotation)
	sinr_cosp := 2 * (q.w*v.X() + v.Y()*v.Z())
	cosr_cosp := 1 - 2*(v.X()*v.X()+v.Y()*v.Y())
	x := math.Atan2(sinr_cosp, cosr_cosp)

	// pitch (y-axis rotation)
	sinp := math.Sqrt(1 + 2*(q.w*v.Y()-v.X()*v.Z()))
	cosp := math.Sqrt(1 - 2*(q.w*v.Y()-v.X()*v.Z()))
	y := 2*math.Atan2(sinp, cosp) - math.Pi/2

	// yaw (z-axis rotation)
	siny_cosp := 2 * (q.w*v.Z() + v.X()*v.Y())
	cosy_cosp := 1 - 2*(v.Y()*v.Y()+v.Z()*v.Z())
	z := math.Atan2(siny_cosp, cosy_cosp)

	return vector3.New(x, y, z)
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

// https://en.wikipedia.org/wiki/Conversion_between_quaternions_and_Euler_angles
func FromEulerAngles(angles vector3.Float64) Quaternion {
	cr := math.Cos(angles.X() * 0.5)
	sr := math.Sin(angles.X() * 0.5)
	cp := math.Cos(angles.Y() * 0.5)
	sp := math.Sin(angles.Y() * 0.5)
	cy := math.Cos(angles.Z() * 0.5)
	sy := math.Sin(angles.Z() * 0.5)

	return Quaternion{
		v: vector3.New(
			sr*cp*cy-cr*sp*sy,
			cr*sp*cy+sr*cp*sy,
			cr*cp*sy-sr*sp*cy,
		),
		w: cr*cp*cy + sr*sp*sy,
	}
}

func FromMatrix(k mat.Matrix4x4) Quaternion {
	var sqrtV float64
	var half float64
	scale := k.X00 + k.X11 + k.X22
	var result Quaternion

	// if scale > 0.0 {
	// 	sqrtV = math.Sqrt(scale + 1.0)
	// 	result.w = sqrtV * 0.5
	// 	sqrtV = 0.5 / sqrtV

	// 	result.v = vector3.New(
	// 		(m.M23-m.M32)*sqrtV,
	// 		(m.M31-m.M13)*sqrtV,
	// 		(m.M12-m.M21)*sqrtV,
	// 	)
	// } else if m.X00 >= m.X11 && m.X00 >= m.X22 {
	// 	sqrtV = math.Sqrt(1.0 + m.X00 - m.X11 - m.X22)
	// 	half = 0.5 / sqrtV

	// 	result.v = vector3.New(
	// 		0.5*sqrtV,
	// 		(m.M12+m.M21)*half,
	// 		(m.M13+m.M31)*half,
	// 	)
	// 	result.w = (m.M23 - m.M32) * half
	// } else if m.X11 > m.X22 {
	// 	sqrtV = math.Sqrt(1.0 + m.X11 - m.X00 - m.X22)
	// 	half = 0.5 / sqrtV

	// 	result.v = vector3.New(
	// 		(m.M21+m.M12)*half,
	// 		0.5*sqrtV,
	// 		(m.M32+m.M23)*half,
	// 	)
	// 	result.w = (m.M31 - m.M13) * half
	// } else {
	// 	sqrtV = math.Sqrt(1.0 + m.X22 - m.X00 - m.X11)
	// 	half = 0.5 / sqrtV

	// 	result.v = vector3.New(
	// 		(m.M31+m.M13)*half,
	// 		(m.M32+m.M23)*half,
	// 		0.5*sqrtV,
	// 	)
	// 	result.w = (m.M12 - m.M21) * half
	// }

	if scale > 0.0 {
		sqrtV = math.Sqrt(scale + 1.0)
		result.w = sqrtV * 0.5
		sqrtV = 0.5 / sqrtV

		result.v = vector3.New(
			(k.X12-k.X21)*sqrtV,
			(k.X20-k.X02)*sqrtV,
			(k.X01-k.X10)*sqrtV,
		)
	} else if k.X00 >= k.X11 && k.X00 >= k.X22 {
		sqrtV = math.Sqrt(1.0 + k.X00 - k.X11 - k.X22)
		half = 0.5 / sqrtV

		result.v = vector3.New(
			0.5*sqrtV,
			(k.X01+k.X10)*half,
			(k.X02+k.X20)*half,
		)
		result.w = (k.X12 - k.X21) * half
	} else if k.X11 > k.X22 {
		sqrtV = math.Sqrt(1.0 + k.X11 - k.X00 - k.X22)
		half = 0.5 / sqrtV

		result.v = vector3.New(
			(k.X10+k.X01)*half,
			0.5*sqrtV,
			(k.X21+k.X12)*half,
		)
		result.w = (k.X20 - k.X02) * half
	} else {
		sqrtV = math.Sqrt(1.0 + k.X22 - k.X00 - k.X11)
		half = 0.5 / sqrtV

		result.v = vector3.New(
			(k.X20+k.X02)*half,
			(k.X21+k.X12)*half,
			0.5*sqrtV,
		)
		result.w = (k.X01 - k.X10) * half
	}

	return result.Normalize()
}
