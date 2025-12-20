package geometry_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestLine3D_ClosestPoint_LineOfZeroLength(t *testing.T) {
	p1 := vector3.New(-1., 0., 0.)
	p2 := vector3.New(-1., 0., 0.)
	line := geometry.NewLine3D(p1, p2)

	p := line.ClosestPointOnLine(vector3.New(12, 12, 12.))
	assert.Equal(t, vector3.New(-1, 0, 0.), p)
}

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

func TestLineStripsFromPoints3D(t *testing.T) {
	tests := map[string]struct {
		points []vector3.Float64
		lines  []geometry.Line3D
	}{
		"nil => nil": {},
		"1 point => nil": {
			points: []vector3.Float64{
				vector3.Zero[float64](),
			},
		},
		"2 points => 1 line": {
			points: []vector3.Float64{
				vector3.Zero[float64](),
				vector3.Up[float64](),
			},
			lines: []geometry.Line3D{
				geometry.NewLine3D(
					vector3.Zero[float64](),
					vector3.Up[float64](),
				),
			},
		},
		"3 points => 2 line": {
			points: []vector3.Float64{
				vector3.Zero[float64](),
				vector3.Up[float64](),
				vector3.New(0., 2., 0.),
			},
			lines: []geometry.Line3D{
				geometry.NewLine3D(
					vector3.Zero[float64](),
					vector3.Up[float64](),
				),
				geometry.NewLine3D(
					vector3.Up[float64](),
					vector3.New(0., 2., 0.),
				),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := geometry.LineStripsFromPoints3D(tc.points)
			assert.Equal(t, tc.lines, result)
		})
	}
}

func TestLinesFromPoints3D(t *testing.T) {
	tests := map[string]struct {
		points []vector3.Float64
		lines  []geometry.Line3D
	}{
		"nil => nil": {},
		"1 point => nil": {
			points: []vector3.Float64{
				vector3.Zero[float64](),
			},
		},
		"2 points => 1 line": {
			points: []vector3.Float64{
				vector3.Zero[float64](),
				vector3.Up[float64](),
			},
			lines: []geometry.Line3D{
				geometry.NewLine3D(
					vector3.Zero[float64](),
					vector3.Up[float64](),
				),
			},
		},
		"3 points => 1 line": {
			points: []vector3.Float64{
				vector3.Zero[float64](),
				vector3.Up[float64](),
				vector3.New(0., 2., 0.),
			},
			lines: []geometry.Line3D{
				geometry.NewLine3D(
					vector3.Zero[float64](),
					vector3.Up[float64](),
				),
			},
		},
		"4 points => 2 line": {
			points: []vector3.Float64{
				vector3.Zero[float64](),
				vector3.Up[float64](),
				vector3.New(0., 2., 0.),
				vector3.New(0., 3., 0.),
			},
			lines: []geometry.Line3D{
				geometry.NewLine3D(
					vector3.Zero[float64](),
					vector3.Up[float64](),
				),
				geometry.NewLine3D(
					vector3.New(0., 2., 0.),
					vector3.New(0., 3., 0.),
				),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := geometry.LinesFromPoints3D(tc.points)
			assert.Equal(t, tc.lines, result)
		})
	}
}
