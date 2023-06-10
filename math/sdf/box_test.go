package sdf_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestBox(t *testing.T) {
	box := sdf.Box(vector3.Zero[float64](), vector3.New(1., 2., 3.))

	tests := map[string]struct {
		pos  vector3.Float64
		want float64
	}{
		"center": {pos: vector3.Zero[float64](), want: -0.5},

		"on left bounds":      {pos: vector3.Left[float64]().Scale(0.5), want: 0.},
		"outside left bounds": {pos: vector3.Left[float64](), want: 0.5},

		"on right bounds":      {pos: vector3.Right[float64]().Scale(0.5), want: 0.},
		"outside right bounds": {pos: vector3.Right[float64](), want: 0.5},

		"on up bounds":      {pos: vector3.Up[float64](), want: 0.},
		"outside up bounds": {pos: vector3.Up[float64]().Scale(2), want: 1.},

		"on down bounds":      {pos: vector3.Down[float64](), want: 0.},
		"outside down bounds": {pos: vector3.Down[float64]().Scale(2), want: 1.},

		"on forward bounds":      {pos: vector3.Forward[float64]().Scale(1.5), want: 0.},
		"outside forward bounds": {pos: vector3.Forward[float64]().Scale(2), want: 0.5},

		"on backwards bounds":      {pos: vector3.Backwards[float64]().Scale(1.5), want: 0.},
		"outside backwards bounds": {pos: vector3.Backwards[float64]().Scale(2), want: 0.5},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, box(tc.pos))
		})
	}
}
