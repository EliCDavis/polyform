package trs_test

import (
	"math"
	"testing"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestTransformArray(t *testing.T) {

	// ARRANGE ================================================================
	transform := trs.New(
		vector3.New(1., 2., 3.),
		quaternion.FromTheta(math.Pi, vector3.Forward[float64]()),
		vector3.New(4., 5., 6.),
	)
	points := []vector3.Float64{
		vector3.Zero[float64](),
		vector3.One[float64](),
	}

	// ACT ====================================================================
	out := transform.TransformArray(points)

	// ASSERT =================================================================
	assert.Len(t, out, len(points))
	assert.Equal(t, vector3.New(1., 2., 3.), out[0])
	assert.InDelta(t, -3., out[1].X(), 0.0000001)
	assert.InDelta(t, -3., out[1].Y(), 0.0000001)
	assert.InDelta(t, 9., out[1].Z(), 0.0000001)

}

func TestTransformInPlace(t *testing.T) {

	// ARRANGE ================================================================
	transform := trs.New(
		vector3.New(1., 2., 3.),
		quaternion.FromTheta(math.Pi, vector3.Forward[float64]()),
		vector3.New(4., 5., 6.),
	)
	points := []vector3.Float64{
		vector3.Zero[float64](),
		vector3.One[float64](),
	}

	// ACT ====================================================================
	transform.TransformInPlace(points)

	// ASSERT =================================================================
	assert.Equal(t, vector3.New(1., 2., 3.), points[0])
	assert.InDelta(t, -3., points[1].X(), 0.0000001)
	assert.InDelta(t, -3., points[1].Y(), 0.0000001)
	assert.InDelta(t, 9., points[1].Z(), 0.0000001)

}

func TestTranslate(t *testing.T) {

	// ARRANGE ================================================================
	transform := trs.New(
		vector3.New(1., 2., 3.),
		quaternion.FromTheta(math.Pi, vector3.Forward[float64]()),
		vector3.New(4., 5., 6.),
	)
	point := vector3.One[float64]()

	// ACT ====================================================================
	translated := transform.Translate(point)
	newPosition := translated.Position()

	// ASSERT =================================================================
	assert.InDelta(t, 2., newPosition.X(), 0.0000001)
	assert.InDelta(t, 3., newPosition.Y(), 0.0000001)
	assert.InDelta(t, 4., newPosition.Z(), 0.0000001)
}

func TestInDelta(t *testing.T) {
	tests := map[string]struct {
		a, b  trs.TRS
		delta float64
		err   string
	}{
		"0 delta, both identity": {
			a:     trs.Identity(),
			b:     trs.Identity(),
			delta: 0,
		},
		"slightly different position, within delta": {
			a:     trs.Position(vector3.New(0., 0., 0.)),
			b:     trs.Position(vector3.New(0., .1, 0.)),
			delta: .2,
		},
		"slightly different position, at delta": {
			a:     trs.Position(vector3.New(0., 0., 0.)),
			b:     trs.Position(vector3.New(0., .2, 0.)),
			delta: .2,
		},
		"slightly different position, outside delta": {
			a:     trs.Position(vector3.New(0., 0., 0.)),
			b:     trs.Position(vector3.New(0., .200000000000001, 0.)),
			delta: .2,
			err:   "expected position.y 0 to be within delta (0.2) of 0.200000000000001",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := tc.a.WithinDelta(tc.b, tc.delta)
			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}
