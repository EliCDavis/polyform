package colors_test

import (
	"image/color"
	"testing"

	"github.com/EliCDavis/polyform/math/colors"
	"github.com/stretchr/testify/assert"
)

func TestRedOps(t *testing.T) {
	tests := map[string]struct {
		input bool
		want  bool
	}{
		"RGBA(4,2,3,4) == 4: true":        {input: colors.RedEqual(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 4), want: true},
		"RGBA(4,2,3,4) == 2: false":       {input: colors.RedEqual(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 2), want: false},
		"NRGBA(255,2,3,255) == 255: true": {input: colors.RedEqual(color.NRGBA{R: 255, G: 2, B: 3, A: 255}, 255), want: true},
		"NRGBA(255,2,3,255) == 4: false":  {input: colors.RedEqual(color.NRGBA{R: 255, G: 2, B: 3, A: 255}, 4), want: false},

		"RGBA(4,2,3,4) <= 5: true":  {input: colors.RedLessThanOrEqual(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 5), want: true},
		"RGBA(4,2,3,4) <= 4: true":  {input: colors.RedLessThanOrEqual(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 4), want: true},
		"RGBA(4,2,3,4) <= 3: false": {input: colors.RedLessThanOrEqual(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 3), want: false},

		"RGBA(4,2,3,4) >= 5: false": {input: colors.RedGreaterThanOrEqual(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 5), want: false},
		"RGBA(4,2,3,4) >= 4: true":  {input: colors.RedGreaterThanOrEqual(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 4), want: true},
		"RGBA(4,2,3,4) >= 3: true":  {input: colors.RedGreaterThanOrEqual(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 3), want: true},

		"RGBA(4,2,3,4) < 5: true":  {input: colors.RedLessThan(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 5), want: true},
		"RGBA(4,2,3,4) < 4: true":  {input: colors.RedLessThan(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 4), want: false},
		"RGBA(4,2,3,4) < 3: false": {input: colors.RedLessThan(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 3), want: false},

		"RGBA(4,2,3,4) > 5: false": {input: colors.RedGreaterThan(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 5), want: false},
		"RGBA(4,2,3,4) > 4: true":  {input: colors.RedGreaterThan(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 4), want: false},
		"RGBA(4,2,3,4) > 3: true":  {input: colors.RedGreaterThan(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 3), want: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.input)
		})
	}
}

func TestGreenOps(t *testing.T) {
	tests := map[string]struct {
		input bool
		want  bool
	}{
		"RGBA(1,4,3,4) == 4: true":        {input: colors.GreenEqual(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 4), want: true},
		"RGBA(1,4,3,4) == 2: false":       {input: colors.GreenEqual(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 2), want: false},
		"NRGBA(1,255,3,255) == 255: true": {input: colors.GreenEqual(color.NRGBA{R: 1, G: 255, B: 3, A: 255}, 255), want: true},
		"NRGBA(1,255,3,255) == 4: false":  {input: colors.GreenEqual(color.NRGBA{R: 1, G: 255, B: 3, A: 255}, 4), want: false},

		"RGBA(1,4,3,4) <= 5: true":  {input: colors.GreenLessThanOrEqual(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 5), want: true},
		"RGBA(1,4,3,4) <= 4: true":  {input: colors.GreenLessThanOrEqual(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 4), want: true},
		"RGBA(1,4,3,4) <= 3: false": {input: colors.GreenLessThanOrEqual(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 3), want: false},

		"RGBA(1,4,3,4) >= 5: false": {input: colors.GreenGreaterThanOrEqual(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 5), want: false},
		"RGBA(1,4,3,4) >= 4: true":  {input: colors.GreenGreaterThanOrEqual(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 4), want: true},
		"RGBA(1,4,3,4) >= 3: true":  {input: colors.GreenGreaterThanOrEqual(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 3), want: true},

		"RGBA(1,4,3,4) < 5: true":  {input: colors.GreenLessThan(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 5), want: true},
		"RGBA(1,4,3,4) < 4: true":  {input: colors.GreenLessThan(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 4), want: false},
		"RGBA(1,4,3,4) < 3: false": {input: colors.GreenLessThan(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 3), want: false},

		"RGBA(1,4,3,4) > 5: false": {input: colors.GreenGreaterThan(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 5), want: false},
		"RGBA(1,4,3,4) > 4: true":  {input: colors.GreenGreaterThan(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 4), want: false},
		"RGBA(1,4,3,4) > 3: true":  {input: colors.GreenGreaterThan(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 3), want: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.input)
		})
	}
}

