package quaternion_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

// A build once reported that this function's X rotation had a flipped
// sign, having inferred the convention from where a rotated part landed
// rather than from the rotation itself. It does not: these are the
// standard right-handed rotations. Pinning each axis here means the next
// build can be pointed at a fact instead of re-deriving one, and means a
// real sign regression would be caught rather than absorbed as folklore.
func TestFromEulerAngleFollowsTheRightHandRule(t *testing.T) {
	const quarter = math.Pi / 2

	for _, tc := range []struct {
		name  string
		euler vector3.Float64
		in    vector3.Float64
		want  vector3.Float64
	}{
		// Thumb along +X, fingers curl +Y -> +Z.
		{"x turns +Y to +Z", vector3.New(quarter, 0., 0.), vector3.New(0., 1., 0.), vector3.New(0., 0., 1.)},
		{"x turns +Z to -Y", vector3.New(quarter, 0., 0.), vector3.New(0., 0., 1.), vector3.New(0., -1., 0.)},
		{"x leaves +X alone", vector3.New(quarter, 0., 0.), vector3.New(1., 0., 0.), vector3.New(1., 0., 0.)},

		// Thumb along +Y, fingers curl +Z -> +X.
		{"y turns +Z to +X", vector3.New(0., quarter, 0.), vector3.New(0., 0., 1.), vector3.New(1., 0., 0.)},
		{"y turns +X to -Z", vector3.New(0., quarter, 0.), vector3.New(1., 0., 0.), vector3.New(0., 0., -1.)},

		// Thumb along +Z, fingers curl +X -> +Y.
		{"z turns +X to +Y", vector3.New(0., 0., quarter), vector3.New(1., 0., 0.), vector3.New(0., 1., 0.)},
		{"z turns +Y to -X", vector3.New(0., 0., quarter), vector3.New(0., 1., 0.), vector3.New(-1., 0., 0.)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := quaternion.FromEulerAngle(tc.euler).Rotate(tc.in)
			assert.InDelta(t, tc.want.X(), got.X(), 1e-9)
			assert.InDelta(t, tc.want.Y(), got.Y(), 1e-9)
			assert.InDelta(t, tc.want.Z(), got.Z(), 1e-9)
		})
	}
}

// The consequence callers actually reason about: a point out along +Z,
// rotated about X, moves to -Y for a positive angle.
func TestFromEulerAngleXMovesForwardPointsDown(t *testing.T) {
	arm := vector3.New(0., 0., 2.)
	for _, angle := range []float64{0.2, 0.5, 1.0} {
		got := quaternion.FromEulerAngle(vector3.New(angle, 0., 0.)).Rotate(arm)
		assert.InDelta(t, -2*math.Sin(angle), got.Y(), 1e-9,
			"y should be -length*sin(angle), not +length*sin(angle)")
		assert.InDelta(t, 2*math.Cos(angle), got.Z(), 1e-9)
	}
}
