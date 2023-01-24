package sample_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestLinearFloatMapping(t *testing.T) {
	tests := map[string]struct {
		fromMin, fromMax, toMin, toMax float64
		input                          float64
		want                           float64
	}{
		"[0, 1] => [1, 2]; f(0.5) => 1.5": {
			fromMin: 0, fromMax: 1,
			toMin: 1, toMax: 2,
			input: 0.5, want: 1.5,
		},

		"[0, 2] => [1, 2]; f(1) => 1.5": {
			fromMin: 0, fromMax: 2,
			toMin: 1, toMax: 2,
			input: 1, want: 1.5,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mapping := sample.LinearFloatMapping(
				tc.fromMin, tc.fromMax,
				tc.toMin, tc.toMax,
			)
			assert.Equal(t, tc.want, mapping(tc.input))
		})
	}
}

func TestLinearVector2Mapping(t *testing.T) {
	tests := map[string]struct {
		fromMin, fromMax float64
		toMin, toMax     vector2.Float64
		input            float64
		want             vector2.Float64
	}{
		"[0, 1] => [(0, 0), (1, 2)]; f(0.5) => (0.5, 1)": {
			fromMin: 0, fromMax: 1,
			toMin: vector2.Zero[float64](), toMax: vector2.New(1., 2.),
			input: 0.5, want: vector2.New(0.5, 1.),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mapping := sample.LinearVector2Mapping(
				tc.fromMin, tc.fromMax,
				tc.toMin, tc.toMax,
			)
			assert.Equal(t, tc.want, mapping(tc.input))
		})
	}
}

func TestLinearVector33Mapping(t *testing.T) {
	tests := map[string]struct {
		fromMin, fromMax float64
		toMin, toMax     vector3.Float64
		input            float64
		want             vector3.Float64
	}{
		"[0, 1] => [(0, 0, 0), (1, 2, 4)]; f(0.5) => (0.5, 1, 2)": {
			fromMin: 0, fromMax: 1,
			toMin: vector3.Zero[float64](), toMax: vector3.New(1., 2., 4.),
			input: 0.5, want: vector3.New(0.5, 1., 2.),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mapping := sample.LinearVector3Mapping(
				tc.fromMin, tc.fromMax,
				tc.toMin, tc.toMax,
			)
			assert.Equal(t, tc.want, mapping(tc.input))
		})
	}
}
