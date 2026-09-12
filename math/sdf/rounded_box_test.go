package sdf_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func surfaceAlong(field func(vector3.Float64) float64, dir vector3.Float64) float64 {
	lo, hi := 0., 100.
	for i := 0; i < 80; i++ {
		mid := (lo + hi) / 2
		if field(dir.Scale(mid)) < 0 {
			lo = mid
		} else {
			hi = mid
		}
	}
	return lo
}

func TestRoundedBoxRoundnessInflates(t *testing.T) {
	size := vector3.New(0.02, 0.30, 0.20)
	roundness := 0.06
	field := sdf.RoundedBox(vector3.Zero[float64](), size, roundness)

	for _, tc := range []struct {
		name string
		dir  vector3.Float64
		half float64
	}{
		{"x", vector3.New(1., 0., 0.), size.X() / 2},
		{"y", vector3.New(0., 1., 0.), size.Y() / 2},
		{"z", vector3.New(0., 0., 1.), size.Z() / 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.InDelta(t, tc.half+roundness, surfaceAlong(field, tc.dir), 0.0001)
		})
	}
}

func TestRoundedBoxStaysSolidWhenRoundnessExceedsHalfSize(t *testing.T) {
	field := sdf.RoundedBox(vector3.Zero[float64](), vector3.New(0.02, 0.30, 0.20), 0.5)

	assert.Negative(t, field(vector3.Zero[float64]()), "center must be inside the solid")
	assert.InDelta(t, 0.51, surfaceAlong(field, vector3.New(1., 0., 0.)), 0.0001)
	assert.Positive(t, field(vector3.New(2., 0., 0.)), "far outside must be outside")
}

func TestRoundedBoxSizingForATargetExtent(t *testing.T) {
	target := vector3.New(0.01, 0.15, 0.10)
	roundness := 0.005
	field := sdf.RoundedBox(
		vector3.Zero[float64](),
		target.Sub(vector3.Fill(roundness)).Scale(2),
		roundness,
	)

	assert.InDelta(t, target.X(), surfaceAlong(field, vector3.New(1., 0., 0.)), 0.0001)
	assert.InDelta(t, target.Y(), surfaceAlong(field, vector3.New(0., 1., 0.)), 0.0001)
	assert.InDelta(t, target.Z(), surfaceAlong(field, vector3.New(0., 0., 1.)), 0.0001)
}
