package colors_test

import (
	"image/color"
	"testing"

	"github.com/EliCDavis/polyform/math/colors"
	"github.com/stretchr/testify/assert"
)

func TestRedEqual(t *testing.T) {
	tests := map[string]struct {
		input   color.Color
		testVal byte
		want    bool
	}{
		"RGBA(1,2,3,4) == 1: true":     {input: color.RGBA{R: 1, G: 2, B: 3, A: 4}, testVal: 1, want: true},
		"RGBA(1,2,3,4) == 2: false":    {input: color.RGBA{R: 1, G: 2, B: 3, A: 4}, testVal: 2, want: false},
		"NRGBA(1,2,3,255) == 1: true":  {input: color.NRGBA{R: 1, G: 2, B: 3, A: 255}, testVal: 1, want: true},
		"NRGBA(1,2,3,255) == 2: false": {input: color.NRGBA{R: 1, G: 2, B: 3, A: 255}, testVal: 2, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, colors.RedEqual(tc.input, tc.testVal))
		})
	}
}

func TestGreenEqual(t *testing.T) {
	tests := map[string]struct {
		input   color.Color
		testVal byte
		want    bool
	}{
		"RGBA(1,2,3,4) == 2: true":     {input: color.RGBA{R: 1, G: 2, B: 3, A: 4}, testVal: 2, want: true},
		"RGBA(1,2,3,4) == 1: false":    {input: color.RGBA{R: 1, G: 2, B: 3, A: 4}, testVal: 1, want: false},
		"NRGBA(1,2,3,255) == 2: true":  {input: color.NRGBA{R: 1, G: 2, B: 3, A: 255}, testVal: 2, want: true},
		"NRGBA(1,2,3,255) == 1: false": {input: color.NRGBA{R: 1, G: 2, B: 3, A: 255}, testVal: 1, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, colors.GreenEqual(tc.input, tc.testVal))
		})
	}
}

func TestBlueEqual(t *testing.T) {
	tests := map[string]struct {
		input   color.Color
		testVal byte
		want    bool
	}{
		"RGBA(1,2,3,4) == 3: true":     {input: color.RGBA{R: 1, G: 2, B: 3, A: 4}, testVal: 3, want: true},
		"RGBA(1,2,3,4) == 2: false":    {input: color.RGBA{R: 1, G: 2, B: 3, A: 4}, testVal: 2, want: false},
		"NRGBA(1,2,3,255) == 3: true":  {input: color.NRGBA{R: 1, G: 2, B: 3, A: 255}, testVal: 3, want: true},
		"NRGBA(1,2,3,255) == 2: false": {input: color.NRGBA{R: 1, G: 2, B: 3, A: 255}, testVal: 2, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, colors.BlueEqual(tc.input, tc.testVal))
		})
	}
}

func TestAlphaEqual(t *testing.T) {
	tests := map[string]struct {
		input   color.Color
		testVal byte
		want    bool
	}{
		"RGBA(1,2,3,4) == 4: true":      {input: color.RGBA{R: 1, G: 2, B: 3, A: 4}, testVal: 4, want: true},
		"RGBA(1,2,3,4) == 2: false":     {input: color.RGBA{R: 1, G: 2, B: 3, A: 4}, testVal: 2, want: false},
		"NRGBA(1,2,3,255) == 255: true": {input: color.NRGBA{R: 1, G: 2, B: 3, A: 255}, testVal: 255, want: true},
		"NRGBA(1,2,3,255) == 4: false":  {input: color.NRGBA{R: 1, G: 2, B: 3, A: 255}, testVal: 4, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, colors.AlphaEqual(tc.input, tc.testVal))
		})
	}
}

func TestAddRGB(t *testing.T) {
	summed := colors.AddRGB(
		color.RGBA{R: 1, G: 2, B: 3, A: 4},
		color.RGBA{R: 4, G: 5, B: 6, A: 4},
		color.NRGBA{R: 7, G: 8, B: 9, A: 255},
	)

	assert.True(t, colors.RedEqual(summed, 12))
	assert.True(t, colors.GreenEqual(summed, 15))
	assert.True(t, colors.BlueEqual(summed, 18))
	assert.True(t, colors.AlphaEqual(summed, 255))
}

func TestMultiplyRGBByConstant(t *testing.T) {
	scaled := colors.MultiplyRGBByConstant(color.RGBA{R: 2, G: 100, B: 150, A: 200}, 0.5)

	assert.True(t, colors.RedEqual(scaled, 1))
	assert.True(t, colors.GreenEqual(scaled, 50))
	assert.True(t, colors.BlueEqual(scaled, 75))
	assert.True(t, colors.AlphaEqual(scaled, 200))
}