func TestBlueOps(t *testing.T) {
	tests := map[string]struct {
		input bool
		want  bool
	}{
		"RGBA(1,2,4,4) == 4: true":        {input: colors.BlueEqual(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 4), want: true},
		"RGBA(1,2,4,4) == 2: false":       {input: colors.BlueEqual(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 2), want: false},
		"NRGBA(1,2,255,255) == 255: true": {input: colors.BlueEqual(color.NRGBA{R: 1, G: 2, B: 255, A: 255}, 255), want: true},
		"NRGBA(1,2,255,255) == 4: false":  {input: colors.BlueEqual(color.NRGBA{R: 1, G: 2, B: 255, A: 255}, 4), want: false},

		"RGBA(1,2,4,4) <= 5: true":  {input: colors.BlueLessThanOrEqual(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 5), want: true},
		"RGBA(1,2,4,4) <= 4: true":  {input: colors.BlueLessThanOrEqual(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 4), want: true},
		"RGBA(1,2,4,4) <= 3: false": {input: colors.BlueLessThanOrEqual(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 3), want: false},

		"RGBA(1,2,4,4) >= 5: false": {input: colors.BlueGreaterThanOrEqual(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 5), want: false},
		"RGBA(1,2,4,4) >= 4: true":  {input: colors.BlueGreaterThanOrEqual(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 4), want: true},
		"RGBA(1,2,4,4) >= 3: true":  {input: colors.BlueGreaterThanOrEqual(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 3), want: true},

		"RGBA(1,2,4,4) < 5: true":  {input: colors.BlueLessThan(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 5), want: true},
		"RGBA(1,2,4,4) < 4: true":  {input: colors.BlueLessThan(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 4), want: false},
		"RGBA(1,2,4,4) < 3: false": {input: colors.BlueLessThan(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 3), want: false},

		"RGBA(1,2,4,4) > 5: false": {input: colors.BlueGreaterThan(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 5), want: false},
		"RGBA(1,2,4,4) > 4: true":  {input: colors.BlueGreaterThan(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 4), want: false},
		"RGBA(1,2,4,4) > 3: true":  {input: colors.BlueGreaterThan(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 3), want: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.input)
		})
	}
}

func TestAlphaOps(t *testing.T) {
	tests := map[string]struct {
		input bool
		want  bool
	}{
		"RGBA(1,2,3,4) == 4: true":      {input: colors.AlphaEqual(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 4), want: true},
		"RGBA(1,2,3,4) == 2: false":     {input: colors.AlphaEqual(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 2), want: false},
		"NRGBA(1,2,3,255) == 255: true": {input: colors.AlphaEqual(color.NRGBA{R: 1, G: 2, B: 3, A: 255}, 255), want: true},
		"NRGBA(1,2,3,255) == 4: false":  {input: colors.AlphaEqual(color.NRGBA{R: 1, G: 2, B: 3, A: 255}, 4), want: false},

		"RGBA(1,2,3,4) <= 5: true":  {input: colors.AlphaLessThanOrEqual(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 5), want: true},
		"RGBA(1,2,3,4) <= 4: true":  {input: colors.AlphaLessThanOrEqual(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 4), want: true},
		"RGBA(1,2,3,4) <= 3: false": {input: colors.AlphaLessThanOrEqual(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 3), want: false},

		"RGBA(1,2,3,4) >= 5: false": {input: colors.AlphaGreaterThanOrEqual(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 5), want: false},
		"RGBA(1,2,3,4) >= 4: true":  {input: colors.AlphaGreaterThanOrEqual(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 4), want: true},
		"RGBA(1,2,3,4) >= 3: true":  {input: colors.AlphaGreaterThanOrEqual(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 3), want: true},

		"RGBA(1,2,3,4) < 5: true":  {input: colors.AlphaLessThan(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 5), want: true},
		"RGBA(1,2,3,4) < 4: true":  {input: colors.AlphaLessThan(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 4), want: false},
		"RGBA(1,2,3,4) < 3: false": {input: colors.AlphaLessThan(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 3), want: false},

		"RGBA(1,2,3,4) > 5: false": {input: colors.AlphaGreaterThan(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 5), want: false},
		"RGBA(1,2,3,4) > 4: true":  {input: colors.AlphaGreaterThan(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 4), want: false},
		"RGBA(1,2,3,4) > 3: true":  {input: colors.AlphaGreaterThan(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 3), want: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.input)
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
