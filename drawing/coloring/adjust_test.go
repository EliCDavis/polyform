package coloring_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRGBHSVRoundTrip(t *testing.T) {
	cases := map[string][3]float64{
		"red":     {1, 0, 0},
		"green":   {0, 1, 0},
		"blue":    {0, 0, 1},
		"white":   {1, 1, 1},
		"black":   {0, 0, 0},
		"grey":    {0.5, 0.5, 0.5},
		"teal":    {0.1, 0.6, 0.55},
		"mustard": {0.8, 0.7, 0.2},
	}

	for name, rgb := range cases {
		t.Run(name, func(t *testing.T) {
			h, s, v := coloring.RGBToHSV(rgb[0], rgb[1], rgb[2])
			r, g, b := coloring.HSVToRGB(h, s, v)

			assert.InDelta(t, rgb[0], r, 1e-9, "red")
			assert.InDelta(t, rgb[1], g, 1e-9, "green")
			assert.InDelta(t, rgb[2], b, 1e-9, "blue")
		})
	}
}

func TestRGBToHSVKnownValues(t *testing.T) {
	h, s, v := coloring.RGBToHSV(1, 0, 0)
	assert.InDelta(t, 0, h, 1e-9, "pure red sits at hue 0")
	assert.InDelta(t, 1, s, 1e-9)
	assert.InDelta(t, 1, v, 1e-9)

	h, _, _ = coloring.RGBToHSV(0, 1, 0)
	assert.InDelta(t, 120, h, 1e-9, "pure green sits at hue 120")

	h, _, _ = coloring.RGBToHSV(0, 0, 1)
	assert.InDelta(t, 240, h, 1e-9, "pure blue sits at hue 240")

	_, s, _ = coloring.RGBToHSV(0.5, 0.5, 0.5)
	assert.InDelta(t, 0, s, 1e-9, "grey has no saturation")
}

func adjustHSV(t *testing.T, in coloring.Color, hue, sat, val float64) coloring.Color {
	t.Helper()
	return nodes.GetNodeOutputPort[coloring.Color](&nodes.Struct[coloring.AdjustHSVNode]{
		Data: coloring.AdjustHSVNode{
			In:         nodes.ConstOutput[coloring.Color]{Val: in},
			Hue:        nodes.ConstOutput[float64]{Val: hue},
			Saturation: nodes.ConstOutput[float64]{Val: sat},
			Value:      nodes.ConstOutput[float64]{Val: val},
		},
	}, "Out").Value()
}

func TestAdjustHSVDerivesAShade(t *testing.T) {
	upholstery := coloring.Color{R: 0.4, G: 0.5, B: 0.7, A: 1}

	darker := adjustHSV(t, upholstery, 0, 1, 0.9)

	assert.Less(t, darker.R, upholstery.R, "a 0.9 value multiplier should darken")
	assert.Less(t, darker.G, upholstery.G)
	assert.Less(t, darker.B, upholstery.B)
	assert.InDelta(t, 1, darker.A, 1e-9, "alpha should survive untouched")

	// Darkening must not shift the hue - it should read as the same color.
	beforeHue, _, _ := coloring.RGBToHSV(upholstery.R, upholstery.G, upholstery.B)
	afterHue, _, _ := coloring.RGBToHSV(darker.R, darker.G, darker.B)
	assert.InDelta(t, beforeHue, afterHue, 1e-6, "darkening should preserve hue")
}

func TestAdjustHSVDesaturates(t *testing.T) {
	vivid := coloring.Color{R: 0.9, G: 0.1, B: 0.1, A: 1}
	washed := adjustHSV(t, vivid, 0, 0.2, 1)

	_, beforeSat, _ := coloring.RGBToHSV(vivid.R, vivid.G, vivid.B)
	_, afterSat, _ := coloring.RGBToHSV(washed.R, washed.G, washed.B)
	assert.Less(t, afterSat, beforeSat, "a saturation multiplier below 1 should wash the color out")
}

func TestAdjustHSVShiftsHueAndWraps(t *testing.T) {
	red := coloring.Color{R: 1, G: 0, B: 0, A: 1}

	shifted := adjustHSV(t, red, 120, 1, 1)
	assert.InDelta(t, 0, shifted.R, 1e-9)
	assert.InDelta(t, 1, shifted.G, 1e-9, "rotating red 120 degrees should land on green")

	// Past a full turn it must wrap rather than clamp.
	wrapped := adjustHSV(t, red, 360, 1, 1)
	assert.InDelta(t, 1, wrapped.R, 1e-9, "a full turn should come back to red")
	assert.InDelta(t, 0, wrapped.G, 1e-9)
}

func TestFromVectorNodeCompletesTheRoundTrip(t *testing.T) {
	original := coloring.Color{R: 0.25, G: 0.5, B: 0.75, A: 1}

	asVector := nodes.GetNodeOutputPort[vector3.Float64](&nodes.Struct[coloring.ToVectorNode]{
		Data: coloring.ToVectorNode{In: nodes.ConstOutput[coloring.Color]{Val: original}},
	}, "Vector 3").Value()

	back := nodes.GetNodeOutputPort[coloring.Color](&nodes.Struct[coloring.FromVectorNode]{
		Data: coloring.FromVectorNode{Vector3: nodes.ConstOutput[vector3.Float64]{Val: asVector}},
	}, "Out").Value()

	require.InDelta(t, original.R, back.R, 1e-9)
	require.InDelta(t, original.G, back.G, 1e-9)
	require.InDelta(t, original.B, back.B, 1e-9)
}

func TestBrightnessNodeClampsAtWhite(t *testing.T) {
	bright := nodes.GetNodeOutputPort[coloring.Color](&nodes.Struct[coloring.BrightnessNode]{
		Data: coloring.BrightnessNode{
			In:     nodes.ConstOutput[coloring.Color]{Val: coloring.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}},
			Amount: nodes.ConstOutput[float64]{Val: 10},
		},
	}, "Out").Value()

	assert.InDelta(t, 1, bright.R, 1e-9, "over-bright channels should clamp, not overflow")
	assert.InDelta(t, 1, bright.G, 1e-9)
	assert.InDelta(t, 1, bright.B, 1e-9)
}
