package curves_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/curves"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestCatmullCurveDistance(t *testing.T) {
	tests := map[string]struct {
		curve curves.CatmullRomCurveParameters
	}{
		"empty": {curve: curves.CatmullRomCurveParameters{}},
		"straight line": {
			curve: curves.CatmullRomCurveParameters{
				P0: vector3.New(0., 0., 0.),
				P1: vector3.New(0., 0., 1.),
				P2: vector3.New(0., 0., 2.),
				P3: vector3.New(0., 0., 3.),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			curve := tc.curve.Curve()

			count := 1000
			inc := (1. / float64(count)) * curve.Length()

			last := curve.Distance(0)
			dist := 0.
			for i := 1; i < count; i++ {
				cur := curve.Distance(float64(i) * inc)
				dist += cur.Distance(last)
				last = cur
			}
			assert.InDelta(t, curve.Length(), dist, inc)
		})
	}
}

func TestCatmullSplineDistance(t *testing.T) {
	tests := map[string]struct {
		curve curves.CatmullRomSplineParameters
	}{
		"empty": {curve: curves.CatmullRomSplineParameters{}},
		"sinlge point line": {
			curve: curves.CatmullRomSplineParameters{
				Points: []vector3.Float64{
					vector3.New(0., 0., 0.),
				},
			},
		},
		"2 point line": {
			curve: curves.CatmullRomSplineParameters{
				Points: []vector3.Float64{
					vector3.New(0., 0., 0.),
					vector3.New(0., 0., 1.),
				},
			},
		},
		"3 point line": {
			curve: curves.CatmullRomSplineParameters{
				Points: []vector3.Float64{
					vector3.New(0., 0., 0.),
					vector3.New(0., 0., 1.),
					vector3.New(0., 0., 2.),
				},
			},
		},
		"straight line": {
			curve: curves.CatmullRomSplineParameters{
				Points: []vector3.Float64{
					vector3.New(0., 0., 0.),
					vector3.New(0., 0., 1.),
					vector3.New(0., 0., 2.),
					vector3.New(0., 0., 3.),
				},
			},
		},
		"little dip": {
			curve: curves.CatmullRomSplineParameters{
				Points: []vector3.Float64{
					vector3.New(0., 0., 0.),
					vector3.New(0., 0., 3.),
					vector3.New(0., -1.1791153159954555, 6.),
					vector3.New(0., 0., 9.),
					vector3.New(0., 0., 12.),
				},
				Alpha: 0.5,
			},
		},
		"little dip shift": {
			curve: curves.CatmullRomSplineParameters{
				Points: []vector3.Float64{
					vector3.New(0.3, 0.1, 0.),
					vector3.New(0.3, 0.1, 3.),
					vector3.New(0.3, -1.0791153159954554, 6.),
					vector3.New(0.3, 0.1, 9.),
					vector3.New(0.3, 0.1, 12.),
				},
				Alpha: 0.5,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			curve := tc.curve.Spline()

			count := 10000
			inc := curve.Length() / float64(count)

			last := curve.At(0)
			dist := 0.
			for i := 1; i <= count; i++ {
				cur := curve.At(float64(i) * inc)
				if i == count {
					cur = curve.At(curve.Length())
				}
				dist += cur.Distance(last)
				last = cur
			}
			assert.InDelta(t, curve.Length(), dist, inc/2)
		})
	}
}
