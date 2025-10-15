package pattern_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/drawing/texturing/pattern"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
)

func TestGrid_Panic(t *testing.T) {
	assert.PanicsWithError(t, "can not create grid: invalid texture dimensions -1x-1", func() {
		pattern.Grid(texturing.Texture[float64]{}, vector2.New(1, 1), vector2.New(-1, -1))
	})

	assert.PanicsWithError(t, "can not repeat texture element on a grid 10x-1 times", func() {
		pattern.Grid(texturing.Texture[float64]{}, vector2.New(10, -1), vector2.New(1, 1))
	})
}

func TestGrid(t *testing.T) {

	tests := map[string]struct {
		element          texturing.Texture[float64]
		repeat           vector2.Int
		targetDimensions vector2.Int
		expected         texturing.Texture[float64]
	}{
		"Simple 3x3": {
			element:          texturing.FromArray([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9}, 3, 3),
			repeat:           vector2.New(3, 3),
			targetDimensions: vector2.New(9, 9),
			expected: texturing.FromArray([]float64{
				1, 2, 3, 1, 2, 3, 1, 2, 3,
				4, 5, 6, 4, 5, 6, 4, 5, 6,
				7, 8, 9, 7, 8, 9, 7, 8, 9,
				1, 2, 3, 1, 2, 3, 1, 2, 3,
				4, 5, 6, 4, 5, 6, 4, 5, 6,
				7, 8, 9, 7, 8, 9, 7, 8, 9,
				1, 2, 3, 1, 2, 3, 1, 2, 3,
				4, 5, 6, 4, 5, 6, 4, 5, 6,
				7, 8, 9, 7, 8, 9, 7, 8, 9,
			}, 9, 9),
		},
	}

	for testName, tc := range tests {
		t.Run(testName, func(t *testing.T) {
			assert.Equal(t, tc.expected, pattern.Grid(tc.element, tc.repeat, tc.targetDimensions))
		})
	}
}
