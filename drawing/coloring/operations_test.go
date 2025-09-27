package coloring_test

import (
	"image/color"
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/stretchr/testify/assert"
)

func TestRedOps(t *testing.T) {
	tests := map[string]struct {
		input bool
		want  bool
	}{
		"RGBA(4,2,3,4) == 4: true":        {input: coloring.RedEqual(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 4), want: true},
		"RGBA(4,2,3,4) == 2: false":       {input: coloring.RedEqual(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 2), want: false},
		"NRGBA(255,2,3,255) == 255: true": {input: coloring.RedEqual(color.NRGBA{R: 255, G: 2, B: 3, A: 255}, 255), want: true},
		"NRGBA(255,2,3,255) == 4: false":  {input: coloring.RedEqual(color.NRGBA{R: 255, G: 2, B: 3, A: 255}, 4), want: false},

		"RGBA(4,2,3,4) <= 5: true":  {input: coloring.RedLessThanOrEqual(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 5), want: true},
		"RGBA(4,2,3,4) <= 4: true":  {input: coloring.RedLessThanOrEqual(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 4), want: true},
		"RGBA(4,2,3,4) <= 3: false": {input: coloring.RedLessThanOrEqual(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 3), want: false},

		"RGBA(4,2,3,4) >= 5: false": {input: coloring.RedGreaterThanOrEqual(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 5), want: false},
		"RGBA(4,2,3,4) >= 4: true":  {input: coloring.RedGreaterThanOrEqual(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 4), want: true},
		"RGBA(4,2,3,4) >= 3: true":  {input: coloring.RedGreaterThanOrEqual(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 3), want: true},

		"RGBA(4,2,3,4) < 5: true":  {input: coloring.RedLessThan(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 5), want: true},
		"RGBA(4,2,3,4) < 4: true":  {input: coloring.RedLessThan(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 4), want: false},
		"RGBA(4,2,3,4) < 3: false": {input: coloring.RedLessThan(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 3), want: false},

		"RGBA(4,2,3,4) > 5: false": {input: coloring.RedGreaterThan(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 5), want: false},
		"RGBA(4,2,3,4) > 4: true":  {input: coloring.RedGreaterThan(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 4), want: false},
		"RGBA(4,2,3,4) > 3: true":  {input: coloring.RedGreaterThan(color.RGBA{R: 4, G: 2, B: 3, A: 4}, 3), want: true},
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
		"RGBA(1,4,3,4) == 4: true":        {input: coloring.GreenEqual(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 4), want: true},
		"RGBA(1,4,3,4) == 2: false":       {input: coloring.GreenEqual(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 2), want: false},
		"NRGBA(1,255,3,255) == 255: true": {input: coloring.GreenEqual(color.NRGBA{R: 1, G: 255, B: 3, A: 255}, 255), want: true},
		"NRGBA(1,255,3,255) == 4: false":  {input: coloring.GreenEqual(color.NRGBA{R: 1, G: 255, B: 3, A: 255}, 4), want: false},

		"RGBA(1,4,3,4) <= 5: true":  {input: coloring.GreenLessThanOrEqual(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 5), want: true},
		"RGBA(1,4,3,4) <= 4: true":  {input: coloring.GreenLessThanOrEqual(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 4), want: true},
		"RGBA(1,4,3,4) <= 3: false": {input: coloring.GreenLessThanOrEqual(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 3), want: false},

		"RGBA(1,4,3,4) >= 5: false": {input: coloring.GreenGreaterThanOrEqual(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 5), want: false},
		"RGBA(1,4,3,4) >= 4: true":  {input: coloring.GreenGreaterThanOrEqual(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 4), want: true},
		"RGBA(1,4,3,4) >= 3: true":  {input: coloring.GreenGreaterThanOrEqual(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 3), want: true},

		"RGBA(1,4,3,4) < 5: true":  {input: coloring.GreenLessThan(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 5), want: true},
		"RGBA(1,4,3,4) < 4: true":  {input: coloring.GreenLessThan(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 4), want: false},
		"RGBA(1,4,3,4) < 3: false": {input: coloring.GreenLessThan(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 3), want: false},

		"RGBA(1,4,3,4) > 5: false": {input: coloring.GreenGreaterThan(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 5), want: false},
		"RGBA(1,4,3,4) > 4: true":  {input: coloring.GreenGreaterThan(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 4), want: false},
		"RGBA(1,4,3,4) > 3: true":  {input: coloring.GreenGreaterThan(color.RGBA{R: 1, G: 4, B: 3, A: 4}, 3), want: true},
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
		"RGBA(1,2,4,4) == 4: true":        {input: coloring.BlueEqual(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 4), want: true},
		"RGBA(1,2,4,4) == 2: false":       {input: coloring.BlueEqual(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 2), want: false},
		"NRGBA(1,2,255,255) == 255: true": {input: coloring.BlueEqual(color.NRGBA{R: 1, G: 2, B: 255, A: 255}, 255), want: true},
		"NRGBA(1,2,255,255) == 4: false":  {input: coloring.BlueEqual(color.NRGBA{R: 1, G: 2, B: 255, A: 255}, 4), want: false},

		"RGBA(1,2,4,4) <= 5: true":  {input: coloring.BlueLessThanOrEqual(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 5), want: true},
		"RGBA(1,2,4,4) <= 4: true":  {input: coloring.BlueLessThanOrEqual(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 4), want: true},
		"RGBA(1,2,4,4) <= 3: false": {input: coloring.BlueLessThanOrEqual(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 3), want: false},

		"RGBA(1,2,4,4) >= 5: false": {input: coloring.BlueGreaterThanOrEqual(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 5), want: false},
		"RGBA(1,2,4,4) >= 4: true":  {input: coloring.BlueGreaterThanOrEqual(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 4), want: true},
		"RGBA(1,2,4,4) >= 3: true":  {input: coloring.BlueGreaterThanOrEqual(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 3), want: true},

		"RGBA(1,2,4,4) < 5: true":  {input: coloring.BlueLessThan(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 5), want: true},
		"RGBA(1,2,4,4) < 4: true":  {input: coloring.BlueLessThan(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 4), want: false},
		"RGBA(1,2,4,4) < 3: false": {input: coloring.BlueLessThan(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 3), want: false},

		"RGBA(1,2,4,4) > 5: false": {input: coloring.BlueGreaterThan(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 5), want: false},
		"RGBA(1,2,4,4) > 4: true":  {input: coloring.BlueGreaterThan(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 4), want: false},
		"RGBA(1,2,4,4) > 3: true":  {input: coloring.BlueGreaterThan(color.RGBA{R: 1, G: 2, B: 4, A: 4}, 3), want: true},
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
		"RGBA(1,2,3,4) == 4: true":      {input: coloring.AlphaEqual(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 4), want: true},
		"RGBA(1,2,3,4) == 2: false":     {input: coloring.AlphaEqual(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 2), want: false},
		"NRGBA(1,2,3,255) == 255: true": {input: coloring.AlphaEqual(color.NRGBA{R: 1, G: 2, B: 3, A: 255}, 255), want: true},
		"NRGBA(1,2,3,255) == 4: false":  {input: coloring.AlphaEqual(color.NRGBA{R: 1, G: 2, B: 3, A: 255}, 4), want: false},

		"RGBA(1,2,3,4) <= 5: true":  {input: coloring.AlphaLessThanOrEqual(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 5), want: true},
		"RGBA(1,2,3,4) <= 4: true":  {input: coloring.AlphaLessThanOrEqual(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 4), want: true},
		"RGBA(1,2,3,4) <= 3: false": {input: coloring.AlphaLessThanOrEqual(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 3), want: false},

		"RGBA(1,2,3,4) >= 5: false": {input: coloring.AlphaGreaterThanOrEqual(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 5), want: false},
		"RGBA(1,2,3,4) >= 4: true":  {input: coloring.AlphaGreaterThanOrEqual(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 4), want: true},
		"RGBA(1,2,3,4) >= 3: true":  {input: coloring.AlphaGreaterThanOrEqual(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 3), want: true},

		"RGBA(1,2,3,4) < 5: true":  {input: coloring.AlphaLessThan(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 5), want: true},
		"RGBA(1,2,3,4) < 4: true":  {input: coloring.AlphaLessThan(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 4), want: false},
		"RGBA(1,2,3,4) < 3: false": {input: coloring.AlphaLessThan(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 3), want: false},

		"RGBA(1,2,3,4) > 5: false": {input: coloring.AlphaGreaterThan(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 5), want: false},
		"RGBA(1,2,3,4) > 4: true":  {input: coloring.AlphaGreaterThan(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 4), want: false},
		"RGBA(1,2,3,4) > 3: true":  {input: coloring.AlphaGreaterThan(color.RGBA{R: 1, G: 2, B: 3, A: 4}, 3), want: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.input)
		})
	}
}

