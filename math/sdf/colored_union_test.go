package sdf_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmoothUnionColoredTwoSpheres(t *testing.T) {
	red := coloring.Color{R: 1, G: 0, B: 0, A: 1}
	blue := coloring.Color{R: 0, G: 0, B: 1, A: 1}

	// Two unit spheres 1.5 apart, so they actually overlap.
	a := sdf.ColoredField{
		Distance: sdf.Sphere(vector3.New(0., 0., 0.), 1),
		Color:    sdf.ConstantColor(red),
	}
	b := sdf.ColoredField{
		Distance: sdf.Sphere(vector3.New(1.5, 0., 0.), 1),
		Color:    sdf.ConstantColor(blue),
	}

	union := sdf.SmoothUnionColored(0.5, a, b)

	tests := map[string]struct {
		pos      vector3.Float64
		wantR    float64
		wantB    float64
		distSign float64 // +1 outside, -1 inside
	}{
		"deep inside A, far from the boundary": {
			pos: vector3.New(0., 0., 0.), wantR: 1, wantB: 0, distSign: -1,
		},
		"deep inside B, far from the boundary": {
			pos: vector3.New(1.5, 0., 0.), wantR: 0, wantB: 1, distSign: -1,
		},
		"exactly on the seam, equidistant from both centers": {
			pos: vector3.New(0.75, 0., 0.), wantR: 0.5, wantB: 0.5, distSign: -1,
		},
		"far outside both, clearly on A's side": {
			// Nearer A's surface than B's, so A's color should win.
			pos: vector3.New(-10., 0., 0.), wantR: 1, wantB: 0, distSign: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			d := union.Distance(tc.pos)
			if tc.distSign < 0 {
				assert.Negative(t, d, "expected inside the union")
			} else {
				assert.Positive(t, d, "expected outside the union")
			}

			c := union.Color(tc.pos)
			assert.InDelta(t, tc.wantR, c.R, 1e-9, "red channel")
			assert.InDelta(t, tc.wantB, c.B, 1e-9, "blue channel")
		})
	}
}

func TestSmoothUnionColoredNearestTwoIgnoresDistantThirdField(t *testing.T) {
	red := coloring.Color{R: 1, G: 0, B: 0, A: 1}
	green := coloring.Color{R: 0, G: 1, B: 0, A: 1}
	blue := coloring.Color{R: 0, G: 0, B: 1, A: 1}

	a := sdf.ColoredField{Distance: sdf.Sphere(vector3.New(0., 0., 0.), 1), Color: sdf.ConstantColor(red)}
	b := sdf.ColoredField{Distance: sdf.Sphere(vector3.New(1.5, 0., 0.), 1), Color: sdf.ConstantColor(green)}
	// Far away - should have zero influence on A/B's blend.
	c := sdf.ColoredField{Distance: sdf.Sphere(vector3.New(0., 0., 100.), 1), Color: sdf.ConstantColor(blue)}

	union := sdf.SmoothUnionColored(0.5, a, b, c)

	got := union.Color(vector3.New(0.75, 0., 0.)) // the A/B seam
	assert.InDelta(t, 0.5, got.R, 1e-9)
	assert.InDelta(t, 0.5, got.G, 1e-9)
	assert.InDelta(t, 0, got.B, 1e-9, "the distant blue sphere must not leak into the A/B seam")
}

func TestUnionColoredTwoSpheres(t *testing.T) {
	red := coloring.Color{R: 1, G: 0, B: 0, A: 1}
	blue := coloring.Color{R: 0, G: 0, B: 1, A: 1}

	a := sdf.ColoredField{Distance: sdf.Sphere(vector3.New(0., 0., 0.), 1), Color: sdf.ConstantColor(red)}
	b := sdf.ColoredField{Distance: sdf.Sphere(vector3.New(1.5, 0., 0.), 1), Color: sdf.ConstantColor(blue)}

	union := sdf.UnionColored(a, b)

	tests := map[string]struct {
		pos  vector3.Float64
		want coloring.Color
	}{
		"deep inside A": {pos: vector3.New(0., 0., 0.), want: red},
		"deep inside B": {pos: vector3.New(1.5, 0., 0.), want: blue},
		"slightly closer to A of the two, inside the overlap": {pos: vector3.New(0.7, 0., 0.), want: red},
		"slightly closer to B of the two, inside the overlap": {pos: vector3.New(0.8, 0., 0.), want: blue},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := union.Color(tc.pos)
			assert.Equal(t, tc.want, got, "hard union should give the nearer field's exact color, never a blend")
		})
	}

	// Distance should match the plain Union of the same fields.
	plainUnion := sdf.Union(a.Distance, b.Distance)
	for _, p := range []vector3.Float64{
		vector3.New(0., 0., 0.), vector3.New(0.75, 0., 0.), vector3.New(10., 10., 10.),
	} {
		assert.InDelta(t, plainUnion(p), union.Distance(p), 1e-9)
	}
}

