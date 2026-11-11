package sh_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/sh"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func coefficients() []vector3.Float64 {
	c := make([]vector3.Float64, 15)
	for i := range c {
		f := float64(i)
		c[i] = vector3.New(0.1*f-0.5, math.Sin(f), 0.3*f*f-1)
	}
	return c
}

func assertCoefficientsEqual(t *testing.T, want, got []vector3.Float64) {
	for i := range want {
		assert.InDelta(t, want[i].X(), got[i].X(), 1e-9, "coefficient %d x", i)
		assert.InDelta(t, want[i].Y(), got[i].Y(), 1e-9, "coefficient %d y", i)
		assert.InDelta(t, want[i].Z(), got[i].Z(), 1e-9, "coefficient %d z", i)
	}
}

func TestIdentityRotationLeavesCoefficientsUnchanged(t *testing.T) {
	original := coefficients()
	rotated := coefficients()
	sh.NewRotation(quaternion.Identity()).Rotate(rotated)
	assertCoefficientsEqual(t, original, rotated)
}

func TestDegreeOneRotatesLikeAVector(t *testing.T) {
	q := quaternion.FromTheta(0.7, vector3.New(1., 2., -0.5).Normalized())
	v := vector3.New(0.3, -1.2, 0.8)

	c := []vector3.Float64{vector3.Fill(-v.Y()), vector3.Fill(v.Z()), vector3.Fill(-v.X())}
	sh.NewRotation(q).Rotate(c)

	want := q.Rotate(v)
	assert.InDelta(t, want.X(), -c[2].X(), 1e-9)
	assert.InDelta(t, want.Y(), -c[0].X(), 1e-9)
	assert.InDelta(t, want.Z(), c[1].X(), 1e-9)
}

func TestRotationMatchesReEvaluatingTheRotatedFunction(t *testing.T) {
	q := quaternion.FromEulerAngle(vector3.New(0.4, -1.1, 2.3))
	inverse := quaternion.New(q.Dir().Scale(-1), q.W())
	original := coefficients()
	rotated := coefficients()
	sh.NewRotation(q).Rotate(rotated)

	for _, dir := range []vector3.Float64{
		vector3.New(0.2, 0.5, -1.).Normalized(),
		vector3.New(-3., 1., 0.1).Normalized(),
		vector3.Up[float64](),
	} {
		for degree := 1; degree <= sh.MaxDegree; degree++ {
			start, n := sh.BandStart(degree), sh.BandSize(degree)
			basisRotated := sh.Band(degree, dir)
			basisOriginal := sh.Band(degree, inverse.Rotate(dir))

			want, got := vector3.Zero[float64](), vector3.Zero[float64]()
			for m := 0; m < n; m++ {
				want = want.Add(original[start+m].Scale(basisOriginal[m]))
				got = got.Add(rotated[start+m].Scale(basisRotated[m]))
			}
			assert.InDelta(t, want.X(), got.X(), 1e-9, "degree %d", degree)
			assert.InDelta(t, want.Y(), got.Y(), 1e-9, "degree %d", degree)
			assert.InDelta(t, want.Z(), got.Z(), 1e-9, "degree %d", degree)
		}
	}
}

func TestComposingTwoRotationsEqualsOne(t *testing.T) {
	a := quaternion.FromTheta(1.3, vector3.New(0., 1., 0.))
	b := quaternion.FromTheta(-0.6, vector3.New(1., 0., 1.).Normalized())

	twice := coefficients()
	sh.NewRotation(a).Rotate(twice)
	sh.NewRotation(b).Rotate(twice)

	once := coefficients()
	sh.NewRotation(b.Multiply(a)).Rotate(once)

	assertCoefficientsEqual(t, once, twice)
}

func TestBandMatricesAreOrthonormal(t *testing.T) {
	r := sh.NewRotation(quaternion.FromEulerAngle(vector3.New(1., 2., 3.)))
	for degree := 1; degree <= sh.MaxDegree; degree++ {
		m := r.Band(degree)
		n := sh.BandSize(degree)
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				dot := 0.
				for k := 0; k < n; k++ {
					dot += m[i][k] * m[j][k]
				}
				want := 0.
				if i == j {
					want = 1
				}
				assert.InDelta(t, want, dot, 1e-9, "degree %d row %d row %d", degree, i, j)
			}
		}
	}
}

func TestPartialBandIsLeftAlone(t *testing.T) {
	original := coefficients()[:5]
	rotated := coefficients()[:5]
	sh.NewRotation(quaternion.FromTheta(1, vector3.Up[float64]())).Rotate(rotated)

	assert.NotEqual(t, original[0], rotated[0])
	assert.Equal(t, original[3], rotated[3])
	assert.Equal(t, original[4], rotated[4])
}