func TestAddRGB(t *testing.T) {
	summed := coloring.AddRGB(
		color.RGBA{R: 1, G: 2, B: 3, A: 4},
		color.RGBA{R: 4, G: 5, B: 6, A: 4},
		color.NRGBA{R: 7, G: 8, B: 9, A: 255},
	)

	assert.True(t, coloring.RedEqual(summed, 12))
	assert.True(t, coloring.GreenEqual(summed, 15))
	assert.True(t, coloring.BlueEqual(summed, 18))
	assert.True(t, coloring.AlphaEqual(summed, 255))
}

func TestMultiplyRGBByConstant(t *testing.T) {
	scaled := coloring.ScaleRGB(color.RGBA{R: 2, G: 100, B: 150, A: 200}, 0.5)

	assert.True(t, coloring.RedEqual(scaled, 1))
	assert.True(t, coloring.GreenEqual(scaled, 50))
	assert.True(t, coloring.BlueEqual(scaled, 75))
	assert.True(t, coloring.AlphaEqual(scaled, 200))
}

func TestSingleComponents(t *testing.T) {
	c := color.RGBA{R: 2, G: 100, B: 150, A: 200}
	assert.Equal(t, byte(2), coloring.Red(c))
	assert.Equal(t, byte(100), coloring.Green(c))
	assert.Equal(t, byte(150), coloring.Blue(c))
	assert.Equal(t, byte(200), coloring.Alpha(c))
}
