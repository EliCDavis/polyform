package trs_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rot(y float64) quaternion.Quaternion {
	return quaternion.FromEulerAngle(vector3.New(0., y, 0.))
}

func TestFromMatrixRefusesShear(t *testing.T) {
	squashed := trs.New(vector3.New(1., 2., 3.), rot(math.Pi/4), vector3.New(0.25, 1., 1.))

	_, err := trs.FromMatrix(squashed.Inverse())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "shear")
}

func TestFromMatrixAcceptsWhatItCanRepresent(t *testing.T) {
	id := trs.Identity().Rotation()

	for _, tc := range []struct {
		name string
		x    trs.TRS
	}{
		{"translate only", trs.New(vector3.New(1., 2., 3.), id, vector3.One[float64]())},
		{"uniform scale with rotation", trs.New(vector3.New(1., 2., 3.), rot(math.Pi/4), vector3.Fill(0.5))},
		{"non-uniform scale, no rotation", trs.New(vector3.New(1., 2., 3.), id, vector3.New(0.25, 1., 1.))},
		{"non-uniform scale, quarter turn", trs.New(vector3.Zero[float64](), rot(math.Pi/2), vector3.New(0.25, 1., 1.))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inverted, err := trs.FromMatrix(tc.x.Inverse())
			require.NoError(t, err)

			for _, p := range []vector3.Float64{
				vector3.New(1., 0., 0.), vector3.New(0., 1., 0.), vector3.New(1., 2., -3.),
			} {
				back := inverted.Transform(tc.x.Transform(p))
				assert.InDelta(t, 0, back.Sub(p).Length(), 1e-9)
			}
		})
	}
}

func TestFromMatrixRefusesACollapsedAxis(t *testing.T) {
	flat := trs.New(vector3.Zero[float64](), trs.Identity().Rotation(), vector3.New(1., 0., 1.))

	_, err := trs.FromMatrix(flat.Matrix())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collapses an axis")
}

func TestMultiplyOrderDecidesRepresentability(t *testing.T) {
	id := trs.Identity().Rotation()
	zero := vector3.Zero[float64]()
	squash := trs.New(zero, id, vector3.New(0.25, 1., 1.))
	turn := trs.New(zero, rot(math.Pi/4), vector3.One[float64]())

	_, err := trs.FromMatrix(turn.Multiply(squash))
	assert.NoError(t, err, "scale innermost stays a TRS")

	_, err = trs.FromMatrix(squash.Multiply(turn))
	assert.Error(t, err, "scale outside a rotation shears")
}

func TestShearScoresZeroOnRepresentableMatrices(t *testing.T) {
	zero := vector3.Zero[float64]()
	squashed := trs.New(zero, rot(math.Pi/4), vector3.New(0.25, 1., 1.))

	assert.LessOrEqual(t, trs.Shear(squashed.Matrix()), trs.ShearTolerance)
	assert.Greater(t, trs.Shear(squashed.Inverse()), 0.5)
}
