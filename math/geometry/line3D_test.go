package geometry_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestLine3D_ClosestPoint_HorizontalLine(t *testing.T) {
	p1 := vector3.New(-1., 0., 0.)
	p2 := vector3.New(1., 0., 0.)
	line := geometry.NewLine3D(p1, p2)

	tests := map[string]struct {
		input vector3.Float64
		want  vector3.Float64
	}{
		"mid point":           {input: vector3.New[float64](0, 0, 0), want: vector3.New[float64](0, 0, 0)},
		"right point":         {input: vector3.New[float64](1, 0, 0), want: vector3.New[float64](1, 0, 0)},
		"left point":          {input: vector3.New[float64](-1, 0, 0), want: vector3.New[float64](-1, 0, 0)},
		"clamped right point": {input: vector3.New[float64](5, 0, 0), want: vector3.New[float64](1, 0, 0)},
		"clamped left point":  {input: vector3.New[float64](-5, 0, 0), want: vector3.New[float64](-1, 0, 0)},

		"+1Y mid point":           {input: vector3.New[float64](0, 1, 0), want: vector3.New[float64](0, 0, 0)},
		"+1Y right point":         {input: vector3.New[float64](1, 1, 0), want: vector3.New[float64](1, 0, 0)},
		"+1Y left point":          {input: vector3.New[float64](-1, 1, 0), want: vector3.New[float64](-1, 0, 0)},
		"+1Y clamped right point": {input: vector3.New[float64](5, 1, 0), want: vector3.New[float64](1, 0, 0)},
		"+1Y clamped left point":  {input: vector3.New[float64](-5, 1, 0), want: vector3.New[float64](-1, 0, 0)},

		"-1Y mid point":           {input: vector3.New[float64](0, -1, 0), want: vector3.New[float64](0, 0, 0)},
		"-1Y right point":         {input: vector3.New[float64](1, -1, 0), want: vector3.New[float64](1, 0, 0)},
		"-1Y left point":          {input: vector3.New[float64](-1, -1, 0), want: vector3.New[float64](-1, 0, 0)},
		"-1Y clamped right point": {input: vector3.New[float64](5, -1, 0), want: vector3.New[float64](1, 0, 0)},
		"-1Y clamped left point":  {input: vector3.New[float64](-5, -1, 0), want: vector3.New[float64](-1, 0, 0)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := line.ClosestPointOnLine(tc.input)
			assert.InDelta(t, tc.want.X(), got.X(), 0.0001)
			assert.InDelta(t, tc.want.Y(), got.Y(), 0.0001)
			assert.InDelta(t, tc.want.Z(), got.Z(), 0.0001)
		})
	}
}

func TestLine3D_ClosestPoint(t *testing.T) {

	tests := map[string]struct {
		point vector3.Float64
		line  geometry.Line3D
		want  vector3.Float64
	}{
		"centered horizontal line": {
			point: vector3.New[float64](0, 0, 0),
			want:  vector3.New[float64](0, 0, 0),
			line:  geometry.NewLine3D(vector3.New(0., -1., 0.), vector3.New(0., 1., 0.)),
		},

		"horizontal line offset +1 +1 / 0, 0, 0": {
			point: vector3.New[float64](0, 0, 0),
			want:  vector3.New[float64](1, 0, 1),
			line:  geometry.NewLine3D(vector3.New(1., -1., 1.), vector3.New(1., 1., 1.)),
		},

		"horizontal line offset -1 +1 / 0, 0, 0": {
			point: vector3.New[float64](0, 0, 0),
			want:  vector3.New[float64](-1, 0, 1),
			line:  geometry.NewLine3D(vector3.New(-1., -1., 1.), vector3.New(-1., 1., 1.)),
		},

		"horizontal line offset +1 -1 / 0, 0, 0": {
			point: vector3.New[float64](0, 0, 0),
			want:  vector3.New[float64](1, 0, -1),
			line:  geometry.NewLine3D(vector3.New(1., -1., -1.), vector3.New(1., 1., -1.)),
		},

		"horizontal line offset -1 -1 / 0, 0, 0": {
			point: vector3.New[float64](0, 0, 0),
			want:  vector3.New[float64](-1, 0, -1),
			line:  geometry.NewLine3D(vector3.New(-1., -1., -1.), vector3.New(-1., 1., -1.)),
		},

		"horizontal line offset +1 +1": {
			point: vector3.New[float64](2, 0.5, 2),
			want:  vector3.New[float64](1, 0.5, 1),
			line:  geometry.NewLine3D(vector3.New(1., -1., 1.), vector3.New(1., 1., 1.)),
		},

		"horizontal line offset -1 +1": {
			point: vector3.New[float64](-2, 0.5, 2),
			want:  vector3.New[float64](-1, 0.5, 1),
			line:  geometry.NewLine3D(vector3.New(-1., -1., 1.), vector3.New(-1., 1., 1.)),
		},

		"horizontal line offset +1 -1": {
			point: vector3.New[float64](2, 0.5, -2),
			want:  vector3.New[float64](1, 0.5, -1),
			line:  geometry.NewLine3D(vector3.New(1., -1., -1.), vector3.New(1., 1., -1.)),
		},

		"horizontal line offset -1 -1": {
			point: vector3.New[float64](-2, 0.5, -2),
			want:  vector3.New[float64](-1, 0.5, -1),
			line:  geometry.NewLine3D(vector3.New(-1., -1., -1.), vector3.New(-1., 1., -1.)),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.line.ClosestPointOnLine(tc.point)
			assert.InDelta(t, tc.want.X(), got.X(), 0.0001)
			assert.InDelta(t, tc.want.Y(), got.Y(), 0.0001)
			assert.InDelta(t, tc.want.Z(), got.Z(), 0.0001)
		})
	}
}
