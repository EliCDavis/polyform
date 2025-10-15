package pattern_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/drawing/texturing/pattern"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
)

func TestCircle_Panic(t *testing.T) {
	assert.PanicsWithError(t, "can not create circle: invalid texture dimensions -1x-1", func() {
		pattern.Circle[int]{}.Texture(vector2.New(-1, -1))
	})

}

func TestCircle(t *testing.T) {

	tests := map[string]struct {
		circle           pattern.Circle[int]
		targetDimensions vector2.Int
		expected         texturing.Texture[int]
	}{
		"Empty": {
			targetDimensions: vector2.New(0, 10),
			expected:         texturing.Empty[int](0, 10),
		},
		"No Radius": {
			targetDimensions: vector2.New(9, 9),
			expected: texturing.FromArray([]int{
				0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0,
			}, 9, 9),
		},
		"Just Radius": {
			circle: pattern.Circle[int]{
				Radius: 0.5,
				Fill:   1,
			},
			targetDimensions: vector2.New(12, 12),
			expected: texturing.FromArray([]int{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0,
				0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0,
				0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0,
				0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0,
				0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}, 12, 12),
		},
		"Outer Border": {
			circle: pattern.Circle[int]{
				Radius:               0.5,
				Fill:                 1,
				OuterBorder:          2,
				OuterBorderThickness: 0.2,
			},
			targetDimensions: vector2.New(12, 12),
			expected: texturing.FromArray([]int{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 2, 2, 2, 0, 0, 0, 0,
				0, 0, 0, 0, 2, 2, 2, 2, 2, 0, 0, 0,
				0, 0, 0, 2, 1, 1, 1, 1, 1, 2, 0, 0,
				0, 0, 2, 2, 1, 1, 1, 1, 1, 2, 2, 0,
				0, 0, 2, 2, 1, 1, 1, 1, 1, 2, 2, 0,
				0, 0, 2, 2, 1, 1, 1, 1, 1, 2, 2, 0,
				0, 0, 0, 2, 1, 1, 1, 1, 1, 2, 0, 0,
				0, 0, 0, 0, 2, 2, 2, 2, 2, 0, 0, 0,
				0, 0, 0, 0, 0, 2, 2, 2, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}, 12, 12),
		},
		"Inner + Outer Border": {
			circle: pattern.Circle[int]{
				Radius:               0.5,
				Fill:                 1,
				OuterBorder:          2,
				OuterBorderThickness: 0.2,
				InnerBorder:          3,
				InnerBorderThickness: 0.1,
			},
			targetDimensions: vector2.New(12, 12),
			expected: texturing.FromArray([]int{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 2, 2, 2, 2, 2, 0, 0, 0,
				0, 0, 0, 2, 2, 3, 3, 3, 2, 2, 0, 0,
				0, 0, 2, 2, 1, 1, 1, 1, 1, 2, 2, 0,
				0, 0, 2, 3, 1, 1, 1, 1, 1, 3, 2, 0,
				0, 0, 2, 3, 1, 1, 1, 1, 1, 3, 2, 0,
				0, 0, 2, 3, 1, 1, 1, 1, 1, 3, 2, 0,
				0, 0, 2, 2, 1, 1, 1, 1, 1, 2, 2, 0,
				0, 0, 0, 2, 2, 3, 3, 3, 2, 2, 0, 0,
				0, 0, 0, 0, 2, 2, 2, 2, 2, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}, 12, 12),
		},
	}

	for testName, tc := range tests {
		t.Run(testName, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.circle.Texture(tc.targetDimensions))
		})
	}
}