func TestUnionColoredNode(t *testing.T) {
	red := coloring.Color{R: 1, G: 0, B: 0, A: 1}
	blue := coloring.Color{R: 0, G: 0, B: 1, A: 1}

	a := sdf.ColoredField{Distance: sdf.Sphere(vector3.New(0., 0., 0.), 1), Color: sdf.ConstantColor(red)}
	b := sdf.ColoredField{Distance: sdf.Sphere(vector3.New(1.5, 0., 0.), 1), Color: sdf.ConstantColor(blue)}

	node := &nodes.Struct[sdf.UnionColoredNode]{
		Data: sdf.UnionColoredNode{
			Fields: []nodes.Output[sdf.ColoredField]{
				nodes.ConstOutput[sdf.ColoredField]{Val: a},
				nodes.ConstOutput[sdf.ColoredField]{Val: b},
			},
		},
	}

	got := nodes.GetNodeOutputPort[sdf.ColoredField](node, "Union").Value()
	assert.Equal(t, red, got.Color(vector3.New(0., 0., 0.)))
	assert.Equal(t, blue, got.Color(vector3.New(1.5, 0., 0.)))
}

func TestUnionColoredNodeRequiresAtLeastOneField(t *testing.T) {
	node := &nodes.Struct[sdf.UnionColoredNode]{}

	var got sdf.ColoredField
	require.NotPanics(t, func() {
		got = nodes.GetNodeOutputPort[sdf.ColoredField](node, "Union").Value()
	})
	require.Nil(t, got.Distance, "no fields provided should capture an error, not set a usable field")
}

func TestWithColorNode(t *testing.T) {
	sphere := sdf.Sphere(vector3.New(0., 0., 0.), 1)
	red := coloring.Color{R: 1, G: 0, B: 0, A: 1}

	node := &nodes.Struct[sdf.WithColorNode]{
		Data: sdf.WithColorNode{
			Field: nodes.ConstOutput[sample.Vec3ToFloat]{Val: sphere},
			Color: nodes.ConstOutput[coloring.Color]{Val: red},
		},
	}

	got := nodes.GetNodeOutputPort[sdf.ColoredField](node, "Out").Value()
	require.NotNil(t, got.Distance)
	require.NotNil(t, got.Color)
	assert.InDelta(t, -1.0, got.Distance(vector3.New(0., 0., 0.)), 1e-9)
	assert.Equal(t, red, got.Color(vector3.New(0., 0., 0.)))
	assert.Equal(t, red, got.Color(vector3.New(5., 5., 5.)), "color is constant everywhere, not just near the field")
}

func TestSmoothUnionColoredNode(t *testing.T) {
	red := coloring.Color{R: 1, G: 0, B: 0, A: 1}
	blue := coloring.Color{R: 0, G: 0, B: 1, A: 1}

	a := sdf.ColoredField{Distance: sdf.Sphere(vector3.New(0., 0., 0.), 1), Color: sdf.ConstantColor(red)}
	b := sdf.ColoredField{Distance: sdf.Sphere(vector3.New(1.5, 0., 0.), 1), Color: sdf.ConstantColor(blue)}

	node := &nodes.Struct[sdf.SmoothUnionColoredNode]{
		Data: sdf.SmoothUnionColoredNode{
			Fields: []nodes.Output[sdf.ColoredField]{
				nodes.ConstOutput[sdf.ColoredField]{Val: a},
				nodes.ConstOutput[sdf.ColoredField]{Val: b},
			},
			Radius: nodes.ConstOutput[float64]{Val: 0.5},
		},
	}

	got := nodes.GetNodeOutputPort[sdf.ColoredField](node, "Union").Value()
	seam := got.Color(vector3.New(0.75, 0., 0.))
	assert.InDelta(t, 0.5, seam.R, 1e-9)
	assert.InDelta(t, 0.5, seam.B, 1e-9)
}

func TestSmoothUnionColoredNodeRequiresAtLeastOneField(t *testing.T) {
	node := &nodes.Struct[sdf.SmoothUnionColoredNode]{}

	var got sdf.ColoredField
	require.NotPanics(t, func() {
		got = nodes.GetNodeOutputPort[sdf.ColoredField](node, "Union").Value()
	})
	require.Nil(t, got.Distance, "no fields provided should capture an error, not set a usable field")
}

func TestColoredFieldDistanceNode(t *testing.T) {
	field := sdf.ColoredField{
		Distance: sdf.Sphere(vector3.New(0., 0., 0.), 1),
		Color:    sdf.ConstantColor(coloring.Color{R: 1, A: 1}),
	}

	node := &nodes.Struct[sdf.ColoredFieldDistanceNode]{
		Data: sdf.ColoredFieldDistanceNode{
			Field: nodes.ConstOutput[sdf.ColoredField]{Val: field},
		},
	}

	got := nodes.GetNodeOutputPort[sample.Vec3ToFloat](node, "Out").Value()
	assert.InDelta(t, -1.0, got(vector3.New(0., 0., 0.)), 1e-9)
}
