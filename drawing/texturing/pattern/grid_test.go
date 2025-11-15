package pattern_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/drawing/texturing/pattern"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
)

func TestGrid(t *testing.T) {

	tests := map[string]struct {
		expected         texturing.Texture[float64]
		targetDimensions vector2.Int
		grid             pattern.Grid[float64]
	}{
		"Empty": {
			targetDimensions: vector2.New(10, 10),
			expected: texturing.FromArray([]float64{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}, 10, 10),
		},
		"Line Filled": {
			targetDimensions: vector2.New(10, 10),
			grid: pattern.Grid[float64]{
				VerticalLineWidth: 1,
				VerticalLines:     1,
				LineValue:         1,
			},
			expected: texturing.FromArray([]float64{
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			}, 10, 10),
		},
		"Single Verticle Line": {
			targetDimensions: vector2.New(10, 10),
			grid: pattern.Grid[float64]{
				VerticalLineWidth: .1,
				VerticalLines:     1,
				LineValue:         1,
			},
			expected: texturing.FromArray([]float64{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}, 10, 10),
		},
		"Single VerticleHorizontal Line": {
			targetDimensions: vector2.New(10, 10),
			grid: pattern.Grid[float64]{
				VerticalLineWidth:   .1,
				VerticalLines:       1,
				HorizontalLineWidth: .1,
				HorizontalLines:     1,
				LineValue:           1,
			},
			expected: texturing.FromArray([]float64{
				0, 0, 0, 0, 0, 1, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 1, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 1, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 1, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 1, 0, 0, 0, 0,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				0, 0, 0, 0, 0, 1, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 1, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 1, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 1, 0, 0, 0, 0,
			}, 10, 10),
		},
		"Two VerticleHorizontal Lines": {
			targetDimensions: vector2.New(10, 10),
			grid: pattern.Grid[float64]{
				VerticalLineWidth:   .1,
				VerticalLines:       2,
				HorizontalLineWidth: .1,
				HorizontalLines:     2,
				LineValue:           1,
			},
			expected: texturing.FromArray([]float64{
				0, 0, 0, 1, 0, 0, 0, 1, 0, 0,
				0, 0, 0, 1, 0, 0, 0, 1, 0, 0,
				0, 0, 0, 1, 0, 0, 0, 1, 0, 0,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				0, 0, 0, 1, 0, 0, 0, 1, 0, 0,
				0, 0, 0, 1, 0, 0, 0, 1, 0, 0,
				0, 0, 0, 1, 0, 0, 0, 1, 0, 0,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				0, 0, 0, 1, 0, 0, 0, 1, 0, 0,
				0, 0, 0, 1, 0, 0, 0, 1, 0, 0,
			}, 10, 10),
		},
	}

	for testName, tc := range tests {
		t.Run(testName, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.grid.Texture(tc.targetDimensions))
		})
	}
}
